package vortex

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/stretchr/testify/require"
)

type testcaseVortex struct {
	M               [][]koalabear.Element
	X               fext.E4
	Ys              []fext.E4
	Alpha           fext.E4
	SelectedColumns []int
	ColumnHash      HashConstructor
	MerkleHash      HashConstructor
}

func TestZeroMatrix(t *testing.T) {

	var (
		numCol = 16
		numRow = 8
	)

	var (
		m               = make([][]koalabear.Element, numRow)
		x               = fext.E4{}
		y               = make([]fext.E4, numRow)
		alpha, _        = new(fext.E4).SetRandom()
		selectedColumns = []int{0, 1, 2, 3}
	)

	for i := range m {
		m[i] = make([]koalabear.Element, numCol)
	}

	runTest(t, &testcaseVortex{
		M:               m,
		X:               x,
		Ys:              y,
		Alpha:           *alpha,
		SelectedColumns: selectedColumns,
	})

}

func TestFullRandom(t *testing.T) {

	var (
		numCol = 16
		numRow = 8
		// #nosec G404 -- test case generation does not require a cryptographic PRNG
		rng = rand.New(rand.NewChaCha8([32]byte{}))
	)

	var (
		m               = make([][]koalabear.Element, numRow)
		x               = randFext(rng)
		ys              = make([]fext.E4, numRow)
		alpha           = randFext(rng)
		selectedColumns = []int{0, 1, 2, 3}
		err             error
	)

	for i := range m {
		m[i] = make([]koalabear.Element, numCol)
		for j := range m[i] {
			m[i][j] = randElement(rng)
		}

		ys[i], err = EvalBasePolyLagrange(m[i], x)
		if err != nil {
			t.Fatal(err)
		}
	}

	runTest(t, &testcaseVortex{
		M:               m,
		X:               x,
		Ys:              ys,
		Alpha:           alpha,
		SelectedColumns: selectedColumns,
	})
}

func TestFullRandomColumnHash(t *testing.T) {

	var (
		numCol = 16
		numRow = 8
		// #nosec G404 -- test case generation does not require a cryptographic PRNG
		rng = rand.New(rand.NewChaCha8([32]byte{}))
	)

	var (
		m               = make([][]koalabear.Element, numRow)
		x               = randFext(rng)
		ys              = make([]fext.E4, numRow)
		alpha           = randFext(rng)
		selectedColumns = []int{0, 1, 2, 3}
		err             error
	)

	for i := range m {
		m[i] = make([]koalabear.Element, numCol)
		for j := range m[i] {
			m[i][j] = randElement(rng)
		}

		ys[i], err = EvalBasePolyLagrange(m[i], x)
		if err != nil {
			t.Fatal(err)
		}
	}

	runTest(t, &testcaseVortex{
		M:               m,
		X:               x,
		Ys:              ys,
		Alpha:           alpha,
		SelectedColumns: selectedColumns,
		ColumnHash:      func() hash.Hash { return sha256.New() },
	})
}

func TestFullRandomNoPoseidon(t *testing.T) {

	var (
		numCol = 16
		numRow = 8
		// #nosec G404 -- test case generation does not require a cryptographic PRNG
		rng = rand.New(rand.NewChaCha8([32]byte{}))
	)

	var (
		m               = make([][]koalabear.Element, numRow)
		x               = randFext(rng)
		ys              = make([]fext.E4, numRow)
		alpha           = randFext(rng)
		selectedColumns = []int{0, 1, 2, 3}
		err             error
	)

	for i := range m {
		m[i] = make([]koalabear.Element, numCol)
		for j := range m[i] {
			m[i][j] = randElement(rng)
		}

		ys[i], err = EvalBasePolyLagrange(m[i], x)
		if err != nil {
			t.Fatal(err)
		}
	}

	runTest(t, &testcaseVortex{
		M:               m,
		X:               x,
		Ys:              ys,
		Alpha:           alpha,
		SelectedColumns: selectedColumns,
		MerkleHash:      func() hash.Hash { return sha256.New() },
	})
}

