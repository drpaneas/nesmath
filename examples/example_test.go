// Package examples contains runnable godoc examples for nesmath.
//
// These are not tutorials in prose - they are executable demonstrations
// of the carry-chain architecture nesmath models. Go's testing framework
// runs them and checks their printed output against the "Output:"
// comment, so they double as regression tests: if the carry-chain
// arithmetic ever behaves differently than the comments in this package
// claim, `go test` fails.
package examples

import (
	"fmt"

	"github.com/drpaneas/nesmath"
)

// ExampleADC demonstrates the single primitive everything else in this
// package builds on: add-with-carry. Adding 0x90 and 0x90 overflows an
// 8-bit byte, wrapping to 0x20 and producing a carry of 1 - exactly what
// the 6502's ADC instruction does, and exactly the mechanism a sub-pixel
// accumulator relies on to promote a whole pixel of movement.
func ExampleADC() {
	a := uint8(0x90)
	carry := nesmath.ADC(&a, 0x90, 0)
	fmt.Printf("a=%#02x carry=%d\n", a, carry)
	// Output: a=0x20 carry=1
}

// Example_q44Split demonstrates splitting a signed 4.4 speed byte
// ([nesmath.Q4_4]) into its whole-pixel and fractional components. 0x19
// is a classic platformer's observed run speed, 1.5625 pixels/frame: a
// whole part of 1 pixel, and a fraction of 0x90 (144/256 = 9/16) that
// accumulates toward a second pixel of movement roughly every other
// frame.
//
// Named with a plain "_suffix" rather than "ExampleQ4_4_Split" because
// Q4_4 itself contains an underscore, which would otherwise be
// misparsed by go vet's Example-name convention as a type named "Q4".
func Example_q44Split() {
	speed := nesmath.Q4_4(0x19)
	whole, frac := speed.Split()
	fmt.Printf("whole=%d frac=%#02x\n", whole, frac)
	// Output: whole=1 frac=0x90
}

// ExampleVerticalMotion_Step demonstrates the asymmetry between vertical
// and horizontal motion: unlike [nesmath.Q4_4], vertical speed is never
// nybble-split. Here Speed starts at -4 (a standing jump's initial
// velocity) and a gravity force of 0x20 accumulates in MoveForce every
// frame. MoveForce is too small to overflow on any single frame, so
// Speed stays at -4 for seven frames - then, on frame 8, MoveForce
// overflows and Speed changes by a whole unit, from -4 to -3. Compare
// this to [ExampleHorizontalMotion_Step], where the sub-pixel
// accumulator affects *position* every other frame while speed itself
// never changes; here it is Speed that changes, in whole-pixel steps,
// while position moves smoothly underneath it via a second, independent
// accumulator (MoveForce's old value feeding Dummy).
func ExampleVerticalMotion_Step() {
	v := nesmath.VerticalMotion{Speed: -4}

	for frame := 1; frame <= 8; frame++ {
		v.Step(0x20, 0x7F, 0)
		fmt.Printf("frame %d: MoveForce=%#02x Speed=%d\n", frame, uint8(v.MoveForce), v.Speed)
	}
	// Output:
	// frame 1: MoveForce=0x20 Speed=-4
	// frame 2: MoveForce=0x40 Speed=-4
	// frame 3: MoveForce=0x60 Speed=-4
	// frame 4: MoveForce=0x80 Speed=-4
	// frame 5: MoveForce=0xa0 Speed=-4
	// frame 6: MoveForce=0xc0 Speed=-4
	// frame 7: MoveForce=0xe0 Speed=-4
	// frame 8: MoveForce=0x00 Speed=-3
}

// ExamplePosition16_AddSigned demonstrates that nesmath does not require
// [nesmath.Q4_4] or [nesmath.Accumulator8] at all for games (or objects)
// that do not need this specific sub-pixel scheme. Many NES
// games move objects by a fixed whole number of pixels per frame with no
// fractional component whatsoever, or add a raw signed velocity byte
// directly into a position - both cases are just repeated calls to
// AddSigned with carryIn always 0, no accumulator feeding it.
//
// Here an object at pixel 253 moves right 4 pixels per frame; on the
// first frame it crosses the page boundary (253+4 wraps past 255), and
// AddSigned's sign-extended carry promotes Page from 0 to 1 exactly the
// same way it would for a fractional-speed object - the mechanism does
// not care whether the delta came from a [nesmath.Q4_4] split or, as
// here, a plain constant.
func ExamplePosition16_AddSigned() {
	pos := nesmath.Position16{Page: 0, Pixel: 253}
	const velocity = 4

	for frame := 1; frame <= 3; frame++ {
		pos.AddSigned(velocity, 0)
		fmt.Printf("frame %d: page=%d pixel=%d\n", frame, pos.Page, pos.Pixel)
	}
	// Output:
	// frame 1: page=1 pixel=1
	// frame 2: page=1 pixel=5
	// frame 3: page=1 pixel=9
}

// ExampleHorizontalMotion_Step demonstrates the full carry pipeline over
// four frames at a classic run speed (0x19). The pixel movement alternates
// 1, 2, 1, 2 - not a rounding bug, but the deterministic result of the
// sub-pixel accumulator (0x90 = 9/16 scaled to /256) overflowing every
// other frame. The long-run average, (1+2)/2 = 1.5, converges to the true
// value of 1.5625 as more frames accumulate.
func ExampleHorizontalMotion_Step() {
	h := nesmath.HorizontalMotion{
		Position: nesmath.Position16{Page: 2, Pixel: 0x40},
		Speed:    0x19,
	}
	for frame := 1; frame <= 4; frame++ {
		delta := h.Step()
		fmt.Printf("frame %d: moved %d pixel(s), pixel now %#02x\n", frame, delta, h.Position.Pixel)
	}
	// Output:
	// frame 1: moved 1 pixel(s), pixel now 0x41
	// frame 2: moved 2 pixel(s), pixel now 0x43
	// frame 3: moved 1 pixel(s), pixel now 0x44
	// frame 4: moved 2 pixel(s), pixel now 0x46
}
