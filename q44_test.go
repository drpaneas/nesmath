package nesmath

import "testing"

// TestQ4_4_Split covers the full matrix of named horizontal speed values
// from the design spec. All of these are horizontal speed values, split
// by a typical horizontal-movement routine exactly as Split implements.
//
// Note on the $FC case: some secondary write-ups describe a game's
// initial standing-jump velocity ($FC) as "-4.0 pixels/frame" by reading
// int8(0xFC) directly as a whole-pixel count. That skips the /16 scaling
// that Q4_4 always applies. The correct nybble split of 0xFC (1111 1100)
// is whole=-1 (high nybble 0xF sign-extended), frac=0xC0 (low nybble 0xC
// shifted into the high byte position) - equivalently, -1 + 0.75 = -0.25,
// which matches int8(0xFC)/16 = -4/16 = -0.25 exactly.
//
// This $FC value is included here only as a Split() math exercise, not as
// a claim about real game behavior: $FC is actually a raw vertical jump
// speed, and vertical speed is never nybble-split at all (see
// [VerticalMotion] and NES_MATH_FOR_PORTERS.md's "Vertical speed is not
// 4.4" section) - a gravity routine adds it to YPosition as a raw signed
// whole byte, so the real in-game value of $FC is -4.0, not -0.25. For
// the same reason, other swimming-jump entries ($FE, $FB, ...) are
// intentionally not tested here: applying Split() to them would imply
// they are meant to be nybble-split, which they are not. The common
// ground-enemy entry below is genuine, though - enemy horizontal speed
// uses the same Q4_4 format as the player's.
func TestQ4_4_Split(t *testing.T) {
	tests := []struct {
		name      string
		speed     Q4_4
		wantWhole int8
		wantFrac  uint8
	}{
		{"walk speed 1.0", 0x10, 1, 0x00},
		{"run speed observed max 1.5625", 0x19, 1, 0x90},
		{"run speed table max 2.5", 0x28, 2, 0x80},
		{"negative speed with fraction", 0xE4, -2, 0x40},
		{"zero", 0x00, 0, 0x00},
		{"maximum negative -8.0", 0x80, -8, 0x00},
		{"maximum positive 7.9375", 0x7F, 7, 0xF0},
		{"0xFC as a Split() math exercise only, not real vertical behavior", 0xFC, -1, 0xC0},
		{"common ground-enemy speed 0.75 (genuine horizontal Q4_4)", 0x0C, 0, 0xC0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whole, frac := tt.speed.Split()
			if whole != tt.wantWhole {
				t.Errorf("whole = %d, want %d", whole, tt.wantWhole)
			}
			if frac != tt.wantFrac {
				t.Errorf("frac = %#02x, want %#02x", frac, tt.wantFrac)
			}
		})
	}
}

func TestQ4_4_IsNegative(t *testing.T) {
	tests := []struct {
		name  string
		speed Q4_4
		want  bool
	}{
		{"positive walk speed", 0x10, false},
		{"zero", 0x00, false},
		{"negative speed", 0xE4, true},
		{"maximum negative", 0x80, true},
		{"maximum positive", 0x7F, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.speed.IsNegative(); got != tt.want {
				t.Errorf("IsNegative() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQ4_4_Abs(t *testing.T) {
	tests := []struct {
		name  string
		speed Q4_4
		want  uint8
	}{
		{"positive max run speed +2.5", 0x28, 0x28},
		{"negative max run speed -2.5", 0xD8, 0x28},
		{"zero", 0x00, 0x00},
		{"maximum negative -8.0", 0x80, 0x80}, // wraps: two's complement negation of 0x80 is itself
		{"positive walk speed", 0x10, 0x10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.speed.Abs(); got != tt.want {
				t.Errorf("Abs() = %#02x, want %#02x", got, tt.want)
			}
		})
	}
}

func TestQ4_4_Negate(t *testing.T) {
	tests := []struct {
		name  string
		speed Q4_4
		want  Q4_4
	}{
		{"positive to negative", 0x28, 0xD8},
		{"negative to positive", 0xD8, 0x28},
		{"zero stays zero", 0x00, 0x00},
		// Two's complement edge case: -8.0 has no positive counterpart in
		// signed 4.4 (the range is -8.0 to +7.9375), so negating it wraps
		// back to itself rather than overflowing. Guards against a
		// regression to a naive "flip sign bit" implementation, which
		// would not reproduce this wraparound.
		{"maximum negative wraps to itself", 0x80, 0x80},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.speed.Negate(); got != tt.want {
				t.Errorf("Negate() = %#02x, want %#02x", uint8(got), uint8(tt.want))
			}
		})
	}
}

// TestQ4_4_Cmp includes two cases (0x80 vs 0x7F, both directions) that
// specifically guard against the 6502 signed-comparison trap documented
// in NES_MATH_FOR_PORTERS.md's "The comparison problem": comparing these
// two raw bytes via CMP + BMI/BPL alone, without also checking the
// overflow flag, gives the WRONG answer, because the subtraction
// 0x7F - 0x80 itself overflows the signed range. If Cmp were ever
// reimplemented using a naive "check the sign bit of a-b" bit trick
// instead of Go's native (and always-correct) int8 comparison operators,
// these two cases would fail and catch the regression immediately.
func TestQ4_4_Cmp(t *testing.T) {
	tests := []struct {
		name string
		a, b Q4_4
		want int
	}{
		{"equal", 0x10, 0x10, 0},
		{"positive greater than positive", 0x28, 0x10, 1},
		{"positive less than positive", 0x10, 0x28, -1},
		{"positive greater than negative", 0x10, 0xE4, 1},
		{"negative less than positive", 0xE4, 0x10, -1},
		{
			name: "signed-overflow trap: -128 vs +127, naive sign-bit-of-(a-b) would say a>b, correct answer is a<b",
			a:    0x80, b: 0x7F, want: -1,
		},
		{
			name: "signed-overflow trap mirrored: +127 vs -128, naive sign-bit-of-(a-b) would say a<b, correct answer is a>b",
			a:    0x7F, b: 0x80, want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Cmp(tt.b); got != tt.want {
				t.Errorf("Cmp() = %d, want %d", got, tt.want)
			}
		})
	}
}
