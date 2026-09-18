// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mamabear

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"math/bits"
	"reflect"
	"strconv"
	"strings"

	"github.com/bits-and-blooms/bitset"
	"github.com/consensys/gnark-crypto/field/hash"
	"github.com/consensys/gnark-crypto/field/pool"
)

// Element represents a field element stored on 1 word (uint64), in Montgomery form.
//
// The Montgomery constant is R = 2^52.  All stored values satisfy:
//
//	canonical: z ∈ [0, p)
//	lazy:      z ∈ [0, R + p)  (allowed as input/output of lazy-reduction ops)
//
// Modulus p =
//
//	p[base10] = 562932773552129
//	p[base16] = 0x1FFFC00000001
//
// # Warning
//
// This code has not been audited and is provided as-is. In particular, there
// are no security guarantees such as constant-time implementation or
// side-channel attack resistance.
type Element [1]uint64

const (
	Limbs = 1  // number of 64-bit words
	Bits  = 49 // number of bits needed to represent an Element
	Bytes = 8  // number of bytes needed to represent an Element (padded to 8)
)

// Field modulus p = 2^49 − 2^34 + 1
const (
	q0 = uint64(562932773552129)
	q  = q0
)

var qElement = Element{q0}

var _modulus big.Int // q stored as big.Int

// Modulus returns p as a big.Int
//
//	p[base10] = 562932773552129
//	p[base16] = 0x1FFFC00000001
func Modulus() *big.Int {
	return new(big.Int).Set(&_modulus)
}

// Montgomery constant R = 2^52; mask for mod R
const (
	rBits   = 52
	rMask   = uint64((1 << rBits) - 1) // 0x000FFFFFFFFFFFFF
	qInvNeg = uint64(562932773552127)  // −p^{-1} mod 2^52;  p * qInvNeg ≡ −1 (mod 2^52)
)

// rSquare = R^2 mod p = 2^104 mod p  (raw, not in Montgomery form)
const rSquareRaw = uint64(15393129233472)

// rSquare as an Element (= R^3 mod p, i.e. R^2 mod p in Montgomery form)
// Used as the starting value in the Inverse algorithm.
var rSquare = Element{448631614111230}

func init() {
	_modulus.SetString("1FFFC00000001", 16)
}

// NewElement returns a new Element from a uint64 value.
//
// It is equivalent to:
//
//	var v Element
//	v.SetUint64(...)
func NewElement(v uint64) Element {
	z := Element{v % q0}
	z.toMont()
	return z
}

// SetUint64 sets z to v and returns z
func (z *Element) SetUint64(v uint64) *Element {
	z[0] = v % q0
	return z.toMont()
}

// SetInt64 sets z to v and returns z
func (z *Element) SetInt64(v int64) *Element {
	m := v >> 63
	z.SetUint64(uint64((v ^ m) - m))
	if m != 0 {
		z.Neg(z)
	}
	return z
}

// Set z = x and returns z
func (z *Element) Set(x *Element) *Element {
	z[0] = x[0]
	return z
}

// SetInterface converts i1 into an Element.
// Supported types: Element, *Element, uint64, int, string, *big.Int, big.Int, []byte.
func (z *Element) SetInterface(i1 any) (*Element, error) {
	if i1 == nil {
		return nil, errors.New("can't set mamabear.Element with <nil>")
	}
	switch c1 := i1.(type) {
	case Element:
		return z.Set(&c1), nil
	case *Element:
		if c1 == nil {
			return nil, errors.New("can't set mamabear.Element with <nil>")
		}
		return z.Set(c1), nil
	case uint8:
		return z.SetUint64(uint64(c1)), nil
	case uint16:
		return z.SetUint64(uint64(c1)), nil
	case uint32:
		return z.SetUint64(uint64(c1)), nil
	case uint:
		return z.SetUint64(uint64(c1)), nil
	case uint64:
		return z.SetUint64(c1), nil
	case int8:
		return z.SetInt64(int64(c1)), nil
	case int16:
		return z.SetInt64(int64(c1)), nil
	case int32:
		return z.SetInt64(int64(c1)), nil
	case int64:
		return z.SetInt64(c1), nil
	case int:
		return z.SetInt64(int64(c1)), nil
	case string:
		return z.SetString(c1)
	case *big.Int:
		if c1 == nil {
			return nil, errors.New("can't set mamabear.Element with <nil>")
		}
		return z.SetBigInt(c1), nil
	case big.Int:
		return z.SetBigInt(&c1), nil
	case []byte:
		return z.SetBytes(c1), nil
	default:
		return nil, errors.New("can't set mamabear.Element from type " + reflect.TypeOf(i1).String())
	}
}

