package config

type Field struct {
	Name    string
	Modulus string
	// RBits overrides the Montgomery radix, so that R = 2^RBits. Zero means the
	// default, R = 2^(NbWords*Word.BitSize).
	RBits uint
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
		Name:    "koalabear",
		Modulus: "0x7f000001",
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
	})
}
