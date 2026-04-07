package strtotime

import (
	"testing"
	"time"
)

func TestCompactTimestamp(t *testing.T) {
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
			result, ok := parseCompactTimestamp(test.input, time.UTC)
			if !ok {
				t.Fatalf("Failed to parse '%s'", test.input)
			}

			// Compare the parsed time with the expected time
			if !result.Equal(test.expected) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
			phpVerify(t, test.input, result, time.Time{}, time.UTC)
		})
	}
}

func TestMonthNameFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{
			"Jan-15-2006",
			time.Date(2006, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			"Feb-28-2023",
			time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			"2006-Jan-15",
			time.Date(2006, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			"2023-Aug-31",
			time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, ok := parseMonthNameFormat(test.input, time.UTC)
			if !ok {
				t.Fatalf("Failed to parse '%s'", test.input)
			}

			// Compare the parsed time with the expected time
			if !result.Equal(test.expected) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
			phpVerify(t, test.input, result, time.Time{}, time.UTC)
		})
	}
}

func TestHTTPLogFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{
			"10/Oct/2000:13:55:36 +0100",
			time.Date(2000, 10, 10, 13, 55, 36, 0, time.FixedZone("", 3600)),
		},
		{
			"01/Jan/2023:00:00:00 -0700",
			time.Date(2023, 1, 1, 0, 0, 0, 0, time.FixedZone("", -25200)),
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, ok := parseHTTPLogFormat(test.input, time.UTC)
			if !ok {
				t.Fatalf("Failed to parse '%s'", test.input)
			}

			// Compare the parsed time with the expected time
			if !result.Equal(test.expected) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
			phpVerify(t, test.input, result, time.Time{}, time.UTC)
		})
	}
}

func TestNumberedWeekday(t *testing.T) {
	// Reference date: 2008-12-01 is a Monday
	reference := time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		input    string
		expected time.Time
	}{
		{
			"1 Monday December 2008",
			time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC), // First Monday in December 2008
		},
		{
			"first Monday December 2008",
			time.Date(2008, 12, 8, 0, 0, 0, 0, time.UTC), // PHP: "first" without "of" skips past first occurrence
		},
		{
			"2 Monday December 2008",
			time.Date(2008, 12, 8, 0, 0, 0, 0, time.UTC), // Second Monday in December 2008
		},
		{
			"second Monday December 2008",
			time.Date(2008, 12, 15, 0, 0, 0, 0, time.UTC), // PHP: "second" without "of" = 2 weeks past first occurrence
		},
		{
			"third Monday December 2008",
			time.Date(2008, 12, 22, 0, 0, 0, 0, time.UTC), // PHP: "third" without "of" = 3 weeks past first occurrence
		},
		{
			"third Monday of December 2008",
			time.Date(2008, 12, 15, 0, 0, 0, 0, time.UTC), // Third Monday in December 2008
		},
		{
			"last Monday December 2008",
			time.Date(2008, 11, 24, 0, 0, 0, 0, time.UTC), // PHP: "last Monday December" = last Mon BEFORE Dec
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, ok := parseNumberedWeekday(test.input, reference, time.UTC)
			if !ok {
				t.Fatalf("Failed to parse '%s'", test.input)
			}

			// Compare the parsed time with the expected time
			if !result.Equal(test.expected) {
				t.Errorf("For input '%s': Expected %v, got %v",
					test.input,
					test.expected.Format("2006-01-02 (Monday)"),
					result.Format("2006-01-02 (Monday)"))
			}
			phpVerify(t, test.input, result, reference, time.UTC)
		})
	}
}

