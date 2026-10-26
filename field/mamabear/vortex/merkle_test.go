package vortex

import (
	"crypto/sha512"
	"hash"
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/stretchr/testify/require"
)

func TestMerkleTree(t *testing.T) {

	posLists := []int{0, 1, 12, 31}
	const q = uint64(562932773552129)

	t.Run("full-zero-leaves", func(t *testing.T) {
		assert := require.New(t)
		leaves := [32]Hash{}

		tree := BuildMerkleTree(leaves[:], func() hash.Hash { return sha512.New384() })

		for _, pos := range posLists {

			proof, err := tree.Open(pos)
			assert.NoError(err)

			err = proof.Verify(pos, leaves[pos], tree.Root(), func() hash.Hash { return sha512.New384() })
			assert.NoError(err)
		}
	})

	t.Run("full-random", func(t *testing.T) {
		assert := require.New(t)

		var (
			rng = rand.New(rand.NewChaCha8([32]byte{}))
		)

		leaves := [32]Hash{}
		for i := range leaves {
			for j := range 4 {
				leaves[i][j] = mamabear.Element{rng.Uint64N(q)}
			}
		}

		tree := BuildMerkleTree(leaves[:], func() hash.Hash { return sha512.New384() })

		for _, pos := range posLists {
			proof, err := tree.Open(pos)
			assert.NoError(err)

			err = proof.Verify(pos, leaves[pos], tree.Root(), func() hash.Hash { return sha512.New384() })
			assert.NoError(err)
		}

	})

	t.Run("full-random-sha384", func(t *testing.T) {
		assert := require.New(t)

		var (
			rng = rand.New(rand.NewChaCha8([32]byte{}))
		)

		leaves := [32]Hash{}
		for i := range leaves {
			for j := range 4 {
				leaves[i][j] = mamabear.Element{rng.Uint64N(q)}
			}
		}

		nh := func() hash.Hash { return sha512.New384() }

		tree := BuildMerkleTree(leaves[:], nh)

		for _, pos := range posLists {
			proof, err := tree.Open(pos)
			assert.NoError(err)

			err = proof.Verify(pos, leaves[pos], tree.Root(), nh)
			assert.NoError(err)
		}

	})

}
