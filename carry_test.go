package nesmath

import "testing"

func TestADC_Overflow(t *testing.T) {
	a := uint8(0x90)
	carry := ADC(&a, 0x90, 0)
	if a != 0x20 {
		t.Errorf("a = %#02x, want %#02x", a, 0x20)
	}
	if carry != 1 {
		t.Errorf("carry = %d, want 1", carry)
	}
}

func TestADC_NoOverflow(t *testing.T) {
	a := uint8(0x10)
	carry := ADC(&a, 0x05, 0)
	if a != 0x15 {
		t.Errorf("a = %#02x, want %#02x", a, 0x15)
	}
	if carry != 0 {
		t.Errorf("carry = %d, want 0", carry)
	}
}

func TestADC_IncomingCarryChains(t *testing.T) {
	// 0xFF + 0x00 + carry(1) overflows even though b is zero - this is
	// the mechanism that lets a zero-force accumulator still promote a
	// pending carry into the next byte in a chain.
	a := uint8(0xFF)
	carry := ADC(&a, 0x00, 1)
	if a != 0x00 {
		t.Errorf("a = %#02x, want %#02x", a, 0x00)
	}
	if carry != 1 {
		t.Errorf("carry = %d, want 1", carry)
	}
}

func TestADC_Table(t *testing.T) {
	tests := []struct {
		name      string
		a, b      uint8
		c         Carry
		wantA     uint8
		wantCarry Carry
	}{
		{"zero plus zero", 0x00, 0x00, 0, 0x00, 0},
		{"max plus one", 0xFF, 0x01, 0, 0x00, 1},
		{"max plus max", 0xFF, 0xFF, 1, 0xFF, 1},
		{"mid values no overflow", 0x40, 0x40, 0, 0x80, 0},
		{"mid values with carry in", 0x40, 0x40, 1, 0x81, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.a
			carry := ADC(&a, tt.b, tt.c)
			if a != tt.wantA {
				t.Errorf("a = %#02x, want %#02x", a, tt.wantA)
			}
			if carry != tt.wantCarry {
				t.Errorf("carry = %d, want %d", carry, tt.wantCarry)
			}
		})
	}
}

// TestSBC_MultiByteChain subtracts 0x01C8 from 0x0200 across two bytes
// using the SEC (carry=1, "no borrow yet") convention, verifying borrow
// propagates from the low byte into the high byte exactly as
// ImposeFriction relies on when decelerating across a fractional/integer
// byte pair.
func TestSBC_MultiByteChain(t *testing.T) {
	low := uint8(0x00)
	lowCarry := SBC(&low, 0xC8, 1)
	if low != 0x38 {
		t.Errorf("low = %#02x, want %#02x", low, 0x38)
	}
	if lowCarry != 0 {
		t.Errorf("lowCarry = %d, want 0 (borrow occurred)", lowCarry)
	}

	high := uint8(0x02)
	highCarry := SBC(&high, 0x01, lowCarry)
	if high != 0x00 {
		t.Errorf("high = %#02x, want %#02x", high, 0x00)
	}
	if highCarry != 1 {
		t.Errorf("highCarry = %d, want 1 (no further borrow)", highCarry)
	}

	result := uint16(high)<<8 | uint16(low)
	if result != 0x0038 {
		t.Errorf("result = %#04x, want %#04x", result, 0x0038)
	}
}

func TestSBC_Borrow(t *testing.T) {
	tests := []struct {
		name      string
		a, b      uint8
		c         Carry
		wantA     uint8
		wantCarry Carry
	}{
		{"no borrow needed", 5, 3, 1, 2, 1},
		{"borrow occurs", 3, 5, 1, 254, 0},
		{"incoming borrow consumed", 5, 3, 0, 1, 1},
		{"incoming borrow causes underflow", 0, 0, 0, 255, 0},
		{"equal values, no incoming borrow", 10, 10, 1, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.a
			carry := SBC(&a, tt.b, tt.c)
			if a != tt.wantA {
				t.Errorf("a = %d, want %d", a, tt.wantA)
			}
			if carry != tt.wantCarry {
				t.Errorf("carry = %d, want %d", carry, tt.wantCarry)
			}
		})
	}
}