func TestFullRandomNoPoseidonColumnHash(t *testing.T) {

	var (
		numCol = 16
		numRow = 8
		// #nosec G404 -- test case generation does not require a cryptographic PRNG
		rng = rand.New(rand.NewChaCha8([32]byte{}))
	)

	var (
		m               = make([][]koalabear.Element, numRow)
		x               = randFext(rng)
		ys              = make([]fext.E4, numRow)
		alpha           = randFext(rng)
		selectedColumns = []int{0, 1, 2, 3}
		err             error
	)

	for i := range m {
		m[i] = make([]koalabear.Element, numCol)
		for j := range m[i] {
			m[i][j] = randElement(rng)
		}

		ys[i], err = EvalBasePolyLagrange(m[i], x)
		if err != nil {
			t.Fatal(err)
		}
	}

	runTest(t, &testcaseVortex{
		M:               m,
		X:               x,
		Ys:              ys,
		Alpha:           alpha,
		SelectedColumns: selectedColumns,
		ColumnHash:      func() hash.Hash { return sha256.New() },
		MerkleHash:      func() hash.Hash { return sha256.New() },
	})
}

func randElement(rng *rand.Rand) koalabear.Element {
	return koalabear.Element{rng.Uint32N(2130706433)}
}

func randFext(rng *rand.Rand) fext.E4 {
	return fext.E4{
		B0: fext.E2{
			A0: randElement(rng),
			A1: randElement(rng),
		},
		B1: fext.E2{
			A0: randElement(rng),
			A1: randElement(rng),
		},
	}
}

func runTest(t *testing.T, tc *testcaseVortex) {

	var (
		numCol             = len(tc.M[0])
		numRow             = len(tc.M)
		reedSolomonInvRate = 2
		numSelectedColumns = len(tc.SelectedColumns)
		sisParams, _       = sis.NewRSis(0, 9, 16, numRow)
		params, _          = NewParams(numCol, numRow, sisParams, reedSolomonInvRate, numSelectedColumns)
	)

	proverState, err := Commit(params, tc.M)
	if err != nil {
		t.Fatal(err)
	}

	proverState.OpenLinComb(tc.Alpha)

	proof, err := proverState.OpenColumns(tc.SelectedColumns)
	if err != nil {
		t.Fatal(err)
	}

	err = params.Verify(VerifierInput{
		Proof:           proof,
		MerkleRoot:      proverState.GetCommitment(),
		ClaimedValues:   tc.Ys,
		EvaluationPoint: tc.X,
		Alpha:           tc.Alpha,
		SelectedColumns: tc.SelectedColumns,
	})

	if err != nil {
		t.Fatal(err)
	}
}

