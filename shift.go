package nesmath

// SignExtend4to8 converts a 4-bit signed value stored in the low nybble of
// a byte into a proper signed int8, replicating the sign bit through the
// upper nybble.
//
// Only the low 4 bits of nybble are considered; any bits above them are
// discarded before the sign check, matching the state of a byte that has
// just been shifted right 4 times (LSR ×4), which always leaves zeros in
// the upper nybble.
//
// It mirrors the sign-extension idiom in MoveObjectHorizontally
// (SMBDIS.ASM:7548-7551):
//
//	cmp #$08
//	bcc SaveXSpd
//	ora #%11110000
//
// If the 4-bit value is 8 or higher, bit 3 was set in the original signed
// nybble, meaning the value is negative; the upper 4 bits are filled with
// 1s to extend the sign into a full 8-bit representation.
func SignExtend4to8(nybble uint8) int8 {
	nybble &= 0x0F
	if nybble >= 0x08 {
		nybble |= 0xF0
	}
	return int8(nybble)
}

// Negate returns the two's complement negation of v.
//
// It mirrors the standard 6502 negation idiom used throughout SMBDIS.ASM
// for absolute value and direction reversal (e.g. line 8473):
//
//	eor #$ff
//	clc
//	adc #$01
//
// Negating 0 returns 0, and negating the minimum representable signed
// value (0x80, i.e. -128) returns itself (0x80), exactly as two's
// complement negation does on real 6502 hardware - there is no separate
// over/underflow bit for this case, it simply wraps.
func Negate(v uint8) uint8 {
	return ^v + 1
}

// ASL16 shifts the 16-bit value formed by high:low left by one bit, in
// place. The bit shifted out of low's bit 7 becomes high's bit 0; the bit
// shifted out of high's bit 7 is discarded.
//
// It mirrors the ASL/ROL pair the 6502 uses to shift a two-byte quantity,
// such as the friction-doubling code during a skid:
//
//	asl FrictionAdderLow
//	rol FrictionAdderHigh
//
// ASL on the low byte multiplies it by two and sets the carry flag to its
// former bit 7. ROL on the high byte then shifts it left and rotates that
// carry into its bit 0. This function performs both steps as a single
// 16-bit operation; the carry out of the high byte (equivalent to the
// 6502's carry flag after the ROL) is discarded, matching call sites that
// do not chain the result further.
func ASL16(low, high *uint8) {
	carryOut := (*low & 0x80) >> 7
	*low <<= 1
	*high = (*high << 1) | carryOut
}
