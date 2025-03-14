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
			time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC), // First Monday in December 2008
		},
		{
			"2 Monday December 2008",
			time.Date(2008, 12, 8, 0, 0, 0, 0, time.UTC), // Second Monday in December 2008
		},
		{
			"second Monday December 2008",
			time.Date(2008, 12, 8, 0, 0, 0, 0, time.UTC), // Second Monday in December 2008
		},
		{
			"third Monday December 2008",
			time.Date(2008, 12, 15, 0, 0, 0, 0, time.UTC), // Third Monday in December 2008
		},
		{
			"third Monday of December 2008",
			time.Date(2008, 12, 15, 0, 0, 0, 0, time.UTC), // Third Monday in December 2008
		},
		{
			"last Monday December 2008",
			time.Date(2008, 12, 29, 0, 0, 0, 0, time.UTC), // Last Monday in December 2008
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
			time.Date(2008, 12, 8, 0, 0, 0, 0, time.UTC),
		},
		{
			"last Monday December 2008",
			time.Date(2008, 12, 29, 0, 0, 0, 0, time.UTC),
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
		})
	}
}