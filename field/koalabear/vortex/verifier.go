package vortex

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

// VerifierInput collects all the inputs to the verifier
// of a vortex opening.
type VerifierInput struct {

	// MerkleRoot is the commitment to the input matrix
	MerkleRoot Hash

	// ClaimedValue value of the leaf. This field is exported
	ClaimedValues []fext.E4

	// EvaluationPoint is the evaluation point
	EvaluationPoint fext.E4

	// SelectedColumns are the positions of the columns sampled by
	// the verifier
	SelectedColumns []int

	// Alpha is the coin sampled by the verifier to compute the
	// linear combination UAlpha
	Alpha fext.E4

	// Proof is the opening proof
	Proof *Proof
}

// Verify implements the verification algorithm for a Vortex opening proof.
func (p *Params) Verify(input VerifierInput) error {

	proof := input.Proof
	if proof == nil {
		return errors.New("invalid proof: missing opening proof")
	}
	n := p.SizeCodeWord()
	if len(proof.UAlpha) != n {
		return fmt.Errorf("invalid proof: uAlpha must have %d elements", n)
	}
	if len(input.SelectedColumns) == 0 || len(input.SelectedColumns) != p.NumSelectedColumns ||
		len(proof.OpenedColumns) != len(input.SelectedColumns) ||
		len(proof.MerkleProofOpenedColumns) != len(input.SelectedColumns) {
		return errors.New("invalid proof: incorrect number of opened columns")
	}
	if len(input.ClaimedValues) == 0 || len(input.ClaimedValues) > p.MaxNbRows {
		return errors.New("invalid proof: incorrect number of claimed rows")
	}
	merkleDepth := log2Ceil(n)
	for i, c := range input.SelectedColumns {
		if c < 0 || c >= n || len(proof.OpenedColumns[i]) != len(input.ClaimedValues) ||
			len(proof.MerkleProofOpenedColumns[i]) != merkleDepth {
			return fmt.Errorf("invalid proof: malformed opening at column %d", i)
		}
	}
	root := input.MerkleRoot

	// This checks the consistency between uAlpha and the claimed value
	uAlphaAtX, err := EvalFextPolyLagrange(proof.UAlpha, input.EvaluationPoint)
	claimsAtAlpha := EvalFextPolyHorner(input.ClaimedValues, input.Alpha)

	if err != nil {
		return fmt.Errorf("invalid proof: could not evaluate uAlpha: %w", err)
	}

	if uAlphaAtX != claimsAtAlpha {
		return errors.New("invalid proof: ualpha and the claim do not match")
	}

	// This checks the reed-solomon member ship of UAlpha
	if !p.IsReedSolomonCodewords(proof.UAlpha) {
		return fmt.Errorf("invalid proof: uAlpha is not a reed-solomon codeword")
	}

	// This checks linear combination of the opened columns matches the requested position of the UAlpha
	if p.checkColLinCombination(input) != nil {
		return fmt.Errorf("invalid proof: uAlpha is not a correct linear combination")
	}

	// This checks the consistency between the proof and the selected columns
	// to the input matrix.
	for i, c := range input.SelectedColumns {

		sisHash := make([]koalabear.Element, p.Key.Degree)
		if err := p.Key.Hash(proof.OpenedColumns[i], sisHash); err != nil {
			return fmt.Errorf("invalid proof: could not hash the column: %w", err)
		}

		leaf := HashPoseidon2(sisHash)

		if err := proof.MerkleProofOpenedColumns[i].Verify(c, leaf, root, p.Conf.merkleHashFunc); err != nil {
			return fmt.Errorf("invalid proof: merkle proof verification failed: %w", err)
		}
	}

	return nil

}

// Check linear combination of the opened columns matches the requested position of the UAlpha
func (p *Params) checkColLinCombination(input VerifierInput) error {
	uAlpha := input.Proof.UAlpha

	for i, selectedColID := range input.SelectedColumns {
		if selectedColID < 0 || selectedColID >= len(uAlpha) {
			return fmt.Errorf("column index %d is out of bounds for the linear combination array of size %d", selectedColID, len(uAlpha))
		}

		// Compute the linear combination of the opened column
		y := EvalBasePolyHorner(input.Proof.OpenedColumns[i], input.Alpha)

		// Check the consistency
		if y != uAlpha[selectedColID] {
			return fmt.Errorf("inconsistent linear combination at index %d (selected column ID %d): expected uAlpha[selectedColID] %s, got %s", i, selectedColID, uAlpha[selectedColID].String(), y.String())
		}
	}

	return nil
}
