package strtotime

import (
	"testing"
	"time"
)

func TestMonthOnly(t *testing.T) {
	tests := []struct {
		input string
		month time.Month
	}{
		{"january", time.January},
		{"February", time.February},
		{"MARCH", time.March},
		{"apr", time.April},
		{"may", time.May},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input)
			if err != nil {
				t.Fatalf("Failed to parse %q: %v", test.input, err)
			}
			
			// Check the month
			if result.Month() != test.month {
				t.Errorf("Expected month to be %s, got %s", test.month, result.Month())
			}
			
			// Day should be 1 for month-only inputs
			if result.Day() != 1 {
				t.Errorf("Expected day to be 1, got %d", result.Day())
			}
			
			// Should be current year
			currentYear := time.Now().Year()
			if result.Year() != currentYear {
				t.Errorf("Expected year to be %d, got %d", currentYear, result.Year())
			}
			
			t.Logf("Successfully parsed %q => %s", test.input, result.Format("2006-01-02 15:04:05"))
		})
	}
}