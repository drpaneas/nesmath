package nesmath

import "testing"

func TestPosition16_AddSigned(t *testing.T) {
	tests := []struct {
		name      string
		start     Position16
		whole     int8
		carryIn   Carry
		want      Position16
		wantCarry Carry
	}{
		{
			name:      "positive move, no pixel overflow",
			start:     Position16{Page: 2, Pixel: 0x40},
			whole:     1,
			carryIn:   0,
			want:      Position16{Page: 2, Pixel: 0x41},
			wantCarry: 0,
		},
		{
			name:      "positive move with incoming carry from sub-pixel",
			start:     Position16{Page: 2, Pixel: 0x41},
			whole:     1,
			carryIn:   1,
			want:      Position16{Page: 2, Pixel: 0x43},
			wantCarry: 0,
		},
		{
			name:      "positive move crosses page boundary",
			start:     Position16{Page: 2, Pixel: 0xFF},
			whole:     1,
			carryIn:   0,
			want:      Position16{Page: 3, Pixel: 0x00},
			wantCarry: 0,
		},
		{
			name:      "negative move crosses page boundary downward",
			start:     Position16{Page: 2, Pixel: 0x00},
			whole:     -1,
			carryIn:   0,
			want:      Position16{Page: 1, Pixel: 0xFF},
			wantCarry: 1,
		},
		{
			name:      "negative move within the same page does not change page",
			start:     Position16{Page: 2, Pixel: 0x05},
			whole:     -1,
			carryIn:   0,
			want:      Position16{Page: 2, Pixel: 0x04},
			wantCarry: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := tt.start
			carry := pos.AddSigned(tt.whole, tt.carryIn)
			if pos != tt.want {
				t.Errorf("position = %+v, want %+v", pos, tt.want)
			}
			if carry != tt.wantCarry {
				t.Errorf("carry = %d, want %d", carry, tt.wantCarry)
			}
		})
	}
}

func TestPosition16_World(t *testing.T) {
	tests := []struct {
		name string
		pos  Position16
		want uint16
	}{
		{"all zero", Position16{Page: 0, Pixel: 0}, 0x0000},
		{"page zero", Position16{Page: 0, Pixel: 0x40}, 0x0040},
		{"page two", Position16{Page: 2, Pixel: 0x40}, 0x0240},
		{"max values", Position16{Page: 0xFF, Pixel: 0xFF}, 0xFFFF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.World(); got != tt.want {
				t.Errorf("World() = %#04x, want %#04x", got, tt.want)
			}
		})
	}
}
