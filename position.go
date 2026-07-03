package nesmath

// Position16 is a page-and-pixel coordinate, not a 16-bit integer.
//
// NES platformers commonly track on-screen position as two separate
// bytes: a page (which 256-pixel-wide screen the object is on) and a
// pixel (the position within that screen). Modeling them as a struct
// rather than a single
// uint16 keeps the type honest about the fact that the game reads and
// compares these bytes independently in most of its code - collision
// checks and rendering read Pixel alone, and only camera/scroll logic
// needs to reason about Page as well.
type Position16 struct {
	Page  uint8 // which 256-pixel screen the object is on
	Pixel uint8 // position within that screen
}

// AddSigned adds a signed whole-pixel delta, plus an incoming carry, to
// the position, mutating it in place, and returns the carry out of the
// page byte.
//
// It mirrors the standard 6502 idiom for adding a signed 8-bit delta to a
// 16-bit quantity: add the delta to the low byte with the incoming carry,
// then add a sign-extension byte (0x00 for a non-negative delta, 0xFF for
// a negative one) to the high byte along with whatever carry the low-byte
// addition produced:
//
//	clc                  ; or carry-in from a previous chain stage
//	lda Delta
//	adc Pixel
//	sta Pixel
//	lda #$00             ; or #$ff if Delta was negative
//	adc Page
//	sta Page
//
// This is exactly the pattern typical horizontal- and vertical-movement
// routines use to fold a whole-pixel movement (already extracted from a
// [Q4_4] via [Q4_4.Split]) into the page/pixel pair,
// including the case where whole is negative: the sign-extension byte
// (0xFF, i.e. -1) at the page level exactly cancels a false carry
// produced when a small negative delta does not actually cross a page
// boundary - see the package tests for a worked example.
//
// The returned carry is the carry out of the page byte, which is 1 only
// when the addition would overflow the full 16-bit page:pixel range. Most
// callers building single-screen or single-world games can ignore it; it
// exists for callers that chain Position16 into a still-wider coordinate
// (e.g. a multi-world distance counter).
func (p *Position16) AddSigned(whole int8, carryIn Carry) Carry {
	pixelCarry := ADC(&p.Pixel, uint8(whole), carryIn)

	signExtension := uint8(0x00)
	if whole < 0 {
		signExtension = 0xFF
	}
	return ADC(&p.Page, signExtension, pixelCarry)
}

// World returns the position collapsed into a single 16-bit value,
// Page*256 + Pixel. This is a convenience for distance and ordering
// comparisons across page boundaries; it is not how a typical NES
// platformer itself stores or compares position, which is why it is a
// derived method rather than the underlying representation.
func (p Position16) World() uint16 {
	return uint16(p.Page)*256 + uint16(p.Pixel)
}
