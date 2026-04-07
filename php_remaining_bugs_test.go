package strtotime

import (
	"testing"
	"time"
)

// TestPHPRemainingBugs tests strtotime parsing issues that cause goro ext/date
// test failures. Each test case is derived from a specific PHP bug report.

// TestBug43452WeekdayExpressions tests "Nth weekday Month Year" expressions.
// PHP bug #43452: "weekday" is not equivalent to "1 weekday"
func TestBug43452WeekdayExpressions(t *testing.T) {
	// Nov 1, 2007 is a Thursday. CET = UTC+1.
	cet := time.FixedZone("CET", 3600)
	base := time.Date(2007, 1, 1, 0, 0, 0, 0, cet)

	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		// Nov 1 2007 is a Thursday.
		// "Thursday Nov 2007" = Nov 1 (first occurrence, no skip)
		{"thursday", "Thursday Nov 2007", time.Date(2007, 11, 1, 0, 0, 0, 0, cet)},
		// "1 Thursday Nov 2007" = Nov 1 (same as bare weekday)
		{"1_thursday", "1 Thursday Nov 2007", time.Date(2007, 11, 1, 0, 0, 0, 0, cet)},
		// "2 Thursday Nov 2007" = Nov 8
		{"2_thursday", "2 Thursday Nov 2007", time.Date(2007, 11, 8, 0, 0, 0, 0, cet)},
		// "3 Thursday Nov 2007" = Nov 15
		{"3_thursday", "3 Thursday Nov 2007", time.Date(2007, 11, 15, 0, 0, 0, 0, cet)},
		// "first/second/third" skip past the current day when it matches.
		// "first Thursday Nov 2007" = Nov 8 (skips Nov 1 which IS a Thursday)
		{"first_thursday", "first Thursday Nov 2007", time.Date(2007, 11, 8, 0, 0, 0, 0, cet)},
		// "second Thursday Nov 2007" = Nov 15
		{"second_thursday", "second Thursday Nov 2007", time.Date(2007, 11, 15, 0, 0, 0, 0, cet)},
		// "third Thursday Nov 2007" = Nov 22
		{"third_thursday", "third Thursday Nov 2007", time.Date(2007, 11, 22, 0, 0, 0, 0, cet)},
		// "+1 week Thursday Nov 2007" = Nov 8 (forward 1 week from Nov 1)
		{"+1week_thursday", "+1 week Thursday Nov 2007", time.Date(2007, 11, 8, 0, 0, 0, 0, cet)},
		// "+2 week Thursday Nov 2007" = Nov 15
		{"+2week_thursday", "+2 week Thursday Nov 2007", time.Date(2007, 11, 15, 0, 0, 0, 0, cet)},
		// "+3 week Thursday Nov 2007" = Nov 22
		{"+3week_thursday", "+3 week Thursday Nov 2007", time.Date(2007, 11, 22, 0, 0, 0, 0, cet)},
		// Friday (Nov 2) — when current day-of-week doesn't match, "first" = first occurrence
		// "Friday Nov 2007" = Nov 2
		{"friday", "Friday Nov 2007", time.Date(2007, 11, 2, 0, 0, 0, 0, cet)},
		// "first Friday Nov 2007" = Nov 2 (doesn't skip because Nov 1 is NOT a Friday)
		{"first_friday", "first Friday Nov 2007", time.Date(2007, 11, 2, 0, 0, 0, 0, cet)},
		// "second Friday Nov 2007" = Nov 9
		{"second_friday", "second Friday Nov 2007", time.Date(2007, 11, 9, 0, 0, 0, 0, cet)},
		// "third Friday Nov 2007" = Nov 16
		{"third_friday", "third Friday Nov 2007", time.Date(2007, 11, 16, 0, 0, 0, 0, cet)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToTime(tc.input, Rel(base), InTZ(cet))
			if err != nil {
				t.Fatalf("StrToTime(%q) error: %v", tc.input, err)
			}
			if !result.Equal(tc.expected) {
				t.Errorf("StrToTime(%q) = %v, want %v", tc.input, result, tc.expected)
			}
			phpVerify(t, tc.input, result, base, cet)
		})
	}
}

