module github.com/consensys/gnark-crypto

go 1.26.8

require (
	github.com/consensys/bavard v0.2.2-0.20260118153501-cba9f5475432
	github.com/leanovate/gopter v0.2.11
	github.com/mmcloughlin/addchain v0.4.0
	github.com/stretchr/testify v1.12.1
	golang.org/x/crypto v0.57.0
	golang.org/x/sync v0.24.0
	golang.org/x/sys v0.49.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/telemetry v0.0.0-20261009151841-97e8ea77aa2a // indirect
	golang.org/x/tools v0.52.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)

tool (
	github.com/klauspost/asmfmt/cmd/asmfmt
	golang.org/x/tools/cmd/goimports
)
