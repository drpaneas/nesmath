package nesmath

// Q4_4 is a signed 4.4 fixed-point speed byte: 4 bits of integer, 4 bits
// of fraction, in two's complement. It is the one type in this package
// with a meaningful decimal interpretation - every other type in nesmath
// models a hardware role (an accumulator, a carry, a page/pixel pair)
// rather than a number.
//
// Q4_4 backs SMBDIS.ASM's SprObject_X_Speed and SprObject_Y_Speed. Its
// value is best understood not as "integer + fraction/16" computed
// independently, but as the whole byte reinterpreted as a signed 8-bit
// integer and divided by 16 - [Split] produces the same result either
// way, because that is how the nybble-split algorithm and two's
// complement scaling agree by construction.
//
// Q4_4 intentionally has no +, -, *, or / operations. SMB never adds two
// speed values directly - speed changes come from the friction and
// gravity systems operating on raw bytes through [ADC]/[SBC], not from
// combining two Q4_4 values. Providing arithmetic here would invite code
// that no NES game actually contains.
type Q4_4 uint8

// Split divides s into its whole-pixel and fractional components.
//
// frac is the low nybble of s shifted into the high nybble of a byte,
// i.e. scaled from a /16 fraction to a /256 fraction so it can be added
// directly to an 8-bit sub-pixel accumulator. whole is the high nybble of
// s, sign-extended via [SignExtend4to8].
//
// This mirrors MoveObjectHorizontally (SMBDIS.ASM:7541-7556) exactly:
//
//	lda SprObject_X_Speed,x
//	asl
//	asl
//	asl
//	asl                  ; frac = low nybble << 4
//	sta $01
//
//	lda SprObject_X_Speed,x
//	lsr
//	lsr
//	lsr
//	lsr                  ; whole = high nybble (unsigned, 0-15)
//	cmp #$08
//	bcc SaveXSpd
//	ora #%11110000       ; sign-extend if the high nybble was negative
//	SaveXSpd:
//	sta $00
//
// Note: a byte with a negative integer part and a nonzero fraction, such
// as 0xFC, does not split into "-4 whole, 0 fraction" - it splits into
// -1 whole and a 0.75 fraction, because -1 + 0.75 = -0.25 = int8(0xFC)/16,
// consistent with two's complement scaling. Treating the raw signed byte
// value as though it were already a whole-pixel count (i.e. reading -4
// directly off int8(0xFC) without dividing by 16) is a common but
// incorrect shortcut; Split always performs the real nybble-split
// algorithm.
func (s Q4_4) Split() (whole int8, frac uint8) {
	raw := uint8(s)
	frac = raw << 4
	whole = SignExtend4to8(raw >> 4)
	return whole, frac
}

// IsNegative reports whether s represents a negative speed, i.e. whether
// bit 7 of the raw byte is set. It mirrors the CMP #$00 / BPL idiom used
// throughout SMBDIS.ASM to branch on a speed's direction.
func (s Q4_4) IsNegative() bool {
	return uint8(s)&0x80 != 0
}

// Abs returns the unsigned magnitude of s as a raw byte, computed by
// negating the whole byte via two's complement if it is negative.
//
// It mirrors the Player_XSpeedAbsolute computation (SMBDIS.ASM:6260),
// which SMB uses to index speed-tier tables and compare against
// thresholds like the run-speed and skid-snap constants regardless of
// direction of travel.
func (s Q4_4) Abs() uint8 {
	raw := uint8(s)
	if raw&0x80 != 0 {
		raw = Negate(raw)
	}
	return raw
}

// Negate returns the two's complement negation of s as a Q4_4, reversing
// its direction while preserving its magnitude.
//
// It mirrors the direction-reversal idiom used for bouncing/turning
// objects such as the Fire Bar and Cheep-Cheep handlers (e.g.
// SMBDIS.ASM:8473):
//
//	eor #$ff
//	clc
//	adc #$01
func (s Q4_4) Negate() Q4_4 {
	return Q4_4(Negate(uint8(s)))
}

// Cmp performs a signed comparison between s and other, returning -1 if s
// is less than other, +1 if s is greater, and 0 if they are equal.
//
// Because a Q4_4 byte's value is monotonic with its signed 8-bit integer
// interpretation (dividing by the constant 16 preserves ordering), Cmp
// compares the raw bytes as signed integers using Go's native int8
// comparison operators, which are always correct for any two values.
//
// This deliberately does not attempt to reproduce a literal 6502
// instruction sequence, because there isn't one safe sequence to
// reproduce: on real hardware, CMP followed by BMI/BPL alone is only a
// valid signed less-than/greater-than test when the implicit subtraction
// does not itself signed-overflow (V=0). A true signed comparison needs
// BMI/BPL combined with a check of the V flag (e.g. an EOR of N and V, a
// pattern most 6502 references gloss over), and SMB's own code sidesteps
// this by using BCC/BCS/BEQ only in places where it already knows both
// operands share a narrow, consistent range where the unsigned and
// signed orderings coincide. Cmp's Go implementation has no such
// restriction - it is correct for the full int8 range - which is why it
// is described here as "a signed comparison," not as a stand-in for any
// specific 6502 branch idiom.
func (s Q4_4) Cmp(other Q4_4) int {
	a := int8(uint8(s))
	b := int8(uint8(other))
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