// TestBug61642WeekdayModifier tests "N weekdays" modifier.
// PHP bug #61642: wrong calculation of +/-N weekdays.
// "0 weekdays" from a Saturday or Sunday should snap to Monday.
func TestBug61642WeekdayModifier(t *testing.T) {
	base := time.Date(2012, 3, 11, 0, 0, 0, 0, time.UTC) // Sunday

	tests := []struct {
		name     string
		input    string
		base     time.Time
		expected time.Time
	}{
		// 0 weekdays from Sunday = Monday
		{"0wd_from_sun", "0 weekdays", time.Date(2012, 3, 11, 0, 0, 0, 0, time.UTC), time.Date(2012, 3, 12, 0, 0, 0, 0, time.UTC)},
		// 0 weekdays from Saturday = Monday
		{"0wd_from_sat", "0 weekdays", time.Date(2012, 3, 10, 0, 0, 0, 0, time.UTC), time.Date(2012, 3, 12, 0, 0, 0, 0, time.UTC)},
		// 0 weekdays from Friday = Friday (already a weekday)
		{"0wd_from_fri", "0 weekdays", time.Date(2012, 3, 9, 0, 0, 0, 0, time.UTC), time.Date(2012, 3, 9, 0, 0, 0, 0, time.UTC)},
		// +5 weekdays from Thursday Mar 8 = next Thursday Mar 15
		{"5wd_from_thu", "+5 weekdays", time.Date(2012, 3, 8, 0, 0, 0, 0, time.UTC), time.Date(2012, 3, 15, 0, 0, 0, 0, time.UTC)},
		// -5 weekdays from Thursday Mar 15 = previous Thursday Mar 8
		{"-5wd_from_thu", "-5 weekdays", time.Date(2012, 3, 15, 0, 0, 0, 0, time.UTC), time.Date(2012, 3, 8, 0, 0, 0, 0, time.UTC)},
	}

	_ = base
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToTime(tc.input, Rel(tc.base), InTZ(time.UTC))
			if err != nil {
				t.Fatalf("StrToTime(%q, base=%v) error: %v", tc.input, tc.base, err)
			}
			if !result.Equal(tc.expected) {
				t.Errorf("StrToTime(%q, base=%v) = %v, want %v",
					tc.input, tc.base, result, tc.expected)
			}
			phpVerify(t, tc.input, result, tc.base, time.UTC)
		})
	}
}

// TestAgoKeyword tests the "ago" relative modifier.
// PHP: strtotime("1 second ago") = now - 1 second.
// Used in oo_002.phpt via DateTime::modify('1 second ago').
func TestAgoKeyword(t *testing.T) {
	base := time.Date(2007, 8, 1, 13, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{"1_second_ago", "1 second ago", time.Date(2007, 8, 1, 12, 59, 59, 0, time.UTC)},
		{"5_minutes_ago", "5 minutes ago", time.Date(2007, 8, 1, 12, 55, 0, 0, time.UTC)},
		{"2_hours_ago", "2 hours ago", time.Date(2007, 8, 1, 11, 0, 0, 0, time.UTC)},
		{"3_days_ago", "3 days ago", time.Date(2007, 7, 29, 13, 0, 0, 0, time.UTC)},
		{"1_month_ago", "1 month ago", time.Date(2007, 7, 1, 13, 0, 0, 0, time.UTC)},
		{"1_year_ago", "1 year ago", time.Date(2006, 8, 1, 13, 0, 0, 0, time.UTC)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToTime(tc.input, Rel(base), InTZ(time.UTC))
			if err != nil {
				t.Fatalf("StrToTime(%q) error: %v", tc.input, err)
			}
			if !result.Equal(tc.expected) {
				t.Errorf("StrToTime(%q) = %v, want %v", tc.input, result, tc.expected)
			}
			phpVerify(t, tc.input, result, base, time.UTC)
		})
	}
}

// TestBug27780DaysAgo tests "N days ago" combined with datetime strings.
// PHP bug #27780: relative date strings with "ago".
func TestBug27780DaysAgo(t *testing.T) {
	chi, _ := time.LoadLocation("America/Chicago")
	base := time.Date(2004, 4, 7, 0, 0, 0, 0, chi)

	tests := []struct {
		name     string
		input    string
		expected int64 // Unix timestamp
	}{
		// "2004-04-07 00:00:00 11 days ago" = 2004-03-27 00:00:00 CST
		{"11_days_ago", "2004-04-07 00:00:00 11 days ago", 1080367200},
		// "-10 day +2 hours" from base = 2004-03-28 02:00:00 CST
		{"-10day_+2hours", "-10 day +2 hours", 1080460800},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToTime(tc.input, Rel(base), InTZ(chi))
			if err != nil {
				t.Fatalf("StrToTime(%q) error: %v", tc.input, err)
			}
			if result.Unix() != tc.expected {
				t.Errorf("StrToTime(%q) = unix %d (%v), want unix %d",
					tc.input, result.Unix(), result, tc.expected)
			}
			phpVerify(t, tc.input, result, base, chi)
		})
	}
}

