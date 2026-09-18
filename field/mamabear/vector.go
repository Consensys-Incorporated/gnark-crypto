// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mamabear

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"slices"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/consensys/gnark-crypto/parallel"
)

// Vector represents a slice of Element.
//
// It implements the following interfaces:
//   - Stringer
//   - io.WriterTo
//   - io.ReaderFrom
//   - encoding.BinaryMarshaler
//   - encoding.BinaryUnmarshaler
//   - sort.Interface
type Vector []Element

// MarshalBinary implements encoding.BinaryMarshaler
func (vector *Vector) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer
	if _, err = vector.WriteTo(&buf); err != nil {
		return
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (vector *Vector) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	_, err := vector.ReadFrom(r)
	return err
}

// WriteTo implements io.WriterTo and writes big-endian encoded Elements.
// Length of the vector is encoded as a uint32 on the first 4 bytes.
func (vector *Vector) WriteTo(w io.Writer) (int64, error) {
	if err := binary.Write(w, binary.BigEndian, uint32(len(*vector))); err != nil {
		return 0, err
	}
	n := int64(4)
	var buf [Bytes]byte
	for i := range len(*vector) {
		BigEndian.PutElement(&buf, (*vector)[i])
		m, err := w.Write(buf[:])
		n += int64(m)
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// AsyncReadFrom reads the vector asynchronously.
// Any error encountered during reading is returned directly; errors during
// validation/conversion are sent on the returned channel.
func (vector *Vector) AsyncReadFrom(r io.Reader) (int64, error, chan error) { //nolint ST1008
	chErr := make(chan error, 1)
	var buf [Bytes]byte
	if read, err := io.ReadFull(r, buf[:4]); err != nil {
		close(chErr)
		return int64(read), err, chErr
	}
	headerSliceLen := uint64(binary.BigEndian.Uint32(buf[:4]))
	if lr, ok := r.(interface{ Len() int }); ok {
		if remaining := lr.Len(); remaining < 0 || headerSliceLen > uint64(remaining/Bytes) {
			close(chErr)
			return 4, io.ErrUnexpectedEOF, chErr
		}
	}

	targetSize := uint64(1 << 32)
	if bits.UintSize == 32 {
		targetSize = uint64(1 << 30)
	}
	maxAllocateSliceLength := targetSize / uint64(Bytes)

	totalRead := int64(4)
	*vector = (*vector)[:0]
	if headerSliceLen == 0 {
		if *vector == nil {
			*vector = []Element{}
		}
		close(chErr)
		return totalRead, nil, chErr
	}

	for i := uint64(0); i < headerSliceLen; i += maxAllocateSliceLength {
		if len(*vector) <= int(i) {
			(*vector) = append(*vector, make([]Element, int(min(headerSliceLen-i, maxAllocateSliceLength)))...)
		}
		bSlice := unsafe.Slice((*byte)(unsafe.Pointer(&(*vector)[i])), int(min(headerSliceLen-i, maxAllocateSliceLength))*Bytes)
		read, err := io.ReadFull(r, bSlice)
		totalRead += int64(read)
		if errors.Is(err, io.ErrUnexpectedEOF) {
			close(chErr)
			return totalRead, fmt.Errorf("less data than expected: read %d elements, expected %d", i+uint64(read)/Bytes, headerSliceLen), chErr
		}
		if err != nil {
			close(chErr)
			return totalRead, err, chErr
		}
	}

	bSlice := unsafe.Slice((*byte)(unsafe.Pointer(&(*vector)[0])), int(headerSliceLen)*Bytes)
	go func() {
		var cptErrors uint64
		parallel.Execute(int(headerSliceLen), func(start, end int) {
			var z Element
			for i := start; i < end; i++ {
				bstart := i * Bytes
				b := bSlice[bstart : bstart+Bytes]
				z[0] = binary.BigEndian.Uint64(b)
				if !z.smallerThanModulus() {
					atomic.AddUint64(&cptErrors, 1)
					return
				}
				z.toMont()
				(*vector)[i] = z
			}
		})
		if cptErrors > 0 {
			chErr <- fmt.Errorf("async read: %d elements failed validation", cptErrors)
		}
		close(chErr)
	}()
	return totalRead, nil, chErr
}

// ReadFrom reads the vector from r.
func (vector *Vector) ReadFrom(r io.Reader) (int64, error) {
	var buf [Bytes]byte
	if read, err := io.ReadFull(r, buf[:4]); err != nil {
		return int64(read), err
	}
	headerSliceLen := uint64(binary.BigEndian.Uint32(buf[:4]))
	if lr, ok := r.(interface{ Len() int }); ok {
		if remaining := lr.Len(); remaining < 0 || headerSliceLen > uint64(remaining/Bytes) {
			return 4, io.ErrUnexpectedEOF
		}
	}

	targetSize := uint64(1 << 32)
	if bits.UintSize == 32 {
		targetSize = uint64(1 << 30)
	}
	maxAllocateSliceLength := targetSize / uint64(Bytes)

	totalRead := int64(4)
	*vector = (*vector)[:0]
	if headerSliceLen == 0 && *vector == nil {
		*vector = []Element{}
	}

	for i := range headerSliceLen {
		read, err := io.ReadFull(r, buf[:])
		totalRead += int64(read)
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return totalRead, fmt.Errorf("less data than expected: read %d elements, expected %d", i, headerSliceLen)
		}
		if err != nil {
			return totalRead, fmt.Errorf("error reading element %d: %w", i, err)
		}
		if uint64(cap(*vector)) <= i {
			(*vector) = slices.Grow(*vector, int(min(headerSliceLen-i, maxAllocateSliceLength)))
		}
		el, err := BigEndian.Element(&buf)
		if err != nil {
			return totalRead, fmt.Errorf("error decoding element %d: %w", i, err)
		}
		*vector = append(*vector, el)
	}
	return totalRead, nil
}

// String implements fmt.Stringer
func (vector Vector) String() string {
	var sbb strings.Builder
	sbb.WriteByte('[')
	for i := range len(vector) {
		sbb.WriteString(vector[i].String())
		if i != len(vector)-1 {
			sbb.WriteByte(',')
		}
	}
	sbb.WriteByte(']')
	return sbb.String()
}

func (vector Vector) Len() int                { return len(vector) }
func (vector Vector) Less(i, j int) bool      { return vector[i].Cmp(&vector[j]) == -1 }
func (vector Vector) Swap(i, j int)           { vector[i], vector[j] = vector[j], vector[i] }
func (vector Vector) Equal(other Vector) bool { return slices.Equal(vector, other) }

func (vector Vector) SetRandom() error {
	for i := range vector {
		if _, err := vector[i].SetRandom(); err != nil {
			return err
		}
	}
	return nil
}

func (vector Vector) MustSetRandom() {
	for i := range vector {
		if _, err := vector[i].SetRandom(); err != nil {
			panic(err)
		}
	}
}

// Exp sets vector[i] = a[i]ᵏ for all i
func (vector Vector) Exp(a Vector, k int64) {
	N := len(a)
	if N != len(vector) {
		panic("vector.Exp: vectors don't have the same length")
	}
	if k == 0 {
		for i := range vector {
			vector[i].SetOne()
		}
		return
	}
	base := a
	exp := k
	if k < 0 {
		base = BatchInvert(a)
		exp = -k
	} else if N > 0 {
		v0 := &vector[0]
		a0 := &a[0]
		if v0 == a0 {
			base = make(Vector, N)
			copy(base, a)
		}
	}
	copy(vector, base)
	for i := bits.Len64(uint64(exp)) - 2; i >= 0; i-- {
		vector.Mul(vector, vector)
		if (uint64(exp)>>uint(i))&1 != 0 {
			vector.Mul(vector, base)
		}
	}
}

// ---- generic implementations (used as fallbacks) ----------------------------

func addVecGeneric(res, a, b Vector) {
	if len(a) != len(b) || len(a) != len(res) {
		panic("vector.Add: vectors don't have the same length")
	}
	for i := range len(a) {
		res[i].Add(&a[i], &b[i])
	}
}

func subVecGeneric(res, a, b Vector) {
	if len(a) != len(b) || len(a) != len(res) {
		panic("vector.Sub: vectors don't have the same length")
	}
	for i := range len(a) {
		res[i].Sub(&a[i], &b[i])
	}
}

func scalarMulVecGeneric(res, a Vector, b *Element) {
	if len(a) != len(res) {
		panic("vector.ScalarMul: vectors don't have the same length")
	}
	for i := range len(a) {
		res[i].Mul(&a[i], b)
	}
}

func sumVecGeneric(res *Element, a Vector) {
	for i := range len(a) {
		res.Add(res, &a[i])
	}
}

func innerProductVecGeneric(res *Element, a, b Vector) {
	if len(a) != len(b) {
		panic("vector.InnerProduct: vectors don't have the same length")
	}
	var tmp Element
	for i := range len(a) {
		tmp.Mul(&a[i], &b[i])
		res.Add(res, &tmp)
	}
}

func mulVecGeneric(res, a, b Vector) {
	if len(a) != len(b) || len(a) != len(res) {
		panic("vector.Mul: vectors don't have the same length")
	}
	for i := range len(a) {
		res[i].Mul(&a[i], &b[i])
	}
}
