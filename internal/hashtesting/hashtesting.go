// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package hashtesting provides test utilities for the hashers of package hash.
package hashtesting

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/consensys/gnark-crypto/hash"
	"github.com/stretchr/testify/require"
)

// Call is one write to a hasher of the Elements.
type Call[F any] struct {
	Elements []F
}

// Marshal returns the concatenation of the marshalled elements.
func Marshal[F any, PF interface {
	*F
	Marshal() []byte
}](e []F) []byte {
	var b []byte
	for i := range e {
		b = append(b, PF(&e[i]).Marshal()...)
	}
	return b
}

// CheckMatchesByteHasher checks a field hasher against a reference hasher over
// bytes. Each call is followed by PadBlock, so that it matches one write of the
// reference, which pads the last block of each write.
//
// It checks that:
//   - Reset discards everything written before.
//   - After each call, the marshalled Sum of the hasher equals the Sum of
//     the reference.
//   - Sum does not change the state.
func CheckMatchesByteHasher[F any, PF interface {
	*F
	Marshal() []byte
}, D any](t *testing.T, newHasher func() hash.FieldHasher[F, D], newRef func() hash.StateStorer, digestBytes func(D) []byte, calls []Call[F]) {
	t.Helper()

	h := newHasher()
	for _, c := range calls {
		writeElements(h, c.Elements)
	}
	h.Reset()

	ref := newRef()
	for i, c := range calls {
		write(t, ref, Marshal[F, PF](c.Elements))
		writeElements(h, c.Elements[:len(c.Elements)/2])
		h.Sum()
		writeElements(h, c.Elements[len(c.Elements)/2:])
		h.PadBlock()

		msg := []any{"call %d of %d", i, len(calls)}
		require.Equal(t, ref.Sum(nil), digestBytes(h.Sum()), msg...)
	}
}

// CheckMerkleDamgard runs checks on calls that write full and partial blocks of
// blockLen elements, including last blocks that do not end on an element
// boundary. In "bytes", the calls are written to the byte hasher, which is
// checked against the reference. In "elements", the calls without a partial
// element are written to the field hasher, see [CheckMatchesByteHasher].
func CheckMerkleDamgard[F any, PF interface {
	*F
	Marshal() []byte
}, D any](t *testing.T, newHasher func() hash.FieldHasher[F, D], newByteHasher, newRef func() hash.StateStorer, digestBytes func(D) []byte, blockLen int, randomElements func(n int) []F) {
	t.Helper()

	var zero F
	elemSize := len(PF(&zero).Marshal())

	// A call writes k bytes of value 1, then n elements. With k != 0 it is written
	// as bytes, and its last block does not end on an element boundary.
	type call struct{ k, n int }
	cases := [][]call{
		{{n: 0}},
		{{n: 1}},
		{{n: blockLen}},
		{{n: blockLen - 1}, {n: 3}},
		{{n: 2*blockLen + 1}, {n: 5}, {n: 0}, {n: blockLen}},
		{{n: 37}},
	}
	for k := 1; k < elemSize; k++ {
		for n := range blockLen {
			cases = append(cases, []call{{n: blockLen}, {k: k, n: n}})
		}
	}

	for _, elements := range []bool{false, true} {
		name := "bytes"
		if elements {
			name = "elements"
		}
		t.Run(name, func(t *testing.T) {
			for _, cs := range cases {
				if elements && slices.ContainsFunc(cs, func(c call) bool { return c.k != 0 }) {
					continue
				}
				t.Run(fmt.Sprint(cs), func(t *testing.T) {
					if elements {
						calls := make([]Call[F], len(cs))
						for i, c := range cs {
							calls[i] = Call[F]{Elements: randomElements(c.n)}
						}
						CheckMatchesByteHasher[F, PF](t, newHasher, newRef, digestBytes, calls)
						return
					}

					inputs := make([][]byte, len(cs))
					for i, c := range cs {
						inputs[i] = append(bytes.Repeat([]byte{1}, c.k), Marshal[F, PF](randomElements(c.n))...)
					}

					h := newByteHasher()
					for _, b := range inputs {
						write(t, h, b)
					}
					h.Reset()

					ref := newRef()
					for i, b := range inputs {
						write(t, ref, b)
						write(t, h, b)

						msg := []any{"call %d of %d", i, len(cs)}
						require.Equal(t, ref.Sum(nil), h.Sum(nil), msg...)
						require.Equal(t, ref.State(), h.State(), msg...)
					}

					restored := newByteHasher()
					// writing the calls twice puts restored in a state other than h's, so that
					// SetState must change it
					for range 2 {
						for _, b := range inputs {
							write(t, restored, b)
						}
					}
					require.NoError(t, restored.SetState(h.State()))

					for _, b := range inputs {
						write(t, h, b)
						write(t, restored, b)
					}
					require.Equal(t, h.Sum(nil), restored.Sum(nil))
				})
			}
		})
	}
}