// TestBug32086DSTBoundary tests strtotime around DST boundaries.
// PHP bug #32086: +1 day near America/Sao_Paulo DST.
func TestBug32086DSTBoundary(t *testing.T) {
	sp, _ := time.LoadLocation("America/Sao_Paulo")

	// strtotime("2004-11-01") in Sao Paulo
	t.Run("nov01", func(t *testing.T) {
		base := time.Date(2004, 1, 1, 0, 0, 0, 0, sp)
		result, err := StrToTime("2004-11-01", Rel(base), InTZ(sp))
		if err != nil {
			t.Fatalf("StrToTime(\"2004-11-01\") error: %v", err)
		}
		expected := int64(1099278000)
		if result.Unix() != expected {
			t.Errorf("StrToTime(\"2004-11-01\") = unix %d (%v), want unix %d",
				result.Unix(), result, expected)
		}
		phpVerify(t, "2004-11-01", result, base, sp)
	})

	// strtotime("+1 day", strtotime("2004-11-01")) — use the PHP result as base
	t.Run("plus1day", func(t *testing.T) {
		base := time.Unix(1099278000, 0).In(sp)
		result, err := StrToTime("+1 day", Rel(base), InTZ(sp))
		if err != nil {
			t.Fatalf("StrToTime(\"+1 day\") error: %v", err)
		}
		expected := int64(1099364400)
		if result.Unix() != expected {
			t.Errorf("StrToTime(\"+1 day\") = unix %d (%v), want unix %d",
				result.Unix(), result, expected)
		}
		phpVerify(t, "+1 day", result, base, sp)
	})
}

// TestBug32555DSTSpringForward tests +1 day across US/Eastern DST spring-forward.
// PHP bug #32555.
func TestBug32555DSTSpringForward(t *testing.T) {
	east, _ := time.LoadLocation("US/Eastern")
	// Sat Apr 2, 2005 01:30:00 EST (before spring forward on Apr 3)
	stamp := time.Date(2005, 4, 2, 1, 30, 0, 0, east)

	result, err := StrToTime("+1 day", Rel(stamp), InTZ(east))
	if err != nil {
		t.Fatalf("StrToTime(\"+1 day\") error: %v", err)
	}
	// PHP: strtotime("+1 day") preserves wall-clock time using AddDate semantics.
	// Apr 3 01:30 is before the 02:00 spring-forward, so it stays EST.
	// Expected: Sun, 03 Apr 2005 01:30:00 -0500 (EST)
	expected := time.Unix(1112509800, 0)
	if !result.Equal(expected) {
		t.Errorf("StrToTime(\"+1 day\", Sat 01:30 EST) = %v (unix %d), want %v (unix %d)",
			result, result.Unix(), expected.In(east), expected.Unix())
	}
	phpVerify(t, "+1 day", result, stamp, east)
}

// TestBug34771AMPM tests AM/PM dot notation ("a.m.", "p.m.").
// PHP bug #34771.
func TestBug34771AMPM(t *testing.T) {
	base := time.Date(2005, 12, 22, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{"12am", "2005-12-22 12am", time.Date(2005, 12, 22, 0, 0, 0, 0, time.UTC)},
		{"1am", "2005-12-22 1am", time.Date(2005, 12, 22, 1, 0, 0, 0, time.UTC)},
		{"12pm", "2005-12-22 12pm", time.Date(2005, 12, 22, 12, 0, 0, 0, time.UTC)},
		{"1pm", "2005-12-22 1pm", time.Date(2005, 12, 22, 13, 0, 0, 0, time.UTC)},
		// Dot notation
		{"12a.m.", "2005-12-22 12a.m.", time.Date(2005, 12, 22, 0, 0, 0, 0, time.UTC)},
		{"1a.m.", "2005-12-22 1a.m.", time.Date(2005, 12, 22, 1, 0, 0, 0, time.UTC)},
		{"12p.m.", "2005-12-22 12p.m.", time.Date(2005, 12, 22, 12, 0, 0, 0, time.UTC)},
		{"1p.m.", "2005-12-22 1p.m.", time.Date(2005, 12, 22, 13, 0, 0, 0, time.UTC)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToTime(tc.input, Rel(base), InTZ(time.UTC))
			if err != nil {
				t.Fatalf("StrToTime(%q) error: %v", tc.input, err)
			}
			if !result.Equal(tc.expected) {
				t.Errorf("StrToTime(%q) = %v, want %v", tc.input, result, tc.expected)
			}
			phpVerify(t, tc.input, result, base, time.UTC)
		})
	}
}

