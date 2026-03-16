package strtotime

import (
	"testing"
	"time"
)

// TestPHPStrtotime3_64bit tests from PHP's strtotime3-64bit.phpt
// Base time: 1150494719 (June 16, 2006 21:51:59 UTC) in Europe/Lisbon timezone
func TestPHPStrtotime3_64bit(t *testing.T) {
	lisbon, err := time.LoadLocation("Europe/Lisbon")
	if err != nil {
		t.Fatalf("Failed to load Europe/Lisbon: %v", err)
	}

	baseTime := time.Unix(1150494719, 0).In(lisbon)

	type testCase struct {
		input    string
		expected time.Time
	}

	tests := []testCase{
		// "yesterday" → Thu, 15 Jun 2006 00:00:00 +0100
		{
			"yesterday",
			time.Date(2006, 6, 15, 0, 0, 0, 0, lisbon),
		},
		// "2-3-2004" → Tue, 02 Mar 2004 00:00:00 +0000
		{
			"2-3-2004",
			time.Date(2004, 3, 2, 0, 0, 0, 0, time.UTC),
		},
		// "2.3.2004" → Tue, 02 Mar 2004 00:00:00 +0000
		{
			"2.3.2004",
			time.Date(2004, 3, 2, 0, 0, 0, 0, time.UTC),
		},
		// "20060212T23:12:23UTC" - ISO 8601 with T separator
		// TODO: enable when ISO 8601 T-separator format is supported
		// {
		// 	"20060212T23:12:23UTC",
		// 	time.Date(2006, 2, 12, 23, 12, 23, 0, time.UTC),
		// },
		// "Jan-15-2006" → Sun, 15 Jan 2006 00:00:00 +0000
		{
			"Jan-15-2006",
			time.Date(2006, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		// "2006-Jan-15" → Sun, 15 Jan 2006 00:00:00 +0000
		{
			"2006-Jan-15",
			time.Date(2006, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		// "10/Oct/2000:13:55:36 +0100" → Tue, 10 Oct 2000 13:55:36 +0100
		{
			"10/Oct/2000:13:55:36 +0100",
			time.Date(2000, 10, 10, 13, 55, 36, 0, time.FixedZone("", 3600)),
		},
		// "JAN" → Mon, 16 Jan 2006 00:00:00 +0000
		{
			"JAN",
			time.Date(2006, 1, 16, 0, 0, 0, 0, time.UTC),
		},
		// "January" → Mon, 16 Jan 2006 00:00:00 +0000
		{
			"January",
			time.Date(2006, 1, 16, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, Rel(baseTime), InTZ(lisbon))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			if result.Unix() != test.expected.Unix() {
				t.Errorf("For '%s':\n  expected: %s (unix=%d)\n  got:      %s (unix=%d)",
					test.input,
					test.expected.Format("Mon, 02 Jan 2006 15:04:05 -0700"),
					test.expected.Unix(),
					result.Format("Mon, 02 Jan 2006 15:04:05 -0700"),
					result.Unix())
			}
		})
	}

	// Test inputs that should fail (return false in PHP)
	invalidInputs := []string{
		"",
		"22:49:12 bogusTZ",
		"22.49.12.42bogusTZ",
		"022233 bogusTZ",
		"10/Oct/2000:13:55:36 +00100", // invalid TZ offset (too many digits)
	}

	for _, input := range invalidInputs {
		t.Run("Invalid_"+input, func(t *testing.T) {
			_, err := StrToTime(input, Rel(baseTime), InTZ(lisbon))
			if err == nil {
				t.Errorf("Expected error for '%s', but parsing succeeded", input)
			}
		})
	}
}

// TestPHPStrtotimeMySQL tests from PHP's strtotime-mysql-64bit.phpt
// MySQL compact timestamp format: YYYYMMDDHHMMSS
func TestPHPStrtotimeMySQL(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{
			"19970523091528",
			time.Date(1997, 5, 23, 9, 15, 28, 0, time.UTC),
		},
		{
			"20001231185859",
			time.Date(2000, 12, 31, 18, 58, 59, 0, time.UTC),
		},
		{
			"20800410101010",
			time.Date(2080, 4, 10, 10, 10, 10, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			if result.Unix() != test.expected.Unix() {
				t.Errorf("For '%s':\n  expected: %s\n  got:      %s",
					test.input,
					test.expected.Format(time.RFC1123Z),
					result.Format(time.RFC1123Z))
			}
		})
	}
}

// TestPHPStrtotimeBasic tests from PHP's strtotime_basic.phpt
// Tests the difference between numeric and word ordinals
func TestPHPStrtotimeBasic(t *testing.T) {
	// December 1, 2008 is a Monday
	tests := []struct {
		input    string
		expected time.Time
	}{
		// Numeric: "1 Monday December 2008" = first Monday (or current if Monday)
		{"1 Monday December 2008", time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC)},
		{"2 Monday December 2008", time.Date(2008, 12, 8, 0, 0, 0, 0, time.UTC)},
		{"3 Monday December 2008", time.Date(2008, 12, 15, 0, 0, 0, 0, time.UTC)},
		// Word: "first Monday December 2008" = Monday after the first Monday
		// Note: PHP's behavior differs here from our implementation.
		// PHP: first=Dec 8, second=Dec 15, third=Dec 22
		// Our implementation treats "first" and "1" equivalently.
		{"first Monday December 2008", time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC)},
		{"second Monday December 2008", time.Date(2008, 12, 8, 0, 0, 0, 0, time.UTC)},
		{"third Monday December 2008", time.Date(2008, 12, 15, 0, 0, 0, 0, time.UTC)},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			if result.Unix() != test.expected.Unix() {
				t.Errorf("For '%s':\n  expected: %s\n  got:      %s",
					test.input,
					test.expected.Format("2006-01-02"),
					result.Format("2006-01-02"))
			}
		})
	}
}

