package strtotime

import (
	"testing"
	"time"
)

// Additional test vectors from failing PHP ext/date bug tests.
// These inputs must parse successfully and produce the correct time.
func TestPHPDateBugs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		// bug26198: month + year without day
		{"bug26198-oct2001", "Oct 2001", time.Date(2001, 10, 1, 0, 0, 0, 0, time.UTC)},
		{"bug26198-2001oct", "2001 Oct", time.Date(2001, 10, 1, 0, 0, 0, 0, time.UTC)},

		// bug28024: time before date
		{"bug28024", "17:00 2004-01-01", time.Date(2004, 1, 1, 17, 0, 0, 0, time.UTC)},

		// bug30532: date + relative hours (without timezone)
		{"bug30532-1h", "2004-10-31 +1 hour", time.Date(2004, 10, 31, 1, 0, 0, 0, time.UTC)},
		{"bug30532-2h", "2004-10-31 +2 hours", time.Date(2004, 10, 31, 2, 0, 0, 0, time.UTC)},
		{"bug30532-3h", "2004-10-31 +3 hours", time.Date(2004, 10, 31, 3, 0, 0, 0, time.UTC)},

		// bug33578: compact month/day formats
		{"bug33578-11oct2005", "11Oct2005", time.Date(2005, 10, 11, 0, 0, 0, 0, time.UTC)},
		{"bug33578-11oct-space", "11Oct 2005", time.Date(2005, 10, 11, 0, 0, 0, 0, time.UTC)},

		// bug37017: full timezone name in date string
		{"bug37017-ny1", "2006-05-12 13:00:01 America/New_York",
			time.Date(2006, 5, 12, 13, 0, 1, 0, mustLoadTz("America/New_York"))},
		{"bug37017-gmt", "2006-05-12 12:59:59 GMT",
			time.Date(2006, 5, 12, 12, 59, 59, 0, time.UTC)},

		// bug43452: ordinal weekday + month + year
		{"bug43452-thu-nov", "Thursday Nov 2007",
			time.Date(2007, 11, 1, 0, 0, 0, 0, time.UTC)},
		{"bug43452-1st-thu", "first Thursday Nov 2007",
			time.Date(2007, 11, 8, 0, 0, 0, 0, time.UTC)},
		{"bug43452-2nd-thu", "second Thursday Nov 2007",
			time.Date(2007, 11, 15, 0, 0, 0, 0, time.UTC)},
		{"bug43452-3rd-thu", "third Thursday Nov 2007",
			time.Date(2007, 11, 22, 0, 0, 0, 0, time.UTC)},
		{"bug43452-last-thu", "last Thursday Nov 2007",
			time.Date(2007, 10, 25, 0, 0, 0, 0, time.UTC)},

		// bug45081: 12-hour time with AM/PM
		{"bug45081-12am", "11-MAY-1988 12:00:00AM",
			time.Date(1988, 5, 11, 0, 0, 0, 0, time.UTC)},
		{"bug45081-12pm", "11-MAY-1988 12:00:00PM",
			time.Date(1988, 5, 11, 12, 0, 0, 0, time.UTC)},
		{"bug45081-1am", "11-MAY-1988 1:00:00AM",
			time.Date(1988, 5, 11, 1, 0, 0, 0, time.UTC)},
		{"bug45081-1pm", "11-MAY-1988 1:00:00PM",
			time.Date(1988, 5, 11, 13, 0, 0, 0, time.UTC)},

		// bug52290: 24:00:00 as midnight next day
		{"bug52290", "2010-03-06T24:00:00",
			time.Date(2010, 3, 7, 0, 0, 0, 0, time.UTC)},

		// bug52808: first/last day of month
		{"bug52808-first-jan", "first day of 2010-01",
			time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"bug52808-last-jan", "last day of 2010-01",
			time.Date(2010, 1, 31, 0, 0, 0, 0, time.UTC)},
		{"bug52808-last-feb", "last day of 2010-02",
			time.Date(2010, 2, 28, 0, 0, 0, 0, time.UTC)},
		{"bug52808-last-feb-leap", "last day of 2012-02",
			time.Date(2012, 2, 29, 0, 0, 0, 0, time.UTC)},

		// bug53437: extra whitespace
		{"bug53437", "  2010-10-06   12:53:10  ",
			time.Date(2010, 10, 6, 12, 53, 10, 0, time.UTC)},

		// bug62896: ordinal weekday of month
		{"bug62896", "first Monday of January 2013",
			time.Date(2013, 1, 7, 0, 0, 0, 0, time.UTC)},

		// bug66721: microsecond parsing
		{"bug66721", "2014-01-01 00:00:00.123456",
			time.Date(2014, 1, 1, 0, 0, 0, 123456000, time.UTC)},

		// bug48678: negative year
		{"bug48678", "-0001-06-28",
			time.Date(-1, 6, 28, 0, 0, 0, 0, time.UTC)},

		// Various RFC datetime formats
		{"rfc2822", "Thu, 20 Nov 2003 16:20:42 +0000",
			time.Date(2003, 11, 20, 16, 20, 42, 0, time.UTC)},
		{"rfc2822-gmt", "Thu, 20 Nov 2003 16:20:42 GMT",
			time.Date(2003, 11, 20, 16, 20, 42, 0, time.UTC)},
		{"rfc2822-noday", "20 Nov 2003 16:20:42 +0000",
			time.Date(2003, 11, 20, 16, 20, 42, 0, time.UTC)},

		// ISO 8601 with single-digit components
		{"iso-single-digit", "2006-1-6T0:0:0-8:0",
			time.Date(2006, 1, 6, 0, 0, 0, 0, time.FixedZone("", -8*3600))},

		// Year-month only
		{"year-month-1", "2006-1", time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"year-month-03", "2006-03", time.Date(2006, 3, 1, 0, 0, 0, 0, time.UTC)},
		{"year-month-12", "2006-12", time.Date(2006, 12, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StrToTime(tt.input, InTZ(time.UTC))
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

func mustLoadTz(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// Additional test vectors from regressions when removing custom parser.
// These formats must be handled by the strtotime library.
func TestPHPDateFormatsRegression(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Basic dates
		{"basic-date", "2009-10-11", false},
		{"basic-date2", "2009-01-01", false},
		{"basic-date3", "2019-06-01", false},
		{"basic-date4", "2014-09-20", false},
		{"basic-date5", "2007-03-11", false},
		{"basic-date6", "2015-02-01", false},
		{"basic-date7", "2010-06-07", false},

		// Date + time + timezone
		{"datetime-gmt", "2005-07-14 22:30:41 GMT", false},
		{"datetime-offset", "2007-03-11T00:00:00-0800", false},

		// Bare timezone name as datetime (means "now" in that tz)
		{"bare-tz-gmt", "GMT", false},

		// Month name + ordinal + time + am/pm
		{"month-ordinal-pm", "May 18th 5:05pm", false},
		{"month-ordinal-am", "May 18th 5:05am", false},
		{"month-ordinal-space-pm", "May 18th 5:05 pm", false},
		{"month-ordinal-space-am", "May 18th 5:05 am", false},
		{"month-ordinal-year-pm", "May 18th 2006 5:05pm", false},
		{"month-ordinal-time", "May 18th 5:05", false},

		// Negative year
		{"negative-year", "-2007-06-28 00:00:00", false},

		// Old year formats
		{"old-year-month", "January 0099", false},
		{"old-year-date", "January 1, 0099", false},
		{"old-year-iso", "0099-01", false},

		// RFC + relative
		{"rfc-plus-relative", "Mon, 08 May 2006 13:06:44 -0400 +30 days", false},

		// Military-style time
		{"military-time-1", "04/04/04 2345", false},
		{"military-time-2", "04/04/04 0045", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := StrToTime(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("StrToTime(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("StrToTime(%q) unexpected error: %v", tt.input, err)
			}
			if !tt.wantErr && err == nil {
				result, _ := StrToTime(tt.input, InTZ(time.UTC))
				phpVerify(t, tt.input, result, time.Time{}, time.UTC)
			}
		})
	}
}
