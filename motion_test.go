package nesmath

import "testing"

// TestHorizontalMotion_Step_Trace reproduces the 4-frame trace from the
// design spec: speed 0x19 (1.5625 px/frame) starting at page=2,
// pixel=0x40, moveForce=0x00, alternating 1,2,1,2 pixels moved per frame
// because the sub-pixel accumulator (0x90 = 9/16 scaled to /256) overflows
// every other frame.
func TestHorizontalMotion_Step_Trace(t *testing.T) {
	h := HorizontalMotion{
		Position: Position16{Page: 2, Pixel: 0x40},
		Speed:    0x19,
	}

	type frameExpectation struct {
		wantMoveForce Accumulator8
		wantPixel     uint8
		wantDelta     int8
	}
	frames := []frameExpectation{
		{wantMoveForce: 0x90, wantPixel: 0x41, wantDelta: 1},
		{wantMoveForce: 0x20, wantPixel: 0x43, wantDelta: 2},
		{wantMoveForce: 0xB0, wantPixel: 0x44, wantDelta: 1},
		{wantMoveForce: 0x40, wantPixel: 0x46, wantDelta: 2},
	}

	for i, want := range frames {
		delta := h.Step()
		if h.MoveForce != want.wantMoveForce {
			t.Errorf("frame %d: MoveForce = %#02x, want %#02x", i+1, uint8(h.MoveForce), uint8(want.wantMoveForce))
		}
		if h.Position.Pixel != want.wantPixel {
			t.Errorf("frame %d: Pixel = %#02x, want %#02x", i+1, h.Position.Pixel, want.wantPixel)
		}
		if h.Position.Page != 2 {
			t.Errorf("frame %d: Page = %d, want 2 (no page crossing expected in this trace)", i+1, h.Position.Page)
		}
		if delta != want.wantDelta {
			t.Errorf("frame %d: delta = %d, want %d", i+1, delta, want.wantDelta)
		}
	}
}

// TestHorizontalMotion_Step_ZeroSpeed is the stationary-object baseline
// (the flagpole flag before collision, static enemies): speed 0x00 splits
// to (0, 0x00), so MoveForce never accumulates and Position never moves.
func TestHorizontalMotion_Step_ZeroSpeed(t *testing.T) {
	h := HorizontalMotion{
		Position: Position16{Page: 2, Pixel: 0x40},
		Speed:    0x00,
	}

	for frame := 1; frame <= 4; frame++ {
		delta := h.Step()
		if delta != 0 {
			t.Errorf("frame %d: delta = %d, want 0", frame, delta)
		}
		if h.Position.Pixel != 0x40 || h.Position.Page != 2 {
			t.Errorf("frame %d: position = %+v, want {Page:2 Pixel:0x40}", frame, h.Position)
		}
		if h.MoveForce != 0x00 {
			t.Errorf("frame %d: MoveForce = %#02x, want 0x00", frame, uint8(h.MoveForce))
		}
	}
}

// TestHorizontalMotion_Step_WalkSpeed is the clean-integer contrast to the
// 0x19 run-speed trace: 0x10 = 1.0 px/frame exactly, frac is always 0x00,
// so MoveForce never accumulates and there is no 1,2,1,2 oscillation -
// every frame moves exactly 1 pixel.
func TestHorizontalMotion_Step_WalkSpeed(t *testing.T) {
	h := HorizontalMotion{
		Position: Position16{Page: 2, Pixel: 0x40},
		Speed:    0x10,
	}

	for frame := 1; frame <= 4; frame++ {
		delta := h.Step()
		if delta != 1 {
			t.Errorf("frame %d: delta = %d, want 1", frame, delta)
		}
		if h.MoveForce != 0x00 {
			t.Errorf("frame %d: MoveForce = %#02x, want 0x00 (frac is always zero at walk speed)", frame, uint8(h.MoveForce))
		}
	}
	if h.Position.Pixel != 0x44 {
		t.Errorf("final Pixel = %#02x, want %#02x", h.Position.Pixel, 0x44)
	}
}

// TestHorizontalMotion_Step_PageCrossRightward verifies carry propagation
// from Pixel overflow into Page during rightward movement: pixel 0xFF
// wrapping to 0x00 increments Page.
func TestHorizontalMotion_Step_PageCrossRightward(t *testing.T) {
	h := HorizontalMotion{
		Position: Position16{Page: 2, Pixel: 0xFE},
		Speed:    0x19,
	}

	h.Step() // frame 1: 0xFE + 1 + 0 = 0xFF, no cross yet
	if h.Position.Pixel != 0xFF || h.Position.Page != 2 {
		t.Errorf("after frame 1: position = %+v, want {Page:2 Pixel:0xFF}", h.Position)
	}

	h.Step() // frame 2: 0xFF + 1 + 1(carry) = 0x01, page crosses to 3
	if h.Position.Pixel != 0x01 || h.Position.Page != 3 {
		t.Errorf("after frame 2: position = %+v, want {Page:3 Pixel:0x01}", h.Position)
	}
}

