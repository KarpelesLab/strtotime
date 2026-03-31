package strtotime

import (
	"testing"
	"time"
)

// Regression tests from specific PHP bug reports that involve strtotime parsing
func TestPHPBugRegressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		base     time.Time
		expected time.Time
		skip     string
	}{
		// bug20382: date formats with day-of-week
		{"bug20382-1", "Tue, 27 May 2003 8:15:25 -0700", time.Time{},
			time.Date(2003, 5, 27, 8, 15, 25, 0, time.FixedZone("", -7*3600)), ""},

		// bug27780: +1 month boundary — CST is fixed UTC-6
		{"bug27780-base", "2004-02-15 00:00:00 CST", time.Time{},
			time.Date(2004, 2, 15, 0, 0, 0, 0, time.FixedZone("CST", -6*3600)), ""},

		// bug32086: date arithmetic around DST
		{"bug32086-nov1", "2004-11-01", time.Date(2004, 11, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2004, 11, 1, 0, 0, 0, 0, time.UTC), ""},
		{"bug32086-nov1+1day", "2004-11-01 +1 day", time.Date(2004, 11, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2004, 11, 2, 0, 0, 0, 0, time.UTC), ""},

		// bug32555: tomorrow with timezone
		{"bug32555-tomorrow", "tomorrow", time.Date(2005, 3, 4, 12, 0, 0, 0, time.UTC),
			time.Date(2005, 3, 5, 0, 0, 0, 0, time.UTC), ""},

		// bug40861: midnight handling
		{"bug40861", "2000-01-01 12:00:00", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), ""},

		// bug44742: timezone in ISO format + relative
		{"bug44742", "2008-07-01T22:35:17+0200 +7 days", time.Time{},
			time.Date(2008, 7, 8, 22, 35, 17, 0, time.FixedZone("", 7200)), ""},

		// bug45599: fractional seconds
		{"bug45599-frac", "2008-07-01T22:35:17.02", time.Date(2008, 7, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2008, 7, 1, 22, 35, 17, 20000000, time.UTC), ""},

		// bug46108: negative timezone offset
		{"bug46108", "2008-11-20 12:00:00-0500", time.Time{},
			time.Date(2008, 11, 20, 12, 0, 0, 0, time.FixedZone("", -5*3600)), ""},

		// bug48678: negative year
		{"bug48678", "-0001-06-28", time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(-1, 6, 28, 0, 0, 0, 0, time.UTC), ""},

		// bug52290: 24:00:00 as midnight next day
		{"bug52290", "2010-03-06T24:00:00", time.Date(2010, 3, 6, 0, 0, 0, 0, time.UTC),
			time.Date(2010, 3, 7, 0, 0, 0, 0, time.UTC), ""},

		// bug52808: first/last day of
		{"bug52808-first", "first day of 2010-01", time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC), ""},
		{"bug52808-last", "last day of 2010-01", time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2010, 1, 31, 0, 0, 0, 0, time.UTC), ""},
		{"bug52808-last-feb", "last day of 2010-02", time.Date(2010, 2, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2010, 2, 28, 0, 0, 0, 0, time.UTC), ""},
		{"bug52808-last-feb-leap", "last day of 2012-02", time.Date(2012, 2, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2012, 2, 29, 0, 0, 0, 0, time.UTC), ""},

		// bug53437: extra whitespace
		{"bug53437", "  2010-10-06   12:53:10  ", time.Date(2010, 10, 6, 0, 0, 0, 0, time.UTC),
			time.Date(2010, 10, 6, 12, 53, 10, 0, time.UTC), ""},

		// bug55397: front-of/back-of (Scottish time expressions)
		{"bug55397-front", "front of 12pm", time.Date(2005, 12, 12, 10, 0, 0, 0, time.UTC),
			time.Date(2005, 12, 12, 11, 45, 0, 0, time.UTC), ""},
		{"bug55397-back", "back of 12pm", time.Date(2005, 12, 12, 10, 0, 0, 0, time.UTC),
			time.Date(2005, 12, 12, 12, 15, 0, 0, time.UTC), ""},

		// bug62896: 'of' keyword
		{"bug62896", "first Monday of January 2013", time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2013, 1, 7, 0, 0, 0, 0, time.UTC), ""},

		// bug66721: microsecond parsing
		{"bug66721", "2014-01-01 00:00:00.123456", time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2014, 1, 1, 0, 0, 0, 123456000, time.UTC), ""},

		// gh9763: this/next with day-of-week
		// "this wednesday" from Monday = this week's wednesday (2 days later)
		{"gh9763-this-wed", "this wednesday", time.Date(2022, 9, 5, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 9, 7, 0, 0, 0, 0, time.UTC), ""},
		// "next wednesday" from Monday = the upcoming wednesday (2 days later)
		{"gh9763-next-wed", "next wednesday", time.Date(2022, 9, 5, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 9, 7, 0, 0, 0, 0, time.UTC), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip != "" {
				t.Skip(tt.skip)
			}
			var opts []Option
			if !tt.base.IsZero() {
				opts = append(opts, Rel(tt.base))
			}
			result, err := StrToTime(tt.input, opts...)
			if err != nil {
				t.Fatalf("Failed to parse %q: %v", tt.input, err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("StrToTime(%q) = %v (unix=%d), want %v (unix=%d)",
					tt.input, result, result.Unix(), tt.expected, tt.expected.Unix())
			}
		})
	}
}
