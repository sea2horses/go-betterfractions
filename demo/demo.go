package main

import (
	"fmt"
	"log"
	"strconv"

	frac "github.com/sea2horses/go-betterfractions"
	fraction "github.com/sea2horses/go-betterfractions"
)

func mustNew(n, d int64) frac.Fraction {
	f, err := frac.New(n, d)
	if err != nil {
		log.Fatalf("New(%d,%d): %v", n, d, err)
	}
	return f
}

func main() {
	// Constructors
	a := mustNew(1, 3)
	b := mustNew(1, 6)
	fmt.Printf("a = %s\nb = %s\n\n", a, b)

	// Parse
	p, err := frac.ParseFracString(" -6/ 8 ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parse(\" -6/ 8 \") = %s\n\n", p) // 3/4

	// Add
	sum, err := frac.Add(a, b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s + %s = %s\n", a, b, sum)

	// Subtract
	diff, err := frac.Subtract(a, b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s - %s = %s\n", a, b, diff)

	// Multiply
	prod, err := frac.Multiply(a, b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s * %s = %s\n", a, b, prod)

	// Divide
	quot, err := frac.Divide(a, b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s / %s = %s\n\n", a, b, quot)

	// Comparisons
	c := mustNew(2, 3)
	fmt.Printf("c = %s\n", c)
	fmt.Printf("Cmp(%s, %s) = %d\n", sum, c, sum.Cmp(c)) // -1, 0, or +1
	fmt.Printf("%s <  %s ? %v\n", sum, c, sum.Less(c))
	fmt.Printf("%s <= %s ? %v\n", sum, c, sum.LessEq(c))
	fmt.Printf("%s >  %s ? %v\n", sum, c, sum.Greater(c))
	fmt.Printf("%s >= %s ? %v\n\n", sum, c, sum.GreaterEq(c))

	// Float conversion
	conv_list := []string{"-0.3", "0.2"}

	for _, c := range conv_list {
		f, err := strconv.ParseFloat(c, 64)
		if err != nil {
			fmt.Printf("Could not parse '%s' to float", c)
		}
		fmt.Printf("%s as float = %g\n", c, f)

		conv, _ := fraction.FromFloat64(f)
		fmt.Printf("%s as fraction = %s\n\n", c, conv.String())
	}

	// Zero, negate, abs
	z := frac.NewI(0)
	fmt.Printf("zero = %s (IsZero=%v, IsNegative=%v)\n", z, z.Equal(frac.Zero()), z.IsNegative())
	neg := frac.Negate(c)
	fmt.Printf("Negate(%s) = %s; Abs(...) = %s\n\n", c, neg, frac.Abs(neg))

	// Divide-by-zero example (error path)
	_, err = frac.Divide(c, z)
	if err != nil {
		fmt.Printf("Divide error (as expected): %v\n", err)
	}
}
