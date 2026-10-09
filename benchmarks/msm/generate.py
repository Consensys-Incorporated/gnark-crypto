#!/usr/bin/env python3
"""Install the same in-package MSM benchmark harness into any checkout."""

import argparse
import pathlib
import subprocess

CURVES = {
    "bls12-381": ("bls12381", list(range(4, 17))),
    "bls12-377": ("bls12377", list(range(4, 17))),
    "bw6-761": ("bw6761", [4, 5, 8, 10, 16]),
}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("checkout", type=pathlib.Path)
    args = parser.parse_args()
    template = pathlib.Path(__file__).with_name("bench_test.go.tmpl").read_text()
    for curve, (package, windows) in CURVES.items():
        source = template.replace("@PACKAGE@", package).replace("@CURVE@", curve)
        source = source.replace("@WINDOWS@", ", ".join(map(str, windows)))
        group_template = source.split("// GROUP_TEMPLATE\n", 1)[1]
        source = source.split("// GROUP_TEMPLATE\n", 1)[0]
        for group in ("G1", "G2"):
            body = group_template.replace("@GROUP@", group)
            generator = group.lower()
            body = body.replace(group + "_GENERATOR", generator)
            body = body.replace(group + "_REFERENCE_ASSIGN", "result.ScalarMultiplication(&" + generator + ", msmBenchCoefficient(scalars))")
            body = body.replace(group + "_POINT_CHECK", "expected.ScalarMultiplication(&" + generator + ", big.NewInt(int64(i+1)))")
            source += body
        source = subprocess.run(["gofmt"], input=source, text=True, capture_output=True, check=True).stdout
        path = args.checkout / "ecc" / curve / "msm_benchmark_test.go"
        if path.exists() and path.read_text() != source:
            raise SystemExit("Refusing to overwrite an existing harness: " + str(path))
        path.write_text(source)
        print(path)
        if "func bestCG1(" in (path.parent / "multiexp.go").read_text():
            selection = pathlib.Path(__file__).with_name("selection_test.go.tmpl").read_text()
            selection = selection.replace("@PACKAGE@", package).replace("@CURVE@", curve)
            selection = subprocess.run(["gofmt"], input=selection, text=True, capture_output=True, check=True).stdout
            selection_path = path.with_name("msm_benchmark_selection_test.go")
            if selection_path.exists() and selection_path.read_text() != selection:
                raise SystemExit("Refusing to overwrite an existing selection harness: " + str(selection_path))
            selection_path.write_text(selection)
            print(selection_path)


if __name__ == "__main__":
    main()
