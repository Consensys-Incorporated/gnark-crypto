package vortex

import (
	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/field/mamabear/sis"
	"github.com/consensys/gnark-crypto/parallel"
)

// transversalHash hashes the columns of codewords, using SIS by default, unless ots (="other than sis") is not nil.
func transversalHash(codewords []mamabear.Element, s *sis.RSis, sizeCodeWord int, ots HashConstructor) []mamabear.Element {
	if ots != nil {
		return transversalHashGeneric(codewords, ots, sizeCodeWord)
	} else {
		return transversalHashSIS(codewords, s, sizeCodeWord)
	}
}

// transversalHashGeneric hashes the columns of the codewords using the provided hash function.
// The result should be read 4 elements at a time (= 32 bytes = SHA-256 output size).
func transversalHashGeneric(codewords []mamabear.Element, newHash HashConstructor, sizeCodeWord int) []mamabear.Element {

	const nbMamabearElementsPerHash = 4

	nbCols := sizeCodeWord
	nbRows := len(codewords) / sizeCodeWord

	res := make([]mamabear.Element, nbCols*nbMamabearElementsPerHash)

	parallel.Execute(nbCols, func(start, end int) {
		h := newHash()
		for i := start; i < end; i++ {
			h.Reset()
			for j := range nbRows {
				curElmt := codewords[j*nbCols+i]
				h.Write(curElmt.Marshal())
			}
			curHash := h.Sum(nil)
			s := i * nbMamabearElementsPerHash
			for j := range nbMamabearElementsPerHash {
				res[s+j].SetBytes(curHash[mamabear.Bytes*j : mamabear.Bytes*(j+1)])
			}
		}
	})
	return res
}

// transversalHashSIS hashes the columns of the codewords using the SIS hash function.
func transversalHashSIS(codewords []mamabear.Element, s *sis.RSis, sizeCodeWord int) []mamabear.Element {

	nbCols := sizeCodeWord
	nbRows := len(codewords) / sizeCodeWord
	sisKeySize := s.Degree

	res := make([]mamabear.Element, nbCols*sisKeySize)

	parallel.Execute(nbCols, func(start, end int) {
		windowSize := 4
		n := end - start
		for n%windowSize != 0 {
			windowSize /= 2
		}
		transposed := make([][]mamabear.Element, windowSize)
		for i := range transposed {
			transposed[i] = make([]mamabear.Element, nbRows)
		}
		for col := start; col < end; col += windowSize {
			for i := range nbRows {
				for j := range transposed {
					transposed[j][i] = codewords[i*sizeCodeWord+col+j]
				}
			}
			for j := range transposed {
				s.Hash(transposed[j], res[(col+j)*sisKeySize:(col+j)*sisKeySize+sisKeySize])
			}
		}
	})

	return res
}