// write writes b to w, and requires no error and that all of b is reported written.
func write(t *testing.T, w io.Writer, b []byte) {
	t.Helper()
	n, err := w.Write(b)
	require.NoError(t, err)
	require.Equal(t, len(b), n)
}

// writeElements writes e to h with WriteElement if it has one element, and with
// WriteElements otherwise.
func writeElements[F, D any](h hash.FieldHasher[F, D], e []F) {
	if len(e) == 1 {
		h.WriteElement(e[0])
	} else {
		h.WriteElements(e...)
	}
}

// CheckNoAllocs checks that writing elements to h and summing does not allocate,
// through a generic caller, as downstream code uses it.
func CheckNoAllocs[F, D any](t *testing.T, h hash.FieldHasher[F, D], e F) {
	t.Helper()
	require.Zero(t, testing.AllocsPerRun(100, func() { writeAndSum[F, D](h, e) }))
}

//go:noinline
func writeAndSum[F, D any, H hash.FieldHasher[F, D]](h H, e F) D {
	h.Reset()
	for range 9 {
		h.WriteElement(e)
	}
	return h.Sum()
}

// CheckByteHasherNoAllocs checks that Reset, Write of 256 bytes of zeros, and Sum
// into a slice with enough capacity do not allocate.
func CheckByteHasherNoAllocs(t *testing.T, h hash.StateStorer) {
	t.Helper()
	input := make([]byte, 256)
	sum := make([]byte, 0, h.Size())
	require.Zero(t, testing.AllocsPerRun(100, func() { h.Reset() }), "Reset")
	require.Zero(t, testing.AllocsPerRun(100, func() { _, _ = h.Write(input) }), "Write")
	require.Zero(t, testing.AllocsPerRun(100, func() { sum = h.Sum(sum[:0]) }), "Sum")
}

// CheckCompressorConcurrent checks that c gives the same results when used from
// several goroutines at once: it computes 64 results sequentially, then computes
// each of them 200 times from each of 8 goroutines sharing c.
func CheckCompressorConcurrent[D comparable](t *testing.T, c hash.FieldCompressor[D], random func() D) {
	t.Helper()

	const nbResults, nbRepeats, nbGoroutines = 64, 200, 8
	left := make([]D, nbResults)
	right := make([]D, nbResults)
	want := make([]D, nbResults)
	for i := range nbResults {
		left[i], right[i] = random(), random()
		c.CompressField(&want[i], &left[i], &right[i])
	}

	var mismatch atomic.Bool
	var wg sync.WaitGroup
	for range nbGoroutines {
		wg.Go(func() {
			for range nbRepeats {
				for i := range nbResults {
					var got D
					c.CompressField(&got, &left[i], &right[i])
					if got != want[i] {
						mismatch.Store(true)
					}
				}
			}
		})
	}
	wg.Wait()
	require.False(t, mismatch.Load(), "concurrent use gives a different result")
}
