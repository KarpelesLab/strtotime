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

			// Day should be preserved from current date, clamped to max days in target month
			currentDay := time.Now().Day()
			expectedDay := currentDay
			maxDays := daysInMonth(time.Now().Year(), test.month)
			if expectedDay > maxDays {
				expectedDay = maxDays
			}
			if result.Day() != expectedDay {
				t.Errorf("Expected day to be %d, got %d", expectedDay, result.Day())
			}

			// Should be current year
			currentYear := time.Now().Year()
			if result.Year() != currentYear {
				t.Errorf("Expected year to be %d, got %d", currentYear, result.Year())
			}

			t.Logf("Successfully parsed %q => %s", test.input, result.Format("2006-01-02 15:04:05"))
			phpVerify(t, test.input, result, time.Time{}, nil)
		})
	}
}
