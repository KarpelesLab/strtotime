package strtotime

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	phpOnce      sync.Once
	phpAvailable bool
)

// hasPHP returns true if the php binary is available on the system.
func hasPHP() bool {
	phpOnce.Do(func() {
		_, err := exec.LookPath("php")
		phpAvailable = err == nil
	})
	return phpAvailable
}

// phpLocationName converts a *time.Location to a string that PHP's
// date_default_timezone_set() understands.
func phpLocationName(loc *time.Location) string {
	if loc == nil {
		// No timezone specified — match Go's time.Local
		loc = time.Local
	}
	if loc == time.UTC {
		return "UTC"
	}
	name := loc.String()
	if name != "" && name != "Local" {
		return name
	}
	// "Local" or unnamed — derive from offset
	now := time.Now().In(loc)
	_, offset := now.Zone()
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	h := offset / 3600
	m := (offset % 3600) / 60
	return fmt.Sprintf("%s%02d:%02d", sign, h, m)
}

// phpVerify cross-checks a Go StrToTime result against PHP's strtotime().
// It is a no-op when PHP is not available.
// base should be the Rel() base time (zero value if none).
// loc should be the InTZ() timezone (nil for default).
func phpVerify(t *testing.T, input string, goResult time.Time, base time.Time, loc *time.Location) {
	t.Helper()
	if !hasPHP() {
		return
	}

	// Skip inputs that PHP cannot handle reliably
	if shouldSkipPHPCheck(input, goResult) {
		return
	}

	phpUnix, ok := runPHPStrtotime(t, input, base, loc)
	if !ok {
		return
	}

	if phpUnix == -1 {
		// Go supports more syntax than PHP — log but don't fail
		t.Logf("PHP cross-check: strtotime(%q) returned false, Go returned %v (unix %d)",
			input, goResult, goResult.Unix())
		return
	}

	diff := goResult.Unix() - phpUnix
	if diff < 0 {
		diff = -diff
	}
	if diff > 1 {
		t.Errorf("PHP cross-check mismatch for %q: Go=%d, PHP=%d (diff=%ds)",
			input, goResult.Unix(), phpUnix, goResult.Unix()-phpUnix)
	}
}

// phpVerifyFail cross-checks that PHP also rejects the given input.
func phpVerifyFail(t *testing.T, input string, base time.Time, loc *time.Location) {
	t.Helper()
	if !hasPHP() {
		return
	}

	phpUnix, ok := runPHPStrtotime(t, input, base, loc)
	if !ok {
		return
	}

	if phpUnix != -1 {
		t.Logf("PHP cross-check note: strtotime(%q) = %d in PHP but Go rejects it",
			input, phpUnix)
	}
}

// runPHPStrtotime executes PHP's strtotime and returns the unix timestamp.
// Returns -1 if PHP's strtotime returned false. Returns ok=false on execution errors.
func runPHPStrtotime(t *testing.T, input string, base time.Time, loc *time.Location) (int64, bool) {
	t.Helper()

	var code strings.Builder
	tz := phpLocationName(loc)
	if tz != "" {
		fmt.Fprintf(&code, "date_default_timezone_set(%q);\n", tz)
	}

	if !base.IsZero() {
		fmt.Fprintf(&code, "$r = @strtotime(%q, %d);\n", input, base.Unix())
	} else {
		fmt.Fprintf(&code, "$r = @strtotime(%q);\n", input)
	}
	code.WriteString("echo ($r === false) ? 'FALSE' : $r;\n")

	cmd := exec.Command("php", "-r", code.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("PHP cross-check: execution error for %q: %v (%s)", input, err, string(output))
		return 0, false
	}

	result := strings.TrimSpace(string(output))
	if result == "FALSE" {
		return -1, true
	}

	phpUnix, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		t.Logf("PHP cross-check: unexpected output for %q: %s", input, result)
		return 0, false
	}

	return phpUnix, true
}

// shouldSkipPHPCheck returns true for inputs that cannot be cross-validated.
func shouldSkipPHPCheck(input string, goResult time.Time) bool {
	// Negative or very large years — PHP may overflow
	y := goResult.Year()
	if y < 0 || y > 9999 {
		return true
	}
	// Zero date — PHP handles specially
	if strings.HasPrefix(input, "0000-00-00") {
		return true
	}
	// Inputs starting with "-" followed by a large year (negative year format)
	if len(input) > 1 && input[0] == '-' && len(input) > 5 {
		// Check if this looks like a negative year: -YYYY-MM-DD
		parts := strings.SplitN(input[1:], "-", 2)
		if len(parts) >= 1 {
			if y, err := strconv.Atoi(parts[0]); err == nil && y > 100 {
				return true
			}
		}
	}
	// Expanded positive year format
	if len(input) > 1 && input[0] == '+' {
		parts := strings.SplitN(input[1:], "-", 2)
		if len(parts) >= 1 {
			if y, err := strconv.Atoi(parts[0]); err == nil && y > 9999 {
				return true
			}
		}
	}
	// Unix timestamps with nanosecond precision (>6 decimals) — PHP rejects these
	if len(input) > 1 && input[0] == '@' {
		if dotIdx := strings.Index(input, "."); dotIdx >= 0 {
			fracLen := len(input) - dotIdx - 1
			if fracLen > 6 {
				return true
			}
		}
	}
	// "eighth day" style ordinal + day: PHP parses differently from Go
	if strings.Contains(strings.ToLower(input), "eighth day") ||
		strings.Contains(strings.ToLower(input), "ninth day") ||
		strings.Contains(strings.ToLower(input), "tenth day") {
		return true
	}
	// Compact DDMon format without space — PHP uses base year, our behavior differs
	lower := strings.ToLower(input)
	if len(lower) >= 4 && len(lower) <= 6 && lower[0] >= '0' && lower[0] <= '9' {
		// Check if it's digit(s)+letters like "11oct"
		i := 0
		for i < len(lower) && lower[i] >= '0' && lower[i] <= '9' {
			i++
		}
		if i > 0 && i < len(lower) {
			rest := lower[i:]
			if _, ok := getMonthByName(rest); ok {
				return true
			}
		}
	}
	return false
}

// getPHPTimezone gets the default timezone that PHP is using.
func getPHPTimezone() (string, error) {
	cmd := exec.Command("php", "-r", "echo date_default_timezone_get();")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// abs returns the absolute value of x.
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
