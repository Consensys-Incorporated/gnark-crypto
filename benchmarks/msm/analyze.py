#!/usr/bin/env python3
"""Summarize raw Go benchmark samples and retain machine-readable comparisons."""

import argparse
import collections
import csv
import json
import math
import pathlib
import re
import statistics
import subprocess

LINE = re.compile(r"^(BenchmarkMSM\S+)\s+(\d+)\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op$")


def samples(path):
    package = None
    values = collections.defaultdict(list)
    for line in path.read_text().splitlines():
        if line.startswith("pkg: "):
            package = line[5:]
        match = LINE.match(line.strip())
        if match:
            if package is None:
                raise ValueError("Missing package header: " + str(path))
            name = re.sub(r"-\d+$", "", match[1])
            values[(package, name)].append(tuple(map(float, match.group(3, 4, 5))))
    return values


def geomean(values):
    return math.exp(statistics.mean(math.log(value) for value in values))


def timing_statistics(path):
    package, timing = None, False
    result = {}
    for row in csv.reader(path.open()):
        if len(row) == 1 and row[0].startswith("pkg: "):
            package = row[0][5:]
        elif len(row) > 1 and not row[0] and row[1] in ("sec/op", "B/op", "allocs/op"):
            timing = row[1] == "sec/op"
        elif timing and len(row) >= 7 and row[0].startswith("MSM"):
            name = "Benchmark" + re.sub(r"-\d+$", "", row[0])
            p = re.search(r"p=([\d.]+)", row[6])
            result[(package, name)] = {"base_time_ci": row[2], "pr_time_ci": row[4], "benchstat_change": row[5], "p_value": float(p[1]) if p else ""}
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("results", type=pathlib.Path)
    parser.add_argument("--preparation", type=pathlib.Path)
    args = parser.parse_args()
    out = args.results
    rows = []
    summary = {}
    for base_path in sorted(out.glob("*-base.txt")):
        profile = base_path.name.removesuffix("-base.txt")
        pr_path = out / (profile + "-pr.txt")
        if not pr_path.exists():
            continue
        base, pr = samples(base_path), samples(pr_path)
        if set(base) != set(pr):
            raise SystemExit("Mismatched benchmark cases in " + profile)
        with (out / (profile + "-benchstat-warnings.txt")).open("w") as warnings:
            with (out / (profile + "-benchstat.txt")).open("w") as handle:
                subprocess.run(["benchstat", str(base_path), str(pr_path)], check=True, stdout=handle, stderr=warnings)
            with (out / (profile + "-benchstat.csv")).open("w") as handle:
                subprocess.run(["benchstat", "-format=csv", str(base_path), str(pr_path)], check=True, stdout=handle, stderr=warnings)
        tests = timing_statistics(out / (profile + "-benchstat.csv"))
        for key in sorted(base):
            package, name = key
            details = dict(re.findall(r"/(n|dist|tasks|c)=([^/]+)", name))
            group = re.search(r"(G[12])", name)[1]
            baseline = statistics.median(v[0] for v in base[key])
            candidate = statistics.median(v[0] for v in pr[key])
            rows.append({"profile": profile, "curve": package.rsplit("/", 1)[-1], "group": group, "n": int(details["n"]), "distribution": details["dist"], "tasks": int(details["tasks"]), "base_samples": len(base[key]), "pr_samples": len(pr[key]), "base_median_ns": baseline, "pr_median_ns": candidate, "delta_percent": 100*(candidate/baseline-1), "base_median_bytes": statistics.median(v[1] for v in base[key]), "pr_median_bytes": statistics.median(v[1] for v in pr[key])})
            if key not in tests:
                raise SystemExit("Missing benchstat timing comparison: " + str(key))
            rows[-1].update(tests[key])
        profile_rows = [row for row in rows if row["profile"] == profile]
        summary[profile] = {"cases": len(profile_rows), "min_samples": min(min(row["base_samples"], row["pr_samples"]) for row in profile_rows), "curve_groups": []}
        for curve, group in sorted({(row["curve"], row["group"]) for row in profile_rows}):
            subset = [row for row in profile_rows if (row["curve"], row["group"]) == (curve, group)]
            entry = {"curve": curve, "group": group, "cases": len(subset), "geomean_delta_percent": 100*(geomean([row["pr_median_ns"]/row["base_median_ns"] for row in subset])-1), "worst": max(subset, key=lambda row: row["delta_percent"]), "best": min(subset, key=lambda row: row["delta_percent"])}
            summary[profile]["curve_groups"].append(entry)
    if rows:
        with (out / "comparisons.csv").open("w") as handle:
            writer = csv.DictWriter(handle, fieldnames=rows[0].keys())
            writer.writeheader()
            writer.writerows(rows)
    fixed_rows = []
    for path in sorted(out.glob("windows-*.txt")):
        version = path.name.removeprefix("windows-").removesuffix(".txt")
        entries = samples(path)
        groups = collections.defaultdict(dict)
        for (package, name), measures in entries.items():
            details = dict(re.findall(r"/(n|dist|tasks|c)=([^/]+)", name))
            group = re.search(r"(G[12])", name)[1]
            key = (package.rsplit("/", 1)[-1], group, int(details["n"]), details["dist"], int(details["tasks"]))
            groups[key][int(details["c"])] = (statistics.median(value[0] for value in measures), len(measures), statistics.median(value[1] for value in measures), statistics.median(value[2] for value in measures))
        for key, windows in sorted(groups.items()):
            best_c = min(windows, key=lambda c: windows[c][0])
            fastest = windows[best_c][0]
            for c, (timing, count, allocated_bytes, allocations) in sorted(windows.items()):
                fixed_rows.append({"version": version, "curve": key[0], "group": key[1], "n": key[2], "distribution": key[3], "tasks": key[4], "c": c, "samples": count, "median_ns": timing, "median_bytes": allocated_bytes, "median_allocations": allocations, "fastest_measured_c": best_c, "overhead_percent": 100*(timing/fastest-1)})
    if fixed_rows:
        with (out / "fixed_windows.csv").open("w") as handle:
            writer = csv.DictWriter(handle, fieldnames=fixed_rows[0].keys())
            writer.writeheader()
            writer.writerows(fixed_rows)
        summary["fixed_windows"] = {"cases": len(fixed_rows), "min_samples": min(row["samples"] for row in fixed_rows)}
        preparation = args.preparation or out.parent.parent / "preparation.json"
        if preparation.exists():
            selections = {}
            prepared = json.loads(preparation.read_text())
            streams = [build["validation_stdout"] for build in prepared["builds"]]
            streams += [record["stdout"] for record in prepared.get("additional_selection_diagnostics", [])]
            for stream in streams:
                for line in stream.splitlines():
                    if line.startswith("selection,"):
                        _, curve, group, n, old_c, new_c = line.split(",")
                        selections[(curve, group, int(n))] = (int(old_c), int(new_c))
            measurements = collections.defaultdict(dict)
            for row in fixed_rows:
                if row["version"] == "pr":
                    measurements[(row["curve"], row["group"], row["n"], row["tasks"])][row["c"]] = row
            selector_rows = []
            for key, windows in sorted(measurements.items()):
                if key[:3] not in selections:
                    continue
                old_c, new_c = selections[key[:3]]
                if old_c not in windows or new_c not in windows:
                    continue
                old, new = windows[old_c], windows[new_c]
                selector_rows.append({"curve": key[0], "group": key[1], "n": key[2], "tasks": key[3], "base_selected_c": old_c, "pr_selected_c": new_c, "fastest_measured_c": new["fastest_measured_c"], "base_selection_overhead_percent": old["overhead_percent"], "pr_selection_overhead_percent": new["overhead_percent"], "pr_vs_base_window_delta_percent": 100*(new["median_ns"]/old["median_ns"]-1), "base_selected_median_bytes": old["median_bytes"], "pr_selected_median_bytes": new["median_bytes"]})
            if selector_rows:
                with (out / "selector_overheads.csv").open("w") as handle:
                    writer = csv.DictWriter(handle, fieldnames=selector_rows[0].keys())
                    writer.writeheader()
                    writer.writerows(selector_rows)
                summary["selector_overheads"] = {"cases": len(selector_rows), "base_mean_overhead_percent": statistics.mean(row["base_selection_overhead_percent"] for row in selector_rows), "pr_mean_overhead_percent": statistics.mean(row["pr_selection_overhead_percent"] for row in selector_rows)}
    (out / "summary.json").write_text(json.dumps(summary, indent=2) + "\n")
    print(json.dumps({profile: {key: value for key, value in data.items() if key != "curve_groups"} for profile, data in summary.items()}, indent=2))


if __name__ == "__main__":
    main()
