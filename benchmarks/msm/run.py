#!/usr/bin/env python3
"""Run compiled Go MSM benchmarks sequentially in alternating A/B blocks."""

import argparse
import datetime
import hashlib
import json
import os
import pathlib
import platform
import random
import subprocess
import threading
import time

CURVES = ["bls12-381", "bls12-377", "bw6-761"]
PROFILES = {
    "primary": {"sizes": "512,1024,1536,2048,4096,8192,16384,32768,65536", "distributions": "uniform", "kind": "Public"},
    "large": {"sizes": "262144,1048576", "distributions": "uniform", "kind": "Public"},
    "windows": {"sizes": "1024,1536,2048,4096,8192,16384,32768", "distributions": "uniform", "kind": "Fixed"},
    "distributions": {"sizes": "4096,32768,262144", "distributions": "small,zeros,repeated", "kind": "Public"},
    "transitions": {"sizes": "1224,1225,1226,1792,1793,1794,4096,4097,4098,6178,6179,6180,9216,9217,9218,20480,20481,20482,24094,24095,24096", "distributions": "uniform", "kind": "Public"},
}


def capture(command, cwd=None):
    return subprocess.run(command, cwd=cwd, check=True, text=True, capture_output=True).stdout.strip()


class ContinuityGuard:
    """Reject a run if the scheduler or wall clock pauses unexpectedly."""

    interval_seconds = 0.25
    maximum_gap_seconds = 2.0

    def __init__(self):
        self.events = []
        self.stop = threading.Event()
        self.last_sample = (time.time(), time.monotonic())
        self.thread = threading.Thread(target=self.monitor, daemon=True)

    def observe(self, wall_gap, monotonic_gap):
        if max(wall_gap, monotonic_gap) > self.maximum_gap_seconds:
            self.events.append({"observed_utc": datetime.datetime.now(datetime.timezone.utc).isoformat(), "wall_gap_seconds": wall_gap, "monotonic_gap_seconds": monotonic_gap})

    def monitor(self):
        wall, monotonic = self.last_sample
        while not self.stop.wait(self.interval_seconds):
            next_wall, next_monotonic = time.time(), time.monotonic()
            self.observe(next_wall-wall, next_monotonic-monotonic)
            wall, monotonic = next_wall, next_monotonic
            self.last_sample = (wall, monotonic)

    def check(self, metadata, metadata_path):
        wall, monotonic = self.last_sample
        self.observe(time.time()-wall, time.monotonic()-monotonic)
        if self.events:
            metadata["continuity_events"] = list(self.events)
            metadata["invalidated_utc"] = datetime.datetime.now(datetime.timezone.utc).isoformat()
            metadata["invalid_reason"] = "Unexpected clock/scheduler gap; exclude this run and repeat it"
            metadata_path.write_text(json.dumps(metadata, indent=2) + "\n")
            raise SystemExit("Timing continuity failed; exclude and repeat " + str(metadata_path.parent))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base", required=True, type=pathlib.Path)
    parser.add_argument("--pr", required=True, type=pathlib.Path)
    parser.add_argument("--scratch", required=True, type=pathlib.Path)
    parser.add_argument("--profiles", default="primary,large,windows,distributions,transitions")
    parser.add_argument("--samples", type=int, default=10)
    parser.add_argument("--block-samples", type=int, default=2)
    parser.add_argument("--benchtime", default="100ms")
    parser.add_argument("--tasks", default="1,%d,0" % os.cpu_count())
    parser.add_argument("--gomaxprocs", type=int, default=os.cpu_count())
    parser.add_argument("--versions", default="base,pr")
    parser.add_argument("--curves", default=",".join(CURVES))
    parser.add_argument("--groups", default="G1|G2")
    parser.add_argument("--windows", help="Comma-separated fixed windows; default is every supported window")
    parser.add_argument("--sizes", help="Override sizes for the chosen profile(s)")
    parser.add_argument("--shuffle-windows", action="store_true", help="Shuffle explicit window order per block with a recorded seed")
    args = parser.parse_args()
    if args.samples < 1 or args.block_samples < 1:
        parser.error("sample counts must be positive")
    profiles = args.profiles.split(",")
    for profile in profiles:
        if profile not in PROFILES:
            parser.error("unknown profile: " + profile)
    stamp = datetime.datetime.now(datetime.timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    out = args.scratch / "results" / (stamp + "-" + "-".join(profiles))
    out.mkdir(parents=True)
    metadata = {
        "started_utc": stamp,
        "platform": platform.platform(),
        "architecture": platform.machine(),
        "logical_cpus": os.cpu_count(),
        "go_version": capture(["go", "version"]),
        "gomaxprocs": args.gomaxprocs,
        "tasks": args.tasks,
        "benchtime": args.benchtime,
        "samples": args.samples,
        "block_samples": args.block_samples,
        "profiles": {p: dict(PROFILES[p], sizes=args.sizes or PROFILES[p]["sizes"]) for p in profiles},
        "curves": args.curves.split(","),
        "groups": args.groups,
        "windows": args.windows or "every supported window",
        "shuffle_windows": args.shuffle_windows,
        "commits": {v: capture(["git", "rev-parse", "HEAD"], cwd=getattr(args, v)) for v in ("base", "pr")},
        "input_seed": 875,
        "bases": "distinct subgroup points (i+1)*generator; blockwise batch normalization",
        "scalar_sampling": "SHA-512 counter stream, canonical rejection sampling",
        "timing": "generation and result checks excluded; identical harness; sequential alternating A/B blocks",
        "harness_sha256": {},
        "binary_sha256": {},
        "commands": [],
        "continuity_guard": {"interval_seconds": ContinuityGuard.interval_seconds, "maximum_gap_seconds": ContinuityGuard.maximum_gap_seconds},
    }
    for key in ("machdep.cpu.brand_string", "hw.ncpu", "hw.physicalcpu", "hw.memsize"):
        try:
            metadata[key] = capture(["sysctl", "-n", key])
        except (subprocess.CalledProcessError, FileNotFoundError):
            metadata[key] = "unavailable"
    metadata["cpu_model"] = metadata["machdep.cpu.brand_string"]
    try:
        metadata["physical_memory_bytes"] = os.sysconf("SC_PHYS_PAGES") * os.sysconf("SC_PAGE_SIZE")
    except (ValueError, OSError):
        metadata["physical_memory_bytes"] = "unavailable"
    if metadata["cpu_model"] == "unavailable" and pathlib.Path("/proc/cpuinfo").exists():
        for line in pathlib.Path("/proc/cpuinfo").read_text().splitlines():
            if line.startswith("model name"):
                metadata["cpu_model"] = line.split(":", 1)[1].strip()
                break
    metadata_path = out / "metadata.json"
    for curve in args.curves.split(","):
        hashes = [hashlib.sha256((getattr(args, v) / "ecc" / curve / "msm_benchmark_test.go").read_bytes()).hexdigest() for v in ("base", "pr")]
        if hashes[0] != hashes[1]:
            raise SystemExit("Harness differs between checkouts: " + curve)
        metadata["harness_sha256"][curve] = hashes[0]
        for version in args.versions.split(","):
            binary = args.scratch / "bin" / (version + "-" + curve + ".test")
            metadata["binary_sha256"][version + "-" + curve] = hashlib.sha256(binary.read_bytes()).hexdigest()
    metadata_path.write_text(json.dumps(metadata, indent=2) + "\n")
    print("Results: " + str(out), flush=True)
    started = time.monotonic()
    guard = ContinuityGuard()
    guard.thread.start()
    for profile in profiles:
        spec = metadata["profiles"][profile]
        env = dict(os.environ, GOMAXPROCS=str(args.gomaxprocs), MSM_BENCH_SIZES=spec["sizes"], MSM_BENCH_DISTRIBUTIONS=spec["distributions"], MSM_BENCH_TASKS=args.tasks)
        env.pop("MSM_BENCH_WINDOWS", None)
        if args.windows:
            env["MSM_BENCH_WINDOWS"] = args.windows
        completed = 0
        block = 0
        while completed < args.samples:
            count = min(args.block_samples, args.samples-completed)
            if args.windows and args.shuffle_windows:
                order = args.windows.split(",")
                random.Random(875+block).shuffle(order)
                env["MSM_BENCH_WINDOWS"] = ",".join(order)
            versions = args.versions.split(",")
            if block % 2:
                versions.reverse()
            for curve in args.curves.split(","):
                for version in versions:
                    guard.check(metadata, metadata_path)
                    binary = args.scratch / "bin" / (version + "-" + curve + ".test")
                    command = [str(binary), "-test.run=^$", "-test.bench=^BenchmarkMSM" + spec["kind"] + "(" + args.groups + ")$", "-test.benchtime=" + args.benchtime, "-test.count=" + str(count), "-test.benchmem", "-test.timeout=60m"]
                    record = {"profile": profile, "curve": curve, "version": version, "block": block, "command": command, "windows_order": env.get("MSM_BENCH_WINDOWS", "every supported window, ascending"), "started_utc": datetime.datetime.now(datetime.timezone.utc).isoformat(), "load_before": os.getloadavg()}
                    metadata["commands"].append(record)
                    metadata_path.write_text(json.dumps(metadata, indent=2) + "\n")
                    print("Running %s %s %s block %d (%d samples)" % (profile, curve, version, block+1, count), flush=True)
                    with (out / (profile + "-" + version + ".txt")).open("a") as combined:
                        combined.write("pkg: github.com/consensys/gnark-crypto/ecc/" + curve + "\n")
                        combined.flush()
                        command_wall, command_monotonic = time.time(), time.monotonic()
                        process = subprocess.run(command, env=env, stdout=combined, stderr=subprocess.STDOUT)
                    record["returncode"] = process.returncode
                    record["wall_elapsed_seconds"] = time.time()-command_wall
                    record["monotonic_elapsed_seconds"] = time.monotonic()-command_monotonic
                    record["finished_utc"] = datetime.datetime.now(datetime.timezone.utc).isoformat()
                    record["load_after"] = os.getloadavg()
                    metadata_path.write_text(json.dumps(metadata, indent=2) + "\n")
                    guard.check(metadata, metadata_path)
                    if process.returncode:
                        raise SystemExit("Benchmark failed; inspect " + str(out))
                    print("Finished in %.1f minutes total" % ((time.monotonic()-started)/60), flush=True)
            completed += count
            block += 1
    guard.stop.set()
    guard.thread.join()
    guard.check(metadata, metadata_path)
    metadata["finished_utc"] = datetime.datetime.now(datetime.timezone.utc).isoformat()
    metadata["elapsed_seconds"] = time.monotonic()-started
    metadata_path.write_text(json.dumps(metadata, indent=2) + "\n")
    print("Completed: " + str(out), flush=True)


if __name__ == "__main__":
    main()
