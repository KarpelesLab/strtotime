package strtotime

import (
	"testing"
	"time"
)

func TestISO8601DateTime(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		// Basic T separator with dashes
		{
			"2023-01-15T14:30:00",
			time.Date(2023, 1, 15, 14, 30, 0, 0, time.UTC),
		},
		// T separator with Z (UTC)
		{
			"2023-01-15T14:30:00Z",
			time.Date(2023, 1, 15, 14, 30, 0, 0, time.UTC),
		},
		// T separator with positive offset (+HH:MM)
		{
			"2023-01-15T14:30:00+05:30",
			time.Date(2023, 1, 15, 14, 30, 0, 0, time.FixedZone("", 5*3600+30*60)),
		},
		// T separator with negative offset (-HH:MM)
		{
			"2023-01-15T14:30:00-07:00",
			time.Date(2023, 1, 15, 14, 30, 0, 0, time.FixedZone("", -7*3600)),
		},
		// T separator with compact offset (+HHMM)
		{
			"2023-01-15T14:30:00+0530",
			time.Date(2023, 1, 15, 14, 30, 0, 0, time.FixedZone("", 5*3600+30*60)),
		},
		// Compact date with T separator (YYYYMMDDTHHMMSS)
		{
			"20060212T231223",
			time.Date(2006, 2, 12, 23, 12, 23, 0, time.UTC),
		},
		// Compact date with T separator and named timezone
		{
			"20060212T231223UTC",
			time.Date(2006, 2, 12, 23, 12, 23, 0, time.UTC),
		},
		// T separator with HH:MM (no seconds)
		{
			"2023-06-15T09:30",
			time.Date(2023, 6, 15, 9, 30, 0, 0, time.UTC),
		},
		// T separator with fractional seconds
		{
			"2023-01-15T14:30:00.123Z",
			time.Date(2023, 1, 15, 14, 30, 0, 123000000, time.UTC),
		},
		// T separator with fractional seconds and offset
		{
			"2023-01-15T14:30:00.500+05:30",
			time.Date(2023, 1, 15, 14, 30, 0, 500000000, time.FixedZone("", 5*3600+30*60)),
		},
		// DATE_ATOM / DATE_RFC3339 format
		{
			"2023-01-15T14:30:45+00:00",
			time.Date(2023, 1, 15, 14, 30, 45, 0, time.FixedZone("", 0)),
		},
		// DATE_ISO8601 format (compact offset)
		{
			"2023-01-15T14:30:45+0000",
			time.Date(2023, 1, 15, 14, 30, 45, 0, time.FixedZone("", 0)),
		},
		// RFC3339 with nanoseconds
		{
			"2023-01-15T14:30:45.123456789Z",
			time.Date(2023, 1, 15, 14, 30, 45, 123456789, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, InTZ(time.UTC))
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

			if result.Nanosecond() != test.expected.Nanosecond() {
				t.Errorf("For '%s': expected nanos %d, got %d",
					test.input, test.expected.Nanosecond(), result.Nanosecond())
			}
			phpVerify(t, test.input, result, time.Time{}, time.UTC)
		})
	}
}

func TestISO8601DateTimeInvalid(t *testing.T) {
	invalid := []string{
		"2023-01-15T25:00:00",       // invalid hour
		"2023-01-15T14:60:00",       // invalid minute
		"20060212T231223 bogusTZ",   // invalid named timezone
		"2023-01-15T14:30:00+25:00", // invalid offset hours
	}

	for _, input := range invalid {
		t.Run(input, func(t *testing.T) {
			_, err := StrToTime(input, InTZ(time.UTC))
			if err == nil {
				t.Errorf("Expected error for '%s', but parsing succeeded", input)
			}
		})
	}
}