// TestHorizontalMotion_Step_PageCrossLeftward verifies borrow propagation
// from Pixel underflow into Page during leftward movement: speed 0xE4
// splits to whole=-2, and moving left from pixel 0x01 crosses the page
// boundary downward via sign extension.
func TestHorizontalMotion_Step_PageCrossLeftward(t *testing.T) {
	h := HorizontalMotion{
		Position: Position16{Page: 2, Pixel: 0x01},
		Speed:    0xE4,
	}

	h.Step()
	if h.Position.Pixel != 0xFF || h.Position.Page != 1 {
		t.Errorf("position = %+v, want {Page:1 Pixel:0xFF}", h.Position)
	}
}

// TestHorizontalMotion_Step_LongRunConvergence confirms the determinism
// property nesmath's design is built around: over 16 frames at 0x19
// (1.5625 px/frame), total pixel movement is exactly 25 - 16*1.5625 -
// with no drift or rounding error, even though the sub-pixel accumulator
// does not repeat its carry pattern on a simple 2-frame cycle (0x90 does
// not evenly divide 256 by a power of 2 smaller than 16, so the true
// cycle is 16 frames with 9 carries, not 8).
func TestHorizontalMotion_Step_LongRunConvergence(t *testing.T) {
	h := HorizontalMotion{
		Position: Position16{Page: 0, Pixel: 0x00},
		Speed:    0x19,
	}

	total := 0
	for frame := 0; frame < 16; frame++ {
		total += int(h.Step())
	}

	if total != 25 {
		t.Errorf("total pixels moved over 16 frames = %d, want 25", total)
	}
}

func TestHorizontalMotion_Step_NegativeSpeed(t *testing.T) {
	// Speed 0xE4 splits into whole=-2, frac=0x40 (see TestQ4_4_Split).
	// Starting MoveForce at 0, the first frame should not overflow
	// (0x00+0x40=0x40), so movement is exactly whole: -2 pixels.
	h := HorizontalMotion{
		Position: Position16{Page: 2, Pixel: 0x40},
		Speed:    0xE4,
	}

	delta := h.Step()
	if delta != -2 {
		t.Errorf("delta = %d, want -2", delta)
	}
	if h.Position.Pixel != 0x3E {
		t.Errorf("Pixel = %#02x, want %#02x", h.Position.Pixel, 0x3E)
	}
	if h.MoveForce != 0x40 {
		t.Errorf("MoveForce = %#02x, want %#02x", uint8(h.MoveForce), 0x40)
	}
}

// TestVerticalMotion_Step_SpeedIsNeverNybbleSplit is a regression guard
// for the bug this type was rewritten to fix: an earlier implementation
// treated vertical Speed as a [Q4_4] and nybble-split it exactly like
// [HorizontalMotion] does. It is not - ImposeGravity (SMBDIS.ASM:7704)
// adds Speed to Position as a raw signed whole byte, with no split at
// all (see NES_MATH_FOR_PORTERS.md's "Vertical speed is not 4.4").
//
// Speed=0xFC (-4) is the specific value that exposes the bug most
// clearly: nybble-split as a Q4_4, 0xFC produces whole=-1 (see
// TestQ4_4_Split), so a Step call would move the object only -1 pixel.
// Added as a raw int8, it must move -4 pixels. force=0 isolates this
// from any accumulator-driven extra pixel, so the returned delta must be
// exactly Speed's raw value.
func TestVerticalMotion_Step_SpeedIsNeverNybbleSplit(t *testing.T) {
	v := VerticalMotion{Speed: -4}

	delta := v.Step(0, 127, 0)

	if delta != -4 {
		t.Errorf("delta = %d, want -4 (if this is -1, Speed is being nybble-split again - see TestQ4_4_Split's 0xFC case)", delta)
	}
	if v.Position.Pixel != 0xFC {
		t.Errorf("Pixel = %#02x, want %#02x", v.Position.Pixel, 0xFC)
	}
}

