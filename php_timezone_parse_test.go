package strtotime

import (
	"testing"
	"time"
)

// Tests from failing PHP ext/date tests — timezone parsing in date strings
func TestPHPTimezoneParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		// date_create-1: parsing with explicit timezone offset should preserve time
		{"2005-07-14T22:30:41-0700", time.Date(2005, 7, 14, 22, 30, 41, 0, time.FixedZone("", -7*3600))},
		{"2005-07-14T22:30:41+0200", time.Date(2005, 7, 14, 22, 30, 41, 0, time.FixedZone("", 2*3600))},
		{"2005-07-14T22:30:41+0000", time.Date(2005, 7, 14, 22, 30, 41, 0, time.UTC)},
		{"2005-07-14T22:30:41+0100", time.Date(2005, 7, 14, 22, 30, 41, 0, time.FixedZone("", 3600))},

		// Various timezone name formats — abbreviations are treated as fixed offsets
		{"2005-07-14 22:30:41 EST", time.Date(2005, 7, 14, 22, 30, 41, 0, time.FixedZone("EST", -5*3600))},
		{"2005-07-14 22:30:41 UTC", time.Date(2005, 7, 14, 22, 30, 41, 0, time.UTC)},
		{"2005-07-14 22:30:41 GMT", time.Date(2005, 7, 14, 22, 30, 41, 0, time.UTC)},

		// bug45554: timezone abbreviation with offset
		{"2008-01-01 00:00:00 CET", time.Date(2008, 1, 1, 0, 0, 0, 0, time.FixedZone("CET", 1*3600))},
		{"2008-06-01 00:00:00 CEST", time.Date(2008, 6, 1, 0, 0, 0, 0, time.FixedZone("CEST", 2*3600))},

		// bug67109: timezone handling with relative
		{"2014-01-01 Asia/Tokyo", time.Date(2014, 1, 1, 0, 0, 0, 0, loadTz("Asia/Tokyo"))},

		// Colon in timezone offset
		{"2008-07-01T22:35:17+02:00", time.Date(2008, 7, 1, 22, 35, 17, 0, time.FixedZone("", 7200))},
		{"2008-07-01T22:35:17-05:30", time.Date(2008, 7, 1, 22, 35, 17, 0, time.FixedZone("", -5*3600-30*60))},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := StrToTime(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse %q: %v", tt.input, err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("StrToTime(%q) = %v (unix=%d), want %v (unix=%d)",
					tt.input, result, result.Unix(), tt.expected, tt.expected.Unix())
			}
			phpVerify(t, tt.input, result, time.Time{}, time.UTC)
		})
	}
}
