//go:build js && wasm

// Command wasm compiles nesmath's interactive exercises to WebAssembly.
// It exports a handful of plain functions on the JS global object; the
// static HTML/JS frontend in web/static calls them directly. There is no
// framework and no build step beyond `go build` with GOOS=js GOARCH=wasm -
// see web/build.sh.
//
// This file only compiles for the js/wasm target (see the build
// constraint below), so it never affects `go build ./...` or
// `go test ./...` for the main nesmath module.
package main

import (
	"syscall/js"

	"github.com/drpaneas/nesmath"
)

func main() {
	js.Global().Set("nesmathADC", js.FuncOf(runADC))
	js.Global().Set("nesmathSBC", js.FuncOf(runSBC))
	js.Global().Set("nesmathChainedAdd16", js.FuncOf(runChainedAdd16))
	js.Global().Set("nesmathNegate", js.FuncOf(runNegate))
	js.Global().Set("nesmathSplit", js.FuncOf(runSplit))
	js.Global().Set("nesmathAccumulatorTrace", js.FuncOf(runAccumulatorTrace))
	js.Global().Set("nesmathSimpleMovement", js.FuncOf(runSimpleMovement))
	js.Global().Set("nesmathHorizontalTrace", js.FuncOf(runHorizontalTrace))
	js.Global().Set("nesmathVerticalTrace", js.FuncOf(runVerticalTrace))
	js.Global().Set("nesmathJumpTrajectory", js.FuncOf(runJumpTrajectory))

	// Keep the program alive: without this, main returns and the exported
	// functions become unreachable the moment the page calls one.
	select {}
}

// runADC exposes nesmath.ADC directly: three bytes in (a, b, carry-in),
// the new value of a and the carry-out come back out. This is the
// exercise that demonstrates the single primitive everything else in the
// library builds on.
func runADC(_ js.Value, args []js.Value) any {
	a := uint8(args[0].Int())
	b := uint8(args[1].Int())
	c := nesmath.Carry(args[2].Int())

	carry := nesmath.ADC(&a, b, c)

	return map[string]any{
		"result": int(a),
		"carry":  int(carry),
	}
}

// runSBC exposes nesmath.SBC directly: three bytes in (a, b, carry-in),
// the new value of a and the carry-out come back out. SBC's carry
// convention is inverted from ADC's: carry=1 means NO borrow occurred,
// carry=0 means a borrow occurred - this exercise exists specifically to
// make that counterintuitive convention visible.
func runSBC(_ js.Value, args []js.Value) any {
	a := uint8(args[0].Int())
	b := uint8(args[1].Int())
	c := nesmath.Carry(args[2].Int())

	carry := nesmath.SBC(&a, b, c)

	return map[string]any{
		"result": int(a),
		"carry":  int(carry),
	}
}

// runChainedAdd16 adds two 16-bit numbers (each given as low/high byte
// pairs) using two chained ADC calls: the low bytes first with carry-in
// 0, then the high bytes using whatever carry the low-byte addition
// produced. This is the general multi-byte pattern underneath
// Position16.AddSigned and every other multi-byte operation in the
// library, shown here unwrapped.
func runChainedAdd16(_ js.Value, args []js.Value) any {
	aLow := uint8(args[0].Int())
	aHigh := uint8(args[1].Int())
	bLow := uint8(args[2].Int())
	bHigh := uint8(args[3].Int())

	lowResult := aLow
	lowCarry := nesmath.ADC(&lowResult, bLow, 0)

	highResult := aHigh
	highCarry := nesmath.ADC(&highResult, bHigh, lowCarry)

	combined16 := uint16(highResult)<<8 | uint16(lowResult)

	return map[string]any{
		"lowResult":  int(lowResult),
		"lowCarry":   int(lowCarry),
		"highResult": int(highResult),
		"highCarry":  int(highCarry),
		"combined16": int(combined16),
	}
}

// runNegate exposes nesmath.Negate along with the intermediate one's
// complement (bit-flip) step, so the frontend can show both stages of
// "invert all bits, then add 1." Entering 0x80 demonstrates the two's
// complement edge case where negation wraps back to itself.
func runNegate(_ js.Value, args []js.Value) any {
	v := uint8(args[0].Int())

	onesComplement := ^v
	result := nesmath.Negate(v)

	return map[string]any{
		"onesComplement": int(onesComplement),
		"result":         int(result),
	}
}

// runSimpleMovement drives Position16.AddSigned directly with a raw
// signed velocity and no accumulator at all - the "Beyond Super Mario
// Bros." pattern from doc.go, for games that don't need SMB's specific
// sub-pixel scheme.
func runSimpleMovement(_ js.Value, args []js.Value) any {
	velocity := int8(args[0].Int())
	startPage := uint8(args[1].Int())
	startPixel := uint8(args[2].Int())
	frameCount := args[3].Int()

	pos := nesmath.Position16{Page: startPage, Pixel: startPixel}

	frames := make([]any, 0, frameCount)
	for frame := 1; frame <= frameCount; frame++ {
		pos.AddSigned(velocity, 0)
		frames = append(frames, map[string]any{
			"frame": frame,
			"page":  int(pos.Page),
			"pixel": int(pos.Pixel),
		})
	}
	return frames
}

