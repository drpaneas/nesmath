package nesmath

// Accumulator8 is a carry-producing residue bucket, not a number.
//
// It models bytes like SprObject_X_MoveForce, SprObject_Y_MoveForce, and
// SprObject_YMF_Dummy: memory locations that exist only to accumulate a
// fractional amount each frame and occasionally overflow, promoting one
// unit into the next tier of a carry chain (see [ADC]). Reading the
// "value" of an Accumulator8 outside of that overflow is meaningless -
// there is no whole number or fraction it represents on its own, which is
// why this type deliberately has no Split, Integer, or Fraction methods.
//
// A byte that is genuinely a number - a speed - is [Q4_4], not
// Accumulator8. Keeping the two types distinct at the type-system level
// prevents code from accidentally reading an accumulator as if it had a
// decimal meaning, or using a Q4_4 as a place to accumulate carry.
type Accumulator8 uint8

// Add adds value into the accumulator in place and returns the carry
// produced by the overflow, exactly as repeatedly executing:
//
//	clc
//	adc value
//
// against the underlying byte would. Most frames, this simply grows the
// residue and returns carry 0. When the residue overflows past 255, Add
// returns carry 1, signaling that one whole unit should be promoted into
// whatever the next tier of the chain is (a pixel, a speed unit, and so
// on). The caller is responsible for feeding that carry into the next
// [ADC] call - Add does not know what the next tier is.
func (a *Accumulator8) Add(value uint8) Carry {
	raw := uint8(*a)
	carry := ADC(&raw, value, 0)
	*a = Accumulator8(raw)
	return carry
}

// Sub subtracts value from the accumulator in place using the SEC
// convention (a fresh subtraction with no incoming borrow), returning the
// resulting carry per [SBC]'s convention: 1 means no borrow occurred, 0
// means a borrow occurred and should be propagated to the next tier.
//
// This mirrors the 6502 idiom:
//
//	sec
//	sbc value
//
// which ImposeGravity uses for the optional upward-deceleration path
// (SMBDIS.ASM:7743-7746), undoing part of a previous [Add] to produce a
// second, opposite-direction carry chain from the same byte.
func (a *Accumulator8) Sub(value uint8) Carry {
	raw := uint8(*a)
	carry := SBC(&raw, value, 1)
	*a = Accumulator8(raw)
	return carry
}

// Value returns the raw byte held by the accumulator. It exists for save
// state serialization and debugging only - the byte has no standalone
// numeric meaning, so callers should not treat Value's result as a
// quantity to compute with.
func (a Accumulator8) Value() uint8 {
	return uint8(a)
}
