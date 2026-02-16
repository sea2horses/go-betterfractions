package fraction_test

import (
	"fmt"
	"testing"

	frac "github.com/sea2horses/go-betterfractions"
)

// --- helpers ---------------------------------------------------------------

func mustNew(t *testing.T, n, d int64) frac.Fraction {
	t.Helper()
	fr, err := frac.New(n, d)
	if err != nil {
		t.Fatalf("New(%d,%d): %v", n, d, err)
	}
	return fr
}

// --- constructors / invariants --------------------------------------------

func TestNew_NormalizesAndSign(t *testing.T) {
	got := mustNew(t, -6, -8) // should be +3/4
	if got.String() != "3/4" || got.IsNegative() {
		t.Fatalf("normalize/sign failed: got %v (neg=%v), want 3/4", got, got.IsNegative())
	}
}

func TestZeroIsCanonical(t *testing.T) {
	z1 := frac.NewI(0)
	z2 := mustNew(t, 0, 7)
	if !z1.Equal(z2) || z1.String() != "0" || z1.IsNegative() {
		t.Fatalf("canonical zero failed: z1=%v, z2=%v", z1, z2)
	}
}

// --- Add / Subtract --------------------------------------------------------

func TestAdd_Basic(t *testing.T) {
	a := mustNew(t, 1, 3)
	b := mustNew(t, 1, 6)
	sum, err := frac.Add(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if sum.String() != "1/2" {
		t.Fatalf("1/3 + 1/6 = %v, want 1/2", sum)
	}
}

func TestAdd_SameDenFastPath(t *testing.T) {
	a := mustNew(t, 3, 8)
	b := mustNew(t, 1, 8)
	sum, err := frac.Add(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if sum.String() != "1/2" {
		t.Fatalf("3/8 + 1/8 = %v, want 1/2", sum)
	}
}

func TestAdd_OppositeSigns(t *testing.T) {
	a := mustNew(t, -1, 3)
	b := mustNew(t, 1, 6)
	sum, err := frac.Add(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if sum.String() != "-1/6" {
		t.Fatalf("-1/3 + 1/6 = %v, want -1/6", sum)
	}
}

func TestSubtract_Basic(t *testing.T) {
	a := mustNew(t, 1, 3)
	b := mustNew(t, 1, 6)
	diff, err := frac.Subtract(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if diff.String() != "1/6" {
		t.Fatalf("1/3 - 1/6 = %v, want 1/6", diff)
	}
}

func TestAdd_ToZeroIsNonNegative(t *testing.T) {
	a := mustNew(t, 5, 7)
	b := mustNew(t, -5, 7)
	sum, err := frac.Add(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !sum.Equal(frac.NewI(0)) || sum.IsNegative() {
		t.Fatalf("sum to zero must be +0/1, got %v (neg=%v)", sum, sum.IsNegative())
	}
}

// --- Multiply / Divide -----------------------------------------------------

func TestMultiply_Basic(t *testing.T) {
	a := mustNew(t, 1, 2)
	b := mustNew(t, 2, 3)
	prod, err := frac.Multiply(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if prod.String() != "1/3" {
		t.Fatalf("1/2 * 2/3 = %v, want 1/3", prod)
	}
}

func TestMultiply_SignsAndCancel(t *testing.T) {
	a := mustNew(t, -3, 5)
	b := mustNew(t, 10, 9)
	prod, err := frac.Multiply(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if prod.String() != "-2/3" {
		t.Fatalf("-3/5 * 10/9 = %v, want -2/3", prod)
	}
}

func TestDivide_Basic(t *testing.T) {
	a := mustNew(t, 3, 4)
	b := mustNew(t, 9, 8)
	q, err := frac.Divide(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if q.String() != "2/3" {
		t.Fatalf("(3/4)/(9/8) = %v, want 2/3", q)
	}
}

func TestDivide_ByZero(t *testing.T) {
	a := mustNew(t, 3, 4)
	_, err := frac.Divide(a, frac.NewI(0))
	if err == nil {
		t.Fatal("expected divide-by-zero error, got nil")
	}
}

// --- Negate / Abs / Float / String ----------------------------------------

func TestNegateAndAbs(t *testing.T) {
	a := mustNew(t, 2, 3)
	na := frac.Negate(a)
	if na.String() != "-2/3" {
		t.Fatalf("Negate(2/3) = %v, want -2/3", na)
	}
	if frac.Abs(na).String() != "2/3" {
		t.Fatalf("Abs(-2/3) != 2/3")
	}
	// Negating zero stays zero
	z := frac.NewI(0)
	if frac.Negate(z).String() != "0" || frac.Negate(z).IsNegative() {
		t.Fatalf("Negate(0) must be +0, got %v", frac.Negate(z))
	}
}

// func TestFromFloat64(t *testing.T) {
// 	cases := map[float64]frac.Fraction{
// 		-0.3: mustNew(t, 3, 10),
// 		0.2:  mustNew(t, 2, 10),
// 		0.5:  mustNew(t, 1, 2),
// 	}

// 	for k, want := range cases {
// 		conv, err := frac.FromFloat64Approx(k)
// 		if err != nil {
// 			t.Fatalf("%g was not able to be converted into fraction", k)
// 		}

// 		if !conv.Equal(want) {
// 			t.Fatalf("Float %g conversion wants %s but returned %s", k, want.String(), conv.String())
// 		}
// 	}
// }

func TestParseDecimal(t *testing.T) {
	cases := map[string]frac.Fraction{
		"-0.3": mustNew(t, 3, 10),
		"0.2":  mustNew(t, 2, 10),
		"0.5":  mustNew(t, 1, 2),
		"2.5":  mustNew(t, 5, 2),
	}

	for k, want := range cases {
		fmt.Printf("t: %v\n", t)
		conv, err := frac.ParseDecimal(k)
		if err != nil {
			t.Fatalf("%s was not able to be converted into fraction, error: %v", k, err)
		}
		if !conv.Equal(want) {
			t.Fatalf("String %s was incorrectly converted into %s", k, conv)
		}
	}
}

func TestFloat64(t *testing.T) {
	a := mustNew(t, 2, 3)
	if v := a.Float64(); !(v > 0.66 && v < 0.67) {
		t.Fatalf("Float64(2/3) = %v, want ~0.666...", v)
	}
	if v := frac.Negate(a).Float64(); !(v < -0.66 && v > -0.67) {
		t.Fatalf("Float64(-2/3) = %v, want ~-0.666...", v)
	}
}

func TestString(t *testing.T) {
	if s := mustNew(t, 4, 2).String(); s != "2" {
		t.Fatalf("4/2.String() = %q, want \"2\"", s)
	}
	if s := mustNew(t, -7, 3).String(); s != "-7/3" {
		t.Fatalf("-7/3.String() = %q, want \"-7/3\"", s)
	}
}

// --- Equal / Cmp -----------------------------------------------------------

func TestEqualAndCmp(t *testing.T) {
	a := mustNew(t, 1, 2)
	b := mustNew(t, 2, 4)
	if !a.Equal(b) {
		t.Fatalf("1/2 should Equal 2/4 after normalization (b=%v)", b)
	}
	c := mustNew(t, 3, 5)
	if a.Cmp(c) != -1 || c.Cmp(a) != 1 || a.Cmp(a) != 0 {
		t.Fatalf("Cmp ordering broken: a=%v c=%v", a, c)
	}
}

// --- Parse -----------------------------------------------------------------

func TestParse_Valid(t *testing.T) {
	cases := map[string]string{
		"3/4":        "3/4",
		"  5/1  ":    "5",
		"-10/7":      "-10/7",
		"0/999":      "0",
		"-0/5":       "0",
		"42":         "42",
		"  12 / 6  ": "2",
	}
	for in, want := range cases {
		fr, err := frac.ParseFracString(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		if fr.String() != want {
			t.Fatalf("Parse(%q) = %v, want %s", in, fr, want)
		}
	}
}

func TestParse_Invalid(t *testing.T) {
	// your Parse does NOT allow negative denominators
	bad := []string{
		"", "/", " / ", "abc", "1//2", "1/2/3", "+", "-", "+/",
		"1/0",   // zero denominator
		"6/-11", // forbidden by your Parse
		"-/7", "1/-",
	}
	for _, in := range bad {
		if _, err := frac.ParseFracString(in); err == nil {
			t.Fatalf("Parse(%q) should fail", in)
		}
	}
}

func TestMethod_NegateAbsInvert(t *testing.T) {
    a := mustNew(t, 2, 3)

    if got := a.Negate(); got.String() != "-2/3" {
        t.Fatalf("Negate() = %v, want -2/3", got)
    }
    if got := a.Negate().Abs(); got.String() != "2/3" {
        t.Fatalf("Abs(Negate()) = %v, want 2/3", got)
    }

    ai, err := a.Invert()
    if err != nil {
        t.Fatalf("Invert() error: %v", err)
    }
    if ai.String() != "3/2" {
        t.Fatalf("Invert() = %v, want 3/2", ai)
    }
}

func TestMethod_InvertZeroError(t *testing.T) {
    z := frac.NewI(0)
    if _, err := z.Invert(); err == nil {
        t.Fatal("Invert(0) should error")
    }
}

func TestChain_Basic(t *testing.T) {
    a := mustNew(t, 1, 2)
    b := mustNew(t, 2, 3)
    c := mustNew(t, 1, 6)

    res, err := frac.Start(a).Sum(b).Sub(c).Result()
    if err != nil {
        t.Fatal(err)
    }
    if res.String() != "1" {
        t.Fatalf("chain result = %v, want 1", res)
    }
}

func TestChain_WithInvertNegateAbs(t *testing.T) {
    // ((-1/2).Invert()).Abs() = 2
    a := mustNew(t, -1, 2)
    res, err := frac.Start(a).Invert().Abs().Result()
    if err != nil {
        t.Fatal(err)
    }
    if res.String() != "2" {
        t.Fatalf("chain result = %v, want 2", res)
    }
}