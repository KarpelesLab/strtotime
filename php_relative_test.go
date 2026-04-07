package strtotime

import (
	"testing"
	"time"
)

// Tests from failing PHP ext/date tests — relative date expressions
func TestPHPRelativeDates(t *testing.T) {
	// Use a fixed base time for reproducible tests
	baseTime := time.Date(2005, 12, 12, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		input    string
		base     time.Time
		expected time.Time
	}{
		// bug30532: EDT + relative hours — EDT is a fixed UTC-4 offset
		{"2004-10-31 EDT +1 hour", time.Time{}, time.Date(2004, 10, 31, 1, 0, 0, 0, time.FixedZone("EDT", -4*3600))},
		{"2004-10-31 EDT +2 hours", time.Time{}, time.Date(2004, 10, 31, 2, 0, 0, 0, time.FixedZone("EDT", -4*3600))},
		{"2004-10-31 EDT +3 hours", time.Time{}, time.Date(2004, 10, 31, 3, 0, 0, 0, time.FixedZone("EDT", -4*3600))},

		// bug32086: date + relative days
		{"2004-11-01 +1 day", time.Time{}, time.Date(2004, 11, 2, 0, 0, 0, 0, time.UTC)},

		// bug35630: compact relative expressions
		{"5 january 2006+3day+1day", time.Time{}, time.Date(2006, 1, 9, 0, 0, 0, 0, time.UTC)},

		// bug52808: first/last day of month
		{"first day of 2010-01", time.Time{}, time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"last day of 2010-01", time.Time{}, time.Date(2010, 1, 31, 0, 0, 0, 0, time.UTC)},
		{"first day of 2010-02", time.Time{}, time.Date(2010, 2, 1, 0, 0, 0, 0, time.UTC)},
		{"last day of 2010-02", time.Time{}, time.Date(2010, 2, 28, 0, 0, 0, 0, time.UTC)},

		// relative with base timestamp (bug35414)
		{"26th Nov", baseTime, time.Date(2005, 11, 26, 0, 0, 0, 0, time.UTC)},

		// bug44742: timezone + relative
		{"2008-07-01T22:35:17+0200 +7 days", time.Time{}, time.Date(2008, 7, 8, 22, 35, 17, 0, time.FixedZone("", 7200))},

		// bug67109: last day of relative month
		{"last day of +1 month", baseTime, time.Date(2006, 1, 31, 10, 0, 0, 0, time.UTC)},

		// Ordinal weekday expressions
		{"third Wednesday of January 2025", time.Time{}, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"first Monday of December 2005", time.Time{}, time.Date(2005, 12, 5, 0, 0, 0, 0, time.UTC)},
		{"last Thursday of November 2005", time.Time{}, time.Date(2005, 11, 24, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			opts := []Option{InTZ(time.UTC)}
			if !tt.base.IsZero() {
				opts = append(opts, Rel(tt.base))
			}
			result, err := StrToTime(tt.input, opts...)
			if err != nil {
				t.Fatalf("Failed to parse %q: %v", tt.input, err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("StrToTime(%q) = %v, want %v", tt.input, result, tt.expected)
			}
			phpVerify(t, tt.input, result, tt.base, time.UTC)
		})
	}
}
