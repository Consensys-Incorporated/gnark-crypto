
<a name="v0.21.0"></a>
## [v0.21.0] - 2026-08-10
### Breaking
- remove `fri` package from all curves; FRI primitives now live in `field/koalabear` ([#856](https://github.com/Consensys/gnark-crypto/issues/856), [#871](https://github.com/Consensys/gnark-crypto/issues/871))
- remove `plookup` package from all curves ([#862](https://github.com/Consensys/gnark-crypto/issues/862))
- `internal/parallel` is now public as `parallel`; the worker pool moved to `utils` ([#819](https://github.com/Consensys/gnark-crypto/issues/819), [#823](https://github.com/Consensys/gnark-crypto/issues/823))
- eddsa `PublicKey.SetBytes` rejects points outside the prime-order subgroup; `PrivateKey.SetBytes` rejects malformed scalars and public keys that don't match the scalar ([#859](https://github.com/Consensys/gnark-crypto/issues/859))

### Security
- eddsa: reject small-subgroup public keys ([#859](https://github.com/Consensys/gnark-crypto/issues/859))
- shplonk, fflonk: check digests and proof points are in the subgroup before verifying ([#858](https://github.com/Consensys/gnark-crypto/issues/858))
- kzg: `MpcSetup.Verify` checked the ratio on the previous SRS instead of the contributed one, and did not bind the G1 update to the proof ([#857](https://github.com/Consensys/gnark-crypto/issues/857))
- add subgroup membership tests to all `twistededwards` curves ([#833](https://github.com/Consensys/gnark-crypto/issues/833))

### Feat
- cube roots in `Fp` and `Fp2` ([#825](https://github.com/Consensys/gnark-crypto/issues/825))
- FFT over degree-6 extension for koalabear and babybear ([#838](https://github.com/Consensys/gnark-crypto/issues/838), [#772](https://github.com/Consensys/gnark-crypto/issues/772))
- generic `BitReverse`/`BitReverseCopy` (naive and Cobra) in `utils` ([#871](https://github.com/Consensys/gnark-crypto/issues/871))
- koalabear: `Compressx16Columns`, `Compressx16ColumnsWithState` and `VectorE6.MulAccByElement` ([#865](https://github.com/Consensys/gnark-crypto/issues/865), [#866](https://github.com/Consensys/gnark-crypto/issues/866), [#871](https://github.com/Consensys/gnark-crypto/issues/871))
- twistededwards: fixed-base `ScalarMultiplicationBase`, used by eddsa ([#839](https://github.com/Consensys/gnark-crypto/issues/839))
- secp256r1: `Fp2` tower and Cardano cube-root solver ([#831](https://github.com/Consensys/gnark-crypto/issues/831))

### Fix
- `Fp2.Sqrt` returned a wrong result on `(non-QR, 0)` inputs ([#845](https://github.com/Consensys/gnark-crypto/issues/845))
- amd64: 4-byte overread in the `innerProdVec` AVX-512 path ([#841](https://github.com/Consensys/gnark-crypto/issues/841))
- restore `ExecuteAligned` behavior lost in the `Execute` refactor ([#836](https://github.com/Consensys/gnark-crypto/issues/836))
- macOS build
- fft/domain: stale `FinerGenerator`/`GeneratorSqrt` comment ([#744](https://github.com/Consensys/gnark-crypto/issues/744))

### Perf
- 4-bit sliding window exponentiation for all fields ([#817](https://github.com/Consensys/gnark-crypto/issues/817))
- lazy-reduction 𝔽p2 multiplication assembly (BLS12-381, BLS12-377, BLS24-315, BLS24-317) ([#861](https://github.com/Consensys/gnark-crypto/issues/861))
- add-sub chains in `Expt`/`mulBySeed` and optimal Karabina 0-runs ([#860](https://github.com/Consensys/gnark-crypto/issues/860))
- hash-to-G1 in the highly 2-adic `Fp` case ([#830](https://github.com/Consensys/gnark-crypto/issues/830))
- inline `E6.Mul`/`E6.Square` for 31-bit primes (2.07x) ([#842](https://github.com/Consensys/gnark-crypto/issues/842))
- AVX-512 batch ops for E6 FFT on koalabear (2.97x) ([#843](https://github.com/Consensys/gnark-crypto/issues/843))
- koalabear: unrolled FFT kernels and faster SIS `LimbIterator` ([#834](https://github.com/Consensys/gnark-crypto/issues/834))
- koalabear: AVX-512 `VectorE6` scalar mul-acc kernels ([#868](https://github.com/Consensys/gnark-crypto/issues/868))
- koalabear: NEON `Compressx16Columns` kernel (arm64) ([#870](https://github.com/Consensys/gnark-crypto/issues/870))
- F31 Poseidon2: `permutation16_avx512` (-17%) ([#829](https://github.com/Consensys/gnark-crypto/issues/829))
- generalize `Compressx16` AVX-512 for variable `colSize` ([#824](https://github.com/Consensys/gnark-crypto/issues/824))
- kzg mpc setup: drop the parallel pool, use batch subgroup checks ([#823](https://github.com/Consensys/gnark-crypto/issues/823))

### Build
- **deps:** bump golang.org/x/crypto to 0.54.0, x/sys to 0.47.0, x/sync to 0.22.0
- pin GitHub Actions to commit SHAs ([#837](https://github.com/Consensys/gnark-crypto/issues/837))
- apply `go fix` modernizations ([#820](https://github.com/Consensys/gnark-crypto/issues/820))

<a name="v0.20.1"></a>
## [v0.20.1] - 2026-03-16
### Fix
- correct IFMA vector mul carry propagation ([#816](https://github.com/Consensys/gnark-crypto/issues/816))
<a name="v0.20.0"></a>
## [v0.20.0] - 2026-03-11
### Build
- **deps:** bump golang.org/x/crypto from 0.41.0 to 0.45.0 in the go_modules group across 1 directory ([#770](https://github.com/Consensys/gnark-crypto/issues/770))
### CI
- enable security code scanner ([#771](https://github.com/Consensys/gnark-crypto/issues/771))
- Standardize imports ([#760](https://github.com/Consensys/gnark-crypto/issues/760))
### Chore
- golangci-lint update ([#813](https://github.com/Consensys/gnark-crypto/issues/813))
- rename sage files ([#806](https://github.com/Consensys/gnark-crypto/issues/806))
- add dependabot configuration for GitHub Actions updates ([#804](https://github.com/Consensys/gnark-crypto/issues/804))
- reintroduce slack notif ([#802](https://github.com/Consensys/gnark-crypto/issues/802))
- update ci workflows to go 1.25 ([#801](https://github.com/Consensys/gnark-crypto/issues/801))
- use self hosted runners ([#763](https://github.com/Consensys/gnark-crypto/issues/763))
- merge v0.19 hotfix branch ([#766](https://github.com/Consensys/gnark-crypto/issues/766))
- clarify PublicKey.SetBytes uses compressed public key ([#732](https://github.com/Consensys/gnark-crypto/issues/732))
### Feat
- add bn254 poseidon2 constants and support t=4/8/12/16 ([#783](https://github.com/Consensys/gnark-crypto/issues/783))
- lattice-based rational reconstructions for gnark scalar decompositions ([#799](https://github.com/Consensys/gnark-crypto/issues/799))
- Expose `NewDefaultPermutation` in Poseidon2 packages ([#761](https://github.com/Consensys/gnark-crypto/issues/761))
- batch subgroup membership testing ([#710](https://github.com/Consensys/gnark-crypto/issues/710))
- add typed field hasher interface in MiMC package ([#752](https://github.com/Consensys/gnark-crypto/issues/752))
- add parallel prefix product for vector e4 ([#750](https://github.com/Consensys/gnark-crypto/issues/750))
- expose vortex.ComputeLagrangeBasisAtX
- vector e4 utils ops ([#749](https://github.com/Consensys/gnark-crypto/issues/749))
- add vector.Exp method
- adds e4 vector Exp method ([#739](https://github.com/Consensys/gnark-crypto/issues/739))
- make bitReverse generic ([#736](https://github.com/Consensys/gnark-crypto/issues/736))
- handle len(p)=0 in kzg ([#730](https://github.com/Consensys/gnark-crypto/issues/730))
- Add P-256 (secp256r1) elliptic curve + ECDSA ([#767](https://github.com/Consensys/gnark-crypto/issues/767))
- add column linear combination check ([#731](https://github.com/Consensys/gnark-crypto/issues/731))
- Jacobian Triple for `j=0` curves ([#715](https://github.com/Consensys/gnark-crypto/issues/715))
### FiatShamir
- creation of new challenges with New() ([#811](https://github.com/Consensys/gnark-crypto/issues/811))
### Fix
- Affine twisted Edwards scalar multiplication edge case ([#774](https://github.com/Consensys/gnark-crypto/issues/774))
- fix cursor bot comments ([#803](https://github.com/Consensys/gnark-crypto/issues/803))
- E2 MustSetRandom all elements ([#793](https://github.com/Consensys/gnark-crypto/issues/793))
- remove duplicate import ([#762](https://github.com/Consensys/gnark-crypto/issues/762))
- use G2Affine in UpdateMonomialsG2 template ([#735](https://github.com/Consensys/gnark-crypto/issues/735))
- make [#740](https://github.com/Consensys/gnark-crypto/issues/740) retro compatible BitReverse generic impl
- reintroduce BitReverse in fft package as deprecated for backward… ([#740](https://github.com/Consensys/gnark-crypto/issues/740))
- [#727](https://github.com/Consensys/gnark-crypto/issues/727) with go generate and cosmetics ([#734](https://github.com/Consensys/gnark-crypto/issues/734))
### Perf
- direct quartic extension for koala and baby bears ([#779](https://github.com/Consensys/gnark-crypto/issues/779))
- optimize (purego) extension for koalabear ([#729](https://github.com/Consensys/gnark-crypto/issues/729))
- optimize pre-computation in fixed-argument pairings by batching double steps ([#798](https://github.com/Consensys/gnark-crypto/issues/798))
- optimize precomputation in fixed-argument pairings ([#797](https://github.com/Consensys/gnark-crypto/issues/797))
- GLV-GLS scalar multiplication in G2 ([#786](https://github.com/Consensys/gnark-crypto/issues/786))
- implement wNAF width=5 GLV scalar mul on G1/2 ([#788](https://github.com/Consensys/gnark-crypto/issues/788))
- use NAF decomposition in `mulWindow` ([#787](https://github.com/Consensys/gnark-crypto/issues/787))
- improve small size fft ([#791](https://github.com/Consensys/gnark-crypto/issues/791))
- use ifma for some 4w vector operations ([#795](https://github.com/Consensys/gnark-crypto/issues/795))
- optimize `Sqrt` in highly 2-adic fields  ([#796](https://github.com/Consensys/gnark-crypto/issues/796))
- adds some more neon implementations for poseidon2 ([#785](https://github.com/Consensys/gnark-crypto/issues/785))
- adds poseidon2 "simd" fashion compress function (avx512) ([#777](https://github.com/Consensys/gnark-crypto/issues/777))
- limit memory allocation during Vector deserialization ([#759](https://github.com/Consensys/gnark-crypto/issues/759))
- twisted Edwards affine arithmetic ([#743](https://github.com/Consensys/gnark-crypto/issues/743))
- use vector ops on FFT E4 ([#746](https://github.com/Consensys/gnark-crypto/issues/746))
- tEd affine scalar mul
- add for small fields int64 for exp ([#738](https://github.com/Consensys/gnark-crypto/issues/738))
- adds E4 vector ops (avx512) ([#733](https://github.com/Consensys/gnark-crypto/issues/733))
- **bw6:** faster direct Fp6 mul ([#776](https://github.com/Consensys/gnark-crypto/issues/776))
- Optimize `Sqrt` in `Fp2` for all fields ([#757](https://github.com/Consensys/gnark-crypto/issues/757))
### Refactor
- use more defines in assembly fine to make files less verbose ([#789](https://github.com/Consensys/gnark-crypto/issues/789))
- code-generate Fp6 tower for BW6 curves ([#792](https://github.com/Consensys/gnark-crypto/issues/792))
### Release
- v0.20.0 ([#815](https://github.com/Consensys/gnark-crypto/issues/815))
<a name="v0.19.2"></a>
## [v0.19.2] - 2025-10-28
### Docs
- add CHANGELOG to v0.19.2
### Fix
- avoid b.Loop in Go 1.23
<a name="v0.19.1"></a>
## [v0.19.1] - 2025-10-28
### Docs
- update CHANGELOG to v0.19.1
### Perf
- limit memory allocation during Vector deserialization ([#759](https://github.com/Consensys/gnark-crypto/issues/759))
<a name="v0.19.0"></a>
## [v0.19.0] - 2025-08-22
### CI
- update linter for the "push to master" flow ([#728](https://github.com/Consensys/gnark-crypto/issues/728))
- Upgrade linter ([#726](https://github.com/Consensys/gnark-crypto/issues/726))
### Feat
- Subgroup membership testing on Bandersnatch ([#708](https://github.com/Consensys/gnark-crypto/issues/708))
- set default size for state ([#700](https://github.com/Consensys/gnark-crypto/issues/700))
### Fix
- refactor asm code generation to avoid use of R15 with global memory accesses ([#712](https://github.com/Consensys/gnark-crypto/issues/712))
- MerkleDamgardHasher.Sum ([#699](https://github.com/Consensys/gnark-crypto/issues/699))
- MerkleDamgardHasher IV ([#701](https://github.com/Consensys/gnark-crypto/issues/701))
- fixes [#697](https://github.com/Consensys/gnark-crypto/issues/697) ([#698](https://github.com/Consensys/gnark-crypto/issues/698))
### Perf
- Doubling and mixed doubling in Jacobians and ext. Jacobians ([#716](https://github.com/Consensys/gnark-crypto/issues/716))
- optimize Eisenstein arithmetic ([#706](https://github.com/Consensys/gnark-crypto/issues/706))
- small scalarmul ([#703](https://github.com/Consensys/gnark-crypto/issues/703))
- Legendre symbol calculation using Pornin20 ([#704](https://github.com/Consensys/gnark-crypto/issues/704))
### Refector
- Add feed-forward step to CompressPoseidon2 function ([#721](https://github.com/Consensys/gnark-crypto/issues/721))
<a name="v0.18.2"></a>
## [v0.18.2] - 2025-10-28
### Docs
- add CHANGELOG to v0.18.2
### Fix
- avoid b.Loop in Go 1.23
<a name="v0.18.1"></a>
## [v0.18.1] - 2025-10-28
### Docs
- add CHANGELOG for 0.18.1
### Perf
- limit memory allocation during Vector deserialization ([#759](https://github.com/Consensys/gnark-crypto/issues/759))
<a name="v0.18.0"></a>
## [v0.18.0] - 2025-06-09
### Build
- **deps:** bump golang.org/x/crypto from 0.33.0 to 0.35.0 ([#677](https://github.com/Consensys/gnark-crypto/issues/677))
### Chore
- remove unused benchmark script ([#675](https://github.com/Consensys/gnark-crypto/issues/675))
### Docs
- replace outdated link to article ([#690](https://github.com/Consensys/gnark-crypto/issues/690))
### Experiment
- vortex on koalabear ([#645](https://github.com/Consensys/gnark-crypto/issues/645))
### Feat
- remove dependency on internal package in ecc.go ([#693](https://github.com/Consensys/gnark-crypto/issues/693))
- make <31 bit field generated using uint32 ([#676](https://github.com/Consensys/gnark-crypto/issues/676))
### Fix
- remove unimplemented path from fft ext
- Eisenstein Half-GCD convergence ([#680](https://github.com/Consensys/gnark-crypto/issues/680))
### Perf
- adds avx512 poseidon2 for small fields ([#665](https://github.com/Consensys/gnark-crypto/issues/665))
- remove useless copies in e4 mul by elem ([#667](https://github.com/Consensys/gnark-crypto/issues/667))
- disable cobra bit reverse for small fields ([#662](https://github.com/Consensys/gnark-crypto/issues/662))
### Refactor
- hash to curve ([#674](https://github.com/Consensys/gnark-crypto/issues/674))
<a name="v0.17.0"></a>
## [v0.17.0] - 2025-03-11
### Chore
- remove useless github workflow
- generify poseidon2 parameters for other curves/fr ([#636](https://github.com/Consensys/gnark-crypto/issues/636))
- add auto close PR workflow ([#638](https://github.com/Consensys/gnark-crypto/issues/638))
### Feat
- baby-bear and koala-bear extensions of degree 4 ([#643](https://github.com/Consensys/gnark-crypto/issues/643))
- poseidon2 for koala-bear, baby-bear and goldilocks ([#629](https://github.com/Consensys/gnark-crypto/issues/629))
- add sis avx512 and fft avx512 for koalabear ([#622](https://github.com/Consensys/gnark-crypto/issues/622))
- add Grumpkin elliptic curve (2-cycle with BN254) ([#625](https://github.com/Consensys/gnark-crypto/issues/625))
- Poseidon2 Hash Instantiation for BLS12-377 ([#623](https://github.com/Consensys/gnark-crypto/issues/623))
### Fix
- poseidon2 templates ([#648](https://github.com/Consensys/gnark-crypto/issues/648))
- ensure fast path is taken only with fixed bound and degree ([#651](https://github.com/Consensys/gnark-crypto/issues/651))
- missing Poseidon2 round keys ([#621](https://github.com/Consensys/gnark-crypto/issues/621))
### Perf
- subgroup membership ([#635](https://github.com/Consensys/gnark-crypto/issues/635))
- Poseidon2 GKR circuit ([#628](https://github.com/Consensys/gnark-crypto/issues/628))
### Refactor
- generify small fields extensions ([#647](https://github.com/Consensys/gnark-crypto/issues/647))
- generate code for poseidon2_test ([#660](https://github.com/Consensys/gnark-crypto/issues/660))
### Style
- better github bot formatting, maybe.
- better github bot formatting, maybe.
- better github bot formatting, maybe.
### Test
- improve NAF decomposition test coverage ([#617](https://github.com/Consensys/gnark-crypto/issues/617))
<a name="v0.16.0"></a>
## [v0.16.0] - 2025-01-31
### Fix
- fixes [#619](https://github.com/Consensys/gnark-crypto/issues/619) help go mod vendor to include files ([#620](https://github.com/Consensys/gnark-crypto/issues/620))
### Perf
- Parallelize subgroup checks in KZG MPC setup ([#615](https://github.com/Consensys/gnark-crypto/issues/615))
<a name="v0.15.0"></a>
## [v0.15.0] - 2025-01-21
### Build
- update to latest bavard, paralellize addchain read ([#609](https://github.com/Consensys/gnark-crypto/issues/609))
- generify
- update ci for slack bot ([#538](https://github.com/Consensys/gnark-crypto/issues/538))
- **deps:** bump golang.org/x/crypto from 0.26.0 to 0.31.0 ([#579](https://github.com/Consensys/gnark-crypto/issues/579))
### Chore
- use errors.New to replace fmt.Errorf with no parameters ([#575](https://github.com/Consensys/gnark-crypto/issues/575))
- fix some function names in comment ([#569](https://github.com/Consensys/gnark-crypto/issues/569))
- fix some function names in comment ([#559](https://github.com/Consensys/gnark-crypto/issues/559))
- ignore gosec false positive
- use golint option to include generated files ([#549](https://github.com/Consensys/gnark-crypto/issues/549))
### Docs
- Change copyright year to 2025 (not autogenerated)
- update element package doc ([#572](https://github.com/Consensys/gnark-crypto/issues/572))
- fix non generated headers ([#571](https://github.com/Consensys/gnark-crypto/issues/571))
- less verbose Apache 2 header, latest bavard ([#570](https://github.com/Consensys/gnark-crypto/issues/570))
- document batchAddAfine func ([#555](https://github.com/Consensys/gnark-crypto/issues/555))
- GSigma -> GSigmaNeg bn254
- update README ([#533](https://github.com/Consensys/gnark-crypto/issues/533))
### Feat
- cleaned iop ([#587](https://github.com/Consensys/gnark-crypto/issues/587))
- restored sis ([#583](https://github.com/Consensys/gnark-crypto/issues/583))
- generate uint32 arith only for babybear and koalabear ([#577](https://github.com/Consensys/gnark-crypto/issues/577))
- restored fft ([#574](https://github.com/Consensys/gnark-crypto/issues/574))
- hash registry with statestorer ([#564](https://github.com/Consensys/gnark-crypto/issues/564))
- add `BabyBear` and `KoalaBear` 31bits fields vanilla Go impl ([#558](https://github.com/Consensys/gnark-crypto/issues/558))
- code gen
- example fflonk
- code gen
- example shplonk
- half-GCD for Eisenstein integers
- implement Eisenstein integers arithmetic
- number of trials for finding the next divisor of r-1 is now limited
- renaming
- fixed size folded polynomials
- renaming
- use GeneratorFullMultiplicativeGroup
- add round trip serialisation test
- renaming gamma -> challenge
- used bufMaxSizePolynomials[:len(polynomials[i])]
- randomise SRS
- code gen
- code gen
- code gen
- code gen
- code gen marshal [][][]frElement
- marshal [][][]frElement
- proof tunes the number of polynomial so it divides r-1
- add dummy polynomial so the number of poly divides r-1
- template marshal
- marshal
- marshal fflonk bn254
- code gen, tests for existence of roots of 1 need to be tuned
- added fflonk in code gen main
- add doc for fflonk
- addition of templates for fflonk
- fixed fflonk folding
- factored extended set
- getIthRootOne ok
- fflonk fold and open
- folding ok
- test roots extraction
- code gen ok
- additions of template files  marshal shplonk
- marshal ok
- generator ok
- test ok
- single point to set in open
- TestBuildZtMinusSi ok
- interpolate ok
- interpolator ok
- debugged neg pairing
- sage debugging, open ok
- wrpime ok
- add poly
- naive mul (for combinings polynomials) ok
- build vanishing poly ok
- division ok
- mul by linear factor ok
- **bw6-761:** Miller loop using direct sextic extension ([#608](https://github.com/Consensys/gnark-crypto/issues/608))
### Fix
- fixed comments and variable naming ([#603](https://github.com/Consensys/gnark-crypto/issues/603))
- re-ran go generate for field/
- buffer sufficient size ([#551](https://github.com/Consensys/gnark-crypto/issues/551))
- use sqrt in eisenstein halfgcd condition
- consider all possible remainders
- makes linter happy
- apply review suggestions
- check error
- check pairing error
- randomised size poly
- raise error when nbDigest != number of polynomials
- fixed panic
- fixed TestGetIthRootOne template
- fixed srs size
- fixed order computation template
- added condition for existence of t-th roots
- used upper bound of the total size in buf
- fixed test TestMultiplyLinearFactor
- fixed vanishing poly
### Perf
- make SetBytes not allocate errors in non-fast path ([#599](https://github.com/Consensys/gnark-crypto/issues/599))
- adds (some) arm NEON assembly for koalabear and babybear ([#588](https://github.com/Consensys/gnark-crypto/issues/588))
- adds avx512 vector ops for koalabear and babybear fields ([#568](https://github.com/Consensys/gnark-crypto/issues/568))
- adds (some) `arm64` assembly for 4 and 6 words modulus ([#554](https://github.com/Consensys/gnark-crypto/issues/554))
- subgroup check ([#557](https://github.com/Consensys/gnark-crypto/issues/557))
- adds AVX512 implementations of vector.Sum, vector.InnerProduct + assembly refactor ([#547](https://github.com/Consensys/gnark-crypto/issues/547))
- add Vector::Add, Sub and ScalarMul assembly (and purego) implementations ([#536](https://github.com/Consensys/gnark-crypto/issues/536))
### Refactor
- clean up SIS, placeholder ([#596](https://github.com/Consensys/gnark-crypto/issues/596))
- remove dependency on ecc Encoder in fft domain marshal ([#565](https://github.com/Consensys/gnark-crypto/issues/565))
- apply review suggestions
- move eisenstein under field/
- up doc
- kill bls12-378 and bw6-756 curves
### Release
- v0.15.0 ([#611](https://github.com/Consensys/gnark-crypto/issues/611))
### Style
- clean comments
- nbInstances -> nbPolynomials
- remaing -> remaining
### Test
- fix windows build
- fix windows build
- half-GCD test with bigger integers
### Typos
- squashed typo fixes ([#610](https://github.com/Consensys/gnark-crypto/issues/610))

### Pull Requests
- Merge pull request [#548](https://github.com/Consensys/gnark-crypto/issues/548) from Consensys/docs/pedersen
- Merge pull request [#498](https://github.com/Consensys/gnark-crypto/issues/498) from Consensys/feat/fflonk
- Merge pull request [#543](https://github.com/Consensys/gnark-crypto/issues/543) from Consensys/feat/eisenstein
- Merge pull request [#534](https://github.com/Consensys/gnark-crypto/issues/534) from Consensys/refactor/kill-some-curves
<a name="v0.14.0"></a>
## [v0.14.0] - 2024-09-03
### Build
- update runner and go version ([#531](https://github.com/Consensys/gnark-crypto/issues/531))
- generify g2.go changes
- generify pedersen_test.go changes
- generify pedersen.go changes
### Chore
- fix some function names ([#526](https://github.com/Consensys/gnark-crypto/issues/526))
- go generate
- go generate
- go generate
### Docs
- describe the prover can provide folded proof
- add documentation examples
- more explanations
- add package documentation
- remove TODO
- expand method documentation
- correct BW6 tower comment
- up E3 mul comment
- **ecc:** refactor comments in g*.go
### Feat
- use curve fold
- define fold method for curve
- use foldable Pedersen commitment verification
- pedersen batch verify bn254
### Fix
- fixes [#522](https://github.com/Consensys/gnark-crypto/issues/522) with bound check before computing twiddles when domain has no precompute set ([#523](https://github.com/Consensys/gnark-crypto/issues/523))
- resolve conflict around kzg
### Perf
- remove uneeded bucket init in affine coordinates fixes [#529](https://github.com/Consensys/gnark-crypto/issues/529) ([#530](https://github.com/Consensys/gnark-crypto/issues/530))
### Refactor
- variable name
- option outputs option
- typos in comments
- correct IsZero comments
### Test
- add test for prefolded pok
- fix example test

### Pull Requests
- Merge pull request [#524](https://github.com/Consensys/gnark-crypto/issues/524) from Consensys/refactor/bls12-MillerLoopFixed
- Merge pull request [#527](https://github.com/Consensys/gnark-crypto/issues/527) from Consensys/feat/pedersen-batchmultivk
- Merge pull request [#517](https://github.com/Consensys/gnark-crypto/issues/517) from Consensys/feat/pedersen-batch-verify
- Merge pull request [#511](https://github.com/Consensys/gnark-crypto/issues/511) from Consensys/docs/up-comments
<a name="v0.13.0"></a>
## [v0.13.0] - 2024-07-19
### CI
- new attempt to fix push to master workflow
- grmpppf
- add missing message field
- use runner.os
- update ci workflows ([#460](https://github.com/Consensys/gnark-crypto/issues/460))
### Chore
- generate
- go generate
- make recoverP internal
- go generate
- **deps:** bump golang.org/x/crypto from 0.10.0 to 0.17.0 ([#473](https://github.com/Consensys/gnark-crypto/issues/473))
### Docs
- describe length extension attack for MiMC hash function. ([#520](https://github.com/Consensys/gnark-crypto/issues/520))
- fix misspelled comment
- skipping inf points in PairFixedQ is implicit
- **ecc:** add comment in IsOnCurve()
- update hash to curve rfc ref and impl comment ([#478](https://github.com/Consensys/gnark-crypto/issues/478))
### Feat
- return constant error for quadratic non residues
- mathfmt doc
- added some docs
- fixed comment, GenFrStar -> GeneratorFullMultiplicativeGroup
- adds byteorder option in mimc `New`  ([#480](https://github.com/Consensys/gnark-crypto/issues/480))
- allow hashing Fiat-Shamir transcript challenge without hashing to field first ([#474](https://github.com/Consensys/gnark-crypto/issues/474))
- kzg.SRS implements UnsafeReadFrom
- added missing WriteRawTo method for kzg.SRS
- added ecc.IDFromString() method
- wrap existing hash-to-field for implementing hash.Hash interface ([#464](https://github.com/Consensys/gnark-crypto/issues/464))
- quick SRS is now balanced
- code gen for quick srs
- avoid batch scalar mul when alpha=1 bn254
### Fix
- m range from 1 in test
- sis limb-decomposition works with log-two_bound > 8
- fixed typo
- template fix
- fixes [#491](https://github.com/Consensys/gnark-crypto/issues/491)
- gkr add gate evaluate ([#490](https://github.com/Consensys/gnark-crypto/issues/490))
- handle edge case where provided domain cardinality is 1 ([#454](https://github.com/Consensys/gnark-crypto/issues/454))
- kzg template for bn254
- remove duplicate gkr template generation
- **bandersnatch:** GLV bounds + test ([#516](https://github.com/Consensys/gnark-crypto/issues/516))
- **kzg:** bw6-633 vk marshal
### Perf
- batchInvert 1/P.Y in fixed-arg pairing
- unsafe.WriteSlice ([#501](https://github.com/Consensys/gnark-crypto/issues/501)) ([#503](https://github.com/Consensys/gnark-crypto/issues/503))
- add kzg.UnsafeToBytes and kzg.UnsafeFromBytes methods ([#499](https://github.com/Consensys/gnark-crypto/issues/499))
- specialize mul by affine lines
- **bls12-377:** precomuting lines in pairing in affine
- **bls12-377:** use precomputed affine lines in kzg
- **bls12-378:** use precomputed affine lines in kzg
- **bls12-378:** precomuting lines in pairing in affine
- **bls12-381:** use precomputed affine lines in kzg
- **bls12-381:** precomuting lines in pairing in affine
- **bls24-315:** precomuting lines in pairing in affine
- **bls24-315:** use precomputed affine lines in kzg
- **bls24-317:** precomuting lines in pairing in affine
- **bls24-317:** use precomputed affine lines in kzg
- **bn254:** precomuting lines in pairing in affine
- **bn254:** use precomputed affine lines in kzg
- **bn254:** optimize affine pairing with doubleAndAdd
- **bw6:** precomuting lines in pairing
- **bw6:** precomuting lines in pairing in affine
- **bw6:** use precomputed affine lines in kzg
- **bw6-633:** remove unnecessary computations in MillerLoop
- **bw6-633:** precomuting lines in pairing in affine
- **bw6-633:** implement faster final exp hard part
- **bw6-756:** implement faster final exp hard part
- **bw6-756:** precomuting lines in pairing in affine + kzg
- **bw6-756:** remove unnecessary computations in MillerLoop
- **bw6-761:** optimize affine pairing with doubleAndAdd
- **bw6-761:** remove unnecessary computations in MillerLoop
- **bw6-761:** implement faster final exp hard part
- **ecc:** faster affine Add
- **ecc:** use DoubleMixed in AddMixed when points are equal
- **ecc:** optimize affine Add, Sub and Double
- **ecc:** scpecialize mul by B in IsOnCurve
- **ecc:** save 1 mul in IsOnCurve
- **kzg:** replace 2 scalarMuls with a JointScalarMul
### Refactor
- remove G1 from KZG verifiying key
- move test util package into isolated one ([#500](https://github.com/Consensys/gnark-crypto/issues/500))
- remove dead ScalarMultiplicationAffine
- make pedersen vk fields public ([#488](https://github.com/Consensys/gnark-crypto/issues/488))
- kzg srs template ([#475](https://github.com/Consensys/gnark-crypto/issues/475))
- re-introduce points in KZG verification key
- apply code review suggestions
- consolidate fixed-argument pairing across curves
- remove E3 Conjugate from BW6
- generify and consolidate new kzg and fixed-arg pairing
- **bw6:** simplify PrecomputeLines
- **bw6-633:** use revisited Ate pairing instead of Tate
- **bw6-633:** clean pairing
- **bw6-756:** use revisited Ate pairing instead of Tate
- **bw6-761:** clean and optimize new pairing
- **bw6-761:** use revisited Ate pairing instead of Tate
### Test
- update the tests
- **ecc:** affine Add involving (0,0)
- **ecc:** benchmark affine ops

### Pull Requests
- Merge pull request [#518](https://github.com/Consensys/gnark-crypto/issues/518) from Consensys/fix/flaky-fri-test
- Merge pull request [#506](https://github.com/Consensys/gnark-crypto/issues/506) from Consensys/perf/affine-pairing
- Merge pull request [#514](https://github.com/Consensys/gnark-crypto/issues/514) from Consensys/refactor/bw6-conj
- Merge pull request [#512](https://github.com/Consensys/gnark-crypto/issues/512) from Consensys/fix/sis-issue-when-logTwoBound-larger-than-8
- Merge pull request [#493](https://github.com/Consensys/gnark-crypto/issues/493) from Consensys/perf/kzg-verify
- Merge pull request [#510](https://github.com/Consensys/gnark-crypto/issues/510) from Consensys/perf/faster-affine-arithmetic
- Merge pull request [#509](https://github.com/Consensys/gnark-crypto/issues/509) from Consensys/perf/affine-arithmetic
- Merge pull request [#497](https://github.com/Consensys/gnark-crypto/issues/497) from Consensys/feat/ecdsa-errors
- Merge pull request [#496](https://github.com/Consensys/gnark-crypto/issues/496) from Consensys/sis/export-fft64
- Merge pull request [#492](https://github.com/Consensys/gnark-crypto/issues/492) from Consensys/feat/gen_fr_star
- Merge pull request [#487](https://github.com/Consensys/gnark-crypto/issues/487) from Consensys/perf/IsOnCurve
- Merge pull request [#469](https://github.com/Consensys/gnark-crypto/issues/469) from Consensys/kzg/srs-vk
- Merge pull request [#466](https://github.com/Consensys/gnark-crypto/issues/466) from Consensys/kzg/srs
- Merge pull request [#439](https://github.com/Consensys/gnark-crypto/issues/439) from Consensys/fix/duplicate-gkr-gen
- Merge pull request [#465](https://github.com/Consensys/gnark-crypto/issues/465) from Consensys/refactor/bw6761-pairing
- Merge pull request [#459](https://github.com/Consensys/gnark-crypto/issues/459) from Consensys/perf/bw6-finalExp
- Merge pull request [#458](https://github.com/Consensys/gnark-crypto/issues/458) from Consensys/feat/kzg_dummy_srs
<a name="v0.12.1"></a>
## [v0.12.1] - 2023-10-05
### Clean
- remove useless snippet in mulWindowed ([#453](https://github.com/Consensys/gnark-crypto/issues/453))
### Ecc
- fix neutral elements cases
- add bench for Jac equality check
- avoid inverses in Jac equality check
### Edwards
- add bench for Projective equality check
- avoid inverses in Projective equality check
### Feat
- discared useless files
- added test for [-s]G=-[s]G
### Fix
- add GT ExpGLV fix to BLS24 + BW6
- merge and fix tests for expGLV
- use max(size(s1), size(s2)) for ExpGLV
- fixed mulWindowed for negative exponents
- fixed ExpGLV in GT
### Perf
- **fft:** introduce cache efficient bit reverse shuffling ([#446](https://github.com/Consensys/gnark-crypto/issues/446))

### Pull Requests
- Merge pull request [#451](https://github.com/Consensys/gnark-crypto/issues/451) from Consensys/fix/glv
- Merge pull request [#450](https://github.com/Consensys/gnark-crypto/issues/450) from jsign/jsign-eq-improv
<a name="v0.12.0"></a>
## [v0.12.0] - 2023-09-27
### CI
- update ci workflows ([#447](https://github.com/Consensys/gnark-crypto/issues/447))
### Chore
- generate
- generate
- generate
- use *big.Int (convention)
- generate
- generate ecdsa
- generate tests
- remove extra method description
- remove extra file
### Edwards
- improve the performance of Add, MixedAdd and IsOnCurve ([#441](https://github.com/Consensys/gnark-crypto/issues/441))
### Feat
- delete playground
- added tests for non malleability of ecdsa
- added tests for non malleability (eddsa)
- ErrShortBuffer -> ErrWrongSizeBuffer
### Fix
- remove test that signature value R.X == 0
- check EdDSA signature values not zero
- compare ecdsa signature values against zero
- removed const error ecdsa
- remove swifft files, removed const error
- fixed malleability issue ecdsa/eddsa
- fixed ambiguous unmarshalling eddsa
### Multiexp
- avoid direct coordinate access to check for zero points ([#414](https://github.com/Consensys/gnark-crypto/issues/414))
### Refactor
- make marshal errors private
### Test
- ensure the test path is taken
- add tests for zero values
- implement malleability tests as subtests

### Pull Requests
- Merge pull request [#449](https://github.com/Consensys/gnark-crypto/issues/449) from Consensys/fix/malleability_sig
- Merge pull request [#442](https://github.com/Consensys/gnark-crypto/issues/442) from jsign/jsign-edwards-extended
<a name="v0.11.2"></a>
## [v0.11.2] - 2023-08-18
### Build
- update go mod deps
- generify gates var
- generify unsafeReadFrom
- generify enc.WriteUint64SliceSlice
- generify marshal changes
- generify
- generify bn254 changes
### Docs
- ConsenSys -> Consensys ([#406](https://github.com/Consensys/gnark-crypto/issues/406))
- explain FoldCommitments
### Edwards
- optimize point negation ([#413](https://github.com/Consensys/gnark-crypto/issues/413))
### Feat
- added size accessor on poly
- added polynomial.SetSize
- add AsyncReadFrom to fr.Vector and fft.Domain ([#424](https://github.com/Consensys/gnark-crypto/issues/424))
- added WriteRawTo to kzg.ProvingKey ([#422](https://github.com/Consensys/gnark-crypto/issues/422))
- more gkr gates
- Gates var for bn254
- added element.Unmarshal closes [#348](https://github.com/Consensys/gnark-crypto/issues/348)
- pedersen.Vk[Bn254].UnsafeReadFrom
- write [][]uint64
- marshal [][]uint64 for bn254
- pedersen(bn254).WriteRawTo
- fold 0 proofs edge case
- pedersen folding proofs
- Marshal [][]fr.Element ([#400](https://github.com/Consensys/gnark-crypto/issues/400))
### Fix
- add explicit operator priority
- ECDSA HashToInt bytes-bits mismatch ([#428](https://github.com/Consensys/gnark-crypto/issues/428))
- dont use internal package in vector.go
- use of MulGate
- remove redundant mulGate
- incorrect semaphore initialization ([#411](https://github.com/Consensys/gnark-crypto/issues/411))
- do not read empty slices as nil
- pedersen folding bug
- more innef assign
- innef assignment in test
### Msm
- semaphore to limit CPUs + better split strategy (up to 25% perf boost on 96cores) ([#403](https://github.com/Consensys/gnark-crypto/issues/403))
### Perf
- dont waste too much memory in fft domain ([#437](https://github.com/Consensys/gnark-crypto/issues/437))
- small memory alloc optimization ([#435](https://github.com/Consensys/gnark-crypto/issues/435))
- fast path for SIS with logTwoBound: 8, logTwoDegree: 6 ([#416](https://github.com/Consensys/gnark-crypto/issues/416))
### Style
- move aux func down
### Test
- check error

### Pull Requests
- Merge pull request [#438](https://github.com/Consensys/gnark-crypto/issues/438) from Consensys/develop
- Merge pull request [#419](https://github.com/Consensys/gnark-crypto/issues/419) from Consensys/feat/gkr-custom-gates
- Merge pull request [#410](https://github.com/Consensys/gnark-crypto/issues/410) from ConsenSys/fix/decode-empty-slice
- Merge pull request [#407](https://github.com/Consensys/gnark-crypto/issues/407) from ConsenSys/feat/fold-pedersen
<a name="v0.11.1"></a>
## [v0.11.1] - 2023-07-11
### Chore
- update CHANGELOG
### Fix
- ECDSA HashToInt bytes-bits mismatch ([#428](https://github.com/Consensys/gnark-crypto/issues/428))

### Pull Requests
- Merge pull request [#429](https://github.com/Consensys/gnark-crypto/issues/429) from Consensys/release/v0.11.1
<a name="v0.11.0"></a>
## [v0.11.0] - 2023-05-02
### Build
- go generate
- generify bn254 changes
- generify bn254 kzg changes
- generify marshal changes
- generify bn254 kzg changes
- bump go1.20
- update ci github action dependencies
### Chore
- PR feedback
### Docs
- make comments more godoc friendly
- remove comment
- remove DO NOT EDIT from non-autogenerated files
### Feat
- fix v computation in ECDSA signature ([#385](https://github.com/Consensys/gnark-crypto/issues/385))
- make `mapToCurve` public to allow for custom cofactor clearing ([#372](https://github.com/Consensys/gnark-crypto/issues/372))
- add Double in affine coordinates
- kzg.Vk.WriteRawTo
- bn254 encoder to support uint64 slices
- **pairing:** return 1 after easy part if result is 1
### Fix
- handle all bitmask in point deserialization
- littleEndian -> bigEndian
- import utils
- don't ignore multiexp error
- minor errors
- generation mistake
- bn254 incorporate evals into kzg batch challenge
- **kzg:** nb of digests in BatchVerifyMultiPoints should be nonzeo
- **linter:** ineffassign in Fpk marshal
### Perf
- **kzg:** remove G2 scalar mul in single verification
### Refactor
- break pedersen key into proving (committing) and verifying
- move roundtrip func to utils
- reflect kzg changes in plookup
- reflect kzg changes in permutation
- kzg tests
- kzg.NewSRS to return two structs
- marshal pk, vk separately
- break up kzg srs; NewSRS for bn254
- export endomorphisms
- **BatchDecompressKarabina:** handle g2=g3=0 case "explicitly"
- **DecompressKarabina:** handle g2=g3=0 case "explicitly"
### Revert
- do not export endomorphisms
- NewSRS to return a pointer again
- revive whole SRS serialization
- reflect kzg.srs revival in other packages
- limited revival of kzg.Srs
### Style
- no loop when serializing one object only

### Pull Requests
- Merge pull request [#391](https://github.com/Consensys/gnark-crypto/issues/391) from ConsenSys/develop
- Merge pull request [#386](https://github.com/Consensys/gnark-crypto/issues/386) from ConsenSys/perf/kzg-verify
- Merge pull request [#384](https://github.com/Consensys/gnark-crypto/issues/384) from ConsenSys/refactor/break-pedersen-notowermod
- Merge pull request [#379](https://github.com/Consensys/gnark-crypto/issues/379) from ConsenSys/feat/encode-uint64-slices
- Merge pull request [#378](https://github.com/Consensys/gnark-crypto/issues/378) from ConsenSys/refactor/break-kzg-srs
- Merge pull request [#382](https://github.com/Consensys/gnark-crypto/issues/382) from ConsenSys/refactor/not-export-endo
- Merge pull request [#380](https://github.com/Consensys/gnark-crypto/issues/380) from omerfirmak/elim-pedersen-alloc
- Merge pull request [#374](https://github.com/Consensys/gnark-crypto/issues/374) from ConsenSys/fix/BatchVerifyMultiPoints-0
- Merge pull request [#376](https://github.com/Consensys/gnark-crypto/issues/376) from ConsenSys/refactor/export-endo
- Merge pull request [#375](https://github.com/Consensys/gnark-crypto/issues/375) from ConsenSys/refactor/FinalExp
- Merge pull request [#377](https://github.com/Consensys/gnark-crypto/issues/377) from ConsenSys/fix/fold-kzg-fs
- Merge pull request [#371](https://github.com/Consensys/gnark-crypto/issues/371) from omerfirmak/elim-pedersen-alloc
- Merge pull request [#369](https://github.com/Consensys/gnark-crypto/issues/369) from ConsenSys/build/updateci
<a name="v0.10.1"></a>
## [v0.10.1] - 2023-07-11
### Chore
- update CHANGELOG
### Fix
- ECDSA HashToInt bytes-bits mismatch ([#428](https://github.com/Consensys/gnark-crypto/issues/428))
<a name="v0.10.0"></a>
## [v0.10.0] - 2023-03-29
### Build
- ran go generate
### Docs
- added Generator docs
- add TODO with small domain warning
- **pairing:** add comments
### Feat
- add ECDSA public key recover from message, signature and recovery info ([#347](https://github.com/Consensys/gnark-crypto/issues/347))
- added comment for bitAt function
- added sis.py
- add reference test
- merge fft branch and cosmetic edits
- parallel.Execute with 1 task fast path
- added fft.WithNbTasks
- FFT signature now takes an option variadic
- expose NaiveMulMod for gnark
- experiment parallel sis
- restored latest fft
- restored fp, fr, etc
### Fix
- code generation fp6
- mods smaller than 5; overestimating nbword
- invalid infinity point encoding
- SIS on 64bit only
- **MillerLoop:** return 1 when size(pairs)=0 after infinity filter
### Perf
- faster init + simplify bit processing loop
- fix todo with small domain path
- optimized sis
- kzg BatchOpenSinglePoint more parallelization
- kept the fastest method for sis
- use bitset in batch invert
- iop ratio more parallelization. reuse cosets
- 3 muls instead of 4 sis
- less memallocs, check for zeroes
- experiment without memcopy not great
- minor optim in iop package
- **bls24-317:** optimize final exp
- **bn254:** use Fuentes et al. instead of Duquesne-Ghammam hard part
- **bw6-756:** optimize GT subgroup membership
- **pairings:** isolate first iteration to avoid a MulByLine
- **pairings:** isolate last iteration to avoid a double/add
- **stark-curve:** no subgroup check on prime-order curve
### Refactor
- expose fft.Generator() method
- export bls12-381 tower
- expose bn254 tower to gnark
- move Hash in _test.go file
- fft.WithCoset() -> fft.OnCoset()
- sis.py -> sis.sage
- **pairings:** make DoubleStep and AddMixedStep private
### Style
- remove dead comments
- code cleaning
- more code cleaning
- code cleaning
- cosmetic edits
### Test
- minor fixes and skip 32bit
- update test_cases.json
- skip tensor commitment test on 32bit arch for now
- added sis sage test case generation
- still good

### Pull Requests
- Merge pull request [#368](https://github.com/Consensys/gnark-crypto/issues/368) from ConsenSys/develop
- Merge pull request [#366](https://github.com/Consensys/gnark-crypto/issues/366) from ConsenSys/refactor/pairings
- Merge pull request [#365](https://github.com/Consensys/gnark-crypto/issues/365) from ConsenSys/fix/gentime-field-bugs
- Merge pull request [#364](https://github.com/Consensys/gnark-crypto/issues/364) from jtraglia/update-tested-go-versions
- Merge pull request [#363](https://github.com/Consensys/gnark-crypto/issues/363) from ConsenSys/fix/infencoding
- Merge pull request [#361](https://github.com/Consensys/gnark-crypto/issues/361) from ConsenSys/perf/plonk
- Merge pull request [#360](https://github.com/Consensys/gnark-crypto/issues/360) from ConsenSys/bls12381-tower-gnark
- Merge pull request [#359](https://github.com/Consensys/gnark-crypto/issues/359) from SherLzp/develop
- Merge pull request [#356](https://github.com/Consensys/gnark-crypto/issues/356) from ConsenSys/perf/bn24317-FinalExp
- Merge pull request [#354](https://github.com/Consensys/gnark-crypto/issues/354) from ConsenSys/bn254-tower-gnark
- Merge pull request [#351](https://github.com/Consensys/gnark-crypto/issues/351) from ConsenSys/perf/bw6-756/GT-subgroup-check
- Merge pull request [#349](https://github.com/Consensys/gnark-crypto/issues/349) from ConsenSys/perf/subgroup-check-stark
- Merge pull request [#344](https://github.com/Consensys/gnark-crypto/issues/344) from ConsenSys/perf/tensor-commitment
- Merge pull request [#345](https://github.com/Consensys/gnark-crypto/issues/345) from ConsenSys/feat/fftopt
- Merge pull request [#263](https://github.com/Consensys/gnark-crypto/issues/263) from AlexandreBelling/perf/tensor-commitment
- Merge pull request [#260](https://github.com/Consensys/gnark-crypto/issues/260) from AlexandreBelling/experimental/tensor-commitment
<a name="v0.9.2"></a>
## [v0.9.2] - 2023-07-11
### Chore
- update CHANGELOG
### Fix
- ECDSA HashToInt bytes-bits mismatch ([#428](https://github.com/Consensys/gnark-crypto/issues/428))
<a name="v0.9.1"></a>
## [v0.9.1] - 2023-02-14
### Build
- go generate
- weird staticcheck rule
- generify
- generify
- generify parallel .Complete
- generify fixes
- generify parallel computeGJ
- generify bn254 mimc changes
- remove G2Exist flag from curves config
### Chore
- remove safeStack
- simpler dst for hash.WriteString
- staticcheck
- more ToBigIntRegular -> BigInt and another little thing
- ToBigIntRegular -> BigInt
- remove codegen temp files
- delete temp files
- setNbOutputs covered by topSort now
- remove two other topological sorts
### Docs
- fix incorrect statement
- innerWork
- update DOI in README.md
### Feat
- SliceToElementSlice to be generic
- some gkr; claim/subclaim in sumcheck seems untenable
- Merkle tree to panic upon hash write error
- pedersen_hash added to stark curve
- signing interface changes and bn254 impl
- add Vector support to ecc marshal encoder ([#336](https://github.com/Consensys/gnark-crypto/issues/336))
- gkr.Circuit.MemoryRequirements
- can pass workers pool in
- parallel versions of computing gate inputs
- more "subtle" parallelization tools
- edDSA to write element by element
- hash.ToField compatible with hash.Hash
- generify parallel.Execute
- pre-hashed message in eddsa
- pre-hashed message in ecdsa
- thread-safe pool
- adds Vector in field package
- first sumcheck attempt. Need specialized fiatshamir
- more usable WriteString instead of "Decompose"
- generify Decomposition solution
- decomposing long challenge names bn254
- generic sumcheck prover and verifier
- start ECDSA on secp256k1
- generate fp with addchain stark-curve
- MSM and BatchScalarMul on secp256k1
- ListHash
- generify MessageCounter fix
- generify gkr test vec gen fixes
- generify gkr fixes
- generify non-test vector gkr test changes
- generify gkr changes
- generify sumcheck changes
- small_rational to and from bytes
- generified gkr tests. test case generation still not working
- some test generification
- generify gkr.tmpl
- edDsa SignFr and VerifyFr
- nuke element.Bit() closes [#306](https://github.com/Consensys/gnark-crypto/issues/306) ([#331](https://github.com/Consensys/gnark-crypto/issues/331))
- generic pool
- generify multi-fan-out input bugfix
- generified benchmark, parallelism and top sort
- add minimal implem of stark curve
- clean ups
- basic benchmarking
- test vector utils for all
- a first attempt at separate test vector utils
- parsing polynomials; more trouble than worth
- remove unused test hash entries
- gkr test vector codegen works for rationals
- autogen gkr tests
- code generation for poly.Text
- better polynomial.Text for bn254
- small_rational to use big.Int underneath
- identity gate, 2 instances test on Q - prover side
- Load test circuit and test hash function
- sumcheck for rationals
- more smallrational features
- polynomial codegen to use simpler FieldDependency data structure
- all necessary functions for the polynomial package
- codegen for gkr
- neg, sub, equal, double
- codegen for sumcheck
- some experimental rational
- just a few lines
- gkr verifier
- Verify stub. shared claim mgmt structure. prover refactor to come
- gkr prover
- gkr sumcheck combine and computeGJ
- some gkr prover stuff
- three kinds of topological sort
- MORE ABSTRACTION 🥲 Evaluating the final claim may be hard
- **ecdsa:** use aes-ctr in signing randomness
- **templates:** ecdsa package for all curves
### Fix
- static checks
- no repeated claims in FinalEvalProof
- mimc pow7
- single multilin test works
- number of rounds for mimc
- sumcheck test
- parallelization bug
- remove mimc printfs
- WriteString returns no error now
- side-stepped thread-safe pool issue
- some debugging
- serialization of stark-curve points
- Multilin tests with toy input-independent hash work
- use ProveFinalEval, make member vars of proof public
- eddsa: separate field element and non-field-element hash inputs
- minor error
- eddsa to use WriteString
- align challenge name, prepending with zeros to avoid field overflow
- ecdsa on secp256k1 working
- merge with develop
- an attempt at input verification
- computeGJ works correctly for mul gate
- mul gate test passes
- package name
- correct field type
- generify the fix
- verifier input building bug
- more noRedundancy bugs
- noRedundancy bugs
- multi-counting outs to the same wire
- ignored errors in sumcheck.setupTranscript
- sumcheck "bad proof accepted"
- accepting bad proofs
- sumcheck claim test
- MessageCounter.Reset
- staticcheck
- uncomment gkr test vec gen
- gkr test-vec gen errors
- generify changes in small_rational maphashes
- fake hash consistency
- staticcheck
- some gkr test vec errors
- no hardcoded fr.Element in sumcheck test case gen
- gkr challenges for SmallRational
- sumcheck testcase gen compiles
- gkr test case generator compiles
- gkr_test files compile
- bn254 sumcheck tests pass
- fiatshamir to write bindings one by one
- single input two identity gates test fixed, but fails
- closes [#316](https://github.com/Consensys/gnark-crypto/issues/316) big int pool doesn't allow nil values
- no expectation of ins/outs lumped together in tests
- topological sort no longer group inputs and outputs
- reintroduce some deleted autogens
- generification
- bn254 seems correct. test vector reordering needed
- gofmt
- yoda, multi-fan-out input bug
- remove "enormousArray" hack, minimize unsafe.Pointer use
- free mallocs
- an "evaluate" missing the pool argument
- errors in rational sumcheck
- remove unreachable return
- writeKey quotation marks
- save new hash entries
- still inconsistency betw rational and bn254
- snapshot for hash inconsistency error
- remove auto-generated test files
- test vec gen works
- manually edited main.go for gkr test vec gen
- snapshot: gkr test vector generation
- wrong type for wireAssignent
- minor ones
- rough edges for rationals
- some corrections for rationals
- restored hash elements incorrectly marked as unnecessary
- more staticcheck fixes
- better string search to make staticcheck happy
- no redeclaring "err" to make staticcheck happy
- all gkr tests pass
- no compile errors, rational cases pass
- setElement more consistent with SetInterface
- setElement method that would work on fr and rationals
- codegen for gkr test cases, compiles
- gkr tests on rationals pass
- no in-place operations
- id circuit verification passes
- fewer gkr test instances by default
- some codegen mistakes
- many instances of small circuits
- remove challengeSeed from sumcheck input
- **EdDSA:** enforce hash function as FS challenge
- **curves:** copy scalars in JointScalarMul
- **ecdsa:** avoid only 1 inverse in verify
- **secp256k1:** bound cRrange by 15 for MSM
- **secp256k1:** code generation for msm
- **stark-curve:** point encoder (a!=0)
- **stark-curve:** svdw hash-to-curve (coeff A!=0)
- **template:** entries var re-declared
### Perf
- parallelize computeGJ more effectively
- parallel Fold
- step 2 Combine
- step 1, use the new pool for computeGJ and next
- minor adjustements to iop package ([#334](https://github.com/Consensys/gnark-crypto/issues/334))
- mimc on bls12-377/fr uses x^17 as a permutation
- multilin.Fold without writing top part in memory
- parallel WireAssignment.Complete for bn254
- simplify gkr pool
- workerPool in gkr
- reduced many memory leaks to two
- parallelism in computeGJ but faulty
- concurrency in computeGJ. to make pool maps thread-safe
- some rudimentary parallelism
- **ecdsa:** avoid 2 inverses in verify (affine scalarMul)
- **ecdsa:** no bigInt allocation
### Refactor
- include secp256k1 in code generation
- move some routines from gkrVectors
- Pedersen hash on stark-curve
- remove Decompose entirely
- plookup.Table -> fr.Vector
- element imports field/pool and field/hash to avoid cycles
- remove dead code (FFT params for secp256k1)
- concile ecdsa and eddsa on signature interface
- smallRational out of gkr
- remove fr/ packages from secp256k1
- generify most gkr test vector ops
- remove "proofEquals" from gkr test vgen
- remove subclaim class
- remove ToField interface
- get rid of ParsedTestCase
- include secp256k1 in code generation
- move all fake hashes to test_vector_utils
- cleaner sumcheck test-vec generation
- clean iop APIs + add Marshal methods ([#337](https://github.com/Consensys/gnark-crypto/issues/337))
- gkr tests to work with fiatshamir.transcript
- remove unused file
- fixing gkr tests
- sumcheck and gkr to use fiatshamir package
- **ecdsa:** follows SEC 1 v-2
- **ecdsa:** make public params exportable to gnark
- **ecdsa:** remove Params struct
- **ecdsa:** marshal + follow interface
- **ecdsa:** make hashToInt accessible for gnark
### Revert
- deleted mimc.Write documentation
- Remove EdDSA SignNum and VerifyNum
- remove SignNum and VerifyNum
- mimc to take in multiple elements
- don't doubly hash finalEvalProof when an input wire is skipped
- roll back cgo in memory management
### Style
- some simplification, remove some dead code
- remove unnecessary variable
- remove some dead code
- format
- rename rational_cases to test_vectors
- don't name unused variables
- improved comments
### Test
- more debug output
- printfs in mimc
- mimc to write out its ins and outs
- a couple of benchmarks with some printfs
- add ecdsa benchmarks
- failing test
- replicate failure in small_rational
- MapHash consistency
- vector with 4 instances
- bigger instances
- add failing verification tests
- trilinear - single claim
- a simple test vector for sumcheck
- rational using big.Int is a failure
- mimc test fails, most likely due to overflow
- mimc; rational overflowing. TODO: Replace int64 with big.Int
- vectors: more
- some extra logging for 16M memory failure
- more tests, mimc circuit fails
- more usable message counter
- fundamental flaw found in computeGJ
- failing test for mul gate
- simplest possible test, two instances of an identity gate

### Pull Requests
- Merge pull request [#342](https://github.com/Consensys/gnark-crypto/issues/342) from ConsenSys/develop
- Merge pull request [#308](https://github.com/Consensys/gnark-crypto/issues/308) from ConsenSys/304-field-agnostic-fiat-shamir-challenge-names
- Merge pull request [#319](https://github.com/Consensys/gnark-crypto/issues/319) from ConsenSys/perf/gkr
- Merge pull request [#328](https://github.com/Consensys/gnark-crypto/issues/328) from ConsenSys/perf/gkrpool
- Merge pull request [#333](https://github.com/Consensys/gnark-crypto/issues/333) from ConsenSys/fix/mimc-pow7
- Merge pull request [#327](https://github.com/Consensys/gnark-crypto/issues/327) from ConsenSys/pedersen-hash
- Merge pull request [#320](https://github.com/Consensys/gnark-crypto/issues/320) from ConsenSys/fix/mimc-rounds
- Merge pull request [#312](https://github.com/Consensys/gnark-crypto/issues/312) from ConsenSys/signature/pre-hashed
- Merge pull request [#311](https://github.com/Consensys/gnark-crypto/issues/311) from ConsenSys/feat/element/vector
- Merge pull request [#310](https://github.com/Consensys/gnark-crypto/issues/310) from ConsenSys/feat/ecdsa
- Merge pull request [#298](https://github.com/Consensys/gnark-crypto/issues/298) from ConsenSys/secp256k1/MSM
- Merge pull request [#309](https://github.com/Consensys/gnark-crypto/issues/309) from ConsenSys/chore/tobigintregular-deprecated
- Merge pull request [#305](https://github.com/Consensys/gnark-crypto/issues/305) from ConsenSys/test/fiatshamir-mimc
- Merge pull request [#299](https://github.com/Consensys/gnark-crypto/issues/299) from ConsenSys/feat/stark-curve
- Merge pull request [#285](https://github.com/Consensys/gnark-crypto/issues/285) from ConsenSys/refac/gkr-fiatshamir
- Merge pull request [#259](https://github.com/Consensys/gnark-crypto/issues/259) from ConsenSys/bench/gkr
<a name="v0.9.0"></a>
## [v0.9.0] - 2023-01-05
### Build
- re-ran gofmt with go1.19, updated ci to that
### Docs
- added audit.pdf and updated link in README
### Feat
- field.Hash ([#271](https://github.com/Consensys/gnark-crypto/issues/271))
- add secp256k1 curve
### Fix
- make BigInt a pointer receiver
- remove generated fr/pedersen package from secp256k1
- rebase on develop
- no compressed marshall because no spare bit
- **secp256k1:** no partitionScalars because there is no spare bit
- typos ([#262](https://github.com/Consensys/gnark-crypto/issues/262))
### Perf
- **MSM:** save 4 mul in ext-Jac add
- **bn254:** faster subgroup membership
### Refactor
- rebase on develop (field api)
- mark ToBigIntRegular as deprecated, introduce BigInt method ([#290](https://github.com/Consensys/gnark-crypto/issues/290))
- clean code generation for endo-based computations ([#281](https://github.com/Consensys/gnark-crypto/issues/281))
- **bn254:** remove unused variable (fixedCoeff)
### Style
- typo

### Pull Requests
- Merge pull request [#297](https://github.com/Consensys/gnark-crypto/issues/297) from ConsenSys/develop
- Merge pull request [#277](https://github.com/Consensys/gnark-crypto/issues/277) from ConsenSys/feat/secp256k1
- Merge pull request [#251](https://github.com/Consensys/gnark-crypto/issues/251) from ConsenSys/perf/IsInSubGroup-BN
- Merge pull request [#250](https://github.com/Consensys/gnark-crypto/issues/250) from ConsenSys/docs/audit
- Merge pull request [#245](https://github.com/Consensys/gnark-crypto/issues/245) from ConsenSys/perf/extJac-add
- Merge pull request [#240](https://github.com/Consensys/gnark-crypto/issues/240) from ConsenSys/go1.19
<a name="v0.8.0"></a>
## [v0.8.0] - 2022-08-04
### Build
- updated go.mod direct deps
- go mod tidy
- update bavard dep
- ran go generate after dev merge
- faster ci path ([#185](https://github.com/Consensys/gnark-crypto/issues/185))
- increase CI timeout
- fix BatchInvert renaming in SSWU templates
### CI
- fix slack integration + adds golanglint-ci ([#184](https://github.com/Consensys/gnark-crypto/issues/184))
- revert most of last commit
- use self hosted runners
### Chore
- bls12-381 vectors: u
- fix some merge issues
- merge develop
- resolve conflicts with non-mont-params
### Clean
- replace modulus generated by constants, add zero-alloc SetRandom ([#194](https://github.com/Consensys/gnark-crypto/issues/194))
- remove uneeded x86 asm and files ([#192](https://github.com/Consensys/gnark-crypto/issues/192))
### Docs
- updated changelog for v0.8.0
- polish readme.md with updated godoc subpackage links ([#235](https://github.com/Consensys/gnark-crypto/issues/235))
- acknowledge that inv(0)==0 in comments as a convention ([#233](https://github.com/Consensys/gnark-crypto/issues/233))
- correct some comments
- added note in pairing godoc - doesn't check inputs are in correct subgroup ([#231](https://github.com/Consensys/gnark-crypto/issues/231))
- add security estimates of implemented curves in comments
- prepare v0.8.0 release notes
- added twitter handle and security policy links
- updated doi
### Feat
- simplified low degree check
- lagrange polys, a couple of poly funcs
- eq folding, unify small and large interfaces
- addition of multiple rounds in fri
- some folding in bn254
- supsub
- closes [#137](https://github.com/Consensys/gnark-crypto/issues/137) moves consensys/goff into field/goff ([#204](https://github.com/Consensys/gnark-crypto/issues/204))
- tests used gopter for polynomials (bls377, no code gen yet)
- multilin for all fields
- field/goldilocks (more efficient 1-limb modulus arith) ([#177](https://github.com/Consensys/gnark-crypto/issues/177))
- fields in fri proofs are exported
- code gen for previous commit
- code gen for previous refactor
- field/generator suppors 1-limb modulus ([#175](https://github.com/Consensys/gnark-crypto/issues/175))
- simplification of the final evaluation check
- simplify deriveQueriesPositions
- removed polynomial package bls24315
- removed polynomial package
- code gen
- simplified deriveQueriesPosition
- code gen for polynomials
- a few polynomial utility functions, separate multilin_tests
- regen polynomial
- addition of ID in proof of proximity (for FiatShamir)
- added finer grained error handling
- exposed Claimed value
- Open returns an error, addition of getter for rho
- code gen
- addition of opening tests
- addition of opening+verification
- merge develop
- code gen for previous fix
- code gen for kzg refactor
- added Fiat Shamir for the (folding) challenges in fri
- code gen for fri
- **E12:** GT torus-based compression/decompression
- **E12:** GT torus-based batch compression/decompression
- **E24:** GT torus-based batch compression/decompression
- **E6:** GT torus-based batch compression/decompression
- **fri:** added check of correctness between rounds, test OK
- **fri:** modified challenge generation so it fits in a snark variable
- **fri:** evaluation field is exported
- **fri:** removed unsused variable
### Fix
- Handle edge case in Karabina decompression ([#219](https://github.com/Consensys/gnark-crypto/issues/219))
- check nbTasks config when running msm, test all possible c-bit windows in when testing.Short not set) ([#226](https://github.com/Consensys/gnark-crypto/issues/226))
- race condition with supportAdx relique in internal/fptower
- element.SetString(_) returns error if invalid input instead of panic
- bavard dependency
- Torus compression exception case
- ToMont takes non-reference
- q in tests requires isogeny
- svdw parameters: z=1 not i
- expand_msg_xmd copy bug, a few tests ([#201](https://github.com/Consensys/gnark-crypto/issues/201))
- "e3" bug and change sign0 to the recent std specification
- closes [#199](https://github.com/Consensys/gnark-crypto/issues/199). Correct bound in eddsa key gen template
- expand_msg_xmd copy bug, a few tests
- 8*Limbs could be too many bytes
- right length argument for ExpandMsgXmd
- remove supportAdx redundant test ([#186](https://github.com/Consensys/gnark-crypto/issues/186))
- evaluation is an array instead of a slice
- twoInv defined in init, removed dead code, unexport nbRounds
- fixed unhandled errors
- fixed conflict
- fixed queries positions
- fixed wrong indexation generator inverse
- rebase on develop
- wrong size for Merkle path opening
- fixed unhandled error
- fixed condition written out of scope, tests with gopter OK
- fixed parity error
- fixed unhandled error
- fri test ok, need to clean and optimize
- TestDeriveQueriesPositions passes
- fixed failing tests (polynomial.go, bls12-381)
- dusted off polynomial.go (bls377, no code gen yet)
### Perf
- remove unecessary inverse in KZG-verify
- faster GLV scalar decompostion
### Refactor
- fft is done in the main loop
- kzg uses DivideByXminusA from polynomial module
- ScalarMul -> ScalarMultiplication
- everything related to multilinear polynomials in the same file
- verbose names
- same codegen for sswu and svdw
- sswu and svdw in "define"s
- ScalarMulUnconverted -> ScalarMultiplicationAffine
- moved divByXminusa to polynomial.go
- BatchScalarMul -> BatchScalarMultiplication
- parameters passed in regular form
- lots of ffts removed
- polynomial -> []frElement in fri
- removed Commit function
- NewPolynomial -> New
- removed test DivPolyByXminusA from kzg, DivPolyBy -> DivBy
- sswuMap -> mapToCurve
- clean comments in curves ([#193](https://github.com/Consensys/gnark-crypto/issues/193))
- DivideByXMinusA returns a pointer
- **polynomial:** Copy() --> GetCopy()
### Style
- remove dead code ([#230](https://github.com/Consensys/gnark-crypto/issues/230))
- inneficient -> inefficient
- cosmetic changes ([#197](https://github.com/Consensys/gnark-crypto/issues/197))
- use bytes package
- removed dead code
- fixed comment
- cleaned comments, removed dead code
### Test
- cleanup
- fix [#205](https://github.com/Consensys/gnark-crypto/issues/205) - msm bench with different bases ([#206](https://github.com/Consensys/gnark-crypto/issues/206))
- bn254 hash to g2
- vectors generated using https://github.com/armfazh/h2c-go-ref
- complete tests for bn254g1, not cross verified
- empty msg, q, q0, q1
- added BitLen test
- reduce load on CI
- **all curves:** compress/decompress pairing result

### Pull Requests
- Merge pull request [#237](https://github.com/Consensys/gnark-crypto/issues/237) from ConsenSys/develop
- Merge pull request [#232](https://github.com/Consensys/gnark-crypto/issues/232) from ConsenSys/docs/comments
- Merge pull request [#229](https://github.com/Consensys/gnark-crypto/issues/229) from ConsenSys/update_deps
- Merge pull request [#227](https://github.com/Consensys/gnark-crypto/issues/227) from ConsenSys/fix/element_setstring
- Merge pull request [#228](https://github.com/Consensys/gnark-crypto/issues/228) from ConsenSys/fix/race/test
- Merge pull request [#224](https://github.com/Consensys/gnark-crypto/issues/224) from ConsenSys/refactor/scalarmul
- Merge pull request [#220](https://github.com/Consensys/gnark-crypto/issues/220) from ConsenSys/perf/kzg-verify
- Merge pull request [#223](https://github.com/Consensys/gnark-crypto/issues/223) from ConsenSys/doc/security-estimates-curves
- Merge pull request [#216](https://github.com/Consensys/gnark-crypto/issues/216) from ConsenSys/feat/poly
- Merge pull request [#217](https://github.com/Consensys/gnark-crypto/issues/217) from ConsenSys/string-utils
- Merge pull request [#215](https://github.com/Consensys/gnark-crypto/issues/215) from ConsenSys/develop
- Merge pull request [#213](https://github.com/Consensys/gnark-crypto/issues/213) from ConsenSys/perf/glv
- Merge pull request [#211](https://github.com/Consensys/gnark-crypto/issues/211) from ConsenSys/develop
- Merge pull request [#129](https://github.com/Consensys/gnark-crypto/issues/129) from ConsenSys/feat/GT-compression
- Merge pull request [#209](https://github.com/Consensys/gnark-crypto/issues/209) from ConsenSys/codegen/svdw-not-e4
- Merge pull request [#203](https://github.com/Consensys/gnark-crypto/issues/203) from ConsenSys/tests/bn254-vectors
- Merge pull request [#196](https://github.com/Consensys/gnark-crypto/issues/196) from ConsenSys/patch/hashToFpGeneric
- Merge pull request [#202](https://github.com/Consensys/gnark-crypto/issues/202) from ConsenSys/gbotrel/issue199
- Merge pull request [#200](https://github.com/Consensys/gnark-crypto/issues/200) from tyGavinZJU/develop
- Merge pull request [#85](https://github.com/Consensys/gnark-crypto/issues/85) from ConsenSys/feat/fri
<a name="v0.7.0"></a>
## [v0.7.0] - 2022-03-25
### Build
- reran go generate
- rebase on develop
- rebase on develop
- add bw6-633 and bw6-756 to kzg constructor
- run go generate
- rebase branch on develop
- add bls12-378 to kzg and hash
- templates for bw6-756
- templates for bw6-756
- add bls12-378 to kzg and hash
### CI
- update workflows
- update ci workflows
- update github action workflows
- named workflows
- develop shorter ci workflow, master longer
- remove circleCI
- updated circleci to latest golang img
- updated github workflow to go 1.18
### Docs
- prepare changelog.md for v0.7.0
- updated hyperelliptic links for twisted ed Add and MixedAdd
- updated DOI
### Feat
- sweet parameters for BLS12-377 G1 SSWU
- add bw6-756 (2-chain w/ bls12-378 GT-strong)
- SSWU for BLS12-378 (GT-strong)
- sweet parameters for BLS12-377 G2 SSWU (23-isogeny)
- sweet parameters for BW6-761 G2 SSWU
- added element.Uint64() method
- sweet parameters for BW6-633 G2 SSWU
- SSWU for BW6-756 (outter to GT-strong)
- add BLS12-378, a GT-strong SNARK-friendly inner curve
- added ecc/twistededwards/ID
- sweet parameters for BW6-633 G1 SSWU
- removed dead code in fft
- sweet parameters for BW6-761 G1 SSWU
- add BLS12-378, a GT-strong SNARK-friendly inner curve
- sweet parameters for BLS24-315 G1 SSWU
- add bw6-756 (2-chain w/ bls12-378 GT-strong)
- **bls12-378:** add companion twisted edwards to GT-strong BLS12-378
- **bls12-378:** add companion twisted edwards to GT-strong BLS12-378
- **bw6-756:** add companion twisted Edwards
- **bw6-756:** add companion twisted Edwards
### Fix
- IsUint64 coherence with other methods, convert from montgomery beforehands
- twisted curve formulae for GT-strong embedded curve (a != -1)
- templating twistededwards for BW6-756 after PR[#160](https://github.com/Consensys/gnark-crypto/issues/160)
- templating twistededwards for BLS12-378 after PR[#160](https://github.com/Consensys/gnark-crypto/issues/160)
- run go generate on new curves
- add bls12-378 to signature package after change
- sswu gopter generators repeatable rng
- increment maxSignatures
- rebase after change
- changing Z to be the qnr instead of isogeny degree seems to work
- increment maxSignatures
- FrMultiplicativeGen overwritten
- **bls12-378:** set root of unity for FFT
- **bls12-378:** set root of unity for FFT
- **bw6-633:** typo FrMultiplicativeGen=13
- **sswu:** specify CoordExtRoot for BLS12-377
### Refactor
- move signature/ constructors to signature/eddsa
- eddsa generated for all twisted ed curve, including bandersnatch
- twistededwards.go -> curve.go
- bandersnatch generated with endo
- generating bandersnatch without endo with same templates
- make twistededwards all template generated
- **sswu template:** NotOne is not always used
- **sswu template:** NotOne is not always used
### Style
- correct comment in config file
- cleaned comments related to cosets
### Test
- add few edge cases to test twistededwards
### Tests
- mark test as Parallel. Check testing.Short() in most tests

### Pull Requests
- Merge pull request [#171](https://github.com/Consensys/gnark-crypto/issues/171) from ConsenSys/test/twistedEdwards
- Merge pull request [#170](https://github.com/Consensys/gnark-crypto/issues/170) from ConsenSys/fix/generators_sswu
- Merge pull request [#167](https://github.com/Consensys/gnark-crypto/issues/167) from ConsenSys/ci/shorter-tests
- Merge pull request [#166](https://github.com/Consensys/gnark-crypto/issues/166) from ConsenSys/ci_new
- Merge pull request [#164](https://github.com/Consensys/gnark-crypto/issues/164) from ConsenSys/ci/go1.18
- Merge pull request [#128](https://github.com/Consensys/gnark-crypto/issues/128) from ConsenSys/feat/GT-strong-BLS12-BW6
- Merge pull request [#127](https://github.com/Consensys/gnark-crypto/issues/127) from ConsenSys/feat/GT-strong-BLS12
- Merge pull request [#160](https://github.com/Consensys/gnark-crypto/issues/160) from ConsenSys/refactor-eddsa
- Merge pull request [#156](https://github.com/Consensys/gnark-crypto/issues/156) from ConsenSys/sswu-all
- Merge pull request [#154](https://github.com/Consensys/gnark-crypto/issues/154) from ConsenSys/fix/fft-mulGen
- Merge pull request [#153](https://github.com/Consensys/gnark-crypto/issues/153) from zhiqiangxu/opt_NewDomain
<a name="v0.6.1"></a>
## [v0.6.1] - 2022-02-15
### Build
- **templates:** fix G2 point template
### Chore
- genericize c-time sswu
- removed the last sswu if
- remove second if, auto-gen removal of first if
- clean up mess regarding inv(0) conflict
- removed redundant function SetHex
- addressed all "small" feedback points
- remove unnecessary benchmarks, remove og equals, rename diff
### Docs
- updated changelog.md for v0.6.1
- updated bibtex citation key
- updated bibtex citation
- added DOI and bibtex citation in readme.md
### Feat
- code gen
- adapted permutation argument to nrw kzg api
- check that the generator is of correct order
- kzg Verify function takes the opening point
- sqrtRatio for any field, needs tests
- precomputed values for q = 1 mod 8
- bls12-381 g1 hash auto generated
- standardize sqrt
- move big int slicing to bavard
- generic isogeny map
- attempt to generate evaluate_poly
- bls12-381 HashToG1 works
- "generify" additional field operations
- BLS12-381 G1 isogeny (very inefficient)
- SqrtRatio. Non-qr case fails
- modified templates + fix plookup table
- adapted plookup using the new fft
- constant time Equal with tests and benchmarks
- mimc constants are exported
- deleted addchain
- modified fft templates
- generic coset on bn254, tests ok
- x64 assembly
- generic selection using bitwise operations
- **twistededwards:** Extended coordinates (a=-1) (faster, not complete)
### Fix
- code generation "oops" with `testPairElement` class name
- fixed fuzz fft
- restored addchain
- forgot to commit bls12377 modifs
- 1st if statement in sswu
- constant-time sqrtRatio
- some feedback addressed
- remove useless mulBy11 test, replace mulBy11 with mulByZ
- all sqrtRatio works
- sqrtRatio works for bls12-377 (p = 1 mod 8)
- some eval_poly debugging
- No isogeny func if no isogeny data
- Generic TestElement0Inverse error
- bls12-381 g1 encode to curve works
- first G1 isogeny test passes
- sqrtRatio for p = 3 mod 4 works
- extended coordinates complete but not unified
- fixed imports in code gen
- removed old addchain files
- fixed some ops in lookup vector
- fixed permutation proofs
- fixed fuzzer
- removed seed from Sum
- removed seed from mimc (fixes [#194](https://github.com/Consensys/gnark-crypto/issues/194))
- clean, remove experimental second version of select
- delete autogen assembly
- delete unused assembly
- errors in reversion
- generate
- fixed review: m is modified locally and returned
- fixes [#126](https://github.com/Consensys/gnark-crypto/issues/126) the domain was not created correctly
- SetInterface returns error if input is nil
- fixed gosec
- mimc is compliant with ethereum
- twisted edwards templates
- **tEdwards:** cofactor ToMont() not FromMont()
### Perf
- init constants in sync.Once in MiMC
- **bandersnatch:** extended coordinates
- **bls24-315:** faster G2 membership test
### Refactor
- property based testing for SqrtRatio, move sgn0 to ecc package
- property based testing of IntToMont
- Move field related funcs to field.Field
- CurveInfo
### Revert
- Inverse0 to have own PR
- no assemly
### Style
- remove excessive logs
- rename `TempForHash` to `HashUtils`
- Cleanup and test vectors for for bls12-381 G1
- cleanup
- Isogeny to be package-private
- removed addchain folders
- match function hierarchy: Select -> select_ -> _selectGeneric
- remove redundant check
### Test
- possible fix
- obnoxious verbosity
- include decimal in error msg
- for BigIntMatchUint64Slice with verbose error messages
- c1 value computed correctly. c2 is wrong
- extracting test data for iso-g1 from faz's implementation
- vectors form standard doc for existing ExpandMsgXmd implementation
- bench: add x=y cases
- match against generic implementation

### Pull Requests
- Merge pull request [#152](https://github.com/Consensys/gnark-crypto/issues/152) from ConsenSys/feat/clean_kzg
- Merge pull request [#145](https://github.com/Consensys/gnark-crypto/issues/145) from ConsenSys/fix/fft_cosets
- Merge pull request [#147](https://github.com/Consensys/gnark-crypto/issues/147) from ConsenSys/sswu-fp-generic-rebased
- Merge pull request [#146](https://github.com/Consensys/gnark-crypto/issues/146) from ConsenSys/perf-mimc-constants
- Merge pull request [#144](https://github.com/Consensys/gnark-crypto/issues/144) from ConsenSys/constant-time-equals
- Merge pull request [#125](https://github.com/Consensys/gnark-crypto/issues/125) from ConsenSys/fix/mimc_miyaguchipreneel
- Merge pull request [#143](https://github.com/Consensys/gnark-crypto/issues/143) from ConsenSys/feat/cmov
- Merge pull request [#140](https://github.com/Consensys/gnark-crypto/issues/140) from ConsenSys/inv(0)=0
- Merge pull request [#110](https://github.com/Consensys/gnark-crypto/issues/110) from ConsenSys/feat/tEd-extended
- Merge pull request [#123](https://github.com/Consensys/gnark-crypto/issues/123) from ConsenSys/perf/BLS24-G2-IsInSubGroup
<a name="v0.6.0"></a>
## [v0.6.0] - 2022-01-03
### Build
- remove duplicate import in template
- add E8 and E24 types to bls24_315.go to export to gnark
- aiming for determinitic addchain generation on CI
- fix gosec unhandled errors
- tell CI to ignore  non-deterministic addchain generated output
- gitignore generated addition chains
- run go mod tidy
- fix marshal template for bls24
### Chore
- generify 32bit fix
- staticcheck, correct commented formula for outer loop iterations
- Take out InverseOld
- generify semicompressed
- Not demanding 64bit arch. TODO: Test correctness on one
- mathfmt, change correctionFac from var to consts, cite Pornin
- Autogen all tests. TODO: bls12-377 fr loops
- signed/unsigned versions of SOS mont for comparison
### Docs
- v0.6.0 draft release notes
- zkteam -> gnark
- update field IsUIint64 doc
### Feat
- addition of test with a tampered proof
- addition of permutation proof in lookup proof (table version)
- addition of proof of permutation
- added UnmarshalJSON and MarshalJSON on fields
- added field.Text(base) to return field element string in a given base, like big.Int
- field.SetString now supports 0b 0o 0x prefixes (base 2, 8 and 16)
- Autogen P20
- ScalarMul-->mulWindowed, modified blinding of scalar in tests
- Three ways of dealing with signed numbers in montgomery reduction
- Signed mont-reduce
- code gen for integrating permutation proof
- Test for corrective factor consistency
- Naive GCD works
- addition of benchmark
- added field.SetInt64
- added quick and dirty benchmarking script
- added support for intX and uintX fixes [#109](https://github.com/Consensys/gnark-crypto/issues/109)
- pre-compute addition chains for Sqrt and Legendre exp functions
- added addchain for all sqrt
- Naive GCD improved
- Full paper implemented, unknown bug
- Full paper implemented, tests passing
- **bls24:** experiment with Fp-Fp2-Fp4-Fp12-Fp24 tower
- **kzg:** test tampered proofs whith quotient set to zero
- **plookup:** challenges are derived using Fiat Shamir
- **plookup:** addition of prover and verifier for tables
- **plookup:** proof generation
- **plookup:** addition of templates
- **plookup:** adding missing files in templates
- **plookup:** generated plookup for all curves
- **plookup:** addition of benchmark
### Fix
- Optimization 3 works, but with many watches
- started adding the missing parts of the quotietn
- fixed bug for 64b
- 32bit compatible assertMatch for bn254/fp
- semi-compressed bn254/fp
- Update factor negation works
- fixed doc file
- number of iterations corrected. integration tests pass
- unbroke the tests
- ensure ecc.Info() is set
- fix neg template to actually use borrow value
- All bn254 tests pass but TestMonReduceNegFixed
- ecc.Info() returns lightweight field info, without calling the whole field generation including addchain generations
- fixes [#104](https://github.com/Consensys/gnark-crypto/issues/104) code generation for saturated modulus like secp256k1 incorrect. added secp256k1 test
- fixed quotient computation
- fixed test generation
- fixed file generation
- removed error for invalid domain size in kzg
- Neg passes, lingering issues with test randomizer
- remove use of R15 for small moduli mul [#113](https://github.com/Consensys/gnark-crypto/issues/113)
- fixed doc generation
- template used file path for doc instead of file name
- fixed exp takes a value, not a pointer. also random field generation test don't generate addchains
- temporaries element in addition chain back to pointers, they will be on the stack anyway, simple template
- fixed exp template takes element name as parameter
- restored randomness generation via Fiat Shamir
- unused code (nSquare)
- computation of last piece of quotient ok
- Non-const t: Precomputation gives little speedup: 1511,1463,1551
- Optimization 3 works, removed debugging code, down to 1879 ns/op
- init addchain cache only when needed
- full proof (without Fiat Shamir) passes
- **bls24:** fix Fp24 test
- **bn254:** correct Expt() addchain
- **plookup:** removed sortByT function, only sort.Sort() is called now
- **plookup:** computation of h0, hn ok
- **plookup:** fixed lookup vector: t must be ordered
- **plookup:** fixed wrong bound for completion of t and f
- **plookup:** computation of h is correct
- **plookup:** computation of Z ok
### Perf
- Branch-free signed non-mont word multiplication
- Field element - Word multiplication implemented
- Replace mulWRegular with faster branched version
- partial rollback for bn254-fp
- Four update factor vars
- fewer helper variables
- Combined updates factor to be signed, next: fewer helper vars
- Field element - Word multiplication implemented
- signed sos ftw
- Batch each 2 u,v updates. Update factors correct result incorrect
- Inlined conversion factor manipulation
- Removed debug logic
- Linear comb w 1 MontRed instead of 2. Slow (debug logic inline)
- field inverse optimizations
- **Miller loops:** specialized mul by curve coeff
- **bls12-381:** faster Miller loop (sparse-sparse mul)
- **bls12-381:** faster final exp (faster expt)
- **bls24:** compute frobenius coefficients
- **bn254:** better short addition chain for Expt()
- **bn254:** addchain with max squares (weighting mul x2.6 cyclosq)
### Refactor
- SOS Montgomery Reduction
- **bn254:** G2 memebership test uses psi directly
### Revert
- remove mathfmt (for now)
### Style
- comments and proofs
- removed debug printing
- more expressive argument name for `approximate`
- comments
- all "//" to be followed with a space
- Compute number of iterations only in field.go only
- hardcoded values shall be consts
- broke inv
- comments
- minor changes
- Some commentary
- Hand-inlined rsh31, comments, single correction factor
- removed commented code used for debugging
- mathfmt
- code cleaning
- separated tables and vectors in two files, cleaned code
- Some more commentary
- **plookup:** changed naming for rows and columns
- **plookup:** removed all the printing
### Test
- BenchInverse to call InverseOld
- Autogen correction factor checking test
- Autogen Montgomery reduction tests
- added property test for addchain based fixed exp
- deterministic sqrt bench
- Consistency check on top

### Pull Requests
- Merge pull request [#121](https://github.com/Consensys/gnark-crypto/issues/121) from ConsenSys/perf/ML-doubling
- Merge pull request [#111](https://github.com/Consensys/gnark-crypto/issues/111) from ConsenSys/field-intX-support
- Merge pull request [#114](https://github.com/Consensys/gnark-crypto/issues/114) from ConsenSys/fix-dynamic-link
- Merge pull request [#108](https://github.com/Consensys/gnark-crypto/issues/108) from ConsenSys/perf/bls12381-pairing
- Merge pull request [#106](https://github.com/Consensys/gnark-crypto/issues/106) from ConsenSys/improvement/field-inv-pornin20
- Merge pull request [#105](https://github.com/Consensys/gnark-crypto/issues/105) from ConsenSys/field-from-json
- Merge pull request [#83](https://github.com/Consensys/gnark-crypto/issues/83) from ConsenSys/experiment/BLS24
- Merge pull request [#102](https://github.com/Consensys/gnark-crypto/issues/102) from ConsenSys/feat/plookup
- Merge pull request [#97](https://github.com/Consensys/gnark-crypto/issues/97) from ConsenSys/feat-addchain
- Merge pull request [#99](https://github.com/Consensys/gnark-crypto/issues/99) from ConsenSys/feat-addchain-expt
<a name="v0.5.3"></a>
## [v0.5.3] - 2021-11-03
### Docs
- updated CHANGELOG.md for v0.5.3
### Feat
- subgroup check optional in decoder, parallel checks on slices
- added element.NewElement(v uint64)
### Fix
- **fp12:** compressed cyclotomic square (receiver == argument)
### Perf
- **bn:** faster G2 membership test
### Style
- cmp(zero) == -1 -> sign() == -1

### Pull Requests
- Merge pull request [#96](https://github.com/Consensys/gnark-crypto/issues/96) from ConsenSys/perf-decode-raw-points
- Merge pull request [#95](https://github.com/Consensys/gnark-crypto/issues/95) from ConsenSys/perf/bn-g2-membership
<a name="v0.5.2"></a>
## [v0.5.2] - 2021-10-26
### Build
- updated CHANGELOG.md for v0.5.2
- updated code generation with Projective parameter in Point
- **templates:** homogenous projective coordinates for G1 (bw6)
### Feat
- add bandersnatch curve (twistedEdwards on bls12-381 with GLV)
- linked info returned by ecc.Info with internal curve config package
- added Bytes per field in ecc.info
- added curveID.Info() which returns constants about a curve
- moved element.Halve into templates
- **bw6:** optimal Tate Miller loop with shared computations
- **bw6-761:** opt. ate with shared squares and shared doublings (alg.2)
### Fix
- halve with full-bits moduli
- **all twistedEdwards:** fix Add() in projective coordinates (issue 89)
- **all twistedEdwards:** remove A as we assume A=-1 (issue 87)
- **fiat-shamir:** added test to ensure len(challenge) > 0
### Perf
- Halve() directly on fp.Element
- **all curves:** Halve() directly on fp.Element
- **bn:** multiply ML external lines 2 by 2 (+multi-ML bench)
- **wip:** Montgomery Rsh instead of mul by 1/2
### Refactor
- **templates:** unify twistedEdwards package across curves
### Style
- correct comments
- factorize field info
- fiat-shamir clean up
- remove dead code (twoInv)
- **tEdwards:** keep jubjub package for backward-compatibility
- **tEdwards:** mulByA inside the package

### Pull Requests
- Merge pull request [#93](https://github.com/Consensys/gnark-crypto/issues/93) from ConsenSys/bandersnatch
- Merge pull request [#90](https://github.com/Consensys/gnark-crypto/issues/90) from ConsenSys/fix/tEdwards-addProj-issue89
- Merge pull request [#82](https://github.com/Consensys/gnark-crypto/issues/82) from ConsenSys/perf/bn254-ML
- Merge pull request [#88](https://github.com/Consensys/gnark-crypto/issues/88) from ConsenSys/issue-87/twistedEdwards
- Merge pull request [#81](https://github.com/Consensys/gnark-crypto/issues/81) from ConsenSys/ML/DoubleStep-Halve
- Merge pull request [#77](https://github.com/Consensys/gnark-crypto/issues/77) from ConsenSys/BW6
<a name="v0.5.1"></a>
## [v0.5.1] - 2021-09-21
### Build
- remove unused code (nSquare Fp24)
- replace go get by go install in CI workflow
- make staticcheck happier
- updated circleCI golang image
### Docs
- updated CHANGELOG.md with v0.5.1 release
- highlight breaking change in twisted edwards and eddsa
### Feat
- reverted to non-asm field inverse
- element.String() special path for uint64 and -uint64 values
- added x86 assembly impl for field.Inverse
- added element.IsUint64()
- added element.Bit(..) to retrieve i-th bit in a field element
- **Fp12:** implements the Karabina cyclotomic square in E12/E6
- **Fp24:** implements the Karabina cyclotomic square in E24/E8
- **Fp6:** implements the Karabina cyclotomic square in E6/E3
- **e12:** implements batch decompression for karabina cyclo square
- **e24:** implements batch decompression for karabina cyclo square
- **experimental:** msm splits first chunk processing if scalar is on one word
### Fix
- use low c bits only for small values in msm
### Perf
- **bls12:** faster G2 membership (eprint 2021/1130 sec.4)
- **bls12-377:** use asm MubBy5 as MulByNonResidue
- **bls24:** mix Karabina+GS+BatchInvert for faster FinalExp (Expt)
- **bw6-633:** fast GT-subgroup check

### Pull Requests
- Merge pull request [#76](https://github.com/Consensys/gnark-crypto/issues/76) from ConsenSys/msm-ones
- Merge pull request [#75](https://github.com/Consensys/gnark-crypto/issues/75) from ConsenSys/feat/karabina
<a name="v0.5.0"></a>
## [v0.5.0] - 2021-08-23
### Build
- added pairing test to bls24-315 generation
- added fft code gen for bls24-315
- update github workflow with go 1.17
- use 1.17rc2 golang while waiting for circleCI image
- fix gosec unhandled errors warnings
- ran go genearte for kzg on bw6-633
- fix gofuzz target for kzg
- prepare changelog for v0.5.0
- re-ran go genearte
- go mod tidy
- add bls24-315 to fiat-shamir and to readme
- added ecc code gen for bls24-315
- replace fr.Bytes in mimc blocksize
- added eddsa code gen for bls24-315
- added twisted edward codegen for bls24-315
- added mimc to code gen path for bls24-315
- added kzg code generation for bls24-315
- re-ran go:generate
- make staticcheck happier
- **bw6-633:** remove twistededwards until script finds coeff
### Clean
- **kzg:** NewSRS returns error if size < 2
### Cleanup
- kzg.Scheme fft.Domain moved into methods that needs it only
### Docs
- fix go report card link
- updated readme with go1.17
- update ecc.md
- add bls24/bw6 to the implemented curves func and doc
- update readme with bw6-633
- cosmetics and minor fixes
- added EdDSA godoc example
- added missing original copyright in merkletree pacakage + package doc
- most package now have a package level doc
- added fft package level doc
- formated and expanded package godoc for field elements
- fix eddsa Bytes doc to indicate returned bytes slice is in compressed form
### Feat
- added Reference benchamrk for continuous benchmarking. fixes [#54](https://github.com/Consensys/gnark-crypto/issues/54)
- added curve level go-fuzz fuzz function
- added ecc.utils.NextPowerOfTwo used in fft and kzg
- kzg Commit takes an optional CPUSemaphore
- MultiExp now takes a ecc.MultiExpConfig to set the CPUSemaphore and scalars repr
- MultiExp returns error if len(points) != len(scalars)
- package doc is not mixed with code generation but inside a dedicated template for better formating and maintenance
- added Bytes() method on Digest (polynomial commitment)
- added ecc.Implemented() that returns list of curve fully implemented
- added code gen for modified fft
- init bw6-633 curve
- added Marshal() in field elements. Simplified interfaces in polynomial/
- addition of kzg for bn254
- kzg NewSRS takes alpha as big.Int
- kzg.SRS is a separate struct
- affine add, remove digest methods
- kzg is now strongly typed with the curve
- removed mock commitment
- added Clone() method on Digests
- addition of kzg for all curves + tests OK
- save allocation when possible when adding 2 polynomials
- ecc encoder now handles []Element so gnark don't have to
- ecc encoders uses binary.Write and binary.Read to support basic types
- addition of methods GetClaimedValue(s) on opening proofs
- Add,Sub,ScalarMul ops added to Digest
- add new curve bls24-315
- code gen corresponding to previous commit
- added mul_by_13 asm impl in field arithmetic
- Element.SetInterface returns an error instead of panicking if unsupported type
- added Fuzz pairing billinearity
- added kzg Fuzz test
- added Fuzz for fft with cosets
- added Fuzz() for fft pacakge
- addition of method Marshal for Kzg proofs
- addition of BatchVerifyMultiPoints in kzg
- added element_fuzz.go with build tag gofuzz to expose non-asm (generic) impl for fuzz testing
- kzg scheme serialization (bn254)
- code gen for kzg scheme serialization + tests
- helper to build a new kzg scheme for testing purpose
- tests for mock commitment (if claimed values are correct, etc)
- addition of test for claimed values (polynomial commitments)
- added subgroup check for proof and digests in kzg
- fixes [#38](https://github.com/Consensys/gnark-crypto/issues/38)
- **all curves:** faster GT memebership
- **bw6:** GT membership
- **bw6-633:** add bw6-633 to hash and fiat-shamir
- **bw6-633:** add twisted edwards companion to bw6-633
- **bw6-633:** optimized frobenius
- **ecc:** multiexp takes nbTasks as parameter instead of semaphore
- **fft:** addition of param to set precomputations in domain struct
- **kzg:** test polynomials of different sizes
- **kzg:** added /kzg with NewSRS(curveID) API for serialization purposes
- **kzg:** kzg objects implement io.ReaderFrom and io.WriterTo
- **kzg:** hash function is passed as parameter
- **kzg:** added optional nbTasks to pass through multiExp api
- **twisted Edwards:** tests use gopter, no more hardcoded values
### Fix
- CommitmentScheme interface matches kzg and mockCommitment
- fixed staticchek
- fixes [#37](https://github.com/Consensys/gnark-crypto/issues/37)
- fixed some errors handling in transcript.go
- fixed unhandled errors (G104)
- fixed conflicts
- handled error pointed by gosec (Fiat Shamir)
- h function is reset after deriving a challenge
- kzg.dividePolyByXminusA doesn't need the fft domain
- fixed fft cosets, tests OK + code gen
- kzg srs size in benchmarks
- restored kzg
- removed old version of polynomials using interface in bw6-633
- fixed Domain serialization
- kzg return type more homogeneous
- mock Digest ScalarMul didn't modify the caller
- fft with coset is now thread safe. style adjustements in code
- e2 x86 asm incorrect offset when x is 0
- fixes [#49](https://github.com/Consensys/gnark-crypto/issues/49)
- proper error handling for polynomial commitments + code gen
- remove Code generated comments in bls24/internal which is not generated
- fixes [#51](https://github.com/Consensys/gnark-crypto/issues/51)
- use crypto/rand instead of math/rand in ecc/../utils.go
- kzg serilization test comparing address instead of value
- **all curves:** IsInSubGroup shouldn't test E12 elements but GT
- **bls12-377:** fix coefficients for Shallue-van de Woestijne hash-to-G2
- **bls24-315:** E2 Mul + no template
- **bw6:** use crypto/rand instead of math/rand
- **bw6:** add GenBigInt and rename utils_test.go to utils.go
- **bw6:** use crypto/rand instead of math/rand
- **bw6-633:** fft
- **bw6-633:** fix final exp
- **bw6-633:** correct coefficients in SWU hash-to-curve
- **bw6-633:** clear cofactor and subgroup membership for G1 and G2
- **bw6-633:** GLV when fr.Limbs is odd
- **kzg:** alpha generation is up to the caller when calling NewScheme
- **kzg:** incorrect poly size in Open
- **msm:** seems doing the inner msm sequentially with little available cpus is bad idea
- **template:** IsInSubGroup shouldn't test E12 elements but GT
- **template:** include GLV case when fr.Limbs is odd
- **template:** code generation for bw6-633 eddsa
- **template:** fix MapToCurve test
- **template:** cyclotomic suqare test
- **template:** Expt test to include negative seed
- **templates:** no psi for bw6 endomorphism test
- **twisted Edwards:** fixed Neg(), and fixes [#57](https://github.com/Consensys/gnark-crypto/issues/57)
### Perf
- minor perf improvments in bw6 tower
- minor serialization chaanges in kzg
- **all curves:** twisted Edwards companions arithmetic with a=-1
- **bls12:** faster G2 clear cofactor
- **bls12:** faster G2 subgroup checks --> psi^2=phi+1
- **bls12:** faster G2 subgroup checks
- **bls12-377:** remove one add, one sub in e2.Square
- **bn:** optimize Expt (no conditional branching)
- **bn254:** Expt in 2-NAF
- **bw6:** replace Inverse and FrobeniusCube by conjugate
- **bw6:** new optimized final exp (hard part)
- **bw6-633:** divide G1 cofactor formula by 4
- **bw6-633:** optimized hard part in final exp
- **fft:** introduced flatten kernel for n==8 and asm impl for butterfly to minimze memory writes
- **kzg:** use syntetic division to divide by x-a
- **kzg:** dividePolyByXminusA parallelism
- **kzg:** dividePolyByXminusA uses precomputed domain twiddle
- **kzg:** merge both loop in divideXpoly into 1
- **kzg:** Open will split the msm in 2 if numCpus > 16
- **kzg:** perform polynomial evaluations in parallel inside batchOpen
- **kzg:** reuse memory for batchopen poly division
- **kzg:** added batch inversion in dividePolyByXminusA
- **kzg:** parallelize s = 1 / (x-a)
- **kzg:** remove useless fft in dividebyxminusa
- **kzg:** simplified dividePolyByXminusA, no need to shift
### Refactor
- removed deprecated MulAssign, AddAssign and SubAssign apis
- moved crypto/* under /
- ported accumulator/ and polynomial/ from gnark
- moved fr/polynomial/kzg to fr/kzg
- ported mock commitment scheme from gnark
- moved duplicated ecc/xxx/CPUSemaphore to ecc/
- remove dead code in pairing
- BatchJacobianToAffineG1Affine -> BatchJacobianToAffineG1
- removed hash functions recorded in transcript.go
- ./crypto/fiat-shamir --> ./fiat-shamir/
- **bls12-377:** change G1 generator to match other libs
- **bls12-377:** change G2 generator (+Fp QNR) to match other libs
- **bls12-377:** remove unused sync.Pool in pairing
- **bw6:** Pairing according to ABLR 2013/722 with Fp6/Fp3
- **kzg:** Proof -> OpeningProof. BatchProofsSinglePoint -> BatchOpeningProof
- **kzg:** removed Scheme, package level methods with SRS and domain as parameter
### Style
- replaced conditionals with template variable to return fuzzed element
- gopter generators are code generated and return values. siimplifies non-sense in templates
- cleaner error message in Element.SetInterface
- cosmetics
- code cleaning in kzg
- go fmt
- code cleaning in polynomial and kzgé
- clean comments
- ComputeChallenge error re-indent
- **fft:** use close(chan)
- **kzg:** cosmetics
- **kzg:** cosmetics
### Test
- added mulGeneric vs mul assembly on E2 in bn254 and bls12-381
- bls24-315 added mulGeneric vs mulAsm E2 test
- **all curves:** test endomorphisms phi and psi
- **curves:** use IsInSubGroup instead IsOnCurve MapToCurveG1Svdw test
- **template:** include bw6 in pairing_test.go
### Tests
- added e2.Neg test in code generation

### Pull Requests
- Merge pull request [#70](https://github.com/Consensys/gnark-crypto/issues/70) from ConsenSys/develop
- Merge pull request [#68](https://github.com/Consensys/gnark-crypto/issues/68) from ConsenSys/fft-cleanup
- Merge pull request [#64](https://github.com/Consensys/gnark-crypto/issues/64) from ConsenSys/feat/bw6-633
- Merge pull request [#65](https://github.com/Consensys/gnark-crypto/issues/65) from ConsenSys/feat/kzgserialization
- Merge pull request [#63](https://github.com/Consensys/gnark-crypto/issues/63) from ConsenSys/feat/kzg_update
- Merge pull request [#61](https://github.com/Consensys/gnark-crypto/issues/61) from ConsenSys/feat/bls24-315
- Merge pull request [#59](https://github.com/Consensys/gnark-crypto/issues/59) from ConsenSys/ci/fuzzandbench
- Merge pull request [#58](https://github.com/Consensys/gnark-crypto/issues/58) from ConsenSys/refactor/bw6-761-pairing
- Merge pull request [#55](https://github.com/Consensys/gnark-crypto/issues/55) from ConsenSys/feat/GT-membership
- Merge pull request [#48](https://github.com/Consensys/gnark-crypto/issues/48) from ConsenSys/feat/kzg
- Merge pull request [#50](https://github.com/Consensys/gnark-crypto/issues/50) from ConsenSys/fix/fft
- Merge pull request [#46](https://github.com/Consensys/gnark-crypto/issues/46) from ConsenSys/fix/domain_precompute
- Merge pull request [#45](https://github.com/Consensys/gnark-crypto/issues/45) from ConsenSys/hotfix/issue_36
- Merge pull request [#44](https://github.com/Consensys/gnark-crypto/issues/44) from ConsenSys/feat/mul_by_13
- Merge pull request [#42](https://github.com/Consensys/gnark-crypto/issues/42) from ConsenSys/feat/fiat_shamir
- Merge pull request [#41](https://github.com/Consensys/gnark-crypto/issues/41) from ConsenSys/docs/godoc
<a name="v0.4.0"></a>
## [v0.4.0] - 2021-03-31
### Build
- updated go.mod
- updated CI and go.mod files
- updated to latest goff
- updated goff to fix incorrect min stack size
- fix slack notification
- added .gitlint file
- run on develop and master only
- test on more archs and targets
- faster path for testing.Short. remove some dead code
- make staticcheck happy, remove dead code
- added github action workflow, wip
### CI
- install asmfmt before test step, now that goff field generation tests are included
- move dep install up
- ignore G204 rule in gosec (process lauched with var)
- testing pr on develop with go 1.15 and go 1.16
### Docs
- added changelog for v0.4.0
- add ecc/ecc.md and field/field.md
- updated README.md
- updated package doc
### Feat
- msm in full extJac
- use add extJac to sum buckets
- e2 bls381 asm mul clean
- e2 asm bls381 square clean
- bls381 e2 asm mul by non residue clean up
- first step in clobbered bp refactor. e2 asm: bn256 stable, bls381 wip
- use add extJac to sum buckets
### Fix
- handle case where numCPU < 4 in precomputeExpTable
- incorrect comment and size returned in twistededwards SetBytes fixes [#34](https://github.com/Consensys/gnark-crypto/issues/34)
- add/dbl exJac
- add/dbl exJac
- point.SetBytes can now be called concurently with same byte slice input
### Perf
- delay coordinates conversion
- delay coordinates conversion
### Refactor
- moved interop tests under github.com/consensys/gnark-tests
- cosmetics
- gurvy -> gnark-crypto
- bls381 -> bls12-381, bls377 -> bls12-377
- curve -> ecc
- moved utils/ into curve/
- e2_bn256 --> e2_bn254
- moved gurvy.go into curve/curve.go
- renamed BN256 to BN254
- moved curves under curve/
- migrated gnark/polynomial under fr/
- ported mimc and eddsa from gnark
- factorized parallelize function and moved asm/amd64 into tower package
- cleaning internal/generator pattern
- checkpoint
- migrated gnark/backend/fft into fft/
- migrated goff packages into /field/...
- moved curves into /curve/...
### Style
- consistent copyright holder and year
- remove some empty lines
- rename point bench functions
- refactor reduceAfterSub... to modReduce...
- rename point bench functions

### Pull Requests
- Merge pull request [#35](https://github.com/Consensys/gnark-crypto/issues/35) from ConsenSys/refactor/monorepo
- Merge pull request [#33](https://github.com/Consensys/gnark-crypto/issues/33) from ConsenSys/msm/full-extJac
- Merge pull request [#32](https://github.com/Consensys/gnark-crypto/issues/32) from ConsenSys/fix/clobbered_bp
<a name="v0.3.8"></a>
## [v0.3.8] - 2021-02-01
### Bls377
- final exp hard part eprint 2020/875
- ML entirely on the twist (ABLR)
- ML entirely on the twist (ABLR)
### Bls381
- final exp hard part eprint 2020/875
- no precompute and go routines in pairing, wip
- ML entirely on the twist (ABLR)
- ML entirely on the twist (ABLR)
- change G1 and G2 generators for interop
### Bn256
- inline lineEval() in MilleLoop
- ML entirely on the twist (ABLR)
- change G1 and G2 generators for interop
### Bw6
- add E6 tests
- correct comments in FinalExp
- add pairing tests
- go fmt
- fix bw6 pairing API to take slices of points and mutualize squares
- go fmt
- fix bw6 pairing API to take slices of points and mutualize squares
### Bw761
- change G1 and G2 generators for interop
- PairingCheck API + ML bench
- PairingCheck API + ML bench
### Curves
- get rid of goroutines in ML
- g1/g2 template add proj to jac conversion
- get rid of goroutines in ML
- g1/g2 template add proj to jac conversion
### Fp12
- improved cyclotomicSquare
### Fp6
- cyclotomic square in Expt()
### TwistedEdwards
- addition of Marshal/Unmarshal+tests, addition of templates, modified GetEdwardsCurve to not expose order (in big.Int)

### Pull Requests
- Merge pull request [#29](https://github.com/Consensys/gnark-crypto/issues/29) from ConsenSys/youssef/bls12-finalExp
- Merge pull request [#27](https://github.com/Consensys/gnark-crypto/issues/27) from ConsenSys/experimental/pairing
- Merge pull request [#26](https://github.com/Consensys/gnark-crypto/issues/26) from ConsenSys/youssef/ML-ABLR
- Merge pull request [#25](https://github.com/Consensys/gnark-crypto/issues/25) from ConsenSys/csquare
- Merge pull request [#23](https://github.com/Consensys/gnark-crypto/issues/23) from ConsenSys/youssef/bw6-API-pairing
<a name="v0.3.7"></a>
## [v0.3.7] - 2021-01-04
<a name="v0.3.6"></a>
## [v0.3.6] - 2020-12-22
### Bls377
- multiple MillerLoops
### Bls381
- PairingCheck func for eth precompile
- 1 MillerLoop func + error handling
- fix MillerLoop (negative seed -> conjugate)
- optimize the final exp hard part (alg2)
- added e2.Mul x86 asm impl. sort of.
- e2 --> added x86 asm impl for squaring
- reverted result.Conjugate in pairing to previous (correct) version
- fix build on non amd64
- fix pairing interop test
- added asm impl for MulByNonResidue
- minor adjustements, wip
- minor adjustements, wip
- added bench in interop
- added failing interop test
- added scalar mul interop tests
- more interop test on GT serialization
### Bls831
- multiple MillerLoops error handling
### Bn256
- multiple miller loops
### Bw761
- added e6 Bytes() and SetBytes()
### Circleci
- ignoring .s files when checking generate output
### Feat
- **bls381:** generalize tripleML to xML
- **bls381:** factorize sq in 3 ML products
### Fix
- substitute == by Equal in e12
- substitute == by Equal in e12
### Marshal
- len slice of points encoded on uint32 instead of uint64.
- remove unecessary buffer allocatiosn in loop, added benchmark for e2.Sqrt
- added BytesRead and BytesWritten to Decoder and Encoder objects
### MultiExp
- the function must be called from an affine point and return an affine poitn
### Template
- pairing test
### Templates
- Set func in g1/2
- fix PairingCheck test
- pairing tests
- correct EFD URL in comments (g*p -> g1p)
### Unmarshal
- now perfoms subgroup check when deserializing points
### Wip
- added Bytes() and SetBytes() method on point with binary encoding and optional point compression

### Pull Requests
- Merge pull request [#21](https://github.com/Consensys/gnark-crypto/issues/21) from ConsenSys/youssef/MillerLoops
- Merge pull request [#19](https://github.com/Consensys/gnark-crypto/issues/19) from ConsenSys/cleanup
<a name="v0.3.5"></a>
## [v0.3.5] - 2020-10-19
### Bls377
- re-exposing GT.MulByXXX methods needed by gnark standard library
<a name="v0.3.4"></a>
## [v0.3.4] - 2020-10-19
### Curves
- curve coeff b is package private
### Multiexp
- renamed MultiExpOpt to CPUSemaphore, and remove data race on the window size attribute. Made staticcheck and linter happy.
<a name="v0.3.3"></a>
## [v0.3.3] - 2020-09-23
<a name="v0.3.0"></a>
## [v0.3.0] - 2020-09-22
<a name="v0.3.0-alpha"></a>
## [v0.3.0-alpha] - 2020-09-15
### BatchJacobianToAffine
- don't allocate new slice to store product, use affine storage space instead
### Bn256
- added test to  ensure e2 mul and square assembly impl matches generic ones
- enabled inlined call to wrappers E2.Mul and E2.Square making G2 point addition 10% faster
### Circleci
- test only main point ops on 32 bit path
- added -short option to tests to skip long or memory greedy tests
### Curve
- added BatchJacobianToAffine using montgomery batch inversion for G1 elements. Skipping G2 on this one
### Curves
- Generators() now return generators of the r-torsion group in Jacobian coordinates
- added Generators() getter on all curves. Added BatchScalarMultiplication method with test
- cleanup unused parameters for old scalar multiplication
### MultiExpOptions
- takes a semaphore (channel) of cpu instead of number of cpus
- fix previous commit.
### Multiexp
- PartitionScalars is now private. CPU semaphore has a lock that's release when all go routines are scheduled in the multiexp
- PartitionScalars now takes MultiExpOptions as optional argument too
- now optionally takes choice of c, max cpus to use, and indicates if scalars are already preprocessed
- code clean up. BatchScalarMultiplication now select window size depending on input length
- renamed ScalarsToDigit to PartitionScalars
- refactored ScalarsToDigit method so that duplicated code in multiExpcXX methods is largely reduced to allocating buckets on the stack
- remove c=20 as it timeouts test on circle ci for now
- api is now sync, like the rest of gx.go apis.
- less samples in generated tests, bw761 generates less multiExp variant of c
- added comments in code. still a lot (too much) template generated and duplicated code
- taking into account available CPUs. slightly better extended addition formula. moved multi exp tests into their own function
- wip, reverting to c divides 64
- wip, added preprocessing of scalars to compute digits in the bucket method
- remove test of c18 that seems to timeout
- cosmetics and documentation
- clean up bucket method, with 30percent gain on bn256 G1. WIP
### Pairing
- sender closes channel, not receiver
### Points
- made p.Neg() inlinable
### ScalarsToDigits
- reused method in BatchScalarMultiplication
### UnsafeFromJacExtended
- cosmetic, doc update
<a name="v0.2.0"></a>
## [v0.2.0] - 2020-08-04
### Bls381
- removed multi exp garbage, cleaned the tests using gopter
- tests for the pairing use gopter, fixed String() on g1, g2 (thanks the templates again), implemented simple double and add for g1 and g2 using big.Int
- got rid of the curve structure
### Circleci
- ensures that generated files are not hand modified
- ensures that generated files are not hand modified.
- new workflow with more insight on unit tests
### Develop
- skip Expt tests for now because the testpoint is not in the cyclotomic subgroup
- merge refactor-api; resolve merge conflicts, TestE12Expt failing for some reason
### WIP
- optimized Miller loop

### Pull Requests
- Merge pull request [#11](https://github.com/Consensys/gnark-crypto/issues/11) from ConsenSys/refactor-cyclotomicsquare
- Merge pull request [#10](https://github.com/Consensys/gnark-crypto/issues/10) from ConsenSys/refactor-constants
- Merge pull request [#5](https://github.com/Consensys/gnark-crypto/issues/5) from ConsenSys/experimental-pairing-gen
- Merge pull request [#4](https://github.com/Consensys/gnark-crypto/issues/4) from ConsenSys/endomul-remove
<a name="v0.1.1"></a>
## [v0.1.1] - 2020-04-08
<a name="v0.1.0"></a>
## [v0.1.0] - 2020-04-07

### Pull Requests
- Merge pull request [#2](https://github.com/Consensys/gnark-crypto/issues/2) from ConsenSys/develop
<a name="v0.0.1"></a>
## v0.0.1 - 2020-03-23
[v0.21.0]: https://github.com/Consensys/gnark-crypto/compare/v0.20.1...v0.21.0
[v0.20.1]: https://github.com/Consensys/gnark-crypto/compare/v0.20.0...v0.20.1
[v0.20.0]: https://github.com/Consensys/gnark-crypto/compare/v0.19.2...v0.20.0
[v0.19.2]: https://github.com/Consensys/gnark-crypto/compare/v0.19.1...v0.19.2
[v0.19.1]: https://github.com/Consensys/gnark-crypto/compare/v0.19.0...v0.19.1
[v0.19.0]: https://github.com/Consensys/gnark-crypto/compare/v0.18.2...v0.19.0
[v0.18.2]: https://github.com/Consensys/gnark-crypto/compare/v0.18.1...v0.18.2
[v0.18.1]: https://github.com/Consensys/gnark-crypto/compare/v0.18.0...v0.18.1
[v0.18.0]: https://github.com/Consensys/gnark-crypto/compare/v0.17.0...v0.18.0
[v0.17.0]: https://github.com/Consensys/gnark-crypto/compare/v0.16.0...v0.17.0
[v0.16.0]: https://github.com/Consensys/gnark-crypto/compare/v0.15.0...v0.16.0
[v0.15.0]: https://github.com/Consensys/gnark-crypto/compare/v0.14.0...v0.15.0
[v0.14.0]: https://github.com/Consensys/gnark-crypto/compare/v0.13.0...v0.14.0
[v0.13.0]: https://github.com/Consensys/gnark-crypto/compare/v0.12.1...v0.13.0
[v0.12.1]: https://github.com/Consensys/gnark-crypto/compare/v0.12.0...v0.12.1
[v0.12.0]: https://github.com/Consensys/gnark-crypto/compare/v0.11.2...v0.12.0
[v0.11.2]: https://github.com/Consensys/gnark-crypto/compare/v0.11.1...v0.11.2
[v0.11.1]: https://github.com/Consensys/gnark-crypto/compare/v0.11.0...v0.11.1
[v0.11.0]: https://github.com/Consensys/gnark-crypto/compare/v0.10.1...v0.11.0
[v0.10.1]: https://github.com/Consensys/gnark-crypto/compare/v0.10.0...v0.10.1
[v0.10.0]: https://github.com/Consensys/gnark-crypto/compare/v0.9.2...v0.10.0
[v0.9.2]: https://github.com/Consensys/gnark-crypto/compare/v0.9.1...v0.9.2
[v0.9.1]: https://github.com/Consensys/gnark-crypto/compare/v0.9.0...v0.9.1
[v0.9.0]: https://github.com/Consensys/gnark-crypto/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/Consensys/gnark-crypto/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/Consensys/gnark-crypto/compare/v0.6.1...v0.7.0
[v0.6.1]: https://github.com/Consensys/gnark-crypto/compare/v0.6.0...v0.6.1
[v0.6.0]: https://github.com/Consensys/gnark-crypto/compare/v0.5.3...v0.6.0
[v0.5.3]: https://github.com/Consensys/gnark-crypto/compare/v0.5.2...v0.5.3
[v0.5.2]: https://github.com/Consensys/gnark-crypto/compare/v0.5.1...v0.5.2
[v0.5.1]: https://github.com/Consensys/gnark-crypto/compare/v0.5.0...v0.5.1
[v0.5.0]: https://github.com/Consensys/gnark-crypto/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/Consensys/gnark-crypto/compare/v0.3.8...v0.4.0
[v0.3.8]: https://github.com/Consensys/gnark-crypto/compare/v0.3.7...v0.3.8
[v0.3.7]: https://github.com/Consensys/gnark-crypto/compare/v0.3.6...v0.3.7
[v0.3.6]: https://github.com/Consensys/gnark-crypto/compare/v0.3.5...v0.3.6
[v0.3.5]: https://github.com/Consensys/gnark-crypto/compare/v0.3.4...v0.3.5
[v0.3.4]: https://github.com/Consensys/gnark-crypto/compare/v0.3.3...v0.3.4
[v0.3.3]: https://github.com/Consensys/gnark-crypto/compare/v0.3.0...v0.3.3
[v0.3.0]: https://github.com/Consensys/gnark-crypto/compare/v0.3.0-alpha...v0.3.0
[v0.3.0-alpha]: https://github.com/Consensys/gnark-crypto/compare/v0.2.0...v0.3.0-alpha
[v0.2.0]: https://github.com/Consensys/gnark-crypto/compare/v0.1.1...v0.2.0
[v0.1.1]: https://github.com/Consensys/gnark-crypto/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/Consensys/gnark-crypto/compare/v0.0.1...v0.1.0
