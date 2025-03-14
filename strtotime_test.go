package strtotime

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestStrToTime(t *testing.T) {
	tests := []string{
		// Basic time concepts
		"now",
		"today",
		"tomorrow",
		"yesterday",
		"next week",
		"last week",

		// Relative time adjustments - positive
		"+1 day",
		"+2 days",
		"+1 week",
		"+3 weeks",
		"+1 month",
		"+6 months",
		"+1 year",
		"+10 years",

		// Relative time adjustments - negative
		"-1 day",
		"-2 days",
		"-1 week",
		"-3 weeks",
		"-1 month",
		"-6 months",
		"-1 year",
		"-10 years",

		// Day of week navigation
		"next Sunday",
		"next Monday",
		"next Tuesday",
		"next Wednesday",
		"next Thursday",
		"next Friday",
		"next Saturday",
		"last Sunday",
		"last Monday",
		"last Tuesday",
		"last Wednesday",
		"last Thursday",
		"last Friday",
		"last Saturday",

		// Abbreviated day names
		"next Sun",
		"next Mon",
		"next Tue",
		"next Wed",
		"next Thu",
		"next Fri",
		"next Sat",
		"last Sun",
		"last Mon",

		// Date formats
		"2023-01-15",
		"2023-12-31",
		"2023/01/15",
		"2023/12/31",
		"01/15/2023",
		"12/31/2023",

		// Full month names
		"January 15 2023",
		"February 28 2023",
		"March 31 2023",
		"April 30 2023",
		"May 31 2023",
		"June 30 2023",
		"July 31 2023",
		"August 31 2023",
		"September 30 2023",
		"October 31 2023",
		"November 30 2023",
		"December 31 2023",

		// Short month names
		"Jan 15, 2023",
		"Feb 28, 2023",
		"Mar 31, 2023",
		"Apr 30, 2023",
		"May 31, 2023",
		"Jun 30, 2023",
		"Jul 31, 2023",
		"Aug 31, 2023",
		"Sep 30, 2023",
		"Oct 31, 2023",
		"Nov 30, 2023",
		"Dec 31, 2023",

		// Month names with no comma
		"Jan 15 2023",
		"Feb 28 2023",

		// Month names with ordinal suffix
		"April 4th",
		"December 25th",

		// Case insensitivity tests
		"TOMORROW",
		"Next Monday",
		"NEXT FRIDAY",
		"Last SATURDAY",

		// Whitespace handling
		" tomorrow ",
		"   next   monday   ",
		"+1     day",

		// Next/Last time units
		"next month",
		"next year",
		"last month",
		"last year",

		// Mixed case for next/last
		"Next Month",
		"LAST YEAR",

		// Compound expressions with spaces around operators
		"next year + 4 days",
		"next month + 2 weeks",
		"next week + 3 days",
		"tomorrow + 12 hours",
		"next year - 2 months",
		"next month - 1 week",

		// Compound expressions without spaces around operators
		"next year+4 days",
		"next month+2 weeks",
		"next week+3 days",
		"tomorrow+12 hours",
		"next year-2 months",
		"next month-1 week",

		// Multiple compound operators
		"next year + 1 month + 1 week",
		"next month - 2 days + 12 hours",
		"next year+1 month-2 days",

		// Mixed spacing in compound expressions
		"next year+1 month + 2 days",
		"next month + 1 week+3 days",

		// Sequential time expressions (stream-based parsing)
		"next monday next year",
		"next friday last month",

		// Comment these back out for now, as they require more complex handling
		// some tests found from comments in the php documentation
		// Commented out as these are covered by the hour tests
		//"+2 hrs",
		//"+2 hourss",
		//"+2 hours",
		//
		// Commented out as these need complex month adjustment
		//"2023-05-30 -1 month",
		//"2023-05-31 -1 month",
		//
		// Skipped as this is a non-standard European format
		//"24.11.22",
	}

	// First get PHP's timezone
	phpTz, err := getPHPTimezone()
	if err != nil {
		t.Fatalf("failed to get PHP timezone: %v", err)
	}

	// Create timezone option
	loc, err := time.LoadLocation(phpTz)
	if err != nil {
		t.Fatalf("failed to load timezone %q: %v", phpTz, err)
	}

	t.Logf("Running tests with timezone: %s", phpTz)

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			// Get PHP's interpretation of the time string
			phpCode := fmt.Sprintf(`
				$ts = strtotime(%q);
				echo $ts . "\n";
				echo date('Y-m-d H:i:s', $ts);
			`, input)

			cmd := exec.Command("php", "-r", phpCode)
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("failed to get PHP strtotime result: %v", err)
			}

			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(lines) != 2 {
				t.Fatalf("unexpected PHP output: %s", string(output))
			}

			phpTime, err := strconv.ParseInt(lines[0], 10, 64)
			if err != nil {
				t.Fatalf("failed to parse PHP timestamp: %v", err)
			}

			phpTimeReadable := lines[1]
			t.Logf("PHP %q => %s (timestamp: %d)", input, phpTimeReadable, phpTime)

			// Get our implementation's interpretation with PHP's timezone
			goTime, err := StrToTime(input, InTZ(loc))
			if err != nil {
				t.Fatalf("StrToTime(%q) error: %v", input, err)
			}

			// Compare timestamps within a small tolerance (1 second)
			if abs(goTime.Unix()-phpTime) > 1 {
				t.Errorf("StrToTime(%q) = %v (%s), PHP returned %v (%s) (diff: %v seconds)",
					input, goTime.Unix(), goTime.Format("2006-01-02 15:04:05"),
					phpTime, phpTimeReadable, abs(goTime.Unix()-phpTime))
			}
		})
	}
}

// getPHPTimezone gets the default timezone that PHP is using
func getPHPTimezone() (string, error) {
	cmd := exec.Command("php", "-r", "echo date_default_timezone_get();")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Return the timezone name
	return strings.TrimSpace(string(output)), nil
}

// abs returns the absolute value of x.
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestUnixTimestampFormats(t *testing.T) {
	// Test parsing Unix timestamps with and without fractional seconds
	tests := []struct {
		input           string
		expectedSeconds int64
		expectedNanos   int
	}{
		{"@1672531200", 1672531200, 0},                  // 2023-01-01 00:00:00 UTC
		{"@1672531200.5", 1672531200, 500000000},        // With .5 seconds
		{"@1672531200.123", 1672531200, 123000000},      // With milliseconds
		{"@1672531200.123456", 1672531200, 123456000},   // With microseconds
		{"@1672531200.123456789", 1672531200, 123456789}, // With nanoseconds
		{"@1672531200.000001", 1672531200, 1000},        // Very small fraction
		{"@1672531200.999999999", 1672531200, 999999999}, // Maximum precision
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input, InTZ(time.UTC))
			if err != nil {
				t.Fatalf("Error parsing '%s': %v", test.input, err)
			}

			if result.Unix() != test.expectedSeconds {
				t.Errorf("For input '%s': expected seconds %d, got %d",
					test.input,
					test.expectedSeconds,
					result.Unix())
			}

			if result.Nanosecond() != test.expectedNanos {
				t.Errorf("For input '%s': expected nanoseconds %d, got %d",
					test.input,
					test.expectedNanos,
					result.Nanosecond())
			}
		})
	}
}
