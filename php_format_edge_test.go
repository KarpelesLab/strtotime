package strtotime

import (
	"testing"
	"time"
)

// Tests from failing PHP ext/date tests — date format edge cases
func TestPHPFormatEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
		skip     string
	}{
		// bug14561: time before date
		{"19:30 Dec 17 2005", time.Date(2005, 12, 17, 19, 30, 0, 0, time.UTC), ""},
		{"Dec 17 19:30 2005", time.Date(2005, 12, 17, 19, 30, 0, 0, time.UTC), ""},

		// bug26198: month + year without day
		{"Oct 2001", time.Date(2001, 10, 1, 0, 0, 0, 0, time.UTC), ""},
		{"2001 Oct", time.Date(2001, 10, 1, 0, 0, 0, 0, time.UTC), ""},

		// bug28024: time before date with timezone
		{"17:00 2004-01-01", time.Date(2004, 1, 1, 17, 0, 0, 0, time.UTC), ""},

		// bug29150: roman numeral month
		{"20 VI. 2005", time.Date(2005, 6, 20, 0, 0, 0, 0, time.UTC), ""},

		// bug33578: various month/day combos
		{"11 Oct 2005", time.Date(2005, 10, 11, 0, 0, 0, 0, time.UTC), ""},
		{"11Oct2005", time.Date(2005, 10, 11, 0, 0, 0, 0, time.UTC), ""},
		{"11Oct 2005", time.Date(2005, 10, 11, 0, 0, 0, 0, time.UTC), ""},

		// bug35414: ordinal dates
		{"Sat 26th Nov 2005 18:18", time.Date(2005, 11, 26, 18, 18, 0, 0, time.UTC), ""},
		{"Dec. 4th, 2005", time.Date(2005, 12, 4, 0, 0, 0, 0, time.UTC), ""},
		{"December 4th, 2005", time.Date(2005, 12, 4, 0, 0, 0, 0, time.UTC), ""},

		// bug35499: AM/PM handling
		{"11/20/2005 8:00 AM", time.Date(2005, 11, 20, 8, 0, 0, 0, time.UTC), ""},

		// bug35887: ISO-8601 with single-digit components
		{"2006-1-6T0:0:0-8:0", time.Date(2006, 1, 6, 0, 0, 0, 0, time.FixedZone("", -8*3600)), ""},

		// bug37017: timezone name in date string
		{"2006-05-12 13:00:01 America/New_York",
			time.Date(2006, 5, 12, 13, 0, 1, 0, loadTz("America/New_York")), ""},

		// bug38229: year-month only
		{"2006-1", time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC), ""},
		{"2006-03", time.Date(2006, 3, 1, 0, 0, 0, 0, time.UTC), ""},
		{"2006-12", time.Date(2006, 12, 1, 0, 0, 0, 0, time.UTC), ""},

		// bug41523: zero date
		{"0000-00-00 00:00:00", time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC), ""},

		// bug45081: 12-hour time edge cases
		{"11-MAY-1988 12:00:00AM", time.Date(1988, 5, 11, 0, 0, 0, 0, time.UTC), ""},
		{"11-MAY-1988 12:00:00PM", time.Date(1988, 5, 11, 12, 0, 0, 0, time.UTC), ""},

		// bug45599: microseconds
		{"2008-07-01T22:35:17.02", time.Date(2008, 7, 1, 22, 35, 17, 20000000, time.UTC), ""},
		{"2008-07-01T22:35:17.03+02:00",
			time.Date(2008, 7, 1, 22, 35, 17, 30000000, time.FixedZone("", 7200)), ""},

		// bug48678: negative year
		{"-0001-06-28", time.Date(-1, 6, 28, 0, 0, 0, 0, time.UTC), ""},

		// bug53437: whitespace in format
		{"  2010-10-06   12:53:10  ", time.Date(2010, 10, 6, 12, 53, 10, 0, time.UTC), ""},

		// rfc-datetime formats
		{"Thu, 20 Nov 2003 16:20:42 +0000",
			time.Date(2003, 11, 20, 16, 20, 42, 0, time.UTC), ""},
		{"Thu, 20 Nov 2003 16:20:42 GMT",
			time.Date(2003, 11, 20, 16, 20, 42, 0, time.UTC), ""},
		{"20 Nov 2003 16:20:42 +0000",
			time.Date(2003, 11, 20, 16, 20, 42, 0, time.UTC), ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if tt.skip != "" {
				t.Skip(tt.skip)
			}
			result, err := StrToTime(tt.input, InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Failed to parse %q: %v", tt.input, err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("StrToTime(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
