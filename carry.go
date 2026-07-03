package nesmath

// Carry models the 6502 carry flag as an arithmetic value rather than a
// boolean condition. It is always 0 or 1.
//
// The 6502's ADC instruction does not branch on carry - it adds it. A
// chain of ADC instructions threads the carry from one byte into the next
// as part of the addition itself:
//
//	CLC
//	ADC SubPixel   ; carry starts at 0 (CLC)
//	ADC Pixel      ; carry here is whatever ADC above produced
//	ADC Page       ; carry here is whatever the previous ADC produced
//
// [ADC] and [SBC] in this package mirror that: the carry is a value you
// pass in and a value you get back, not a condition you test with an if
// statement.
type Carry uint8

// ADC adds b and c into *a, in place, and returns the resulting carry.
//
// It mirrors the 6502 ADC instruction (without decimal mode, which the
// NES's 2A03 does not implement): *a, b, and c are added as unsigned
// 8-bit quantities, the truncated 8-bit result is stored back into *a,
// and the carry out of bit 7 is returned.
//
// Passing c as 0 corresponds to executing CLC before ADC. Passing c as 1
// (typically the carry returned by a previous ADC or SBC call) corresponds
// to chaining additions across multiple bytes without an intervening CLC,
// exactly as real NES code does for multi-byte position and force
// updates.
//
// ADC does not distinguish signed from unsigned interpretation of a or b,
// matching the 6502: the bit pattern and the carry behavior are identical
// either way. The caller decides how to interpret the result.
func ADC(a *uint8, b uint8, c Carry) Carry {
	sum := uint16(*a) + uint16(b) + uint16(c)
	*a = uint8(sum)
	return Carry(sum >> 8)
}

// SBC subtracts b and the borrow (1-c) from *a, in place, and returns the
// resulting carry, which the 6502 convention treats as an inverted borrow:
// c=1 out means no borrow occurred, c=0 out means a borrow occurred.
//
// This is the single most counterintuitive part of 6502 arithmetic: unlike
// most modern subtract-with-borrow semantics, carry=1 is the "good" state
// (no borrow needed), and it is what SEC (set carry) before a subtraction
// chain establishes. Friction routines and similar code in real NES games
// rely on this convention when subtracting across chained bytes.
//
// SBC is implemented the same way the 6502 hardware implements it
// internally: as an ADC of the bitwise complement of b. This is not a
// simplification - it is the actual mechanism, and it is why SBC and ADC
// share an adder circuit on real 6502 silicon.
func SBC(a *uint8, b uint8, c Carry) Carry {
	return ADC(a, ^b, c)
}