// TestVerticalMotion_Step_GravityTrace reproduces the 8-frame gravity
// trace from the design spec: force 0x20, starting speed 0xFC (-4).
// MoveForce (not Dummy) is what accumulates the incoming force, and it
// overflows on frame 8, incrementing the raw speed byte by one whole
// unit, from -4 to -3.
func TestVerticalMotion_Step_GravityTrace(t *testing.T) {
	v := VerticalMotion{Speed: -4}

	wantMoveForce := []Accumulator8{0x20, 0x40, 0x60, 0x80, 0xA0, 0xC0, 0xE0, 0x00}
	wantSpeed := []int8{-4, -4, -4, -4, -4, -4, -4, -3}

	for frame := 0; frame < 8; frame++ {
		v.Step(0x20, 0x7F, 0)
		if v.MoveForce != wantMoveForce[frame] {
			t.Errorf("frame %d: MoveForce = %#02x, want %#02x", frame+1, uint8(v.MoveForce), uint8(wantMoveForce[frame]))
		}
		if v.Speed != wantSpeed[frame] {
			t.Errorf("frame %d: Speed = %d, want %d", frame+1, v.Speed, wantSpeed[frame])
		}
	}
}

// TestVerticalMotion_Step_Clamp verifies the clamp fires only once BOTH
// conditions from SMBDIS.ASM:7727-7735 are true: Speed has reached/passed
// maxSpeed, AND MoveForce's fractional half has also reached 0x80.
// MoveForce=0xFF plus force=0x81 overflows to exactly 0x80 (>=0x80), so
// the clamp fires this frame.
func TestVerticalMotion_Step_Clamp(t *testing.T) {
	v := VerticalMotion{Speed: 4, MoveForce: 0xFF}

	v.Step(0x81, 4, 0)

	if v.Speed != 4 {
		t.Errorf("Speed = %d, want 4 (clamped)", v.Speed)
	}
	if v.MoveForce != 0x00 {
		t.Errorf("MoveForce = %#02x, want %#02x (cleared by clamp)", uint8(v.MoveForce), 0x00)
	}
}

// TestVerticalMotion_Step_ClampSkippedWhenMoveForceBelowHalf is the
// subtle behavior the earlier, oversimplified fix missed: even though
// Speed reaches 5 (past maxSpeed=4) this frame, the clamp does NOT fire
// because MoveForce only overflowed to 0x00 (< 0x80). This is a
// deliberate one-frame rounding tolerance in the ROM, not a bug.
func TestVerticalMotion_Step_ClampSkippedWhenMoveForceBelowHalf(t *testing.T) {
	v := VerticalMotion{Speed: 4, MoveForce: 0xFF}

	v.Step(0x01, 4, 0)

	if v.Speed != 5 {
		t.Errorf("Speed = %d, want 5 (not yet clamped)", v.Speed)
	}
	if v.MoveForce != 0x00 {
		t.Errorf("MoveForce = %#02x, want %#02x", uint8(v.MoveForce), 0x00)
	}
}

func TestVerticalMotion_Step_NoClampWhenUnderLimit(t *testing.T) {
	v := VerticalMotion{Speed: 3, MoveForce: 0xFF}

	v.Step(0x01, 10, 0)

	if v.Speed != 4 {
		t.Errorf("Speed = %d, want 4 (unclamped increment)", v.Speed)
	}
}

// TestVerticalMotion_Step_DummyUsesOldMoveForce is the crux of the bug
// this type was rewritten for: Dummy must accumulate MoveForce's value
// from BEFORE this frame's update (0x50), not after (0x60).
func TestVerticalMotion_Step_DummyUsesOldMoveForce(t *testing.T) {
	v := VerticalMotion{MoveForce: 0x50}

	v.Step(0x10, 0x7F, 0)

	if v.Dummy != 0x50 {
		t.Errorf("Dummy = %#02x, want %#02x (should have used MoveForce's pre-update value)", uint8(v.Dummy), 0x50)
	}
	if v.MoveForce != 0x60 {
		t.Errorf("MoveForce = %#02x, want %#02x (post-update value)", uint8(v.MoveForce), 0x60)
	}
}

// TestVerticalMotion_Step_UpwardForce exercises the optional SBC-based
// deceleration path (SMBDIS.ASM:7736-7758): subtracting more than
// MoveForce currently holds produces a borrow that decrements Speed.
func TestVerticalMotion_Step_UpwardForce(t *testing.T) {
	v := VerticalMotion{Speed: 0, MoveForce: 0x05}

	v.Step(0x00, 0x7F, 0x0A)

	if v.MoveForce != 0xFB {
		t.Errorf("MoveForce = %#02x, want %#02x", uint8(v.MoveForce), 0xFB)
	}
	if v.Speed != -1 {
		t.Errorf("Speed = %d, want -1 (decremented by the borrow)", v.Speed)
	}
}