// An unchecked suffix changes the polynomial evaluated at x without changing
// the Reed-Solomon codeword or any opened column.
func TestVerifyRejectsOversizedUAlpha(t *testing.T) {
	const numColumns, numRows = 4, 4
	sisParams, err := sis.NewRSis(0, 9, 16, numRows)
	require.NoError(t, err)
	params, err := NewParams(numColumns, numRows, sisParams, 2, 1)
	require.NoError(t, err)

	matrix := make([][]koalabear.Element, numRows)
	for row := range matrix {
		matrix[row] = make([]koalabear.Element, numColumns)
		for col := range matrix[row] {
			matrix[row][col] = koalabear.NewElement(uint64(100 + 10*row + col))
		}
	}
	x := fext.E4{B0: fext.E2{A0: koalabear.NewElement(123), A1: koalabear.NewElement(456)},
		B1: fext.E2{A0: koalabear.NewElement(789), A1: koalabear.NewElement(1011)}}
	alpha := fext.E4{B0: fext.E2{A0: koalabear.NewElement(17), A1: koalabear.NewElement(19)},
		B1: fext.E2{A0: koalabear.NewElement(23), A1: koalabear.NewElement(29)}}
	claims := make([]fext.E4, numRows)
	for row := range matrix {
		claims[row], err = EvalBasePolyLagrange(matrix[row], x)
		require.NoError(t, err)
	}
	state, err := Commit(params, matrix)
	require.NoError(t, err)
	state.OpenLinComb(alpha)
	proof, err := state.OpenColumns([]int{0})
	require.NoError(t, err)
	input := VerifierInput{
		Proof: proof, MerkleRoot: state.GetCommitment(), ClaimedValues: claims,
		EvaluationPoint: x, Alpha: alpha, SelectedColumns: []int{0},
	}
	require.NoError(t, params.Verify(input))
	input.Proof = nil
	require.Error(t, params.Verify(input), "missing proof must be rejected")
	short := *proof
	short.UAlpha = short.UAlpha[:len(short.UAlpha)-1]
	input.Proof = &short
	require.Error(t, params.Verify(input), "short codeword must be rejected without panicking")
	input.Proof = proof

	falseClaims := append([]fext.E4(nil), claims...)
	var one fext.E4
	one.SetOne()
	falseClaims[0].Add(&falseClaims[0], &one)
	input.ClaimedValues = falseClaims
	require.Error(t, params.Verify(input))

	n := params.SizeCodeWord()
	forged := *proof
	forged.UAlpha = make([]fext.E4, 2*n)
	copy(forged.UAlpha, proof.UAlpha)
	forged.UAlpha[n] = EvalBasePolyHorner(forged.OpenedColumns[0], alpha)
	prefixValue, err := EvalFextPolyLagrange(forged.UAlpha, x)
	require.NoError(t, err)
	target := EvalFextPolyHorner(falseClaims, alpha)
	var correction fext.E4
	correction.Sub(&target, &prefixValue)
	corrected := false
	for index := n + 1; index < len(forged.UAlpha); index++ {
		basis := make([]fext.E4, len(forged.UAlpha))
		basis[index].SetOne()
		weight, evalErr := EvalFextPolyLagrange(basis, x)
		require.NoError(t, evalErr)
		if weight.IsZero() {
			continue
		}
		var inverse fext.E4
		inverse.Inverse(&weight)
		forged.UAlpha[index].Mul(&correction, &inverse)
		corrected = true
		break
	}
	require.True(t, corrected, "no nonzero suffix Lagrange coordinate")
	input.Proof = &forged
	require.Error(t, params.Verify(input), "forged false claim must be rejected")
}

