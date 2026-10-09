#!/usr/bin/env python3
"""Build a reviewable local artifact bundle from completed MSM benchmark runs."""

import argparse
import csv
import json
import pathlib
import shutil
import statistics
import subprocess
import sys


def write_csv(path, rows):
    if rows:
        with path.open("w") as handle:
            writer = csv.DictWriter(handle, fieldnames=rows[0].keys())
            writer.writeheader()
            writer.writerows(rows)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--scratch", type=pathlib.Path, required=True)
    parser.add_argument("--output", type=pathlib.Path, required=True)
    args = parser.parse_args()
    analyze = pathlib.Path(__file__).with_name("analyze.py")
    args.output.mkdir(parents=True, exist_ok=True)
    metadata = []
    comparisons, fixed, selectors = [], [], []
    for source in sorted((args.scratch / "results").iterdir()):
        meta_path = source / "metadata.json"
        if not source.is_dir() or not meta_path.exists():
            continue
        meta = json.loads(meta_path.read_text())
        if not meta.get("finished_utc") or "invalidated_utc" in meta or meta.get("continuity_events"):
            raise SystemExit("Incomplete benchmark run: " + str(source))
        if any(command.get("returncode") != 0 for command in meta["commands"]):
            raise SystemExit("Unsuccessful benchmark process: " + str(source))
        if "continuity_guard" in meta:
            maximum_gap = meta["continuity_guard"]["maximum_gap_seconds"]
            if any(abs(c["wall_elapsed_seconds"]-c["monotonic_elapsed_seconds"]) > maximum_gap for c in meta["commands"]):
                raise SystemExit("Wall/monotonic clock discontinuity: " + str(source))
        subprocess.run([sys.executable, str(analyze), str(source), "--preparation", str(args.scratch / "preparation.json")], check=True, stdout=subprocess.DEVNULL)
        profiles = list(meta["profiles"])
        if profiles == ["primary"]:
            name = "primary"
        elif profiles == ["windows"]:
            name = "windows/" + source.name
        else:
            name = source.name
        for file, rows in (("comparisons.csv", comparisons), ("fixed_windows.csv", fixed), ("selector_overheads.csv", selectors)):
            path = source / file
            if path.exists():
                for row in csv.DictReader(path.open()):
                    sample_fields = ("base_samples", "pr_samples") if file == "comparisons.csv" else ("samples",) if file == "fixed_windows.csv" else ()
                    if any(int(row[field]) != meta["samples"] for field in sample_fields):
                        raise SystemExit("Unexpected sample count: " + str(path))
                    # Keep the exhaustive exploratory sweep separate from the repeated pass.
                    if file != "comparisons.csv" and meta["samples"] == 1:
                        continue
                    rows.append(dict(row, run=name))
        target = args.output / name
        shutil.copytree(source, target, dirs_exist_ok=True)
        metadata.append({"archive": name, "metadata": meta})
    if not metadata:
        raise SystemExit("No completed benchmark runs found")
    write_csv(args.output / "comparisons.csv", comparisons)
    write_csv(args.output / "fixed_windows_refined.csv", fixed)
    write_csv(args.output / "selector_overheads_refined.csv", selectors)
    shutil.copy2(args.scratch / "preparation.json", args.output / "preparation.json")
    exclusions = args.scratch / "exclusions.json"
    if exclusions.exists():
        shutil.copy2(exclusions, args.output / "exclusions.json")
    for path in (args.scratch / "results").glob("*-refinement.json"):
        record = json.loads(path.read_text())
        if "finished_utc" not in record:
            raise SystemExit("Incomplete window refinement: " + str(path))
        shutil.copy2(path, args.output / path.name)
    (args.output / "index.json").write_text(json.dumps(metadata, indent=2) + "\n")
    first = metadata[0]["metadata"]
    cpu = first.get("cpu_model", first.get("machdep.cpu.brand_string", "unavailable"))
    cores = first.get("logical_cpus", first.get("hw.ncpu", "unavailable"))
    memory = first.get("physical_memory_bytes", first.get("hw.memsize", "unavailable"))
    try:
        memory_text = "%.1f GiB RAM" % (int(memory)/2**30)
    except (ValueError, TypeError):
        memory_text = "RAM size unavailable"
    hardware = "Measurements were run on %s, %s, %s logical CPUs, %s, using %s." % (cpu, first["architecture"], cores, memory_text, first["go_version"])
    identities = {(m["metadata"]["architecture"], m["metadata"].get("cpu_model", m["metadata"].get("machdep.cpu.brand_string")), m["metadata"]["go_version"], tuple(sorted(m["metadata"]["commits"].items()))) for m in metadata}
    if len(identities) != 1:
        raise SystemExit("Use a separate artifact bundle for each hardware/toolchain/commit combination")
    if first["architecture"] not in ("x86_64", "amd64"):
        hardware += " Actual Intel or AMD x86-64 measurements remain outstanding."
    text = ["# PR 875 benchmark results", "", hardware, "", "Base commit: `" + first["commits"]["base"] + "`. PR commit: `" + first["commits"]["pr"] + "`. Fitted constants and production code were unchanged during these experiments.", "", "Inputs are distinct valid subgroup points `(i+1)*generator` and deterministic scalars. Input generation, known-coefficient oracle calculation, and result verification are outside timing. Public timings use Jacobian output. All six binaries passed fixture and existing MSM tests in short mode; every timed case checks its result.", "", "## Coverage", "", "| Run | Profiles | Curves | Groups | Samples per case |", "| --- | --- | --- | --- | --- |"]
    for record in metadata:
        meta = record["metadata"]
        text.append("| [" + record["archive"] + "](" + record["archive"] + "/metadata.json) | " + ", ".join(meta["profiles"]) + " | " + ", ".join(meta["curves"]) + " | " + meta["groups"].replace("|", ", ") + " | " + str(meta["samples"]) + " |")
    text += ["", "## Public API comparisons", "", "[Combined comparisons](comparisons.csv) retain medians, sample counts, allocation bytes, and percentage changes. Negative deltas mean the PR was faster. Each public comparison run also contains its per-case `benchstat` report, confidence intervals, and p-values. Geometric mean ratios below weight each configuration equally; they do not represent a proving workload.", "", "| Profile | Curve | Group | Geometric mean median-time change |", "| --- | --- | --- | --- |"]
    for record in metadata:
        summary = json.loads((args.output / record["archive"] / "summary.json").read_text())
        for profile, details in summary.items():
            for group in details.get("curve_groups", []):
                text.append("| %s | %s | %s | %+.2f%% |" % (profile, group["curve"], group["group"], group["geomean_delta_percent"]))
    if comparisons:
        text += ["", "Per-case timing tests below use benchstat's default significance threshold. They are not corrected for multiple comparisons; many cases are tested, and repeated sizes across profiles are not independent evidence.", "", "| Profile | Cases | Faster | Slower | Inconclusive |", "| --- | --- | --- | --- | --- |"]
        for profile in sorted({row["profile"] for row in comparisons}):
            subset = [row for row in comparisons if row["profile"] == profile]
            faster = sum(row["benchstat_change"] != "~" and float(row["delta_percent"]) < 0 for row in subset)
            slower = sum(row["benchstat_change"] != "~" and float(row["delta_percent"]) > 0 for row in subset)
            text.append("| %s | %d | %d | %d | %d |" % (profile, len(subset), faster, slower, len(subset)-faster-slower))
        regressions = sorted((row for row in comparisons if row["benchstat_change"] != "~" and float(row["delta_percent"]) > 0), key=lambda row: float(row["delta_percent"]), reverse=True)
        if regressions:
            write_csv(args.output / "timing_regressions.csv", regressions)
            text += ["", "### Observed timing regressions", "", "[All per-case regressions](timing_regressions.csv) are retained. The table lists up to twelve largest median increases detected by the individual timing tests. The p-values use rounded benchstat output and are not corrected for multiple comparisons.", "", "| Profile | Curve | Group | Points | Scalars | Tasks | Median-time increase | p |", "| --- | --- | --- | --- | --- | --- | --- | --- |"]
            for row in regressions[:12]:
                p_value = float(row["p_value"])
                p_text = "<0.001" if p_value == 0 else "%.3f" % p_value
                text.append("| %s | %s | %s | %s | %s | %s | %+.2f%% | %s |" % (row["profile"], row["curve"], row["group"], row["n"], row["distribution"], row["tasks"], float(row["delta_percent"]), p_text))
    allocation_cases = []
    for curve, group in sorted({(row["curve"], row["group"]) for row in comparisons}):
        candidates = [row for row in comparisons if (row["curve"], row["group"]) == (curve, group) and float(row["base_median_bytes"]) > 0]
        if candidates:
            largest = max(candidates, key=lambda row: float(row["pr_median_bytes"])/float(row["base_median_bytes"]))
            if float(largest["pr_median_bytes"])/float(largest["base_median_bytes"]) >= 2:
                allocation_cases.append(largest)
    if allocation_cases:
        text += ["", "## Allocation tradeoffs", "", "The table shows the largest measured allocation increase for each curve/group with an increase of at least 2x. These are allocated bytes per operation. They are not peak resident memory, and a timing delta alone does not establish statistical significance. All configurations remain available in the combined comparisons.", "", "| Profile | Curve | Group | Points | Scalars | Tasks | Base MiB/op | PR MiB/op | Allocation ratio | Median-time change |", "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |"]
        for row in allocation_cases:
            old, new = float(row["base_median_bytes"]), float(row["pr_median_bytes"])
            text.append("| %s | %s | %s | %s | %s | %s | %.3f | %.3f | %.2fx | %+.2f%% |" % (row["profile"], row["curve"], row["group"], row["n"], row["distribution"], row["tasks"], old/2**20, new/2**20, new/old, float(row["delta_percent"])))
    if selectors:
        text += ["", "## Window-selection overhead", "", "The exhaustive sweep used one exploratory sample for all supported windows. The repeated pass used ten samples for the union of each configuration's three fastest first-pass windows and both heuristics' selected windows, with shuffled window order per block. The fastest repeated window is the fastest within that candidate set, not a statistically proven global optimum.", "", "[Repeated window timings](fixed_windows_refined.csv) and [selector overheads](selector_overheads_refined.csv) compare both selectors on the same unsplit PR kernels. The base selector's local closure is reproduced using its exact arithmetic; the PR selectors are called directly. Public API results also include recursive splitting and cannot be attributed to window selection alone.", "", "| Curve | Group | Configurations | Base mean overhead | PR mean overhead |", "| --- | --- | --- | --- | --- |"]
        if any(record["metadata"]["profiles"].get("windows", {}).get("sizes") == "262144" for record in metadata):
            text[-2:-2] = ["A targeted follow-up also compared windows 14, 15, and 16 at 262144 points with NbTasks=1 on both BLS curves, using ten samples per window. Those four additional curve/group configurations are included in the averages below.", ""]
        for curve, group in sorted({(row["curve"], row["group"]) for row in selectors}):
            subset = [row for row in selectors if (row["curve"], row["group"]) == (curve, group)]
            old = statistics.mean(float(row["base_selection_overhead_percent"]) for row in subset)
            new = statistics.mean(float(row["pr_selection_overhead_percent"]) for row in subset)
            text.append("| %s | %s | %d | %.2f%% | %.2f%% |" % (curve, group, len(subset), old, new))
        large_selectors = [row for row in selectors if int(row["n"]) == 262144 and int(row["tasks"]) == 1]
        if large_selectors:
            text += ["", "### Large-input window follow-up", "", "These ten-sample uniform-input comparisons use the same unsplit PR kernels at 262144 points and NbTasks=1. They help assess window choice in the observed public API regressions. The ratios are descriptive median comparisons; no per-window significance test was performed, and public API splitting is excluded.", "", "| Curve | Group | Base selected c | PR selected c | Fastest tested c | PR-window median change | PR/base allocated bytes |", "| --- | --- | --- | --- | --- | --- | --- |"]
            for row in large_selectors:
                allocation_ratio = float(row["pr_selected_median_bytes"])/float(row["base_selected_median_bytes"])
                text.append("| %s | %s | %s | %s | %s | %+.2f%% | %.2fx |" % (row["curve"], row["group"], row["base_selected_c"], row["pr_selected_c"], row["fastest_measured_c"], float(row["pr_vs_base_window_delta_percent"]), allocation_ratio))
    loads = [command["load_before"][0] for record in metadata for command in record["metadata"]["commands"]]
    if exclusions.exists():
        text += ["", "## Recovery from device shutdown", "", "The initial scalar-distribution run spanned a reported device shutdown. Its entire distribution profile was excluded, including samples before and after the shutdown. A fresh ten-sample comparison was run with temporary macOS sleep prevention and a clock/scheduler continuity guard. The already-completed large-input profile was recovered byte-for-byte from before the interruption. [The exclusion record](exclusions.json) identifies the original archive, which is retained outside the repository for audit."]
    text += ["", "## Limits", "", "CPU affinity, power management, and thermal settings were not controlled by the runner. Recorded one-minute load averages before benchmark processes ranged from %.2f to %.2f. Those averages include prior benchmark activity. Many main-run timing confidence intervals were wide, so small changes and observed regressions may be inconclusive. Use the per-case statistical reports rather than treating every median difference as established." % (min(loads), max(loads)), "", "Synthetic consecutive generator multiples are not captured prover inputs. Full-width uniform, small, zero-heavy, and repeated scalar profiles should be interpreted separately. Allocation measurements are allocated bytes per operation, not peak resident memory. These checks share the library's field and scalar-multiplication primitives and do not audit those primitives externally.", "", "The copied run metadata preserves original local command paths and UTC timestamps. Reproduction instructions are in [the harness README](../../README.md).", ""]
    (args.output / "README.md").write_text("\n".join(text))
    print("Artifact bundle: " + str(args.output))


if __name__ == "__main__":
    main()