func TestISOWeekDate(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		// 2023-W01 = Monday January 2, 2023
		{
			"2023-W01",
			time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		// 2023-W01-1 = Monday January 2, 2023
		{
			"2023-W01-1",
			time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		// 2023-W01-5 = Friday January 6, 2023
		{
			"2023-W01-5",
			time.Date(2023, 1, 6, 0, 0, 0, 0, time.UTC),
		},
		// 2023-W01-7 = Sunday January 8, 2023
		{
			"2023-W01-7",
			time.Date(2023, 1, 8, 0, 0, 0, 0, time.UTC),
		},
		// Compact form: 2023W01
		{
			"2023W01",
			time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		// Compact form with day: 2023W015
		{
			"2023W015",
			time.Date(2023, 1, 6, 0, 0, 0, 0, time.UTC),
		},
		// 2008-W01 = Monday December 31, 2007 (week 1 of 2008 starts in Dec 2007!)
		{
			"2008-W01",
			time.Date(2007, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		// 2008-W52 = Monday December 22, 2008
		{
			"2008-W52",
			time.Date(2008, 12, 22, 0, 0, 0, 0, time.UTC),
		},
		// 2004-W53 (2004 has 53 weeks) = Monday December 27, 2004
		{
			"2004-W53",
			time.Date(2004, 12, 27, 0, 0, 0, 0, time.UTC),
		},
		// Week with single digit
		{
			"2023-W3",
			time.Date(2023, 1, 16, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			if !result.Equal(test.expected) {
				t.Errorf("For '%s':\n  expected: %s\n  got:      %s",
					test.input,
					test.expected.Format("2006-01-02 (Monday)"),
					result.Format("2006-01-02 (Monday)"))
			}
			phpVerify(t, test.input, result, time.Time{}, time.UTC)
		})
	}
}

func TestISOWeekDateInvalid(t *testing.T) {
	invalid := []string{
		"2023-W00",   // week 0 invalid
		"2023-W54",   // week 54 invalid
		"2023-W01-0", // day 0 invalid
		"2023-W01-8", // day 8 invalid
	}

	for _, input := range invalid {
		t.Run(input, func(t *testing.T) {
			_, err := StrToTime(input, InTZ(time.UTC))
			if err == nil {
				t.Errorf("Expected error for '%s', but parsing succeeded", input)
			}
		})
	}
}

func TestNumericTimezoneOffset(t *testing.T) {
	tests := []struct {
		input        string
		expectedUnix int64
	}{
		// Space-separated datetime with Z
		{"2023-01-15 14:30:00Z", 1673793000},
		// DateTime with +HH:MM offset
		{"2023-01-15T14:30:00+05:30", 1673773200},
		// DateTime with -HHMM offset
		{"2023-01-15T14:30:00-0500", 1673811000},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			if result.Unix() != test.expectedUnix {
				t.Errorf("For '%s': expected unix %d, got %d (%s)",
					test.input, test.expectedUnix, result.Unix(),
					result.Format(time.RFC3339))
			}
			phpVerify(t, test.input, result, time.Time{}, time.UTC)
		})
	}
}

// TestPHPStrtotime2DateConstants tests the formats from PHP's strtotime2.phpt
// which verifies that date(FORMAT, t) can be round-tripped through strtotime
func TestPHPStrtotime2DateConstants(t *testing.T) {
	// Use a fixed reference time
	refUnix := int64(1700000000) // 2023-11-14 22:13:20 UTC
	ref := time.Unix(refUnix, 0).In(time.UTC)

	tests := []struct {
		name   string
		format string // Go format string equivalent
	}{
		// DATE_ATOM / DATE_RFC3339 / DATE_W3C: "2023-11-14T22:13:20+00:00"
		{"DATE_ATOM", "2006-01-02T15:04:05-07:00"},
		// DATE_ISO8601: "2023-11-14T22:13:20+0000"
		{"DATE_ISO8601", "2006-01-02T15:04:05-0700"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Format the reference time using the format
			formatted := ref.Format(test.format)
			t.Logf("Formatted: %s", formatted)

			// Parse it back
			result, err := StrToTime(formatted, InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Failed to parse %s format '%s': %v", test.name, formatted, err)
			}

			if result.Unix() != refUnix {
				t.Errorf("Round-trip failed for %s: formatted '%s', expected unix %d, got %d",
					test.name, formatted, refUnix, result.Unix())
			}
			phpVerify(t, formatted, result, time.Time{}, time.UTC)
		})
	}
}
