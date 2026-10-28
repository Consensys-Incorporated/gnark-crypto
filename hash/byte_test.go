// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash_test

import (
	"bytes"
	"testing"

	kb "github.com/consensys/gnark-crypto/field/koalabear"
	p2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/stretchr/testify/require"
)

func TestByteHasherRejectedBlock(t *testing.T) {
	var block [8]kb.Element
	for i := range block {
		block[i].SetUint64(uint64(i + 1))
	}
	var valid []byte
	for i := range block {
		valid = append(valid, block[i].Marshal()...)
	}
	invalid := bytes.Clone(valid)
	// Reject a noncanonical element after valid elements in the same block.
	copy(invalid[len(invalid)-kb.Bytes:], bytes.Repeat([]byte{0xff}, kb.Bytes))
	h := p2.NewMerkleDamgardHasher()
	initial := h.State()
	n, err := h.Write(invalid)
	require.Error(t, err)
	require.Zero(t, n)
	require.Equal(t, initial, h.State())

	n, err = h.Write(append(bytes.Clone(valid), invalid...))
	require.Error(t, err)
	require.Equal(t, len(valid), n)
	ref := p2.NewMerkleDamgardHasher()
	_, err = ref.Write(valid)
	require.NoError(t, err)
	require.Equal(t, ref.State(), h.State())
	_, err = h.Write(valid)
	require.NoError(t, err)
	_, err = ref.Write(valid)
	require.NoError(t, err)
	require.Equal(t, ref.State(), h.State())
}

func TestByteHasherStateEncoding(t *testing.T) {
	h := p2.NewMerkleDamgardHasher()
	require.NoError(t, h.SetState([]byte{1, 2, 3}))
	state := make([]byte, h.Size())
	copy(state[len(state)-3:], []byte{1, 2, 3})
	require.Equal(t, state, h.State())
	require.Error(t, h.SetState(make([]byte, h.Size()+1)))
	require.Equal(t, state, h.State())
	bad := bytes.Clone(state)
	copy(bad[:kb.Bytes], bytes.Repeat([]byte{0xff}, kb.Bytes))
	require.Error(t, h.SetState(bad))
	require.Equal(t, state, h.State())
	// Returned state is caller-owned, not the live chaining state.
	returned := h.State()
	clear(returned)
	require.Equal(t, state, h.State())
}
