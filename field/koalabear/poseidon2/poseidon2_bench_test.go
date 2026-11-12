package poseidon2

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/koalabear"
)

func benchmarkCompressx16ColumnsWithState(b *testing.B, accelerated bool) {
	const colSize = 512
	state := make([]fr.Element, 16*8)
	matrix := make([]fr.Element, 16*colSize)
	for i := range state {
		state[i].SetUint64(uint64(i*40503 + 17))
	}
	for i := range matrix {
		matrix[i].SetUint64(uint64(i*2654435761 + 23))
	}
	result := make([][8]fr.Element, 16)
	h := NewPermutation(16, 6, 21)
	if accelerated {
		if !h.params.hasFast16_6_21 {
			b.Skip("Poseidon2 accelerator unavailable")
		}
	} else {
		h.disableAVX512()
	}

	b.ResetTimer()
	for b.Loop() {
		h.Compressx16ColumnsWithState(state, matrix, colSize, result)
	}
}

func BenchmarkCompressx16ColumnsWithState512(b *testing.B) {
	b.Run("accelerated", func(b *testing.B) {
		benchmarkCompressx16ColumnsWithState(b, true)
	})
	b.Run("generic", func(b *testing.B) {
		benchmarkCompressx16ColumnsWithState(b, false)
	})
}