// SetZero z = 0
func (z *Element) SetZero() *Element {
	z[0] = 0
	return z
}

// SetOne z = 1 (in Montgomery form: R mod p)
func (z *Element) SetOne() *Element {
	z[0] = 137438953464 // R mod p
	return z
}

// Div z = x * y⁻¹ (mod p)
func (z *Element) Div(x, y *Element) *Element {
	var yInv Element
	yInv.Inverse(y)
	z.Mul(x, &yInv)
	return z
}

// Equal returns z == x; constant-time
func (z *Element) Equal(x *Element) bool {
	return z.NotEqual(x) == 0
}

// NotEqual returns 0 if and only if z == x; constant-time
func (z *Element) NotEqual(x *Element) uint64 {
	return z[0] ^ x[0]
}

// IsZero returns z == 0
func (z *Element) IsZero() bool {
	return z[0] == 0
}

// IsOne returns z == 1 (i.e. the Montgomery form 1 = R mod p)
func (z *Element) IsOne() bool {
	return z[0] == 137438953464
}

// IsUint64 reports whether z can be represented as a uint64.
func (z *Element) IsUint64() bool {
	return true
}

// Uint64 returns the uint64 representation of z.
func (z *Element) Uint64() uint64 {
	return z.Bits()[0]
}

// FitsOnOneWord reports whether z (except the least significant word) is 0.
func (z *Element) FitsOnOneWord() bool {
	return true
}

// Cmp compares z and x lexicographically and returns −1, 0, or +1.
func (z *Element) Cmp(x *Element) int {
	_z := z.Bits()
	_x := x.Bits()
	if _z[0] > _x[0] {
		return 1
	} else if _z[0] < _x[0] {
		return -1
	}
	return 0
}

// LexicographicallyLargest returns true if z > (p−1)/2.
func (z *Element) LexicographicallyLargest() bool {
	_z := z.Bits()
	var b uint64
	_, b = bits.Sub64(_z[0], 281466386776065, 0) // (p-1)/2 + 1
	return b == 0
}

// SetRandom sets z to a uniform random value in [0, p).
func (z *Element) SetRandom() (*Element, error) {
	const l = 8
	const bitLen = Bits
	const k = (bitLen + 7) / 8

	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}

	var buf [l]byte
	for {
		if _, err := io.ReadFull(rand.Reader, buf[:k]); err != nil {
			return nil, err
		}
		buf[k-1] &= uint8(int(1<<b) - 1)
		z[0] = binary.LittleEndian.Uint64(buf[:])
		if z.smallerThanModulus() {
			return z.toMont(), nil
		}
	}
}

// MustSetRandom sets z to a uniform random value in [0, p); panics on error.
func (z *Element) MustSetRandom() *Element {
	if _, err := z.SetRandom(); err != nil {
		panic(err)
	}
	return z
}

// smallerThanModulus returns true if z < p
func (z *Element) smallerThanModulus() bool {
	return z[0] < q
}

// One returns 1 (in Montgomery form)
func One() Element {
	var one Element
	one.SetOne()
	return one
}

// Halve sets z to z / 2 (mod p)
func (z *Element) Halve() {
	if z[0]&1 == 1 {
		// z[0] < p < 2^49, so z[0] + p < 2^50 — no overflow
		z[0] += q
	}
	z[0] >>= 1
}

