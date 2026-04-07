package strtotime

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestCSV loads test cases from testdata/strtotime_tests.csv and verifies
// each one against StrToTime and optionally PHP's strtotime.
func TestCSV(t *testing.T) {
	records := loadCSV(t, "testdata/strtotime_tests.csv")

	for i, rec := range records {
		input := rec[0]
		baseUnix, err := strconv.ParseInt(rec[1], 10, 64)
		if err != nil {
			t.Fatalf("line %d: bad base_unix %q: %v", i+2, rec[1], err)
		}
		tz := rec[2]
		expectedUnix, err := strconv.ParseInt(rec[3], 10, 64)
		if err != nil {
			t.Fatalf("line %d: bad expected_unix %q: %v", i+2, rec[3], err)
		}

		name := fmt.Sprintf("%d/%s", i, input)
		t.Run(name, func(t *testing.T) {
			var opts []Option
			var base time.Time
			if baseUnix != 0 {
				base = time.Unix(baseUnix, 0).UTC()
				opts = append(opts, Rel(base))
			}
			var loc *time.Location
			if tz != "" {
				var lerr error
				loc, lerr = csvLoadLocation(tz)
				if lerr != nil {
					t.Fatalf("bad timezone %q: %v", tz, lerr)
				}
				opts = append(opts, InTZ(loc))
			}

			result, perr := StrToTime(input, opts...)
			if perr != nil {
				t.Fatalf("StrToTime(%q) error: %v", input, perr)
			}
			if result.Unix() != expectedUnix {
				t.Errorf("StrToTime(%q) = %d (%s), want %d (%s) [diff=%ds]",
					input, result.Unix(), result.UTC().Format("2006-01-02 15:04:05"),
					expectedUnix, time.Unix(expectedUnix, 0).UTC().Format("2006-01-02 15:04:05"),
					result.Unix()-expectedUnix)
			}
			phpVerify(t, input, result, base, loc)
		})
	}
}

// TestCSVInvalid loads test cases from testdata/strtotime_invalid.csv and
// verifies each one returns an error from StrToTime.
func TestCSVInvalid(t *testing.T) {
	records := loadCSV(t, "testdata/strtotime_invalid.csv")

	for i, rec := range records {
		input := rec[0]
		baseUnix, err := strconv.ParseInt(rec[1], 10, 64)
		if err != nil {
			t.Fatalf("line %d: bad base_unix %q: %v", i+2, rec[1], err)
		}
		tz := rec[2]

		name := fmt.Sprintf("%d/%s", i, input)
		t.Run(name, func(t *testing.T) {
			var opts []Option
			var base time.Time
			if baseUnix != 0 {
				base = time.Unix(baseUnix, 0).UTC()
				opts = append(opts, Rel(base))
			}
			var loc *time.Location
			if tz != "" {
				var lerr error
				loc, lerr = csvLoadLocation(tz)
				if lerr != nil {
					t.Fatalf("bad timezone %q: %v", tz, lerr)
				}
				opts = append(opts, InTZ(loc))
			}

			_, perr := StrToTime(input, opts...)
			if perr == nil {
				t.Errorf("StrToTime(%q) should have returned error", input)
			}
			phpVerifyFail(t, input, base, loc)
		})
	}
}

// loadCSV reads a CSV file and returns all records (skipping the header).
func loadCSV(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comment = '#'
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if len(records) < 2 {
		t.Fatalf("%s: no test cases (only %d rows)", path, len(records))
	}
	return records[1:] // skip header
}

// csvLoadLocation parses a timezone string for CSV test cases.
// Uses the package's loadLocation for IANA names, and handles offset formats.
func csvLoadLocation(tz string) (*time.Location, error) {
	if tz == "" {
		return nil, nil
	}
	// Try offset format like "+05:00" or "-07:00" first
	if len(tz) >= 5 && (tz[0] == '+' || tz[0] == '-') {
		cleaned := strings.Replace(tz, ":", "", 1)
		if len(cleaned) >= 5 {
			h, herr := strconv.Atoi(cleaned[1:3])
			m, merr := strconv.Atoi(cleaned[3:5])
			if herr == nil && merr == nil {
				offset := h*3600 + m*60
				if tz[0] == '-' {
					offset = -offset
				}
				return time.FixedZone(tz, offset), nil
			}
		}
	}
	// Try IANA/abbreviation via package function
	return loadLocation(tz)
}
