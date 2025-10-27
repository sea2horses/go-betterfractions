package fraction

// EDITED VERSION OF THE NETHRUSTER LIBRARY, ORIGINAL HERE:
// https://github.com/nethruster/go-fraction/blob/main/fraction.go

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"
)

// Fraction represents a fraction. It is an immutable type.
//
// It is always a valid fraction (never x/0) and it is always simplified.
type Fraction struct {
	numerator   uint64
	denominator uint64
	negative    bool // false: positive, true: negative
}

// integer is a generic interface that represents all the integer types of Go.
type integer interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

var (
	// ErrDivideByZero is returned when trying to divide by a fraction with a value of 0.
	ErrDivideByZero = errors.New("denominator cannot be zero")
	// ErrInvalid is returned when trying to get a fraction from a NaN float or when trying to parse an invalid fractions string
	ErrInvalid = errors.New("invalid conversion")
	// ErrOutOfRange is returned when trying to get a fraction from a float that is out of the range that this library
	// can represent.
	ErrOutOfRange = errors.New("the number is out of range for this library")
	// ErrZeroDenominator is returned when trying to create a new fraction with 0 as a denominator.
	ErrZeroDenominator = errors.New("denominator cannot be zero")

	// zeroValue is the simplified representation of a fraction with value 0.
	zeroValue = Zero()
)

// Creates a 'Zero' fraction
func Zero() Fraction {
	return NewI(0)
}

// Creates a 'One' fraction
func One() Fraction {
	return NewI(1)
}

// New creates a new fraction with the given numerator and denominator.
//
// It always simplifies the fraction. It returns ErrZeroDenominator if the value of the denominator is 0.
func New[T, K integer](numerator T, denominator K) (Fraction, error) {
	if denominator == 0 {
		return zeroValue, ErrZeroDenominator
	}
	if numerator == 0 {
		return zeroValue, nil
	}

	// Sign law, only negative when one of them is negative
	sign := (numerator < 0) != (denominator < 0)

	n := uint64(abs(numerator))
	d := uint64(abs(denominator))
	gcf := gcd(n, d)

	return Fraction{
		numerator:   n / gcf,
		denominator: d / gcf,
		negative:    sign,
	}, nil
}

// NewI creates a fraction just based from an integer, the old `fractions` package did not have a function like this, so any time
// you would want to compare a fraction to an integer, you would have to use New(1, 1) for example, now, you can just use
// NewI(1) to simplify this process
func NewI[T integer](numerator T) Fraction {
	sign := numerator < 0

	n := uint64(abs(numerator))

	return Fraction{
		numerator:   n,
		denominator: 1,
		negative:    sign,
	}
}

// FromFloat64 tries to create an exact fraction from the Float64 provided. Keep in mind that the range of numbers that
// floats can represent are bigger than what this fraction type that uses int64 internally can; in that case,
// ErrOutOfRange will be returned. Also keep in mind that floats usually represent approximations to a number; this
// function will try to approximate it as much as possible, but some precision may be lost.
// WARNING: Due to how floating point works, you may see some strange results like -0.3 becoming -5404319552844595/18014398509481984, for better control over parsing, you can use FromDecimalString to correctly get -3/10 or advise the user to write in fraction form
//
// If a NaN is provided, ErrInvalid will be returned.
func FromFloat64(f float64) (Fraction, error) {
	if math.IsNaN(f) {
		return zeroValue, ErrInvalid
	}
	if f < -9.223372036854775e+18 || f > 9.223372036854775e+18 {
		return zeroValue, ErrOutOfRange
	}
	if f > -2.168404344971009e-19 && f < 2.168404344971009e-19 {
		return zeroValue, nil
	}

	// Decompose float64
	bits := math.Float64bits(f)
	isNegative := bits&(1<<63) != 0
	exp := int64((bits>>52)&(1<<11-1)) - 1023
	mantissa := (bits & (1<<52 - 1)) | 1<<52 // Since we discarded tiny values, it'll never be denormalized.

	// Amount of times to shift the mantissa to the right to compensate for the exponent
	shift := 52 - exp

	// Reduce shift and mantissa as far as we can
	for mantissa&1 == 0 && shift > 0 {
		mantissa >>= 1
		shift--
	}

	// Choose whether to shift the numerator or denominator
	var shiftN, shiftD int64 = 0, 0
	if shift > 0 {
		shiftD = shift
	} else {
		shiftN = shift
	}

	// Shift that require larger shifts that what an int64 can hold, or larger than the mantissa itself, will be
	// approximated splitting it between the numerator and denominator.
	if shiftD > 62 {
		shiftD = 62
		shiftN = shift - 62
	} else if shiftN > 52 {
		shiftN = 52
		shiftD = shift - 52
	}

	numerator, denominator := int64(mantissa), int64(1)
	denominator <<= shiftD
	if shiftN < 0 {
		numerator <<= -shiftN
	} else {
		numerator >>= shiftN
	}

	if isNegative {
		numerator *= -1
	}
	return New(numerator, denominator)
}