// TestVerticalMotion_Step_UpwardForceClamp mirrors
// TestVerticalMotion_Step_Clamp for the upward direction: the clamp only
// fires once Speed has passed the negated maxSpeed AND MoveForce's
// fractional half has dropped below 0x80 (SMBDIS.ASM:7750-7758), the
// mirror image of the downward clamp's ">=0x80" condition.
func TestVerticalMotion_Step_UpwardForceClamp(t *testing.T) {
	v := VerticalMotion{Speed: -4, MoveForce: 0x05}

	v.Step(0, 4, 0x90)

	if v.Speed != -4 {
		t.Errorf("Speed = %d, want -4 (clamped to negated maxSpeed)", v.Speed)
	}
	if v.MoveForce != 0xFF {
		t.Errorf("MoveForce = %#02x, want %#02x (cleared by clamp)", uint8(v.MoveForce), 0xFF)
	}
}

// TestVerticalMotion_Step_UpwardForceSkippedWhenZero confirms upForce=0
// leaves Speed and MoveForce exactly as the downward-only steps computed,
// with no side effects from the upward section at all.
func TestVerticalMotion_Step_UpwardForceSkippedWhenZero(t *testing.T) {
	v := VerticalMotion{Speed: 2, MoveForce: 0x10}

	v.Step(0x05, 0x7F, 0)

	if v.Speed != 2 {
		t.Errorf("Speed = %d, want 2 (unaffected by force=0x05 since no overflow occurred)", v.Speed)
	}
	if v.MoveForce != 0x15 {
		t.Errorf("MoveForce = %#02x, want %#02x", uint8(v.MoveForce), 0x15)
	}
}

func TestVerticalMotion_Step_NoPageOverflow(t *testing.T) {
	v := VerticalMotion{
		Position: Position16{Page: 0, Pixel: 0xFF},
		Speed:    1, // raw whole-pixel speed, not split
	}

	delta := v.Step(0, 0x7F, 0)

	if delta != 1 {
		t.Errorf("delta = %d, want 1", delta)
	}
	if v.Position.Pixel != 0x00 {
		t.Errorf("Pixel = %#02x, want %#02x", v.Position.Pixel, 0x00)
	}
	if v.Position.Page != 1 {
		t.Errorf("Page = %d, want 1", v.Position.Page)
	}
	if v.HighPos != 0x00 {
		t.Errorf("HighPos = %#02x, want %#02x (page did not overflow)", v.HighPos, 0x00)
	}
}

func TestVerticalMotion_Step_NegativeSpeedUsesSignExtension(t *testing.T) {
	// Speed is a raw -2 here (0xFE), so the sign-extension byte fed into
	// both Page and HighPos is 0xFF. Starting at Page=0, Pixel=0, moving
	// -2 pixels lands on World()-1 semantics: Page wraps down to 0xFF,
	// Pixel to 0xFE (0xFF00+0xFE = 0xFFFE = -2 mod 65536). Page's
	// resulting carry out is 0 (0x00+0xFF+0 does not overflow), so
	// HighPos - primed at 1 to make the effect observable - absorbs that
	// same 0xFF sign extension and overflows from 1 to 0.
	v := VerticalMotion{
		Position: Position16{Page: 0x00, Pixel: 0x00},
		Speed:    -2,
		HighPos:  0x01,
	}

	v.Step(0, 0x7F, 0)

	if v.Position.Page != 0xFF {
		t.Errorf("Page = %#02x, want %#02x", v.Position.Page, 0xFF)
	}
	if v.Position.Pixel != 0xFE {
		t.Errorf("Pixel = %#02x, want %#02x", v.Position.Pixel, 0xFE)
	}
	if v.HighPos != 0x00 {
		t.Errorf("HighPos = %#02x, want %#02x (overflowed by the negative sign extension)", v.HighPos, 0x00)
	}
}

func TestVerticalMotion_Step_PageOverflowPromotesHighPos(t *testing.T) {
	v := VerticalMotion{
		Position: Position16{Page: 0xFF, Pixel: 0xFF},
		Speed:    1,
	}

	v.Step(0, 0x7F, 0)

	if v.Position.Page != 0x00 {
		t.Errorf("Page = %#02x, want %#02x (wrapped)", v.Position.Page, 0x00)
	}
	if v.HighPos != 0x01 {
		t.Errorf("HighPos = %#02x, want %#02x (promoted by page overflow)", v.HighPos, 0x01)
	}
}
