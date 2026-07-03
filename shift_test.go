package nesmath

import "testing"

func TestSignExtend4to8(t *testing.T) {
	tests := []struct {
		name   string
		nybble uint8
		want   int8
	}{
		{"zero", 0x0, 0},
		{"positive max in range", 0x7, 7},
		{"boundary negative", 0x8, -8},
		{"mid negative", 0xE, -2},
		{"all ones", 0xF, -1},
		{"upper bits ignored", 0xFE, -2},
		{"upper bits ignored, positive", 0xF3, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SignExtend4to8(tt.nybble)
			if got != tt.want {
				t.Errorf("SignExtend4to8(%#02x) = %d, want %d", tt.nybble, got, tt.want)
			}
		})
	}
}

func TestNegate(t *testing.T) {
	tests := []struct {
		name string
		v    uint8
		want uint8
	}{
		{"zero", 0x00, 0x00},
		{"one", 0x01, 0xFF},
		{"max byte", 0xFF, 0x01},
		{"minimum signed value wraps to itself", 0x80, 0x80},
		{"positive four", 0x04, 0xFC},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Negate(tt.v)
			if got != tt.want {
				t.Errorf("Negate(%#02x) = %#02x, want %#02x", tt.v, got, tt.want)
			}
		})
	}
}

// TestASL16_FrictionDoubling reproduces the exact friction-doubling
// values SMB uses during a skid reversal: FrictionData entries are
// doubled via ASL/ROL before being applied, and these two results
// (0x01A0 for walking, 0x01C8 for running) are the specific 16-bit values
// cited in the wiki's friction-doubling section.
func TestASL16_FrictionDoubling(t *testing.T) {
	tests := []struct {
		name         string
		low, high    uint8
		wantL, wantH uint8
	}{
		{"walk friction 0xD0 doubles to 0x01A0", 0xD0, 0x00, 0xA0, 0x01},
		{"run friction 0xE4 doubles to 0x01C8", 0xE4, 0x00, 0xC8, 0x01},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			low, high := tt.low, tt.high
			ASL16(&low, &high)
			if low != tt.wantL || high != tt.wantH {
				t.Errorf("ASL16(%#02x, %#02x) = (%#02x, %#02x), want (%#02x, %#02x)",
					tt.low, tt.high, low, high, tt.wantL, tt.wantH)
			}
		})
	}
}

func TestASL16(t *testing.T) {
	tests := []struct {
		name         string
		low, high    uint8
		wantL, wantH uint8
	}{
		{"no carry propagation", 0x01, 0x00, 0x02, 0x00},
		{"low overflow carries into high", 0x80, 0x00, 0x00, 0x01},
		{"high overflow discarded", 0x00, 0x80, 0x00, 0x00},
		{"both bytes shift with carry chain", 0xFF, 0x7F, 0xFE, 0xFF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			low, high := tt.low, tt.high
			ASL16(&low, &high)
			if low != tt.wantL {
				t.Errorf("low = %#02x, want %#02x", low, tt.wantL)
			}
			if high != tt.wantH {
				t.Errorf("high = %#02x, want %#02x", high, tt.wantH)
			}
		})
	}
}