// TestBug51987ISOOrdinalDate tests ISO 8601 ordinal date format (YYYY-DDD).
// PHP bug #51987.
func TestBug51987ISOOrdinalDate(t *testing.T) {
	base := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		// 1985-102 = April 12, 1985 (day 102 of 1985)
		{"1985-102", "1985-102", time.Date(1985, 4, 12, 0, 0, 0, 0, time.UTC)},
		// 2007-001 = January 1, 2007
		{"2007-001", "2007-001", time.Date(2007, 1, 1, 0, 0, 0, 0, time.UTC)},
		// 2007-365 = December 31, 2007
		{"2007-365", "2007-365", time.Date(2007, 12, 31, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToTime(tc.input, Rel(base), InTZ(time.UTC))
			if err != nil {
				t.Fatalf("StrToTime(%q) error: %v", tc.input, err)
			}
			if result.Year() != tc.expected.Year() || result.Month() != tc.expected.Month() || result.Day() != tc.expected.Day() {
				t.Errorf("StrToTime(%q) = %v, want %v", tc.input, result.Format("2006-01-02"), tc.expected.Format("2006-01-02"))
			}
			phpVerify(t, tc.input, result, base, time.UTC)
		})
	}
}

// TestBug73942ThisWeek tests "dayname this week" modifier.
// PHP bug #73942.
func TestBug73942ThisWeek(t *testing.T) {
	// Sunday Jan 8, 2017 — ISO week 1, Monday=Jan 2, Friday=Jan 6
	base := time.Date(2017, 1, 8, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		// "Monday this week" from Sunday = Monday Jan 2
		{"monday", "Monday this week", time.Date(2017, 1, 2, 0, 0, 0, 0, time.UTC)},
		// "Friday this week" from Sunday = Friday Jan 6
		{"friday", "Friday this week", time.Date(2017, 1, 6, 0, 0, 0, 0, time.UTC)},
		// "Sunday this week" from Sunday = Sunday Jan 8 (same day)
		{"sunday", "Sunday this week", time.Date(2017, 1, 8, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToTime(tc.input, Rel(base), InTZ(time.UTC))
			if err != nil {
				t.Fatalf("StrToTime(%q) error: %v", tc.input, err)
			}
			if !result.Equal(tc.expected) {
				t.Errorf("StrToTime(%q, base=Sun Jan 8) = %v (%s), want %v (%s)",
					tc.input,
					result, result.Format("Mon Jan 2"),
					tc.expected, tc.expected.Format("Mon Jan 2"))
			}
			phpVerify(t, tc.input, result, base, time.UTC)
		})
	}
}

// TestBug74057SaturdayThisWeek tests "saturday this week" with explicit reference.
// PHP bug #74057: strtotime("saturday this week") ignores reference timestamp.
func TestBug74057SaturdayThisWeek(t *testing.T) {
	tests := []struct {
		name     string
		base     time.Time
		expected time.Time
	}{
		// From Monday Apr 3 → Saturday Apr 8
		{"from_mon", time.Date(2017, 4, 3, 0, 0, 0, 0, time.UTC), time.Date(2017, 4, 8, 0, 0, 0, 0, time.UTC)},
		// From Tuesday Apr 4 → Saturday Apr 8
		{"from_tue", time.Date(2017, 4, 4, 0, 0, 0, 0, time.UTC), time.Date(2017, 4, 8, 0, 0, 0, 0, time.UTC)},
		// From Wednesday Apr 5 → Saturday Apr 8
		{"from_wed", time.Date(2017, 4, 5, 0, 0, 0, 0, time.UTC), time.Date(2017, 4, 8, 0, 0, 0, 0, time.UTC)},
		// From Thursday Apr 6 → Saturday Apr 8
		{"from_thu", time.Date(2017, 4, 6, 0, 0, 0, 0, time.UTC), time.Date(2017, 4, 8, 0, 0, 0, 0, time.UTC)},
		// From Friday Apr 7 → Saturday Apr 8
		{"from_fri", time.Date(2017, 4, 7, 0, 0, 0, 0, time.UTC), time.Date(2017, 4, 8, 0, 0, 0, 0, time.UTC)},
		// From Saturday Apr 8 → Saturday Apr 8 (same day)
		{"from_sat", time.Date(2017, 4, 8, 0, 0, 0, 0, time.UTC), time.Date(2017, 4, 8, 0, 0, 0, 0, time.UTC)},
		// From Sunday Apr 9 → Saturday Apr 8 (previous Saturday in same ISO week)
		// PHP considers Sunday the end of the ISO week
		{"from_sun", time.Date(2017, 4, 9, 0, 0, 0, 0, time.UTC), time.Date(2017, 4, 8, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToTime("saturday this week", Rel(tc.base), InTZ(time.UTC))
			if err != nil {
				t.Fatalf("StrToTime(\"saturday this week\", base=%v) error: %v", tc.base.Format("Mon Jan 2"), err)
			}
			if !result.Equal(tc.expected) {
				t.Errorf("StrToTime(\"saturday this week\", base=%v) = %v (%s), want %v (%s)",
					tc.base.Format("Mon Jan 2"),
					result, result.Format("Mon Jan 2"),
					tc.expected, tc.expected.Format("Mon Jan 2"))
			}
			phpVerify(t, "saturday this week", result, tc.base, time.UTC)
		})
	}
}