func TestVerifyRejectsMalformedOpening(t *testing.T) {
	const numColumns, numRows = 4, 4
	sisParams, err := sis.NewRSis(0, 9, 16, numRows)
	require.NoError(t, err)
	params, err := NewParams(numColumns, numRows, sisParams, 2, 1)
	require.NoError(t, err)
	matrix := make([][]koalabear.Element, numRows)
	for i := range matrix {
		matrix[i] = make([]koalabear.Element, numColumns)
		matrix[i][0] = koalabear.NewElement(uint64(i + 1))
	}
	var x, alpha fext.E4
	x.B0.A0 = koalabear.NewElement(123)
	alpha.B0.A0 = koalabear.NewElement(17)
	claims := make([]fext.E4, numRows)
	for i := range matrix {
		claims[i], err = EvalBasePolyLagrange(matrix[i], x)
		require.NoError(t, err)
	}
	state, err := Commit(params, matrix)
	require.NoError(t, err)
	state.OpenLinComb(alpha)
	proof, err := state.OpenColumns([]int{0})
	require.NoError(t, err)
	input := VerifierInput{
		Proof: proof, MerkleRoot: state.GetCommitment(), ClaimedValues: claims,
		EvaluationPoint: x, Alpha: alpha, SelectedColumns: []int{0},
	}
	require.NoError(t, params.Verify(input))

	t.Run("missing selection permits an unbound claim", func(t *testing.T) {
		bad := input
		bad.SelectedColumns = nil
		bad.Proof = &Proof{UAlpha: make([]fext.E4, params.SizeCodeWord())}
		bad.ClaimedValues = make([]fext.E4, numRows)
		require.Error(t, params.Verify(bad))
	})
	t.Run("missing opened column", func(t *testing.T) {
		bad := *proof
		bad.OpenedColumns = nil
		input.Proof = &bad
		require.Error(t, params.Verify(input))
	})
	t.Run("missing merkle proof", func(t *testing.T) {
		bad := *proof
		bad.MerkleProofOpenedColumns = nil
		input.Proof = &bad
		require.Error(t, params.Verify(input))
	})
	t.Run("short opened column", func(t *testing.T) {
		bad := *proof
		bad.OpenedColumns = [][]koalabear.Element{proof.OpenedColumns[0][:numRows-1]}
		input.Proof = &bad
		require.Error(t, params.Verify(input))
	})
	t.Run("oversized opened column", func(t *testing.T) {
		bad := *proof
		bad.OpenedColumns = [][]koalabear.Element{append(append([]koalabear.Element(nil), proof.OpenedColumns[0]...), koalabear.Element{})}
		input.Proof = &bad
		require.Error(t, params.Verify(input))
	})
	t.Run("out of range Merkle alias", func(t *testing.T) {
		bad := input
		bad.Proof = proof
		bad.SelectedColumns = []int{params.SizeCodeWord()}
		require.Error(t, params.Verify(bad))
	})
	t.Run("negative column", func(t *testing.T) {
		bad := input
		bad.SelectedColumns = []int{-1}
		require.Error(t, params.Verify(bad))
	})
	t.Run("missing claims", func(t *testing.T) {
		bad := input
		bad.ClaimedValues = nil
		require.Error(t, params.Verify(bad))
	})
	t.Run("extra claims", func(t *testing.T) {
		bad := input
		bad.ClaimedValues = append(append([]fext.E4(nil), claims...), fext.E4{})
		require.Error(t, params.Verify(bad))
	})
	t.Run("extra opened column", func(t *testing.T) {
		bad := *proof
		bad.OpenedColumns = append(append([][]koalabear.Element(nil), proof.OpenedColumns...), proof.OpenedColumns[0])
		input.Proof = &bad
		require.Error(t, params.Verify(input))
	})
	t.Run("extra Merkle proof", func(t *testing.T) {
		bad := *proof
		bad.MerkleProofOpenedColumns = append(append([]MerkleProof(nil), proof.MerkleProofOpenedColumns...), proof.MerkleProofOpenedColumns[0])
		input.Proof = &bad
		require.Error(t, params.Verify(input))
	})
	t.Run("missing Merkle sibling", func(t *testing.T) {
		bad := *proof
		bad.MerkleProofOpenedColumns = []MerkleProof{proof.MerkleProofOpenedColumns[0][:0]}
		input.Proof = &bad
		require.Error(t, params.Verify(input))
	})
	t.Run("extra Merkle sibling", func(t *testing.T) {
		bad := *proof
		bad.MerkleProofOpenedColumns = []MerkleProof{append(append(MerkleProof(nil), proof.MerkleProofOpenedColumns[0]...), Hash{})}
		input.Proof = &bad
		require.Error(t, params.Verify(input))
	})
}

