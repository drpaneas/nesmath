# nesmath

Go library for NES-accurate fixed-point arithmetic: add-with-carry,
sub-pixel accumulators, and the two's-complement tricks that classic NES
platformers use to move things around the screen with an 8-bit CPU that
has no multiply, no divide, and no floating point.

It is not a fixed-point number library. There is no `+` or `*` here.
The NES does not compute physics with numbers; it chains single-byte
additions through the carry flag. This package models that chain
directly, one hardware role at a time, instead of pretending it is
ordinary arithmetic with a decimal point somewhere in the middle.

See [NES_MATH_FOR_PORTERS.md](NES_MATH_FOR_PORTERS.md) for the reasoning,
worked frame-by-frame traces, and the porting checklist. Or try it live:
https://drpaneas.github.io/nesmath/

## Install

```
go get github.com/drpaneas/nesmath
```

Requires Go 1.22.

## Example

```go
package main

import (
	"fmt"

	"github.com/drpaneas/nesmath"
)

func main() {
	// Add-with-carry: the one arithmetic primitive the 6502 has.
	a := uint8(0x90)
	carry := nesmath.ADC(&a, 0x90, 0)
	fmt.Printf("a=%#02x carry=%d\n", a, carry) // a=0x20 carry=1

	// A signed 4.4 speed byte splits into whole pixels and a fraction.
	speed := nesmath.Q4_4(0x19)
	whole, frac := speed.Split()
	fmt.Printf("whole=%d frac=%#02x\n", whole, frac) // whole=1 frac=0x90
}
```

## What's here

- `ADC`, `SBC` - add-with-carry and subtract-with-borrow on a byte.
- `Carry` - the carry flag, threaded through calls instead of branched on.
- `Q4_4` - a signed 4.4 fixed-point speed byte, with `Split`, `Abs`, `Negate`, `Cmp`.
- `Accumulator8` - a carry-producing residue bucket. Not a number; has no value of its own.
- `Position16` - a page-and-pixel coordinate, with `AddSigned` for the standard multi-byte carry idiom.
- `HorizontalMotion`, `VerticalMotion` - the two- and three-tier carry chains that turn a speed byte into on-screen movement each frame.
- `SignExtend4to8`, `Negate`, `ASL16` - the small bit-level helpers everything above is built from.

Full API docs: https://pkg.go.dev/github.com/drpaneas/nesmath

## Test

```
go test ./...
```

Every function is checked against hand-traced fixtures, not just
round-trip properties - the numbers in the tests are the same numbers
you would get single-stepping the real 6502 arithmetic.
