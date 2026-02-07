# go-betterfractions

`go-betterfractions` is a library for the go programming language aimed at making working with fractions easy and safe.

This is a rewrite to the original `go-fraction` library by **Miguel-Dorta** and **claudio4** found <a href="https://github.com/nethruster/go-fraction">here</a>

What does this package contain and what does it add over the original?

## Usage
To use the package add the next line to your imports:
```go
import (
    // ... your other imports
	fraction "github.com/sea2horses/go-betterfractions"
)
```
*You can also change the alias, so you can use `frac` instead of `fraction` for convenience*

---

Like the original package you can create a new fraction using the `New` function
```go
f1, err := fraction.New(1,2) // 1/2, nil
f2, err := fraction.New(2,3) // 2/3, nil
_, err := fraction.New(1,0)  // ErrZeroDenominator
```
**Addition:** This package also lets you create a fraction from an integer using the `NewI` function, and also provides built-in exports for `0` and `1`, done for easier use when cross-operating with integers and fractions
```go
f1 := fraction.NewI(5) // Equivalent to fraction.New(5,1)
f2 := fraction.Zero() // Equivalent to fraction.New(0,1) or fraction.NewI(0)
f3 := fraction.One() // You get the gist
```

---

You can perform operations either using the fraction's built-in methods or the versions that the library provides
```go
f1, _ := fraction.NewI(3) // 3 in Fraction Form
f2, _ := fraction.New(3, 2) // 3/2, nil

// You can either do
res, _ := f1.Add(f2)
// or
res, _ := fraction.Add(f1, f2)
```
These operations include:
- **Add** - Adds two fractions together
- **Subtract** - Subtracts two fractions together
- **Multiply** - Multiplies two fractions together
- **Divide** - Divides two fractions together
- **Equals** - Tests two fraction equality
- **Abs** - Returns a fraction without its negative component

---

These were already present in the original library, but `betterfractions` includes some other methods to make your development life easier
- **Negate** - Inverts the negative component of a fraction, turns `3/2` into `-3/2`
- **Invert** - Swaps a fraction's numerator with its denominator, for example, turns `-4/3` into `-3/4`
- **String** - Converts a fraction into a string
- **Comparators** - This includes `Less`, `LessEq`, `Greater` and `GreaterEq` which are overflow-safe

The library also brings two functions:
- **Cmp** - Compares two fractions and returns `-1`, `0` or `1` depending on their comparator, the Comparator methods are built on this
- **Parse** - Parses a string into a fraction

An example would be this:
```go
reader := bufio.NewReader(os.Stdin)
fmt.Print("Enter a fraction: ")
str, _ := reader.ReadString('\n')
// Use fraction.Parse to get the fraction from the string
f, err := fraction.Parse(str)
if err != nil {
    fmt.Println("The fraction provided is invalid.")
} else {
    fmt.Println("Resulting fraction: ", f.String())
}
```

Also, as a courtesy from the original package, there's also a conversions to and from `float64`, props to the original.
```go
floatValue := f1.Float64() // 0.5
f7, err := fraction.FromFloat64(0.5) // 1/2, nil
```

## Safety
`betterfractions` tackles a point mentioned in the original `go-fraction` package, **overflow and bound checking**, this library is completely **overflow-safe**, although most likely you won't need the overflow-safeness, it is good to have!

`betterfractions` also fundamentally changes the fraction struct, instead of being two `int32` as the numerator and denominator, like in the original, `betterfractions` handles the negative sign as a separate value, like an actual fraction, which means now both the numerator and denominator are `unsigned 64 bit integers` leaving a lot more room for usage!

## Demo

You can checkout the `betterfractions` demo by cloning/downloading the repository, going to demo and running the file.

## To-Do
- Link the tests with github actions

---

[MIT license](LICENSE)