func TestNumberedWeekdayRelative(t *testing.T) {
	tests := []struct {
		input     string
		reference time.Time
		expected  time.Time
	}{
		{
			// First Monday of January 2009 = Jan 5
			"first monday of next month",
			time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			// Last Friday of November 2008 = Nov 28
			"last friday of last month",
			time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2008, 11, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			// Second Tuesday of February 2025 = Feb 11
			"second tuesday of next month",
			time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 2, 11, 0, 0, 0, 0, time.UTC),
		},
		{
			// First Wednesday of February 2025 = Feb 5
			"first wednesday of last month",
			time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			// PHP: weekday resolved in base year (2008), then year+1 applied.
			// Dec 1 2008=Mon, next Sunday=Dec 7. +1 year → 2009-12-07 (Mon).
			"first sunday of next year",
			time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2009, 12, 7, 0, 0, 0, 0, time.UTC),
		},
		{
			// PHP: weekday resolved in base year (2009 April), then year-1 applied.
			// April 1 2009=Wed, next Fri=April 3, -7=March 27. -1 year → 2008-03-27.
			"last friday of last year",
			time.Date(2009, 3, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2008, 3, 27, 0, 0, 0, 0, time.UTC),
		},
		{
			// Third Thursday of next month (Feb 2025) = Feb 20
			"third thursday of next month",
			time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, ok := parseNumberedWeekday(test.input, test.reference, time.UTC)
			if !ok {
				t.Fatalf("Failed to parse '%s'", test.input)
			}
			if !result.Equal(test.expected) {
				t.Errorf("For input '%s': Expected %s, got %s",
					test.input,
					test.expected.Format("2006-01-02 (Monday)"),
					result.Format("2006-01-02 (Monday)"))
			}
			phpVerify(t, test.input, result, test.reference, time.UTC)
		})
	}
}

func TestDayOfMonth(t *testing.T) {
	reference := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		input    string
		expected time.Time
	}{
		{
			"first day of december 2025",
			time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"last day of february 2024",
			time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), // leap year
		},
		{
			"last day of february 2025",
			time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC), // non-leap year
		},
		{
			"third day of january 2025",
			time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			"fifth day of march",
			time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			"last day of december 2025",
			time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, ok := parseNumberedWeekday(test.input, reference, time.UTC)
			if !ok {
				t.Fatalf("Failed to parse '%s'", test.input)
			}
			if !result.Equal(test.expected) {
				t.Errorf("For input '%s': Expected %s, got %s",
					test.input,
					test.expected.Format("2006-01-02"),
					result.Format("2006-01-02"))
			}
			phpVerify(t, test.input, result, reference, time.UTC)
		})
	}
}

func TestDayOfMonthRelative(t *testing.T) {
	tests := []struct {
		input     string
		reference time.Time
		expected  time.Time
	}{
		{
			"first day of next month",
			time.Date(2008, 12, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"last day of last month",
			time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			"last day of next month",
			time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), // leap year
		},
		{
			// PHP: "next year" keeps month (June). First day of June 2026
			"first day of next year",
			time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			// PHP: "last year" keeps month (June). Last day of June 2024
			"last day of last year",
			time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, ok := parseNumberedWeekday(test.input, test.reference, time.UTC)
			if !ok {
				t.Fatalf("Failed to parse '%s'", test.input)
			}
			if !result.Equal(test.expected) {
				t.Errorf("For input '%s': Expected %s, got %s",
					test.input,
					test.expected.Format("2006-01-02"),
					result.Format("2006-01-02"))
			}
			phpVerify(t, test.input, result, test.reference, time.UTC)
		})
	}
}

func TestStrToTimeWithExtendedFormats(t *testing.T) {
	// Reference time: December 1, 2008
	reference := time.Date(2008, 12, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		input    string
		expected time.Time
	}{
		{
			"19970523091528",
			time.Date(1997, 5, 23, 9, 15, 28, 0, time.UTC),
		},
		{
			"Jan-15-2006",
			time.Date(2006, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			"2006-Jan-15",
			time.Date(2006, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			"10/Oct/2000:13:55:36 +0100",
			time.Date(2000, 10, 10, 13, 55, 36, 0, time.FixedZone("", 3600)),
		},
		{
			"1 Monday December 2008",
			time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"second Monday December 2008",
			time.Date(2008, 12, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			"last Monday December 2008",
			time.Date(2008, 11, 24, 0, 0, 0, 0, time.UTC),
		},
		{
			"first monday of next month",
			time.Date(2009, 1, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			"last friday of last month",
			time.Date(2008, 11, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			"first day of next month",
			time.Date(2009, 1, 1, 10, 0, 0, 0, time.UTC), // PHP: preserves base time
		},
		{
			"last day of december 2008",
			time.Date(2008, 12, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, Rel(reference), InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			// Compare using Unix timestamps to avoid timezone issues
			expectedTS := test.expected.Unix()
			resultTS := result.Unix()

			if expectedTS != resultTS {
				t.Errorf("For input '%s': Expected %s, got %s",
					test.input,
					test.expected.Format("2006-01-02 15:04:05 MST"),
					result.Format("2006-01-02 15:04:05 MST"))
			}
			phpVerify(t, test.input, result, reference, time.UTC)
		})
	}
}