func FuzzVortex(f *testing.F) {
	const (
		sisLog2Degree = 4
		sisLog2Bound  = 8
	)

	f.Add(uint16(128), uint16(128), uint16(4), int64(0), int64(0), false)
	f.Add(uint16(64), uint16(64), uint16(126), int64(43), int64(42), true)
	f.Add(uint16(64), uint16(1), uint16(1), int64(43), int64(42), false)
	f.Add(uint16(3), uint16(116), uint16(6), int64(26), int64(63), true)

	f.Fuzz(func(t *testing.T,
		_numCol, _numRow, _numSelectedColumns uint16,
		rngSeed, sisSeed int64,
		invRate8 bool,
	) {
		assert := require.New(t)
		numCol := int(_numCol)
		numRow := int(_numRow)
		numSelectedColumns := int(_numSelectedColumns)

		invRate := 2
		if invRate8 {
			invRate = 8
		}

		numCol = nextPowerOfTwo(numCol)
		if numCol == 0 || numRow == 0 || numSelectedColumns == 0 {
			t.Skip()
		}
		if numCol > 1<<11 || numRow > 1<<11 || numSelectedColumns > numCol*invRate-1 {
			t.Skip()
		}

		var seed [32]byte
		binary.PutVarint(seed[:], rngSeed)
		// #nosec G404 -- fuzz does not require a cryptographic PRNG
		rng := rand.New(rand.NewChaCha8(seed))

		sisParams, err := sis.NewRSis(sisSeed, sisLog2Degree, sisLog2Bound, numRow)
		assert.NoError(err, "failed to create SIS params")

		params, err := NewParams(numCol, numRow, sisParams, invRate, numSelectedColumns)
		assert.NoError(err, "failed to create vortex params")

		alpha := randFext(rng)
		x := randFext(rng)
		ys := make([]fext.E4, numRow)
		selectedColumns := make([]int, numSelectedColumns)
		m := make([][]koalabear.Element, numRow)

		for i := range selectedColumns {
			selectedColumns[i] = rng.IntN(numCol*invRate - 1)
		}

		for row := range m {
			m[row] = make([]koalabear.Element, numCol)

			for j := range m[row] {
				m[row][j] = randElement(rng)
			}

			ys[row], err = EvalBasePolyLagrange(m[row], x)
			assert.NoError(err, "failed to evaluate polynomial")
		}

		proverState, err := Commit(params, m)
		assert.NoError(err, "failed to commit")

		proverState.OpenLinComb(alpha)
		proof, err := proverState.OpenColumns(selectedColumns)
		assert.NoError(err, "failed to open columns")

		err = params.Verify(VerifierInput{
			Proof:           proof,
			MerkleRoot:      proverState.GetCommitment(),
			ClaimedValues:   ys,
			EvaluationPoint: x,
			Alpha:           alpha,
			SelectedColumns: selectedColumns,
		})
		assert.NoError(err, "failed to verify proof")
	})
}

// BenchmarkVortexReal benchmarks Vortex in (estimated) production conditions for the
// zkEVM. We aim to have it commit to 4GiB of data. So about 1<<30 koalabear elements.
func BenchmarkVortexReal(b *testing.B) {

	var (
		numCol             = 1 << 19
		numRow             = 1 << 11
		invRate            = 2
		numSelectedColumns = 256
		wg                 sync.WaitGroup
		sisParams, _       = sis.NewRSis(0, 9, 16, numRow)
		params, _          = NewParams(numCol, numRow, sisParams, invRate, numSelectedColumns)
		// #nosec G404 -- test case generation does not require a cryptographic PRNG
		topRng          = rand.New(rand.NewChaCha8([32]byte{}))
		alpha           = randFext(topRng)
		selectedColumns = make([]int, 256)
	)

	for i := range selectedColumns {
		selectedColumns[i] = topRng.IntN(numCol * 2)
	}

	// Generating the matrix and filling it with PRNG elements on a single-thread would
	// be very time-consuming so we parallelize it, giving it different seeds for each
	// row.
	m := make([][]koalabear.Element, numRow)
	for row := range m {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			m[row] = make([]koalabear.Element, numCol)
			seed := [32]byte{}
			binary.PutVarint(seed[:], int64(row))

			// #nosec G404 -- test case generation does not require a cryptographic PRNG
			rng := rand.New(rand.NewChaCha8(seed))
			for j := range m[row] {
				m[row][j] = randElement(rng)
			}
		}(row)
	}

	wg.Wait()

	var (
		proverState *ProverState
		err         error
	)

	b.Run("committing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			proverState, err = Commit(params, m)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	_ = proverState
	_ = alpha

	b.Run("opening-alpha", func(b *testing.B) {
		proverState, err = Commit(params, m)
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			proverState.OpenLinComb(alpha)
		}
	})

	b.Run("opening-columns", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := proverState.OpenColumns(selectedColumns)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

}