// fromMont converts z in place from Montgomery to regular representation
func (z *Element) fromMont() *Element {
	fromMont(z)
	return z
}

// Add z = x + y (mod p)
func (z *Element) Add(x, y *Element) *Element {
	t := x[0] + y[0]
	if t >= q {
		t -= q
	}
	z[0] = t
	return z
}

// Double z = x + x (mod p)
func (z *Element) Double(x *Element) *Element {
	t := x[0] << 1
	if t >= q {
		t -= q
	}
	z[0] = t
	return z
}

// Sub z = x − y (mod p)
func (z *Element) Sub(x, y *Element) *Element {
	t := x[0] - y[0]
	if t > q { // underflow occurred (unsigned wrap)
		t += q
	}
	z[0] = t
	return z
}

// Neg z = p − x
func (z *Element) Neg(x *Element) *Element {
	if x.IsZero() {
		z.SetZero()
		return z
	}
	z[0] = q - x[0]
	return z
}

// Select is a constant-time conditional move.
// If c == 0, z = x0. Else z = x1.
func (z *Element) Select(c int, x0 *Element, x1 *Element) *Element {
	cC := uint64((int64(c) | -int64(c)) >> 63) // 0 if c==0, ^0 otherwise
	z[0] = x0[0] ^ cC&(x0[0]^x1[0])
	return z
}

func _fromMontGeneric(z *Element) {
	r := montMul(z[0], 1)
	if r >= q {
		r -= q
	}
	z[0] = r
}

func _reduceGeneric(z *Element) {
	if !z.smallerThanModulus() {
		z[0] -= q
	}
}

