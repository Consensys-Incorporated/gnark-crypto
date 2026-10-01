package vortex

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/field/mamabear"
	fext "github.com/consensys/gnark-crypto/field/mamabear/extensions"
)

// VerifierInput collects all the inputs to the verifier of a vortex opening.
type VerifierInput struct {
	// MerkleRoot is the commitment to the input matrix.
	MerkleRoot Hash

	// ClaimedValues is the value of the leaf.
	ClaimedValues []fext.E3

	// EvaluationPoint is the evaluation point.
	EvaluationPoint fext.E3

	// SelectedColumns are the positions of the columns sampled by the verifier.
	SelectedColumns []int

	// Alpha is the coin sampled by the verifier to compute the linear combination UAlpha.
	Alpha fext.E3

	// Proof is the opening proof.
	Proof *Proof
}

// Verify implements the verification algorithm for a Vortex opening proof.
func (p *Params) Verify(input VerifierInput) error {

	proof := input.Proof
	root := input.MerkleRoot

	// Check consistency between uAlpha and the claimed value.
	uAlphaAtX, err := EvalFextPolyLagrange(input.Proof.UAlpha, input.EvaluationPoint)
	claimsAtAlpha := EvalFextPolyHorner(input.ClaimedValues, input.Alpha)

	if err != nil {
		return fmt.Errorf("invalid proof: could not evaluate uAlpha: %w", err)
	}

	if uAlphaAtX != claimsAtAlpha {
		return errors.New("invalid proof: ualpha and the claim do not match")
	}

	// Check reed-solomon membership of UAlpha.
	if !p.IsReedSolomonCodewords(proof.UAlpha) {
		return fmt.Errorf("invalid proof: uAlpha is not a reed-solomon codeword")
	}

	// Check linear combination of the opened columns matches UAlpha.
	if p.checkColLinCombination(input) != nil {
		return fmt.Errorf("invalid proof: uAlpha is not a correct linear combination")
	}

	// Check consistency between the proof and the selected columns.
	nbBytes := mamabear.Bytes
	for i, c := range input.SelectedColumns {

		sisHash := make([]mamabear.Element, p.Key.Degree)
		if err := p.Key.Hash(proof.OpenedColumns[i], sisHash); err != nil {
			return fmt.Errorf("invalid proof: could not hash the column: %w", err)
		}

		// Hash the SIS output into a Merkle leaf using the Merkle hash function.
		var leaf Hash
		if p.Conf.merkleHashFunc == nil {
			leaf = HashPoseidon2(sisHash)
		} else {
			h := p.Conf.merkleHashFunc()
			for _, e := range sisHash {
				h.Write(e.Marshal())
			}
			bs := h.Sum(nil)
			for j := range 4 {
				leaf[j].SetBytes(bs[nbBytes*j : nbBytes*j+nbBytes])
			}
		}

		if err := proof.MerkleProofOpenedColumns[i].Verify(c, leaf, root, p.Conf.merkleHashFunc); err != nil {
			return fmt.Errorf("invalid proof: merkle proof verification failed: %w", err)
		}
	}

	return nil
}

// checkColLinCombination checks that the linear combination of opened columns matches UAlpha.
func (p *Params) checkColLinCombination(input VerifierInput) error {
	uAlpha := input.Proof.UAlpha

	for i, selectedColID := range input.SelectedColumns {
		if selectedColID < 0 || selectedColID >= len(uAlpha) {
			return fmt.Errorf("column index %d is out of bounds for the linear combination array of size %d", selectedColID, len(uAlpha))
		}

		y := EvalBasePolyHorner(input.Proof.OpenedColumns[i], input.Alpha)

		if y != uAlpha[selectedColID] {
			return fmt.Errorf("inconsistent linear combination at index %d (selected column ID %d): expected uAlpha[selectedColID] %s, got %s", i, selectedColID, uAlpha[selectedColID].String(), y.String())
		}
	}

	return nil
}
