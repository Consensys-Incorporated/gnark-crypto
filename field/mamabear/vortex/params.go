package vortex

import (
	"errors"
	"hash"

	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/field/mamabear/fft"
	"github.com/consensys/gnark-crypto/field/mamabear/sis"
	"github.com/consensys/gnark-crypto/utils"
)

var (
	ErrWrongSizeHash = errors.New("the hash size should be 32 bytes")
)

// HashConstructor a functions returning a hash. Hash functions are stored this way, to allocate
// them when needed and parallelise the execution when possible.
type HashConstructor = func() hash.Hash

// Configuration options of the vortex prover
type Config struct {
	// hash function used to build the Merkle tree. By default (nil), poseidon2 is used.
	merkleHashFunc HashConstructor
	// hash function used to hash the stacked codewords. By default, this hash function is SIS.
	columnHash HashConstructor
}

// Option provides options for altering the default behavior of the vortex prover.
type Option func(opt *Config) error

// WithMerkleHash specifies the hash function used to build the Merkle tree of the hashed
// columns of the stacked codewords.
func WithMerkleHash(h hash.Hash) Option {
	return func(opt *Config) error {
		bs := h.Size()
		if bs != 32 {
			return ErrWrongSizeHash
		}
		opt.merkleHashFunc = func() hash.Hash { return h }
		return nil
	}
}

// WithColumnHash specifies the hash function used to hash the columns of the stacked codewords.
func WithColumnHash(h hash.Hash) Option {
	return func(opt *Config) error {
		bs := h.Size()
		if bs != 32 {
			return ErrWrongSizeHash
		}
		opt.columnHash = func() hash.Hash { return h }
		return nil
	}
}

func defaultConfig() Config {
	return Config{merkleHashFunc: nil, columnHash: nil}
}

// Params collects the public parameters of the commitment scheme.
type Params struct {
	// RSis stores the public parameters of the ring-SIS instance in use to hash the columns.
	Key *sis.RSis
	// ReedSolomonInvRate corresponds to the inverse-rate of the Reed-Solomon code.
	ReedSolomonInvRate int
	// Domain[0]: domain for FFT^-1 of size NbColumns.
	// Domain[1]: domain for FFT of size BlowUp * NbColumns.
	Domains [2]*fft.Domain
	// NbColumns number of columns of the matrix storing the polynomials.
	NbColumns int
	// MaxNbRows number of rows of the matrix.
	MaxNbRows int
	// NumSelectedColumns indicates the number of columns to open.
	NumSelectedColumns int

	// Coset table of the small domain, bit reversed
	CosetTableBitReverse mamabear.Vector

	// Conf is used to provide some customisation.
	Conf Config
}

// NewParams constructs a new set of public parameters.
func NewParams(
	numColumns int,
	maxNumRow int,
	sisParams *sis.RSis,
	reedSolomonInvRate int,
	numSelectedColumns int,
	opts ...Option,
) (*Params, error) {
	if numColumns < 1 || !isPowerOfTwo(numColumns) {
		return nil, errors.New("number of columns must be a power of two")
	}

	if reedSolomonInvRate != 2 && reedSolomonInvRate != 4 && reedSolomonInvRate != 8 {
		return nil, errors.New("reed solomon rate must be 2, 4 or 8")
	}

	conf := defaultConfig()
	if len(opts) != 0 {
		for _, opt := range opts {
			err := opt(&conf)
			if err != nil {
				return nil, err
			}
		}
	}

	shift, err := mamabear.Generator(uint64(numColumns * reedSolomonInvRate))
	if err != nil {
		return nil, err
	}

	smallDomain := fft.NewDomain(uint64(numColumns), fft.WithShift(shift))
	cosetTable, err := smallDomain.CosetTable()
	if err != nil {
		return nil, err
	}
	cosetTableBitReverse := make(mamabear.Vector, len(cosetTable))
	copy(cosetTableBitReverse, cosetTable)
	utils.BitReverse(cosetTableBitReverse)
	bigDomain := fft.NewDomain(uint64(numColumns * reedSolomonInvRate))

	return &Params{
		Key: sisParams,
		Domains: [2]*fft.Domain{
			smallDomain,
			bigDomain,
		},
		ReedSolomonInvRate:   reedSolomonInvRate,
		NbColumns:            numColumns,
		MaxNbRows:            maxNumRow,
		NumSelectedColumns:   numSelectedColumns,
		CosetTableBitReverse: cosetTableBitReverse,
		Conf:                 conf,
	}, nil
}

// SizeCodeWord returns the number of columns after encoding.
func (p *Params) SizeCodeWord() int {
	return p.NbColumns * p.ReedSolomonInvRate
}