// FromFloat64Approx returns a reduced fraction ~= f with denominator <= maxDen.
// It uses continued fractions (convergents). If f is NaN or maxDen==0, returns ErrInvalid.
// This function is specially useful since some floating points are not stored exactly in binary,
// if you want -0.3 to become -3/10 and not -5404319552844595/18014398509481984, use this function or
// alternatively, use the ParseDecimal() function which more accurately translates a decimal
// number into a fraction or the Parse() function which also coverts fraction strings like "3/2"
func FromFloat64Approx(f float64, maxDen uint64) (Fraction, error) {
	if math.IsNaN(f) || maxDen == 0 {
		return zeroValue, ErrInvalid
	}
	if f == 0 {
		return zeroValue, nil
	}
	neg := f < 0
	if neg {
		f = -f
	}

	// Continued fraction expansion
	// p/q are the current convergent, pPrev/qPrev the previous
	var pPrev, qPrev uint64 = 0, 1
	var p, q uint64 = 1, 0

	x := f
	for range 1000 { // safety bound
		a := uint64(math.Floor(x))
		// next convergent = a*(p/q) + (pPrev/qPrev)
		// new p = a*p + pPrev ; new q = a*q + qPrev (check overflow)
		if a != 0 {
			if p > math.MaxUint64/a || q > math.MaxUint64/a {
				break // overflow if we take this step; stop at previous
			}
		}
		newP := a*p + pPrev
		newQ := a*q + qPrev
		if newQ == 0 || newQ > maxDen {
			break
		}
		pPrev, qPrev = p, q
		p, q = newP, newQ

		fracPart := x - float64(a)
		if fracPart == 0 {
			break
		}
		x = 1.0 / fracPart
	}

	// p/q is our convergent
	res := Fraction{numerator: p, denominator: q, negative: neg}.normalize()
	return res, nil
}

// Parses a string either containing a fraction or a decimal number into
// the fraction struct
// Makes use of ParseFracString and ParseDecimal under the hood
func Parse(s string) (Fraction, error) {
	if strings.Contains(s, "/") {
		return ParseFracString(s)
	} else {
		return ParseDecimal(s)
	}
}

// ParseDecimal translates the string of a decimal number into a fraction
// -0.3 returns -3/10
// 0.2 returns 2/10
// 2.5 returns 5/2
func ParseDecimal(s string) (Fraction, error) {
	// Trim leftover spaces
	str := strings.TrimSpace(s)
	negative := false

	// Get the sign
	if str[0] == '-' {
		negative = true
		// Remove negative sign
		str = str[1:]
	}

	// Now get both parts of the number
	parts := strings.Split(str, ".")

	if len(parts) > 2 {
		return zeroValue, errors.New("too much dots")
	}

	var lhs uint64

	if parts[0] == "" {
		return zeroValue, errors.New("no leading numeral at left hand side of decimal")
	}

	fmt.Println("Parsing LHS")
	lhs, err := strconv.ParseUint(parts[0], 10, 64)

	if err != nil {
		return zeroValue, err
	}

	fmt.Println("LHS Parsed")

	if len(parts) == 1 {

		fmt.Println("Only numerator")
		return Fraction{
			numerator:   lhs,
			denominator: 1,
			negative:    negative,
		}, err
	}

	fmt.Println("Parsing RHS")
	rhs, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return zeroValue, err
	}

	fmt.Println("RHS Parsed")

	fmt.Println("Getting fractional...")
	fracpart, err := New(rhs, uint64(math.Pow(10, float64(getintsize(rhs)))))
	fmt.Printf("Fractional obtained: %s\n", fracpart.String())
	if err != nil {
		return zeroValue, err
	}

	return NewI(lhs).Add(fracpart)
}

