#!/usr/bin/env python3
"""Repeat first-pass MSM window candidates and both heuristics' selected windows."""

import argparse
import csv
import datetime
import json
import pathlib
import subprocess
import sys


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--coarse", type=pathlib.Path, required=True)
    parser.add_argument("--base", type=pathlib.Path, required=True)
    parser.add_argument("--pr", type=pathlib.Path, required=True)
    parser.add_argument("--scratch", type=pathlib.Path, required=True)
    parser.add_argument("--samples", type=int, default=10)
    parser.add_argument("--benchtime", default="100ms")
    args = parser.parse_args()
    directory = pathlib.Path(__file__).parent.resolve()
    subprocess.run([sys.executable, str(directory / "analyze.py"), str(args.coarse)], check=True)
    metadata = json.loads((args.coarse / "metadata.json").read_text())
    selections = {}
    preparation = json.loads((args.scratch / "preparation.json").read_text())
    for build in preparation["builds"]:
        for line in build["validation_stdout"].splitlines():
            if line.startswith("selection,"):
                _, curve, group, n, old_c, new_c = line.split(",")
                selections.setdefault((curve, int(n)), set()).update((int(old_c), int(new_c)))
    measurements = {}
    for row in csv.DictReader((args.coarse / "fixed_windows.csv").open()):
        key = (row["curve"], int(row["n"]), row["group"], int(row["tasks"]))
        measurements.setdefault(key, []).append((float(row["median_ns"]), int(row["c"])))
    candidates = {}
    for (curve, n, group, tasks), timings in sorted(measurements.items()):
        key = (curve, n)
        candidates.setdefault(key, set()).update(c for _, c in sorted(timings)[:3])
        candidates[key].update(selections[key])
    plans = [{"curve": curve, "n": n, "windows": sorted(windows)} for (curve, n), windows in sorted(candidates.items())]
    stamp = datetime.datetime.now(datetime.timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    manifest = {"coarse": str(args.coarse), "samples": args.samples, "method": "Union of the three fastest first-pass windows for each group/task configuration and the base and PR selectors at each size; repeated on unchanged PR kernels", "plans": plans, "results": []}
    manifest_path = args.scratch / "results" / (stamp + "-refinement.json")
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n")
    for index, plan in enumerate(plans):
        print("Refinement %d/%d: %s n=%d windows=%s" % (index+1, len(plans), plan["curve"], plan["n"], plan["windows"]), flush=True)
        command = [sys.executable, str(directory / "run.py"), "--base", str(args.base), "--pr", str(args.pr), "--scratch", str(args.scratch), "--profiles", "windows", "--versions", "pr", "--curves", plan["curve"], "--sizes", str(plan["n"]), "--windows", ",".join(map(str, plan["windows"])), "--shuffle-windows", "--samples", str(args.samples), "--block-samples", "2", "--benchtime", args.benchtime, "--tasks", metadata["tasks"], "--gomaxprocs", str(metadata["gomaxprocs"])]
        process = subprocess.Popen(command, stdout=subprocess.PIPE, text=True, bufsize=1)
        result_dir = None
        for line in process.stdout:
            print(line, end="", flush=True)
            if line.startswith("Results: "):
                result_dir = pathlib.Path(line.strip()[9:])
        if process.wait():
            raise SystemExit("Refinement failed; inspect " + str(manifest_path))
        subprocess.run([sys.executable, str(directory / "analyze.py"), str(result_dir)], check=True)
        manifest["results"].append(str(result_dir))
        manifest_path.write_text(json.dumps(manifest, indent=2) + "\n")
    manifest["finished_utc"] = datetime.datetime.now(datetime.timezone.utc).isoformat()
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n")
    print("Refinement complete: " + str(manifest_path), flush=True)


if __name__ == "__main__":
    main()