// BatchInvert returns a new slice with every element inverted.
// Uses Montgomery batch inversion.
func BatchInvert(a []Element) []Element {
	res := make([]Element, len(a))
	if len(a) == 0 {
		return res
	}
	zeroes := bitset.New(uint(len(a)))
	accumulator := One()
	for i := range len(a) {
		if a[i].IsZero() {
			zeroes.Set(uint(i))
			continue
		}
		res[i] = accumulator
		accumulator.Mul(&accumulator, &a[i])
	}
	accumulator.Inverse(&accumulator)
	for i := len(a) - 1; i >= 0; i-- {
		if zeroes.Test(uint(i)) {
			continue
		}
		res[i].Mul(&res[i], &accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}
	return res
}

func _butterflyGeneric(a, b *Element) {
	t := *a
	a.Add(a, b)
	b.Sub(&t, b)
}

// BitLen returns the minimum number of bits needed to represent z.
func (z *Element) BitLen() int {
	return bits.Len64(z[0])
}

// Hash msg to count prime field elements.
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-5.2
func Hash(msg, dst []byte, count int) ([]Element, error) {
	const L = 16 + (Bits+7)/8
	lenInBytes := count * L
	pseudoRandomBytes, err := hash.ExpandMsgXmd(msg, dst, lenInBytes)
	if err != nil {
		return nil, err
	}
	vv := pool.BigInt.Get()
	res := make([]Element, count)
	for i := range count {
		vv.SetBytes(pseudoRandomBytes[i*L : (i+1)*L])
		res[i].SetBigInt(vv)
	}
	pool.BigInt.Put(vv)
	return res, nil
}

// Exp z = xᵏ (mod p)
func (z *Element) Exp(x Element, k *big.Int) *Element {
	if k.IsInt64() {
		return z.ExpInt64(x, k.Int64())
	}
	e := k
	if k.Sign() == -1 {
		x.Inverse(&x)
		e = pool.BigInt.Get()
		defer pool.BigInt.Put(e)
		e.Neg(k)
	}
	z.Set(&x)
	for i := e.BitLen() - 2; i >= 0; i-- {
		z.Square(z)
		if e.Bit(i) == 1 {
			z.Mul(z, &x)
		}
	}
	return z
}

// ExpInt64 z = xᵏ (mod p)
func (z *Element) ExpInt64(x Element, k int64) *Element {
	if k == 0 {
		return z.SetOne()
	}
	if k < 0 {
		x.Inverse(&x)
		k = -k
	}
	e := uint64(k)
	z.Set(&x)
	for i := int(bits.Len64(e)) - 2; i >= 0; i-- {
		z.Square(z)
		if (e>>i)&1 == 1 {
			z.Mul(z, &x)
		}
	}
	return z
}

// toMont converts z to Montgomery form: z = z * R mod p
func (z *Element) toMont() *Element {
	// montMul(z, R^2 mod p) = z * R^2 * R^{-1} mod p = z * R mod p
	r := montMul(z[0], rSquareRaw)
	if r >= q {
		r -= q
	}
	z[0] = r
	return z
}

// String returns the decimal representation of z.
func (z *Element) String() string {
	return z.Text(10)
}

// toBigInt returns z as a big.Int in Montgomery form (internal helper)
func (z *Element) toBigInt(res *big.Int) *big.Int {
	var b [Bytes]byte
	binary.BigEndian.PutUint64(b[:], z[0])
	return res.SetBytes(b[:])
}

// Text returns the string representation of z in the given base.
func (z *Element) Text(base int) string {
	if base < 2 || base > 36 {
		panic("invalid base")
	}
	if z == nil {
		return "<nil>"
	}
	const maxUint32 = 4294967295
	if base == 10 {
		var zzNeg Element
		zzNeg.Neg(z)
		zzNeg.fromMont()
		if zzNeg[0] <= maxUint32 && zzNeg[0] != 0 {
			return "-" + strconv.FormatUint(zzNeg[0], base)
		}
	}
	zz := z.Bits()
	return strconv.FormatUint(zz[0], base)
}

// BigInt sets and returns z as a *big.Int
func (z *Element) BigInt(res *big.Int) *big.Int {
	_z := *z
	_z.fromMont()
	return _z.toBigInt(res)
}

// ToBigIntRegular returns z as a big.Int in regular form.
//
// Deprecated: use BigInt(*big.Int) instead.
func (z Element) ToBigIntRegular(res *big.Int) *big.Int {
	z.fromMont()
	return z.toBigInt(res)
}

// Bits provides access to z by returning its value as a little-endian [1]uint64 array
// in regular (non-Montgomery) form.
func (z *Element) Bits() [1]uint64 {
	_z := *z
	fromMont(&_z)
	return _z
}

// Bytes returns the value of z as a big-endian byte array
func (z *Element) Bytes() (res [Bytes]byte) {
	BigEndian.PutElement(&res, *z)
	return
}

// Marshal returns the value of z as a big-endian byte slice
func (z *Element) Marshal() []byte {
	b := z.Bytes()
	return b[:]
}

// Unmarshal is an alias for SetBytes.
func (z *Element) Unmarshal(e []byte) {
	z.SetBytes(e)
}

// SetBytes interprets e as the bytes of a big-endian unsigned integer,
// sets z to that value, and returns z.
func (z *Element) SetBytes(e []byte) *Element {
	if len(e) == Bytes {
		v, err := BigEndian.Element((*[Bytes]byte)(e))
		if err == nil {
			*z = v
			return z
		}
	}
	vv := pool.BigInt.Get()
	vv.SetBytes(e)
	z.SetBigInt(vv)
	pool.BigInt.Put(vv)
	return z
}

// SetBytesCanonical interprets e as a big-endian 8-byte integer.
// Returns an error if e does not encode a value in [0, p).
func (z *Element) SetBytesCanonical(e []byte) error {
	if len(e) != Bytes {
		return errors.New("invalid mamabear.Element encoding")
	}
	v, err := BigEndian.Element((*[Bytes]byte)(e))
	if err != nil {
		return err
	}
	*z = v
	return nil
}

// SetBigInt sets z to v and returns z
func (z *Element) SetBigInt(v *big.Int) *Element {
	z.SetZero()
	var zero big.Int
	c := v.Cmp(&_modulus)
	if c == 0 {
		return z
	} else if c != 1 && v.Cmp(&zero) != -1 {
		return z.setBigInt(v)
	}
	vv := pool.BigInt.Get()
	vv.Mod(v, &_modulus)
	z.setBigInt(vv)
	pool.BigInt.Put(vv)
	return z
}

// setBigInt assumes 0 ⩽ v < p
func (z *Element) setBigInt(v *big.Int) *Element {
	vBits := v.Bits()
	if len(vBits) > 0 {
		z[0] = uint64(vBits[0])
	} else {
		z[0] = 0
	}
	return z.toMont()
}

// SetString creates a big.Int with number and calls SetBigInt on z.
func (z *Element) SetString(number string) (*Element, error) {
	vv := pool.BigInt.Get()
	if _, ok := vv.SetString(number, 0); !ok {
		return nil, errors.New("Element.SetString failed -> can't parse number into a big.Int " + number)
	}
	z.SetBigInt(vv)
	pool.BigInt.Put(vv)
	return z, nil
}

// MarshalJSON returns json encoding of z (z.Text(10))
func (z *Element) MarshalJSON() ([]byte, error) {
	if z == nil {
		return []byte("null"), nil
	}
	const maxSafeBound = 15
	s := z.Text(10)
	if len(s) <= maxSafeBound {
		return []byte(s), nil
	}
	var sbb strings.Builder
	sbb.WriteByte('"')
	sbb.WriteString(s)
	sbb.WriteByte('"')
	return []byte(sbb.String()), nil
}

// UnmarshalJSON accepts numbers and strings as input.
func (z *Element) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) > Bits*3 {
		return errors.New("value too large (max = Element.Bits * 3)")
	}
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	vv := pool.BigInt.Get()
	if _, ok := vv.SetString(s, 0); !ok {
		return errors.New("can't parse into a big.Int: " + s)
	}
	z.SetBigInt(vv)
	pool.BigInt.Put(vv)
	return nil
}