// Fast Addition module when both fractions denominators are the same
func fastAdd(f1, f2 Fraction) (Fraction, error) {
	a := f1.numerator
	b := f2.numerator

	var num uint64
	var neg bool
	if f1.negative == f2.negative {
		if a > math.MaxUint64-b {
			return zeroValue, ErrOutOfRange
		}
		num = a + b
		neg = f1.negative
	} else {
		if a >= b {
			num = a - b
			neg = f1.negative
		} else {
			num = b - a
			neg = f2.negative
		}
	}

	return Fraction{numerator: num, denominator: f1.denominator, negative: neg}.normalize(), nil
}

// Add adds both fractions and returns the result.
//
// Can return ErrOutOfRange if sum overflows the uint64 limit
func Add(f1, f2 Fraction) (Fraction, error) {
	if f1.isZero() {
		return f2.normalize(), nil
	}
	if f2.isZero() {
		return f1.normalize(), nil
	}

	if f1.denominator == f2.denominator {
		return fastAdd(f1, f2)
	}

	g := gcd(f1.denominator, f2.denominator)

	scale1 := f2.denominator / g
	scale2 := f1.denominator / g

	// check a = n1*scale1, b = n2*scale2
	if f1.numerator > math.MaxUint64/scale1 || f2.numerator > math.MaxUint64/scale2 {
		return zeroValue, ErrOutOfRange
	}
	a := f1.numerator * scale1
	b := f2.numerator * scale2

	// den = (d1/g) * d2
	den := f1.denominator / g
	if den > math.MaxUint64/f2.denominator {
		return zeroValue, ErrOutOfRange
	}
	den *= f2.denominator

	var num uint64
	var neg bool
	if f1.negative == f2.negative {
		if a > math.MaxUint64-b { // sum overflow
			return zeroValue, ErrOutOfRange
		}
		num = a + b
		neg = f1.negative
	} else {
		if a >= b {
			num = a - b
			neg = f1.negative
		} else {
			num = b - a
			neg = f2.negative
		}
	}
	return (Fraction{numerator: num, denominator: den, negative: neg}).normalize(), nil
}

// Negates a fraction, turning it from negative to positive or positive to negative
func Negate(f1 Fraction) Fraction {
	if f1.numerator == 0 {
		return zeroValue
	}

	return Fraction{numerator: f1.numerator, denominator: f1.denominator, negative: !f1.negative}
}

// Inverts a fraction's numerator with its denominator
//
// Can return ErrZeroDenominator if fraction's numerator is 0
func Invert(f1 Fraction) (Fraction, error) {
	if f1.numerator == 0 {
		return zeroValue, ErrZeroDenominator
	}

	return Fraction{numerator: f1.denominator, denominator: f1.numerator, negative: f1.negative}, nil
}

// Returns a fraction without its negative component
func Abs(f Fraction) Fraction {
	return Fraction{
		numerator:   f.numerator,
		denominator: f.denominator,
		negative:    false,
	}
}

// Returns the subtraction of two fractions
// Since this function uses Add under the hood, it can also return ErrOutOfRange if the resulting numerator breaks the uint64 limit
func Subtract(f1 Fraction, f2 Fraction) (Fraction, error) {
	return Add(f1, Negate(f2))
}

