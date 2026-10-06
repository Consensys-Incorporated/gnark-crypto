package config

type Field struct {
	Name    string
	Modulus string
	// RBits overrides the Montgomery radix, so that R = 2^RBits. Zero means the
	// default, R = 2^(NbWords*Word.BitSize).
	RBits uint
	// HandwrittenVectorASMAMD64 declares that this field ships hand-written
	// amd64 vector and FFT kernels next to its generated output.
	HandwrittenVectorASMAMD64 bool
	// PolynomialExtensions lists the degrees of the extensions of this field
	// over which the polynomial package (Polynomial, MultiLin, Pool, ...) is
	// generated. Empty means no polynomial package.
	PolynomialExtensions []int
}

var Fields []Field

func addField(f Field) {
	Fields = append(Fields, f)
}

func init() {
	addField(Field{
		Name:    "goldilocks",
		Modulus: "0xFFFFFFFF00000001",
	})
	addField(Field{
		Name:                 "koalabear",
		Modulus:              "0x7f000001",
		PolynomialExtensions: []int{6},
	})
	addField(Field{
		Name:    "babybear",
		Modulus: "0x78000001",
	})
	addField(Field{
		// 2^49 - 2^34 + 1, co-designed with AVX-512IFMA: the Montgomery radix
		// is 2^52 to match the VPMADD52 operand width, which leaves 3 bits of
		// headroom for lazy reduction.
		Name:    "mamabear",
		Modulus: "0x1FFFC00000001",
		RBits:   52,
		// field/mamabear/element_amd64.s and fft/kernel_amd64.{go,s} are
		// hand-written; the asm generator has no regime for a single-word
		// field with a sub-word radix.
		HandwrittenVectorASMAMD64: true,
		PolynomialExtensions:      []int{3},
	})
}