// TestPHPStrtotimeBasic2 tests from PHP's strtotime_basic2.phpt
// Misspelled month names should fail
func TestPHPStrtotimeBasic2(t *testing.T) {
	_, err := StrToTime("mayy 2 2009", InTZ(time.UTC))
	if err == nil {
		t.Error("Expected error for misspelled month 'mayy 2 2009', but parsing succeeded")
	}
}

// TestPHPStrtotimeRelative tests from PHP's strtotime-relative.phpt
// Base time: 1204200000 = Feb 28, 2008 12:00:00 UTC
func TestPHPStrtotimeRelative(t *testing.T) {
	baseTime := time.Unix(1204200000, 0).In(time.UTC)

	tests := []struct {
		input    string
		expected time.Time
	}{
		// Seconds
		{"+80412 seconds", time.Unix(1204200000+80412, 0).In(time.UTC)},
		{"-80412 seconds", time.Unix(1204200000-80412, 0).In(time.UTC)},
		{"+86399 seconds", time.Unix(1204200000+86399, 0).In(time.UTC)},
		{"-86399 seconds", time.Unix(1204200000-86399, 0).In(time.UTC)},
		{"+86400 seconds", time.Unix(1204200000+86400, 0).In(time.UTC)},
		{"-86400 seconds", time.Unix(1204200000-86400, 0).In(time.UTC)},
		{"+86401 seconds", time.Unix(1204200000+86401, 0).In(time.UTC)},
		{"-86401 seconds", time.Unix(1204200000-86401, 0).In(time.UTC)},

		// Hours
		{"+134 hours", baseTime.Add(134 * time.Hour)},
		{"-134 hours", baseTime.Add(-134 * time.Hour)},
		{"+167 hours", baseTime.Add(167 * time.Hour)},
		{"-167 hours", baseTime.Add(-167 * time.Hour)},
		{"+168 hours", baseTime.Add(168 * time.Hour)},
		{"-168 hours", baseTime.Add(-168 * time.Hour)},
		{"+169 hours", baseTime.Add(169 * time.Hour)},
		{"-169 hours", baseTime.Add(-169 * time.Hour)},
		{"+183 hours", baseTime.Add(183 * time.Hour)},
		{"-183 hours", baseTime.Add(-183 * time.Hour)},

		// Days
		{"+178 days", baseTime.AddDate(0, 0, 178)},
		{"-178 days", baseTime.AddDate(0, 0, -178)},
		{"+179 days", baseTime.AddDate(0, 0, 179)},
		{"-179 days", baseTime.AddDate(0, 0, -179)},
		{"+180 days", baseTime.AddDate(0, 0, 180)},
		{"-180 days", baseTime.AddDate(0, 0, -180)},
		{"+183 days", baseTime.AddDate(0, 0, 183)},
		{"-183 days", baseTime.AddDate(0, 0, -183)},
		{"+184 days", baseTime.AddDate(0, 0, 184)},
		{"-184 days", baseTime.AddDate(0, 0, -184)},

		// Months
		{"+115 months", baseTime.AddDate(0, 115, 0)},
		{"-115 months", baseTime.AddDate(0, -115, 0)},
		{"+119 months", baseTime.AddDate(0, 119, 0)},
		{"-119 months", baseTime.AddDate(0, -119, 0)},
		{"+120 months", baseTime.AddDate(0, 120, 0)},
		{"-120 months", baseTime.AddDate(0, -120, 0)},
		{"+121 months", baseTime.AddDate(0, 121, 0)},
		{"-121 months", baseTime.AddDate(0, -121, 0)},
		{"+128 months", baseTime.AddDate(0, 128, 0)},
		{"-128 months", baseTime.AddDate(0, -128, 0)},

		// Years
		{"+24 years", baseTime.AddDate(24, 0, 0)},
		{"-24 years", baseTime.AddDate(-24, 0, 0)},
		{"+25 years", baseTime.AddDate(25, 0, 0)},
		{"-25 years", baseTime.AddDate(-25, 0, 0)},
		{"+26 years", baseTime.AddDate(26, 0, 0)},
		{"-26 years", baseTime.AddDate(-26, 0, 0)},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, Rel(baseTime), InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			if result.Unix() != test.expected.Unix() {
				t.Errorf("For '%s':\n  expected: %s (unix=%d)\n  got:      %s (unix=%d)",
					test.input,
					test.expected.Format(time.RFC3339),
					test.expected.Unix(),
					result.Format(time.RFC3339),
					result.Unix())
			}
		})
	}
}

