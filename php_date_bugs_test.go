package strtotime

import (
	"testing"
	"time"
)

// Additional test vectors from failing PHP ext/date bug tests.
// These inputs must parse successfully and produce the correct time.
func TestPHPDateBugs(t *testing.T) {
	utcRef := time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    string
		base     time.Time
		expected time.Time
	}{
		// bug26198: month + year without day
		{"bug26198-oct2001", "Oct 2001", utcRef, time.Date(2001, 10, 1, 0, 0, 0, 0, time.UTC)},
		{"bug26198-2001oct", "2001 Oct", utcRef, time.Date(2001, 10, 1, 0, 0, 0, 0, time.UTC)},

		// bug28024: time before date
		{"bug28024", "17:00 2004-01-01", utcRef, time.Date(2004, 1, 1, 17, 0, 0, 0, time.UTC)},

		// bug30532: date + relative hours (without timezone)
		{"bug30532-1h", "2004-10-31 +1 hour", utcRef, time.Date(2004, 10, 31, 1, 0, 0, 0, time.UTC)},
		{"bug30532-2h", "2004-10-31 +2 hours", utcRef, time.Date(2004, 10, 31, 2, 0, 0, 0, time.UTC)},
		{"bug30532-3h", "2004-10-31 +3 hours", utcRef, time.Date(2004, 10, 31, 3, 0, 0, 0, time.UTC)},

		// bug33578: compact month/day formats
		{"bug33578-11oct2005", "11Oct2005", utcRef, time.Date(2005, 10, 11, 0, 0, 0, 0, time.UTC)},
		{"bug33578-11oct-space", "11Oct 2005", utcRef, time.Date(2005, 10, 11, 0, 0, 0, 0, time.UTC)},

		// bug37017: full timezone name in date string
		{"bug37017-ny1", "2006-05-12 13:00:01 America/New_York", time.Time{},
			time.Date(2006, 5, 12, 13, 0, 1, 0, mustLoadTz("America/New_York"))},
		{"bug37017-gmt", "2006-05-12 12:59:59 GMT", time.Time{},
			time.Date(2006, 5, 12, 12, 59, 59, 0, time.UTC)},

		// bug43452: ordinal weekday + month + year
		{"bug43452-thu-nov", "Thursday Nov 2007", utcRef,
			time.Date(2007, 11, 1, 0, 0, 0, 0, time.UTC)},
		{"bug43452-1st-thu", "first Thursday Nov 2007", utcRef,
			time.Date(2007, 11, 1, 0, 0, 0, 0, time.UTC)},
		{"bug43452-2nd-thu", "second Thursday Nov 2007", utcRef,
			time.Date(2007, 11, 8, 0, 0, 0, 0, time.UTC)},
		{"bug43452-3rd-thu", "third Thursday Nov 2007", utcRef,
			time.Date(2007, 11, 15, 0, 0, 0, 0, time.UTC)},
		{"bug43452-last-thu", "last Thursday Nov 2007", utcRef,
			time.Date(2007, 11, 29, 0, 0, 0, 0, time.UTC)},

		// bug45081: 12-hour time with AM/PM
		{"bug45081-12am", "11-MAY-1988 12:00:00AM", utcRef,
			time.Date(1988, 5, 11, 0, 0, 0, 0, time.UTC)},
		{"bug45081-12pm", "11-MAY-1988 12:00:00PM", utcRef,
			time.Date(1988, 5, 11, 12, 0, 0, 0, time.UTC)},
		{"bug45081-1am", "11-MAY-1988 1:00:00AM", utcRef,
			time.Date(1988, 5, 11, 1, 0, 0, 0, time.UTC)},
		{"bug45081-1pm", "11-MAY-1988 1:00:00PM", utcRef,
			time.Date(1988, 5, 11, 13, 0, 0, 0, time.UTC)},

		// bug52290: 24:00:00 as midnight next day
		{"bug52290", "2010-03-06T24:00:00", utcRef,
			time.Date(2010, 3, 7, 0, 0, 0, 0, time.UTC)},

		// bug52808: first/last day of month
		{"bug52808-first-jan", "first day of 2010-01", utcRef,
			time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"bug52808-last-jan", "last day of 2010-01", utcRef,
			time.Date(2010, 1, 31, 0, 0, 0, 0, time.UTC)},
		{"bug52808-last-feb", "last day of 2010-02", utcRef,
			time.Date(2010, 2, 28, 0, 0, 0, 0, time.UTC)},
		{"bug52808-last-feb-leap", "last day of 2012-02", utcRef,
			time.Date(2012, 2, 29, 0, 0, 0, 0, time.UTC)},

		// bug53437: extra whitespace
		{"bug53437", "  2010-10-06   12:53:10  ", utcRef,
			time.Date(2010, 10, 6, 12, 53, 10, 0, time.UTC)},

		// bug62896: ordinal weekday of month
		{"bug62896", "first Monday of January 2013", utcRef,
			time.Date(2013, 1, 7, 0, 0, 0, 0, time.UTC)},

		// bug66721: microsecond parsing
		{"bug66721", "2014-01-01 00:00:00.123456", utcRef,
			time.Date(2014, 1, 1, 0, 0, 0, 123456000, time.UTC)},

		// bug48678: negative year
		{"bug48678", "-0001-06-28", utcRef,
			time.Date(-1, 6, 28, 0, 0, 0, 0, time.UTC)},

		// Various RFC datetime formats
		{"rfc2822", "Thu, 20 Nov 2003 16:20:42 +0000", time.Time{},
			time.Date(2003, 11, 20, 16, 20, 42, 0, time.UTC)},
		{"rfc2822-gmt", "Thu, 20 Nov 2003 16:20:42 GMT", time.Time{},
			time.Date(2003, 11, 20, 16, 20, 42, 0, time.UTC)},
		{"rfc2822-noday", "20 Nov 2003 16:20:42 +0000", time.Time{},
			time.Date(2003, 11, 20, 16, 20, 42, 0, time.UTC)},

		// ISO 8601 with single-digit components
		{"iso-single-digit", "2006-1-6T0:0:0-8:0", utcRef,
			time.Date(2006, 1, 6, 0, 0, 0, 0, time.FixedZone("", -8*3600))},

		// Year-month only
		{"year-month-1", "2006-1", utcRef, time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"year-month-03", "2006-03", utcRef, time.Date(2006, 3, 1, 0, 0, 0, 0, time.UTC)},
		{"year-month-12", "2006-12", utcRef, time.Date(2006, 12, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func mustLoadTz(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
