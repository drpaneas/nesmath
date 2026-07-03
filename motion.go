package nesmath

// HorizontalMotion is the two-tier carry chain that turns a [Q4_4] speed
// into pixel movement: Speed splits into a whole-pixel delta and a
// fractional crumb, the crumb accumulates in MoveForce until it
// overflows, and the resulting carry - combined with the whole-pixel
// delta - is folded into Position.
//
// Field names describe roles; the SMBDIS.ASM addresses they replace are
// noted for cross-referencing against the disassembly.
type HorizontalMotion struct {
	Position  Position16
	MoveForce Accumulator8 // SMB: SprObject_X_MoveForce ($0400+x)
	Speed     Q4_4         // SMB: SprObject_X_Speed ($0057+x)
}

// Step executes one frame of horizontal movement and returns the net
// number of whole pixels the object moved this frame.
//
// It mirrors MoveObjectHorizontally (SMBDIS.ASM:7541 onward) exactly as a
// two-stage carry pipeline:
//
//	frac's carry into pixel  := MoveForce.Add(frac)
//	pixel/page update        := Position.AddSigned(whole, thatCarry)
//
// The returned value is not always equal to Speed's whole-pixel part: when
// the sub-pixel accumulator overflows, the carry adds one more pixel of
// movement this frame (see [Q4_4.Split] and the package-level example for
// the resulting 1,2,1,2 alternation at a fractional speed like 0x19).
func (h *HorizontalMotion) Step() int8 {
	whole, frac := h.Speed.Split()
	carry := h.MoveForce.Add(frac)
	h.Position.AddSigned(whole, carry)
	return int8(uint8(whole) + uint8(carry))
}

// VerticalMotion is the carry structure SMB uses for gravity. Unlike
// [HorizontalMotion], vertical speed is never nybble-split - Speed is a
// raw signed whole-pixel value, added directly to Position. Sub-pixel
// precision instead comes from two accumulators that share a byte across
// two different points in its lifecycle each frame:
//
//   - MoveForce accumulates the incoming force every frame; its overflow
//     promotes Speed by one whole unit (not 1/16 - there is no fraction
//     to promote, because Speed was never split).
//   - Dummy accumulates MoveForce's value from *before* that frame's
//     update; its overflow is folded into the Position update alongside
//     Speed's raw byte, producing the extra +1/-1 pixel of movement that
//     smooths motion between the (comparatively rare) frames where Speed
//     itself changes.
//
// Field names describe roles; the SMBDIS.ASM addresses they replace are
// noted for cross-referencing against the disassembly.
type VerticalMotion struct {
	Position  Position16
	MoveForce Accumulator8 // SMB: SprObject_Y_MoveForce ($0433+x)
	Dummy     Accumulator8 // SMB: SprObject_YMF_Dummy ($0416+x)
	Speed     int8         // SMB: SprObject_Y_Speed ($009F+x) - raw signed byte, NOT Q4_4
	HighPos   uint8        // SMB: SprObject_Y_HighPos - off-screen detection
}

// Step executes one frame of vertical movement and returns the net number
// of whole pixels the object moved this frame. It mirrors ImposeGravity
// (SMBDIS.ASM:7704-7759) exactly, including the order of operations,
// which matters: the 6502's carry flag survives silently across the LDY,
// LDA, BPL, and DEY/STY instructions between the Dummy update and the
// Position update, so the carry consumed by Position's addition is the
// one produced by Dummy accumulating MoveForce's OLD (pre-update) value -
// not any split of Speed, which is never split at all for vertical
// motion.
//
// force is the gravity/force value to accumulate into MoveForce this
// frame (e.g. VerticalForce for the player, or a table-driven per-object
// value for enemies - the specific values are game data supplied by the
// caller, not part of this package). maxSpeed caps the downward speed;
// the clamp only fires once MoveForce's fractional half has also reached
// 0x80 (SMBDIS.ASM:7727-7735), a deliberate rounding behavior that avoids
// clamping one frame earlier than the ROM would - not a simplification of
// it.
//
// upForce is the optional simultaneous upward-deceleration force used by
// a few objects that need two-directional gravity in one call (e.g. Red
// Koopa Paratroopa, moving platforms) - never by the player, whose
// ascent/descent force switching is handled by the caller choosing which
// force value to pass, not by this built-in mechanism. Passing upForce=0
// skips this section entirely (SMBDIS.ASM:7736-7758).
func (v *VerticalMotion) Step(force uint8, maxSpeed int8, upForce uint8) int8 {
	carryA := v.Dummy.Add(v.MoveForce.Value())

	delta := v.Speed
	carryB := v.Position.AddSigned(delta, carryA)
	signExtension := uint8(0x00)
	if delta < 0 {
		signExtension = 0xFF
	}
	ADC(&v.HighPos, signExtension, carryB)

	carryC := v.MoveForce.Add(force)
	speedRaw := uint8(v.Speed)
	ADC(&speedRaw, 0, carryC)
	v.Speed = int8(speedRaw)

	if v.Speed >= maxSpeed && v.MoveForce.Value() >= 0x80 {
		v.Speed = maxSpeed
		v.MoveForce = 0
	}

	if upForce != 0 {
		carryD := v.MoveForce.Sub(upForce)
		speedRaw = uint8(v.Speed)
		SBC(&speedRaw, 0, carryD)
		v.Speed = int8(speedRaw)

		negatedMax := int8(Negate(uint8(maxSpeed)))
		if v.Speed < negatedMax && v.MoveForce.Value() < 0x80 {
			v.Speed = negatedMax
			v.MoveForce = 0xFF
		}
	}

	return int8(uint8(delta) + uint8(carryA))
}
