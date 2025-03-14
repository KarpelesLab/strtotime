package strtotime

import (
	"testing"
	"time"
)

func TestAprilFourth(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"april 4th"},
		{"April 4th"},
		{"April 4"},
		{"apr 4th"},
		{"apr 4"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input)
			if err != nil {
				t.Fatalf("Failed to parse %q: %v", test.input, err)
			}

			// The year should be the current year
			currentYear := time.Now().Year()

			// Month should be April (4)
			if result.Month() != time.April {
				t.Errorf("Expected month to be April, got %s", result.Month())
			}

			// Day should be 4
			if result.Day() != 4 {
				t.Errorf("Expected day to be 4, got %d", result.Day())
			}

			// Should be current year
			if result.Year() != currentYear {
				t.Errorf("Expected year to be %d, got %d", currentYear, result.Year())
			}

			t.Logf("Successfully parsed %q => %s", test.input, result.Format("2006-01-02 15:04:05"))
		})
	}
}
