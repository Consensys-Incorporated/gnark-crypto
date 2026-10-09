# MSM benchmarks for PR 875

This harness compares the PR base commit with the PR head without changing
production code or fitted constants. It covers BLS12-381, BLS12-377, and BW6-761,
both G1 and G2, public `MultiExp`, and `_innerMsm` at fixed windows.
Public timings use `G1Jac.MultiExp` and `G2Jac.MultiExp`; final output
normalization into affine coordinates is excluded.

The published M2 measurements compare base commit
`0a975d747c416958f20ad4e785bcf102449a81c9` with PR commit
`19d23dbd7dc0b508a8f0cbca768bf65fee39321b`, using Go 1.25.7.
The [PR findings](https://github.com/Consensys-Incorporated/gnark-crypto/pull/875#issuecomment-6048325387)
link to the [original benchmark bundle](https://github.com/user-attachments/files/33180517/pr875-apple-m2-benchmarks.zip),
including raw results, metadata, and the harness source used for those runs.
Local result bundles under `results/` are ignored by Git.

The generated `ecc/*/msm_benchmark_test.go` files are separate from the existing
generated tests. `generate.py` installs identical benchmark source in both
checkouts. An additional selection test in the PR checkout calls its actual
selectors and records the old heuristic's reproduced arithmetic. Run generation
again after changing `bench_test.go.tmpl` only after removing
the old custom harness files. It refuses to overwrite different existing files.

## Inputs and correctness

All bases are distinct valid subgroup points: base `i` is `(i+1)*generator`.
Normalization happens in blocks of 4096 to bound temporary memory. These are
synthetic inputs, not captured proving workloads or a trusted setup.

Uniform scalars use a SHA-512 counter stream with seed 875 and canonical
rejection sampling. BW6-761 scalars fill its entire scalar field. Other
distributions are positive values cycling through 1 to 65535, 50% zeros mixed
with uniform values, and uniform values repeated in contiguous blocks of 100.
At 32768 points, the small-scalar distribution is exactly 1 through 32768.

For these bases the expected MSM is a single scalar multiplication by
`sum((i+1)*scalar[i]) mod r`. Every case checks this result before and after
timing. The fixture tests also compare with a sum of individual scalar
multiplications, check subgroup membership, and exercise every supported window.
These checks share the library's field and scalar-multiplication primitives;
they are not an external audit of those primitives.

## Reproduce

Use a separate scratch directory for each hardware/toolchain/commit combination.
Use two clean checkouts at the exact comparison commits, the same Go toolchain,
and a scratch directory outside the repository. The scripts require Python 3.9
or later, Go and `gofmt`, and Git. Analysis also requires `benchstat` on `PATH`.
Replace the example paths below with the paths to your comparison checkouts.
Copy `benchmarks/msm` from this commit into the PR checkout if reproducing the
recorded comparison at its earlier PR head. `prepare.py` then installs the same
benchmark source in both checkouts.

For a direct benchmark of the BW6-761 small-scalar case in one checkout:

```sh
MSM_BENCH_SIZES=32768 MSM_BENCH_DISTRIBUTIONS=small \
  MSM_BENCH_TASKS=0 GOMAXPROCS=8 \
  go test ./ecc/bw6-761 -run '^$' -bench '^BenchmarkMSMPublicG2$' \
  -benchmem -benchtime=1s -count=10
```

This measures one version. Use the alternating A/B runner below to compare
the base and PR checkouts.

`prepare.py` installs the harness, compiles all six binaries, runs fixture and
existing MSM tests, and records validation output and binary hashes:

```sh
python3 benchmarks/msm/prepare.py \
  --base /tmp/gnark-crypto-pr875-base \
  --pr /path/to/pr-checkout \
  --scratch /tmp/gnark-msm-pr875
```

Then run sequential alternating A/B blocks:

```sh
python3 benchmarks/msm/run.py \
  --base /tmp/gnark-crypto-pr875-base \
  --pr /path/to/pr-checkout \
  --scratch /tmp/gnark-msm-pr875 \
  --profiles primary,large,distributions \
  --samples 10 --block-samples 2 --benchtime 100ms \
  --tasks 1,8,0 --gomaxprocs 8
```

Choose task counts and `GOMAXPROCS` for the actual machine. `tasks=0` means the
public API's default. Fixed-window runs explicitly resolve it to twice the CPU
count. `GOMAXPROCS` stays the same across configurations. Task count controls
the algorithm's configuration, not process CPU affinity. For an additional
strict single-P experiment use `--tasks 1 --gomaxprocs 1`.

The runner records commits, harness hashes, hardware, toolchain, process start
and finish times, load averages, and commands. It keeps raw samples. It does
not change CPU affinity, power management, or thermal settings.

For local macOS runs, prefix the runner command with `caffeinate -i -s` to
temporarily prevent idle sleep while the command is active. Forced shutdown
can still interrupt a run. The runner monitors clock/scheduler continuity every
250 milliseconds and rejects a run after any gap longer than two seconds.
Unfinished or invalidated runs must be excluded and repeated before reporting.

Analyze a completed result directory with:

```sh
python3 benchmarks/msm/analyze.py /tmp/gnark-msm-pr875/results/RESULT_DIRECTORY
```

`benchstat` must be available on `PATH`. Outputs include its statistical reports,
raw median comparisons, and fixed-window timings relative to the fastest
measured window. Negative comparison deltas mean the PR was faster.

For an exploratory exhaustive sweep use `--profiles windows --versions pr
--samples 1 --benchtime 1x`. Repeat competitive windows with `--windows` and
`--samples 10`. `--sizes`, `--curves`, `--groups`, and `--tasks` support targeted
follow-up measurements. Single-sample sweeps are exploratory and do not provide
statistical confidence in small differences between windows.

`refine.py` automates the repeated pass. At each curve and size it takes the
union of the three fastest first-pass windows for each group/task configuration
and both heuristics' selected windows. It repeats that candidate set with ten
samples, shuffles window order per block with a recorded seed, and records its
plans and output directories in a manifest:

```sh
python3 benchmarks/msm/refine.py \
  --coarse /tmp/gnark-msm-pr875/results/COARSE_DIRECTORY \
  --base /tmp/gnark-crypto-pr875-base \
  --pr /path/to/pr-checkout \
  --scratch /tmp/gnark-msm-pr875
```

The refined fastest window is the fastest within the repeated candidate set.
The full single-sample sweep remains available separately. This approach
provides more measurements for close decisions without repeatedly timing every
clearly oversized window.

After all runs complete, `report.py --scratch SCRATCH --output STAGING_DIRECTORY`
builds a combined report and artifact bundle. Scan the staging directory for
credentials before copying it into the repository. The report preserves the
exhaustive exploratory sweep separately from repeated candidate measurements.

## Profiles

| Profile | Sizes | Distributions | Measurement |
| --- | --- | --- | --- |
| primary | 512 through 65536, including 1536 | Uniform | Public API |
| large | 262144 and 1048576 | Uniform | Public API |
| windows | 1024, 1536, 2048, 4096, 8192, 16384, 32768 | Uniform | Every supported fixed window |
| distributions | 4096, 32768, 262144 | Small, zeros, repeated | Public API |
| transitions | Adjacent sizes around original M2 window boundaries | Uniform | Public API |

The transition profile uses boundaries from the PR and base descriptions. It is
not an automatic discovery of every curve-specific boundary.

For this run, the dense transition follow-up targets G1 around selected original
BLS12-381 boundaries and the BW6-761 base selector's 2817-point transition:

```sh
python3 benchmarks/msm/run.py \
  --base /tmp/gnark-crypto-pr875-base \
  --pr /path/to/pr-checkout \
  --scratch /tmp/gnark-msm-pr875 \
  --profiles transitions --groups G1 \
  --sizes 1224,1225,1226,2816,2817,2818,4096,4097,4098,6178,6179,6180,24094,24095,24096 \
  --samples 10 --block-samples 2 --benchtime 100ms \
  --tasks 1,8,0 --gomaxprocs 8
```

The large-input regression follow-up compares the old selected window (15),
the PR selected window (14), and window 16 on the same PR kernels. This
measurement isolates window choice from public API splitting:

```sh
python3 benchmarks/msm/run.py \
  --base /tmp/gnark-crypto-pr875-base \
  --pr /path/to/pr-checkout \
  --scratch /tmp/gnark-msm-pr875 \
  --profiles windows --curves bls12-381,bls12-377 \
  --sizes 262144 --windows 14,15,16 \
  --tasks 1 --gomaxprocs 8 --versions pr \
  --samples 10 --block-samples 2 --benchtime 100ms --shuffle-windows
```

## Interpretation

Fixed-window results exclude recursive input splitting but include scalar
partitioning, per-window processor selection, overweight-chunk scheduling,
and final reduction. Public API results include window selection and recursive
splitting. A public API improvement cannot be attributed to window selection
alone.

Use statistical comparisons and report regressions as well as improvements.
The minimum of repeated samples is not the reported statistic. A local ARM64
run is evidence for that hardware only. Repeating on actual Intel and AMD
x86-64 machines remains necessary for the maintainer's portability question.
