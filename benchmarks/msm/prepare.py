#!/usr/bin/env python3
"""Install, compile, and validate both benchmark checkouts with one Go toolchain."""

import argparse
import datetime
import hashlib
import json
import os
import pathlib
import subprocess
import sys


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base", required=True, type=pathlib.Path)
    parser.add_argument("--pr", required=True, type=pathlib.Path)
    parser.add_argument("--scratch", required=True, type=pathlib.Path)
    args = parser.parse_args()
    binary_dir = args.scratch.resolve() / "bin"
    binary_dir.mkdir(parents=True, exist_ok=True)
    metadata = {"go_version": subprocess.check_output(["go", "version"], text=True).strip(), "builds": []}
    generator = pathlib.Path(__file__).with_name("generate.py").resolve()
    for version in ("base", "pr"):
        checkout = getattr(args, version).resolve()
        subprocess.run([sys.executable, str(generator), str(checkout)], check=True)
        for curve in ("bls12-381", "bls12-377", "bw6-761"):
            binary = binary_dir / (version + "-" + curve + ".test")
            command = ["go", "test", "-c", "-o", str(binary), "./ecc/" + curve]
            print("Building " + version + " " + curve, flush=True)
            subprocess.run(command, cwd=checkout, check=True)
            test_command = [str(binary), "-test.run=^Test(MSMBenchmark|BestC|MsmBatchSize|MultiExp|CrossMultiExp)", "-test.short", "-test.v"]
            validation_environment = {"MSM_BENCH_SIZES": "1024,1536,2048,4096,8192,16384,32768,262144"}
            validation = subprocess.run(test_command, env=dict(os.environ, **validation_environment), check=True, capture_output=True, text=True)
            record = {"version": version, "curve": curve, "commit": subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=checkout, text=True).strip(), "build_command": command, "validation_command": test_command, "validation_stdout": validation.stdout, "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(), "finished_utc": datetime.datetime.now(datetime.timezone.utc).isoformat()}
            metadata["builds"].append(record)
            record["validation_environment"] = validation_environment
            print("Validated " + version + " " + curve, flush=True)
    (args.scratch / "preparation.json").write_text(json.dumps(metadata, indent=2) + "\n")
    print("Ready: " + str(binary_dir))


if __name__ == "__main__":
    main()
