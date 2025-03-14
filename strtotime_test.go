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
	tests := []struct {
		input string
		want  string // Human readable expected result for debugging
	}{
		{"now", "now"},
		{"today", "today"},
		{"tomorrow", "tomorrow"},
		{"yesterday", "yesterday"},
		{"next week", "next week"},
		{"last week", "last week"},
		{"+1 day", "+1 day"},
		{"+1 week", "+1 week"},
		{"+1 month", "+1 month"},
		{"+1 year", "+1 year"},
		{"next Thursday", "next Thursday"},
		{"last Monday", "last Monday"},
		{"2023-01-15", "2023-01-15"},
		{"2023/01/15", "2023/01/15"},
		{"01/15/2023", "01/15/2023"},
		{"January 15 2023", "January 15 2023"},
		{"Jan 15, 2023", "Jan 15, 2023"},
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

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Get PHP's interpretation of the time string
			phpCode := fmt.Sprintf(`
				$ts = strtotime(%q);
				echo $ts . "\n";
				echo date('Y-m-d H:i:s', $ts);
			`, tt.input)

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
			t.Logf("PHP %q => %s (timestamp: %d)", tt.input, phpTimeReadable, phpTime)
			
			// Get our implementation's interpretation with PHP's timezone
			goTime, err := StrToTime(tt.input, TZ{Location: loc})
			if err != nil {
				t.Fatalf("StrToTime(%q) error: %v", tt.input, err)
			}

			// Compare timestamps within a small tolerance (1 second)
			if abs(goTime.Unix()-phpTime) > 1 {
				t.Errorf("StrToTime(%q) = %v (%s), PHP returned %v (%s) (diff: %v seconds)",
					tt.input, goTime.Unix(), goTime.Format("2006-01-02 15:04:05"), 
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