// A ByteOrder specifies how to convert byte slices into an Element.
type ByteOrder interface {
	Element(*[Bytes]byte) (Element, error)
	PutElement(*[Bytes]byte, Element)
	String() string
}

var errInvalidEncoding = errors.New("invalid mamabear.Element encoding")

// BigEndian is the big-endian implementation of ByteOrder.
var BigEndian bigEndian

type bigEndian struct{}

func (bigEndian) Element(b *[Bytes]byte) (Element, error) {
	var z Element
	z[0] = binary.BigEndian.Uint64((*b)[:])
	if !z.smallerThanModulus() {
		return Element{}, errInvalidEncoding
	}
	z.toMont()
	return z, nil
}

func (bigEndian) PutElement(b *[Bytes]byte, e Element) {
	e.fromMont()
	binary.BigEndian.PutUint64((*b)[:], e[0])
}

func (bigEndian) String() string { return "BigEndian" }

// LittleEndian is the little-endian implementation of ByteOrder.
var LittleEndian littleEndian

type littleEndian struct{}

func (littleEndian) Element(b *[Bytes]byte) (Element, error) {
	var z Element
	z[0] = binary.LittleEndian.Uint64((*b)[:])
	if !z.smallerThanModulus() {
		return Element{}, errInvalidEncoding
	}
	z.toMont()
	return z, nil
}

func (littleEndian) PutElement(b *[Bytes]byte, e Element) {
	e.fromMont()
	binary.LittleEndian.PutUint64((*b)[:], e[0])
}

func (littleEndian) String() string { return "LittleEndian" }

// Cube sets z to x^3 and returns z
func (z *Element) Cube(x *Element) *Element {
	var t Element
	t.Square(x).Mul(&t, x)
	z.Set(&t)
	return z
}

