package nesmath

import "testing"

func TestAccumulator8_Add(t *testing.T) {
	tests := []struct {
		name      string
		start     Accumulator8
		add       uint8
		wantValue Accumulator8
		wantCarry Carry
	}{
		{"grows without overflow", 0x00, 0x90, 0x90, 0},
		{"repeated add still no overflow", 0x90, 0x20, 0xB0, 0},
		{"overflow produces carry", 0xF0, 0x20, 0x10, 1},
		{"exact fill to 256 overflows", 0x80, 0x80, 0x00, 1},
		{"adding zero never overflows", 0xFF, 0x00, 0xFF, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := tt.start
			carry := acc.Add(tt.add)
			if acc != tt.wantValue {
				t.Errorf("accumulator = %#02x, want %#02x", uint8(acc), uint8(tt.wantValue))
			}
			if carry != tt.wantCarry {
				t.Errorf("carry = %d, want %d", carry, tt.wantCarry)
			}
		})
	}
}

// TestAccumulator8_GravityTrace reproduces the 8-frame gravity force trace
// from the design spec: a force of 0x20 added each frame overflows on the
// 8th frame, which is the mechanism that lets a force too small to move a
// speed byte on any single frame still produce smooth motion over time.
func TestAccumulator8_GravityTrace(t *testing.T) {
	var force Accumulator8
	const gravityForce = 0x20

	wantValues := []Accumulator8{0x20, 0x40, 0x60, 0x80, 0xA0, 0xC0, 0xE0, 0x00}
	wantCarries := []Carry{0, 0, 0, 0, 0, 0, 0, 1}

	for frame := 0; frame < 8; frame++ {
		carry := force.Add(gravityForce)
		if force != wantValues[frame] {
			t.Errorf("frame %d: force = %#02x, want %#02x", frame+1, uint8(force), uint8(wantValues[frame]))
		}
		if carry != wantCarries[frame] {
			t.Errorf("frame %d: carry = %d, want %d", frame+1, carry, wantCarries[frame])
		}
	}
}

// TestAccumulator8_Add_StandalonePattern exercises the accumulator alone,
// without any [HorizontalMotion] wrapper, adding the run-speed fraction
// (0x90) repeatedly and confirming carries land on every other frame
// (frames 2, 4, 6, 8). This is the exact mechanism behind the 1,2,1,2
// pixel oscillation, isolated from motion/position concerns.
func TestAccumulator8_Add_StandalonePattern(t *testing.T) {
	var acc Accumulator8

	wantValues := []Accumulator8{0x90, 0x20, 0xB0, 0x40, 0xD0, 0x60, 0xF0, 0x80}
	wantCarries := []Carry{0, 1, 0, 1, 0, 1, 0, 1}

	for frame := 0; frame < 8; frame++ {
		carry := acc.Add(0x90)
		if acc != wantValues[frame] {
			t.Errorf("frame %d: acc = %#02x, want %#02x", frame+1, uint8(acc), uint8(wantValues[frame]))
		}
		if carry != wantCarries[frame] {
			t.Errorf("frame %d: carry = %d, want %d", frame+1, carry, wantCarries[frame])
		}
	}
}

func TestAccumulator8_Sub(t *testing.T) {
	tests := []struct {
		name      string
		start     Accumulator8
		sub       uint8
		wantValue Accumulator8
		wantCarry Carry
	}{
		{"no borrow needed", 0x05, 0x03, 0x02, 1},
		{"borrow occurs", 0x03, 0x05, 0xFE, 0},
		{"equal values, no borrow", 0x10, 0x10, 0x00, 1},
		{"subtracting zero never borrows", 0x00, 0x00, 0x00, 1},
		{"undoes a previous Add", 0xE0, 0x0A, 0xD6, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := tt.start
			carry := acc.Sub(tt.sub)
			if acc != tt.wantValue {
				t.Errorf("accumulator = %#02x, want %#02x", uint8(acc), uint8(tt.wantValue))
			}
			if carry != tt.wantCarry {
				t.Errorf("carry = %d, want %d", carry, tt.wantCarry)
			}
		})
	}
}

func TestAccumulator8_Value(t *testing.T) {
	acc := Accumulator8(0x42)
	if got := acc.Value(); got != 0x42 {
		t.Errorf("Value() = %#02x, want %#02x", got, 0x42)
	}
}
