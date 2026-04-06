package strtotime

import (
	"testing"
	"time"
)

// TestPHPBugRegressions2 tests strtotime parsing for various PHP bug reports
// that still fail.
func TestPHPBugRegressions2(t *testing.T) {
	base := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    string
		base     time.Time
		expected time.Time
	}{
		// bug40861: Multiple +/- on relative units
		// "+60 minutes" from 12:00:00 → 13:00:00
		{
			name:     "bug40861_plus60min",
			input:    "+60 minutes",
			base:     time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2000, 1, 1, 13, 0, 0, 0, time.UTC),
		},
		// "+-60 minutes" = plus then minus-60 = PHP interprets as -60 min = 11:00:00
		// (the +- is a sign combination, the negative 60 wins)
		{
			name:     "bug40861_plus_minus_60min",
			input:    "+-60 minutes",
			base:     base,
			expected: time.Date(2000, 1, 1, 11, 0, 0, 0, time.UTC),
		},
		// "--60 minutes" = double-negative = +60 min = 13:00:00
		// This is the actual bug: "--60" fails to parse in current implementation
		{
			name:     "bug40861_double_minus_60min",
			input:    "--60 minutes",
			base:     base,
			expected: time.Date(2000, 1, 1, 13, 0, 0, 0, time.UTC),
		},

		// bug43452: weekday expressions with week offsets in Norway timezone
		// Nov 1, 2007 is a Thursday (CET = UTC+1)
		// "+1 week Thursday Nov 2007" = advance 1 week from first Thursday = Nov 8
		{
			name:     "bug43452_plus1_week_thursday_nov2007",
			input:    "+1 week Thursday Nov 2007",
			base:     time.Date(2007, 1, 1, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
			expected: time.Date(2007, 11, 8, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
		},
		// "+2 week Thursday Nov 2007" = Nov 15
		{
			name:     "bug43452_plus2_week_thursday_nov2007",
			input:    "+2 week Thursday Nov 2007",
			base:     time.Date(2007, 1, 1, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
			expected: time.Date(2007, 11, 15, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
		},
		// "+3 week Thursday Nov 2007" = Nov 22
		{
			name:     "bug43452_plus3_week_thursday_nov2007",
			input:    "+3 week Thursday Nov 2007",
			base:     time.Date(2007, 1, 1, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
			expected: time.Date(2007, 11, 22, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
		},
		// "first Thursday Nov 2007" = next Thursday after first Thursday = Nov 8
		{
			name:     "bug43452_first_thursday_nov2007",
			input:    "first Thursday Nov 2007",
			base:     time.Date(2007, 1, 1, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
			expected: time.Date(2007, 11, 8, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
		},
		// "second Thursday Nov 2007" = Nov 15
		{
			name:     "bug43452_second_thursday_nov2007",
			input:    "second Thursday Nov 2007",
			base:     time.Date(2007, 1, 1, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
			expected: time.Date(2007, 11, 15, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
		},
		// "third Thursday Nov 2007" = Nov 22
		{
			name:     "bug43452_third_thursday_nov2007",
			input:    "third Thursday Nov 2007",
			base:     time.Date(2007, 1, 1, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
			expected: time.Date(2007, 11, 22, 0, 0, 0, 0, time.FixedZone("CET", 3600)),
		},

		// oo_002: Time with month name and GMT
		// "1pm Aug 1 GMT 2007" = Aug 1, 2007 13:00:00 UTC
		{
			name:     "oo_002_time_month_gmt",
			input:    "1pm Aug 1 GMT 2007",
			base:     base,
			expected: time.Date(2007, 8, 1, 13, 0, 0, 0, time.UTC),
		},

		// bug37368: Date with offset and relative modifier
		// "Mon, 08 May 2006 13:06:44 -0400 +30 days" = Wed, 07 Jun 2006 17:06:44 +0000
		// unix 1149700004: May 8 13:06:44-0400 is 17:06:44 UTC; +30 days = Jun 7 13:06:44 -0400
		{
			name:  "bug37368_june",
			input: "Mon, 08 May 2006 13:06:44 -0400 +30 days",
			base:  base,
			// PHP output: Wed, 07 Jun 2006 17:06:44 +0000 = unix 1149700004
			expected: time.Unix(1149700004, 0).UTC(),
		},

		// bug54597: 4-digit year 0099 preservation
		// "January 1, 0099" should keep year as 99, not expand to 1999
		{
			name:     "bug54597_year_0099",
			input:    "January 1, 0099",
			base:     base,
			expected: time.Date(99, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		// "January 0099" (no day)
		{
			name:     "bug54597_january_0099",
			input:    "January 0099",
			base:     base,
			expected: time.Date(99, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		// "0099-01" ISO year-month
		{
			name:     "bug54597_0099_01",
			input:    "0099-01",
			base:     base,
			expected: time.Date(99, 1, 1, 0, 0, 0, 0, time.UTC),
		},

		// bug74173: Shortened UTC offset "+00:0" (3 digits total)
		// "2016-10-30T00:00:00+00:0" should parse as UTC
		{
			name:     "bug74173_short_tz_offset",
			input:    "2016-10-30T00:00:00+00:0",
			base:     base,
			expected: time.Date(2016, 10, 30, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := []Option{Rel(tc.base), InTZ(tc.base.Location())}
			result, err := StrToTime(tc.input, opts...)
			if err != nil {
				t.Errorf("StrToTime(%q) error: %v", tc.input, err)
				return
			}
			if !result.Equal(tc.expected) {
				t.Errorf("StrToTime(%q) = %v (unix=%d), want %v (unix=%d)",
					tc.input, result, result.Unix(), tc.expected, tc.expected.Unix())
			}
		})
	}
}

// TestPHPStrtotime3FailingFormats tests formats from strtotime3-64bit.phpt that
// currently fail. Base time: Unix 1150494719 = Fri Jun 16, 2006 22:51:59 +0100 (Europe/Lisbon summer).
func TestPHPStrtotime3FailingFormats(t *testing.T) {
	lisbon, err := time.LoadLocation("Europe/Lisbon")
	if err != nil {
		t.Fatalf("Failed to load Europe/Lisbon: %v", err)
	}
	baseTime := time.Unix(1150494719, 0).In(lisbon)

	tests := []struct {
		input    string
		expected time.Time
	}{
		// "22.49.12.42GMT" = time 22:49:12 with fraction in GMT (UTC+0)
		// → Fri, 16 Jun 2006 23:49:12 +0100 (= 22:49:12 UTC displayed in Lisbon +0100)
		{
			"22.49.12.42GMT",
			time.Unix(1150498152, 0).In(lisbon),
		},
		// "t0222" = PHP time prefix format: time 02:22:00 in local tz
		// → Fri, 16 Jun 2006 02:22:00 +0100
		{
			"t0222",
			time.Unix(1150420920, 0).In(lisbon),
		},
		// "022233" = hhmmss compact time format: 02:22:33 in local tz
		// → Fri, 16 Jun 2006 02:22:33 +0100
		{
			"022233",
			time.Unix(1150420953, 0).In(lisbon),
		},
		// "2006167" = pgydotd format: year 2006, day-of-year 167
		// Day 167 of 2006 = Jun 16, 2006
		// → Fri, 16 Jun 2006 00:00:00 +0100
		{
			"2006167",
			time.Unix(1150412400, 0).In(lisbon),
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := StrToTime(tc.input, Rel(baseTime), InTZ(lisbon))
			if err != nil {
				t.Errorf("StrToTime(%q) error: %v", tc.input, err)
				return
			}
			if result.Unix() != tc.expected.Unix() {
				t.Errorf("StrToTime(%q) = %v (unix=%d), want %v (unix=%d)",
					tc.input,
					result.Format("Mon, 02 Jan 2006 15:04:05 -0700"),
					result.Unix(),
					tc.expected.Format("Mon, 02 Jan 2006 15:04:05 -0700"),
					tc.expected.Unix())
			}
		})
	}
}

// TestPHPStrtotime3InvalidFormats tests inputs from strtotime3-64bit.phpt that should
// return false (error) but currently parse incorrectly.
func TestPHPStrtotime3InvalidFormats(t *testing.T) {
	lisbon, err := time.LoadLocation("Europe/Lisbon")
	if err != nil {
		t.Fatalf("Failed to load Europe/Lisbon: %v", err)
	}
	baseTime := time.Unix(1150494719, 0).In(lisbon)

	// "t0222 t0222" is a double time specification which should fail
	// (the test checks what the implementation does - PHP returns year 0222 date)
	// This is here to document the expected behavior even though it's unusual
	invalidInputs := []string{
		// These must fail (return error):
		"22:49:12 bogusTZ",
		"22.49.12.42bogusTZ",
		"022233 bogusTZ",
		"20060212T23:12:23 bogusTZ",
		"10/Oct/2000:13:55:36 +00100",
	}

	for _, input := range invalidInputs {
		input := input // capture
		t.Run("must_fail_"+input, func(t *testing.T) {
			_, err := StrToTime(input, Rel(baseTime), InTZ(lisbon))
			if err == nil {
				t.Errorf("StrToTime(%q) should fail but succeeded", input)
			}
		})
	}
}

// TestPHPStrtotimeBasicWordOrdinals tests from strtotime_basic.phpt.
// "first/second/third Monday December 2008" uses word ordinals which PHP treats
// differently from numeric ordinals: "first" skips past the current matching weekday.
// December 1, 2008 is a Monday. PHP's expected output:
//
//	first Monday December 2008  = 2008-12-08 (skip past Dec 1 to Dec 8)
//	second Monday December 2008 = 2008-12-15
//	third Monday December 2008  = 2008-12-22
func TestPHPStrtotimeBasicWordOrdinals(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		// Word ordinals: skip past current weekday to next occurrence
		{"first Monday December 2008", time.Date(2008, 12, 8, 0, 0, 0, 0, time.UTC)},
		{"second Monday December 2008", time.Date(2008, 12, 15, 0, 0, 0, 0, time.UTC)},
		{"third Monday December 2008", time.Date(2008, 12, 22, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result, err := StrToTime(tc.input, InTZ(time.UTC))
			if err != nil {
				t.Errorf("StrToTime(%q) error: %v", tc.input, err)
				return
			}
			if result.Unix() != tc.expected.Unix() {
				t.Errorf("StrToTime(%q) = %v (unix=%d), want %v (unix=%d)",
					tc.input,
					result.Format("2006-01-02"),
					result.Unix(),
					tc.expected.Format("2006-01-02"),
					tc.expected.Unix())
			}
		})
	}
}

// TestPHPStrtotime2DateFormats tests from strtotime2.phpt.
// DATE_COOKIE, DATE_ISO8601_EXPANDED, and DATE_RFC850 formats must round-trip
// correctly through strtotime.
//
// DATE_COOKIE format:  "l, d-M-Y H:i:s T"  e.g. "Thursday, 24-Jan-2019 10:00:00 UTC"
// DATE_ISO8601_EXPANDED: "+Y-m-d\TH:i:sP"   e.g. "+2019-01-24T10:00:00+00:00"
// DATE_RFC850 format: "l, d-M-y H:i:s T"   e.g. "Thursday, 24-Jan-19 10:00:00 UTC"
func TestPHPStrtotime2DateFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64 // unix timestamp
	}{
		// DATE_COOKIE: "Thursday, 24-Jan-2019 10:00:00 UTC"
		{
			name:     "DATE_COOKIE_Thu_24_Jan_2019",
			input:    "Thursday, 24-Jan-2019 10:00:00 UTC",
			expected: 1548324000,
		},
		// DATE_COOKIE with different date
		{
			name:     "DATE_COOKIE_Sat_15_Jul_2023",
			input:    "Saturday, 15-Jul-2023 14:30:00 UTC",
			expected: 1689431400,
		},
		// DATE_ISO8601_EXPANDED with leading +: "+2019-01-24T10:00:00+00:00"
		{
			name:     "DATE_ISO8601_EXPANDED_2019",
			input:    "+2019-01-24T10:00:00+00:00",
			expected: 1548324000,
		},
		// DATE_ISO8601_EXPANDED with leading +: "+2023-07-15T14:30:00+00:00"
		{
			name:     "DATE_ISO8601_EXPANDED_2023",
			input:    "+2023-07-15T14:30:00+00:00",
			expected: 1689431400,
		},
		// DATE_RFC850: "Thursday, 24-Jan-19 10:00:00 UTC" (2-digit year)
		{
			name:     "DATE_RFC850_Thu_24_Jan_19",
			input:    "Thursday, 24-Jan-19 10:00:00 UTC",
			expected: 1548324000,
		},
		// DATE_RFC850 with different date
		{
			name:     "DATE_RFC850_Sat_15_Jul_23",
			input:    "Saturday, 15-Jul-23 14:30:00 UTC",
			expected: 1689431400,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToTime(tc.input, InTZ(time.UTC))
			if err != nil {
				t.Errorf("StrToTime(%q) error: %v", tc.input, err)
				return
			}
			if result.Unix() != tc.expected {
				t.Errorf("StrToTime(%q) = unix %d (%v), want unix %d",
					tc.input, result.Unix(), result, tc.expected)
			}
		})
	}
}

// TestPHPBug72719WeekdayNextWeek tests from bug72719.phpt.
// "Monday next week 13:00" should set the day-of-week, then apply "next week",
// then set the time. All days must work (not just Sunday).
func TestPHPBug72719WeekdayNextWeek(t *testing.T) {
	tests := []struct {
		input   string
		weekday time.Weekday
	}{
		{"Monday next week 13:00", time.Monday},
		{"Tuesday next week 14:00", time.Tuesday},
		{"Wednesday next week 14:00", time.Wednesday},
		{"Thursday next week 15:00", time.Thursday},
		{"Friday next week 16:00", time.Friday},
		{"Saturday next week 17:00", time.Saturday},
		{"Sunday next week 18:00", time.Sunday},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result, err := StrToTime(tc.input, InTZ(time.UTC))
			if err != nil {
				t.Errorf("StrToTime(%q) error: %v", tc.input, err)
				return
			}
			if result.Weekday() != tc.weekday {
				t.Errorf("StrToTime(%q) = %v (weekday=%v), want weekday=%v",
					tc.input, result.Format("2006-01-02"), result.Weekday(), tc.weekday)
			}
		})
	}
}

// TestPHPBug61642Weekdays tests from bug61642.phpt.
// N weekdays modification must work correctly for all multiples, including 5, 10.
// From a Thursday (2012-03-29): +5 weekdays = +1 week = 2012-04-05 Thu.
// From a Saturday (2012-03-31): +5 weekdays = 5 business days forward = 2012-04-06 Fri.
func TestPHPBug61642Weekdays(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		base     time.Time
		expected time.Time
	}{
		// From Thursday 2012-03-29: +5 weekdays = Thu 2012-04-05
		{
			name:     "thu_plus5_weekdays",
			input:    "+5 weekdays",
			base:     time.Date(2012, 3, 29, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2012, 4, 5, 0, 0, 0, 0, time.UTC),
		},
		// From Friday 2012-03-30: +5 weekdays = Fri 2012-04-06
		{
			name:     "fri_plus5_weekdays",
			input:    "+5 weekdays",
			base:     time.Date(2012, 3, 30, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2012, 4, 6, 0, 0, 0, 0, time.UTC),
		},
		// From Saturday 2012-03-31: +5 weekdays = Mon 2012-04-02 (sat+sun skipped, then 5 business days)
		// Actually PHP: from Sat/Sun, +N weekdays first snaps to Mon then adds N-1 more weekdays
		// PHP expected: Fri 2012-04-06
		{
			name:     "sat_plus5_weekdays",
			input:    "+5 weekdays",
			base:     time.Date(2012, 3, 31, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2012, 4, 6, 0, 0, 0, 0, time.UTC),
		},
		// From Sunday 2012-04-01: +5 weekdays = Fri 2012-04-06
		{
			name:     "sun_plus5_weekdays",
			input:    "+5 weekdays",
			base:     time.Date(2012, 4, 1, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2012, 4, 6, 0, 0, 0, 0, time.UTC),
		},
		// From Monday 2012-04-02: +5 weekdays = Mon 2012-04-09
		{
			name:     "mon_plus5_weekdays",
			input:    "+5 weekdays",
			base:     time.Date(2012, 4, 2, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2012, 4, 9, 0, 0, 0, 0, time.UTC),
		},
		// From Thursday 2012-03-29: -5 weekdays = Thu 2012-03-22
		{
			name:     "thu_minus5_weekdays",
			input:    "-5 weekdays",
			base:     time.Date(2012, 3, 29, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2012, 3, 22, 0, 0, 0, 0, time.UTC),
		},
		// From Thursday 2012-03-29: +10 weekdays = Thu 2012-04-12
		{
			name:     "thu_plus10_weekdays",
			input:    "+10 weekdays",
			base:     time.Date(2012, 3, 29, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2012, 4, 12, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := StrToTime(tc.input, Rel(tc.base), InTZ(time.UTC))
			if err != nil {
				t.Errorf("StrToTime(%q) error: %v", tc.input, err)
				return
			}
			if result.Unix() != tc.expected.Unix() {
				t.Errorf("StrToTime(%q) base=%v: got %v (unix=%d), want %v (unix=%d)",
					tc.input,
					tc.base.Format("2006-01-02 Mon"),
					result.Format("2006-01-02 Mon"),
					result.Unix(),
					tc.expected.Format("2006-01-02 Mon"),
					tc.expected.Unix())
			}
		})
	}
}