// runSplit exposes Q4_4.Split, plus the raw nybbles so the frontend can
// render the byte-splitting visually without reimplementing the bit
// manipulation in JavaScript.
func runSplit(_ js.Value, args []js.Value) any {
	raw := uint8(args[0].Int())
	speed := nesmath.Q4_4(raw)

	whole, frac := speed.Split()
	highNybble := raw >> 4
	lowNybble := raw & 0x0F

	return map[string]any{
		"highNybble": int(highNybble),
		"lowNybble":  int(lowNybble),
		"whole":      int(whole),
		"frac":       int(frac),
		// The real decimal value, computed the way the article insists on:
		// the whole signed byte divided by 16, not whole+frac assembled
		// independently. int8(raw)/16 and whole+frac/256 agree by
		// construction; exposing the direct division here doubles as a
		// sanity check visible in the browser.
		"decimal": float64(int8(raw)) / 16,
	}
}

// runAccumulatorTrace drives Accumulator8.Add for frameCount frames with a
// fixed add value, returning one entry per frame. This is the exercise
// behind the "why does movement alternate 1,2,1,2" question - the
// frontend renders the carry column as a bar chart.
func runAccumulatorTrace(_ js.Value, args []js.Value) any {
	addValue := uint8(args[0].Int())
	frameCount := args[1].Int()

	var acc nesmath.Accumulator8
	frames := make([]any, 0, frameCount)
	for frame := 1; frame <= frameCount; frame++ {
		carry := acc.Add(addValue)
		frames = append(frames, map[string]any{
			"frame": frame,
			"value": int(acc.Value()),
			"carry": int(carry),
		})
	}
	return frames
}

// runHorizontalTrace drives HorizontalMotion.Step for frameCount frames at
// a fixed Q4_4 speed, returning per-frame position and sub-pixel state so
// the frontend can animate an object crossing the page boundary.
func runHorizontalTrace(_ js.Value, args []js.Value) any {
	speed := nesmath.Q4_4(uint8(args[0].Int()))
	startPage := uint8(args[1].Int())
	startPixel := uint8(args[2].Int())
	frameCount := args[3].Int()

	h := nesmath.HorizontalMotion{
		Position: nesmath.Position16{Page: startPage, Pixel: startPixel},
		Speed:    speed,
	}

	frames := make([]any, 0, frameCount)
	for frame := 1; frame <= frameCount; frame++ {
		delta := h.Step()
		frames = append(frames, map[string]any{
			"frame":     frame,
			"moveForce": int(h.MoveForce.Value()),
			"page":      int(h.Position.Page),
			"pixel":     int(h.Position.Pixel),
			"delta":     int(delta),
		})
	}
	return frames
}

// runVerticalTrace drives VerticalMotion.Step for frameCount frames with a
// fixed gravity force, returning per-frame speed/position/accumulator
// state so the frontend can plot a jump arc and mark its peak (the frame
// where Speed crosses from negative to non-negative). upForce is the
// optional mirrored upward-deceleration force (SMBDIS.ASM:7736-7758);
// pass 0 to skip that section entirely, matching the player.
func runVerticalTrace(_ js.Value, args []js.Value) any {
	initialSpeed := int8(args[0].Int())
	force := uint8(args[1].Int())
	maxSpeed := int8(args[2].Int())
	upForce := uint8(args[3].Int())
	frameCount := args[4].Int()

	v := nesmath.VerticalMotion{Speed: initialSpeed}

	frames := make([]any, 0, frameCount)
	for frame := 1; frame <= frameCount; frame++ {
		delta := v.Step(force, maxSpeed, upForce)
		frames = append(frames, map[string]any{
			"frame":     frame,
			"moveForce": int(v.MoveForce.Value()),
			"dummy":     int(v.Dummy.Value()),
			"speed":     int(v.Speed),
			"page":      int(v.Position.Page),
			"pixel":     int(v.Position.Pixel),
			"delta":     int(delta),
		})
	}
	return frames
}

// runJumpTrajectory drives a HorizontalMotion and a VerticalMotion
// together, frame by frame, returning cumulative X/Y displacement so the
// frontend can plot the combined trajectory as a single curve. This is
// the only exercise that composes both motion types in one simulation.
func runJumpTrajectory(_ js.Value, args []js.Value) any {
	hSpeed := nesmath.Q4_4(uint8(args[0].Int()))
	vSpeed := int8(args[1].Int())
	vForce := uint8(args[2].Int())
	vMaxSpeed := int8(args[3].Int())
	frameCount := args[4].Int()

	h := nesmath.HorizontalMotion{Speed: hSpeed}
	v := nesmath.VerticalMotion{Speed: vSpeed}

	var x, y int
	frames := make([]any, 0, frameCount)
	for frame := 1; frame <= frameCount; frame++ {
		x += int(h.Step())
		y += int(v.Step(vForce, vMaxSpeed, 0))
		frames = append(frames, map[string]any{
			"frame": frame,
			"x":     x,
			"y":     y,
		})
	}
	return frames
}