// Inverse z = x⁻¹ (mod p)
//
// If x == 0, sets and returns z = 0.
// Uses Algorithm 16 from "Efficient Software-Implementation of Finite Fields
// with Applications to Cryptography" (Savas & Koç, 2010).
func (z *Element) Inverse(x *Element) *Element {
	if x.IsZero() {
		z.SetZero()
		return z
	}

	var r, s, u, v uint64
	u = q
	s = rSquareRaw // R^2 mod p; aligns invariant so output is in Montgomery form
	r = 0
	v = x[0]

	var carry, borrow uint64

	for (u != 1) && (v != 1) {
		for v&1 == 0 {
			v >>= 1
			if s&1 == 0 {
				s >>= 1
			} else {
				s, carry = bits.Add64(s, q, 0)
				s >>= 1
				if carry != 0 {
					s |= (1 << 63)
				}
			}
		}
		for u&1 == 0 {
			u >>= 1
			if r&1 == 0 {
				r >>= 1
			} else {
				r, carry = bits.Add64(r, q, 0)
				r >>= 1
				if carry != 0 {
					r |= (1 << 63)
				}
			}
		}
		if v >= u {
			v -= u
			s, borrow = bits.Sub64(s, r, 0)
			if borrow == 1 {
				s += q
			}
		} else {
			u -= v
			r, borrow = bits.Sub64(r, s, 0)
			if borrow == 1 {
				r += q
			}
		}
	}

	if u == 1 {
		z[0] = r
	} else {
		z[0] = s
	}
	return z
}

// Legendre returns the Legendre symbol of z (either +1, −1, or 0).
//
// Uses the binary GCD algorithm.
func (z *Element) Legendre() int {
	// We work with the regular (non-Montgomery) value.
	a := z.Bits()[0]
	b := q0
	l := 1

	for a != 0 {
		pow2 := bits.TrailingZeros64(a)
		a >>= pow2
		if bMod8 := b % 8; pow2%2 == 1 && (bMod8 == 3 || bMod8 == 5) {
			l = -l
		}
		s, borrow := bits.Sub64(a, b, 0)
		if borrow == 1 {
			if b%4 == 3 && a%4 == 3 {
				l = -l
			}
			a, b = b-a, a
		} else {
			a = s
		}
	}
	if b == 1 {
		return l
	}
	return 0
}

// MulBy3 x *= 3 (mod p)
func MulBy3(x *Element) {
	var y Element
	y.Double(x)
	x.Add(x, &y)
}

// MulBy5 x *= 5 (mod p)
func MulBy5(x *Element) {
	var y Element
	y.SetUint64(5)
	x.Mul(x, &y)
}

// MulBy13 x *= 13 (mod p)
func MulBy13(x *Element) {
	var y Element
	y.SetUint64(13)
	x.Mul(x, &y)
}

// Mul2ExpNegN multiplies x by 2^{−n} (mod p) using the Montgomery representation.
//
// Since the Montgomery constant is R = 2^52, the Montgomery form of 2^{−n}
// is R / 2^n = 2^{52−n}.  Valid for 0 < n ≤ 52.
func (z *Element) Mul2ExpNegN(x *Element, n uint32) *Element {
	z[0] = montMul(x[0], uint64(1)<<(rBits-n))
	return z
}

// LazyAdd returns a + b without reduction; result ∈ [0, 2p).
// Callers are responsible for reduction before any operation requiring [0, p).
func LazyAdd(z, x, y *Element) {
	z[0] = x[0] + y[0]
}

// ConSubP performs z = min(z, z − p), i.e. a conditional subtract of p.
// Equivalent to: if z >= p { z -= p }.
func ConSubP(z *Element) {
	if z[0] >= q {
		z[0] -= q
	}
}

// ReduceFast reduces x ∈ [0, 2^64) to [0, 2^50) using the sparse form of p.
//
// Since p = 2^49 − 2^34 + 1, we have 2^49 ≡ 2^34 − 1 (mod p), so:
//
//	x = x_lo + x_hi · 2^49 ≡ x_lo + x_hi · (2^34 − 1)   (mod p)
//
// The result is NOT fully reduced; call ConSubP to bring it into [0, p).
func ReduceFast(x uint64) uint64 {
	xHi := x >> 49
	xLo := x & ((1 << 49) - 1)
	t := (xHi << 34) - xHi // xHi * (2^34 − 1)
	return xLo + t         // ∈ [0, 2^50) for x ∈ [0, R + p)
}