// Multiply takes two fractions and then multiplies them
// it uses a different algorithm than the original fractions package to reduce overflow risk
func Multiply(f1, f2 Fraction) (Fraction, error) {
	if f1.numerator == 0 || f2.numerator == 0 {
		return zeroValue, nil
	}

	// cross-cancel to reduce overflow risk
	g1 := gcd(f1.numerator, f2.denominator)
	g2 := gcd(f2.numerator, f1.denominator)

	n1 := f1.numerator / g1
	d2 := f2.denominator / g1
	n2 := f2.numerator / g2
	d1 := f1.denominator / g2

	if n1 > math.MaxUint64/n2 || d1 > math.MaxUint64/d2 {
		return zeroValue, ErrOutOfRange
	}
	num := n1 * n2
	den := d1 * d2
	neg := f1.negative != f2.negative

	return (Fraction{numerator: num, denominator: den, negative: neg}).normalize(), nil
}

func Divide(f1 Fraction, f2 Fraction) (Fraction, error) {
	f2i, err := Invert(f2)
	if err != nil {
		return zeroValue, err
	}
	return Multiply(f1, f2i)
}

// Checks two fractions equality
//
// Although New() already disregards sign as positive if fraction is 0, this function also disregards denominator and sign if both fractions numerators are 0
func Equal(f1 Fraction, f2 Fraction) bool {
	if f1.numerator == 0 && f2.numerator == 0 {
		return true
	} else {
		return f1.numerator == f2.numerator && f1.denominator == f2.denominator && f1.negative == f2.negative
	}
}

// Add adds both fractions and returns the result.
//
// Can return ErrOutOfRange if sum overflows the uint64 limit
func (f1 Fraction) Add(f2 Fraction) (Fraction, error) {
	return Add(f1, f2)
}

// Subtracts both fractions and returns the result
//
// Can return ErrOutOfRange if subtraction overflows the uint64 limit
func (f1 Fraction) Subtract(f2 Fraction) (Fraction, error) {
	return Subtract(f1, f2)
}

// Multiply multiplies both fractions and returns the result.
//
// It returns ErrDivideByZero if it tries to divide by a fraction with value 0.
// Also can return ErrOutOfRange if multiplication is out of the uint64 range
func (f1 Fraction) Multiply(f2 Fraction) (Fraction, error) {
	return Multiply(f1, f2)
}

// Divide divides both fractions and returns the result.
//
// It returns ErrDivideByZero if it tries to divide by a fraction with value 0.
func (f1 Fraction) Divide(f2 Fraction) (Fraction, error) {
	return Divide(f1, f2)
}

// Equal compares the value of both fractions, returning true if they are equals, and false otherwise.
func (f1 Fraction) Equal(f2 Fraction) bool {
	return Equal(f1, f2)
}

// Float64 returns the value of the fraction as a float64.
func (f1 Fraction) Float64() float64 {
	val := float64(f1.numerator) / float64(f1.denominator)

	if f1.negative && f1.numerator != 0 {
		return -val
	}
	return val
}

// Denominator returns the fraction denominator.
func (f1 Fraction) Denominator() uint64 {
	return f1.denominator
}

// Numerator returns the fraction numerator.
func (f1 Fraction) Numerator() uint64 {
	return f1.numerator
}

// Sign returns the sign of the fraction
func (f1 Fraction) IsNegative() bool {
	return f1.negative
}

// Returns if fraction is 0
func (f1 Fraction) isZero() bool {
	return f1.numerator == 0
}

// To string method
func (f1 Fraction) String() string {
	if f1.numerator == 0 {
		return "0"
	}

	var str strings.Builder

	if f1.negative {
		str.WriteRune('-')
	}
	str.WriteString(fmt.Sprint(f1.numerator))
	if f1.denominator != 1 {
		str.WriteRune('/')
		str.WriteString(fmt.Sprint(f1.denominator))
	}

	return str.String()
}

// Cmp returns -1 if a<b, 0 if a==b, +1 if a>b.
func Cmp(f1 Fraction, f2 Fraction) int {
	// Fast path: equal zeros (your invariant ensures canonical 0/1/positive).
	if f1.numerator == 0 && f2.numerator == 0 {
		return 0
	}

	// Different signs: negatives are smaller.
	if f1.negative != f2.negative {
		if f1.negative {
			return -1
		}
		return 1
	}

	// Same sign: compare magnitudes safely without overflow.
	// Compare a/b ? c/d by cross-multiplying with gcd reduction:
	// a * (d/g) ? c * (b/g), where g = gcd(b, d)
	g := gcd(f1.denominator, f2.denominator)
	lmul := f2.denominator / g
	rmul := f1.denominator / g

	// 128-bit products via bits.Mul64 (hi, lo)
	ahi, alo := bits.Mul64(f1.numerator, lmul)
	bhi, blo := bits.Mul64(f2.numerator, rmul)

	c := cmp128(ahi, alo, bhi, blo)

	// If both negative, reverse the comparison
	if f1.negative { // and b.negative is also true here
		return -c
	}
	return c
}

