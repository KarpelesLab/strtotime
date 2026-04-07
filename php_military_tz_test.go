package strtotime

import (
	"testing"
	"time"
)

// Tests from PHP ext/date bug26317.phpt — military timezone single-letter codes
// PHP supports single-letter military timezones: A=+1, B=+2, ..., M=+12, N=-1, ..., Y=-12, Z=UTC
func TestPHPMilitaryTimezones(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		// bug26317: "2003-11-19 16:20:42 Z" should parse Z as UTC
		{"2003-11-19 16:20:42 Z", time.Date(2003, 11, 19, 16, 20, 42, 0, time.UTC)},
		// bug26317: "2003-11-19 09:20:42 T" — T = UTC-7
		{"2003-11-19 09:20:42 T", time.Date(2003, 11, 19, 9, 20, 42, 0, time.FixedZone("T", -7*3600))},
		// bug26317: "2003-11-19 19:20:42 C" — C = UTC+3
		{"2003-11-19 19:20:42 C", time.Date(2003, 11, 19, 19, 20, 42, 0, time.FixedZone("C", 3*3600))},
		// A = UTC+1
		{"2003-11-19 17:20:42 A", time.Date(2003, 11, 19, 17, 20, 42, 0, time.FixedZone("A", 1*3600))},
		// M = UTC+12
		{"2003-11-19 04:20:42 M", time.Date(2003, 11, 19, 4, 20, 42, 0, time.FixedZone("M", 12*3600))},
		// N = UTC-1
		{"2003-11-19 15:20:42 N", time.Date(2003, 11, 19, 15, 20, 42, 0, time.FixedZone("N", -1*3600))},
		// Y = UTC-12
		{"2003-11-19 04:20:42 Y", time.Date(2003, 11, 19, 4, 20, 42, 0, time.FixedZone("Y", -12*3600))},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := StrToTime(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse %q: %v", tt.input, err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("StrToTime(%q) = %v, want %v", tt.input, result, tt.expected)
			}
			phpVerify(t, tt.input, result, time.Time{}, time.UTC)
		})
	}
}
