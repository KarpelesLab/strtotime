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
			goTime, err := StrToTime(input, TZ{Location: loc})
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