// Method form for convenience.
func (f Fraction) Cmp(g Fraction) int { return Cmp(f, g) }

// Every single comparator that you'll ever need

func (f Fraction) Less(g Fraction) bool      { return f.Cmp(g) < 0 }
func (f Fraction) LessEq(g Fraction) bool    { return f.Cmp(g) <= 0 }
func (f Fraction) Greater(g Fraction) bool   { return f.Cmp(g) > 0 }
func (f Fraction) GreaterEq(g Fraction) bool { return f.Cmp(g) >= 0 }

// ParseFracString a string to a fraction
// This can return ErrInvalid if parsing was unsuccesful or ErrZeroDenominator if the denominator is, well, zero
func ParseFracString(str string) (Fraction, error) {
	s := strings.TrimSpace(str)

	if s == "" {
		return zeroValue, errors.New("empty fraction")
	}

	sign := false
	if s[0] == '-' {
		sign = true
		s = strings.TrimSpace(s[1:])

		if s == "" {
			return zeroValue, errors.New("no leading numeral (no numbers after sign)")
		}
	}

	parts := strings.Split(s, "/")

	if len(parts) > 2 {
		return zeroValue, errors.New("to many fraction separators '/'")
	}

	numeratorStr := strings.TrimSpace(parts[0])
	if numeratorStr == "" {
		return zeroValue, errors.New("numerator cannot be empty")
	}

	num, err := strconv.ParseUint(numeratorStr, 10, 64)
	if err != nil {
		return zeroValue, errors.New("numerator could not be parsed to unsigned 64 bit int")
	}

	den := uint64(1)
	if len(parts) == 2 {
		denominatorStr := strings.TrimSpace(parts[1])
		if denominatorStr == "" {
			return zeroValue, errors.New("fraction separator found but numerator empty")
		}

		den, err = strconv.ParseUint(denominatorStr, 10, 64)
		if err != nil {
			return zeroValue, errors.New("denominator could not be parsed to unsigned 64 bit int")
		}

		if den == 0 {
			return zeroValue, ErrZeroDenominator
		}
	}

	f := Fraction{numerator: num, denominator: den, negative: sign}
	return f.normalize(), nil
}

// Normalizes (simplifies) a fraction
func (f Fraction) normalize() Fraction {
	if f.numerator == 0 {
		return Fraction{numerator: 0, denominator: 1, negative: false}
	}
	g := gcd(f.numerator, f.denominator)
	return Fraction{
		numerator:   f.numerator / g,
		denominator: f.denominator / g,
		negative:    f.negative,
	}
}

// abs returns the absolute value of an integer.
func abs[T integer](n T) T {
	if n < 0 {
		return -n
	}
	return n
}

// gcd returns the greatest common divisor of the two numbers. It assumes that both numbers are positive integers.
func gcd(n1, n2 uint64) uint64 {
	for n2 != 0 {
		n1, n2 = n2, n1%n2
	}
	return n1
}

// cmp128 compares two unsigned 128-bit integers represented as (hi, lo).
// Returns -1 if x<y, 0 if x==y, +1 if x>y.
func cmp128(xhi, xlo, yhi, ylo uint64) int {
	if xhi < yhi {
		return -1
	}
	if xhi > yhi {
		return 1
	}
	if xlo < ylo {
		return -1
	}
	if xlo > ylo {
		return 1
	}
	return 0
}

func getintsize(i uint64) uint8 {
	if i == 0 {
		return 1
	}

	var size uint8 = 0
	for c := i; c > 0; c /= 10 {
		size += 1
	}

	return size
}
