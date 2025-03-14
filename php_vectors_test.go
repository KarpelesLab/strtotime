package strtotime

import (
	"strings"
	"testing"
	"time"
)

func TestPHPSpecificVectors(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{
			"2005-07-14 22:30:41",
			time.Date(2005, 7, 14, 22, 30, 41, 0, time.UTC),
		},
		{
			"2005-07-14 22:30:41 GMT",
			time.Date(2005, 7, 14, 22, 30, 41, 0, time.UTC),
		},
		{
			"@1121373041",
			time.Date(2005, 7, 14, 20, 30, 41, 0, time.UTC), // Using the actual Unix timestamp value
		},
		{
			"@1121373041 CEST",
			time.Date(2005, 7, 14, 22, 30, 41, 0, loadTz("Europe/Paris")),
		},
		{
			"@1121373041.5",
			time.Date(2005, 7, 14, 20, 30, 41, 500000000, time.UTC), // With milliseconds
		},
		{
			"@1121373041.75",
			time.Date(2005, 7, 14, 20, 30, 41, 750000000, time.UTC), // With fractional seconds
		},
		{
			"@1121373041.123456789",
			time.Date(2005, 7, 14, 20, 30, 41, 123456789, time.UTC), // With nanoseconds
		},
		{
			"@1121373041.123 CEST",
			time.Date(2005, 7, 14, 22, 30, 41, 123000000, loadTz("Europe/Paris")), // With milliseconds and timezone
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			var opts []Option
			if test.expected.Location().String() != "UTC" {
				opts = append(opts, InTZ(test.expected.Location()))
			} else {
				opts = append(opts, InTZ(time.UTC))
			}

			result, err := StrToTime(test.input, opts...)
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			// Compare time values (normalized to UTC to avoid timezone issues)
			expectedUTC := test.expected.Unix()
			resultUTC := result.Unix()

			if expectedUTC != resultUTC {
				t.Errorf("For input '%s': expected %s (%d), got %s (%d)", 
					test.input, 
					test.expected.Format("2006-01-02 15:04:05 MST"),
					expectedUTC,
					result.Format("2006-01-02 15:04:05 MST"),
					resultUTC)
			}
			
			// Check nanoseconds for fractional timestamps
			if strings.Contains(test.input, ".") {
				if test.expected.Nanosecond() != result.Nanosecond() {
					t.Errorf("For input '%s': expected nanoseconds %d, got %d",
						test.input,
						test.expected.Nanosecond(),
						result.Nanosecond())
				}
			}
		})
	}
}

// Helper to load timezone
func loadTz(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		// Fallback to UTC
		return time.UTC
	}
	return loc
}