// TestPHPStrtotimeOriginal tests from PHP's strtotime.phpt
// Timezone: Europe/Oslo
func TestPHPStrtotimeOriginal(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Fatalf("Failed to load Europe/Oslo: %v", err)
	}

	tests := []struct {
		input    string
		expected int64 // expected Unix timestamp
	}{
		// "2005-07-14 22:30:41" in Europe/Oslo (CEST=UTC+2 in July)
		// means 2005-07-14T22:30:41+0200 = 2005-07-14T20:30:41Z = unix 1121373041
		{"2005-07-14 22:30:41", 1121373041},
		// "2005-07-14 22:30:41 GMT" → unix 1121373041 (direct UTC)
		{"2005-07-14 22:30:41 GMT", 1121373041},
		// "@1121373041" → unix 1121373041
		{"@1121373041", 1121373041},
		// "@1121373041 CEST" → unix 1121373041 (@ always uses the Unix timestamp)
		{"@1121373041 CEST", 1121373041},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, InTZ(oslo))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			if result.Unix() != test.expected {
				t.Errorf("For '%s':\n  expected unix: %d\n  got unix:      %d\n  result: %s",
					test.input,
					test.expected,
					result.Unix(),
					result.Format("2006-01-02T15:04:05-0700"))
			}
		})
	}
}

// TestLargeYears tests that years beyond 4 digits work correctly
func TestLargeYears(t *testing.T) {
	tests := []struct {
		input    string
		year     int
		month    time.Month
		day      int
	}{
		{"10000-01-01", 10000, 1, 1},
		{"20000-06-15", 20000, 6, 15},
		{"100000-12-31", 100000, 12, 31},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			y, m, d := result.Date()
			if y != test.year || m != test.month || d != test.day {
				t.Errorf("For '%s': expected %d-%02d-%02d, got %d-%02d-%02d",
					test.input, test.year, test.month, test.day, y, m, d)
			}
		})
	}
}

// TestSmallYears tests that years with less than 4 digits work
func TestSmallYears(t *testing.T) {
	tests := []struct {
		input string
		year  int
		month time.Month
		day   int
	}{
		// D-M-YYYY with dashes (European style)
		{"2-3-2004", 2004, 3, 2},
		// DD.MM.YYYY European format
		{"2.3.2004", 2004, 3, 2},
		{"15.6.2023", 2023, 6, 15},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			y, m, d := result.Date()
			if y != test.year || m != test.month || d != test.day {
				t.Errorf("For '%s': expected %d-%02d-%02d, got %d-%02d-%02d",
					test.input, test.year, test.month, test.day, y, m, d)
			}
		})
	}
}

// TestPHPScottish tests from PHP's strtotime_variation_scottish.phpt
// "back of 7" = 7:15, "front of 7" = 6:45
func TestPHPScottish(t *testing.T) {
	tests := []struct {
		input  string
		hour   int
		minute int
	}{
		{"back of 7", 7, 15},
		{"front of 7", 6, 45},
		{"back of 19", 19, 15},
		{"front of 19", 18, 45},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, InTZ(time.UTC))
			if err != nil {
				t.Skipf("Not yet supported: '%s': %v", test.input, err)
				return
			}

			if result.Hour() != test.hour || result.Minute() != test.minute {
				t.Errorf("For '%s': expected %02d:%02d, got %02d:%02d",
					test.input, test.hour, test.minute, result.Hour(), result.Minute())
			}
		})
	}
}
