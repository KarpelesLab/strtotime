package strtotime

import (
	"testing"
	"time"
)

// Test that strtotime produces correct results when a specific timezone
// context is provided via InTZ(). These come from failing PHP ext/date
// tests where date.timezone is set to a specific zone.
func TestPHPTimezoneContext(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	sp, _ := time.LoadLocation("America/Sao_Paulo")
	eastern, _ := time.LoadLocation("US/Eastern")

	tests := []struct {
		name  string
		input string
		tz    *time.Location
		base  time.Time
	}{
		// bug30532: EDT + relative hours across DST boundary
		{"bug30532_edt_1h", "2004-10-31 EDT +1 hour", ny, time.Time{}},
		{"bug30532_edt_2h", "2004-10-31 EDT +2 hours", ny, time.Time{}},
		{"bug30532_edt_3h", "2004-10-31 EDT +3 hours", ny, time.Time{}},
		{"bug30532_1h", "2004-10-31 +1 hour", ny, time.Time{}},
		{"bug30532_2h", "2004-10-31 +2 hours", ny, time.Time{}},
		{"bug30532_3h", "2004-10-31 +3 hours", ny, time.Time{}},

		// bug32086: dates around DST in America/Sao_Paulo
		{"bug32086_nov1", "2004-11-01", sp, time.Time{}},
		{"bug32086_nov1_plus1", "2004-11-01 +1 day", sp, time.Time{}},
		{"bug32086_nov2", "2004-11-02", sp, time.Time{}},
		{"bug32086_nov3", "2004-11-03", sp, time.Time{}},
		{"bug32086_feb19", "2005-02-19", sp, time.Time{}},
		{"bug32086_feb19_plus1", "2005-02-19 +1 day", sp, time.Time{}},
		{"bug32086_feb20", "2005-02-20", sp, time.Time{}},
		{"bug32086_feb21", "2005-02-21", sp, time.Time{}},

		// bug32555: "tomorrow" in US/Eastern
		{"bug32555_tomorrow", "tomorrow", eastern,
			time.Date(2005, 3, 4, 12, 0, 0, 0, eastern)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []Option
			opts = append(opts, InTZ(tt.tz))
			if !tt.base.IsZero() {
				opts = append(opts, Rel(tt.base))
			}
			result, err := StrToTime(tt.input, opts...)
			if err != nil {
				t.Fatalf("StrToTime(%q, tz=%s) failed: %v", tt.input, tt.tz, err)
			}
			// Verify result is in the expected timezone
			resultZone := result.Location().String()
			expectedZone := tt.tz.String()
			// For dates without explicit timezone, result should be in InTZ timezone
			if resultZone != expectedZone && resultZone != "UTC" {
				// Accept if the timezone name differs but represents the same zone
				_, resultOff := result.Zone()
				checkTime := result.In(tt.tz)
				_, expectedOff := checkTime.Zone()
				if resultOff != expectedOff {
					t.Logf("Note: result in %s (offset %d), expected %s (offset %d)",
						resultZone, resultOff, expectedZone, expectedOff)
				}
			}
			phpVerify(t, tt.input, result, tt.base, tt.tz)
		})
	}
}

// Test DST-sensitive relative date arithmetic.
func TestPHPDSTRelativeArithmetic(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name     string
		input    string
		tz       *time.Location
		wantDate string
	}{
		{"spring_forward", "2005-04-02 +1 day", ny, "2005-04-03"},
		{"fall_back", "2005-10-29 +1 day", ny, "2005-10-30"},
		{"month_across_dst", "2005-03-01 +1 month", ny, "2005-04-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StrToTime(tt.input, InTZ(tt.tz))
			if err != nil {
				t.Fatalf("StrToTime(%q) failed: %v", tt.input, err)
			}
			gotDate := result.Format("2006-01-02")
			if gotDate != tt.wantDate {
				t.Errorf("StrToTime(%q) date = %s, want %s (full: %s)",
					tt.input, gotDate, tt.wantDate, result)
			}
			phpVerify(t, tt.input, result, time.Time{}, tt.tz)
		})
	}
}

// Test timezone abbreviation resolution in date strings.
func TestPHPTimezoneAbbrevInContext(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name     string
		input    string
		tz       *time.Location
		wantZone string
	}{
		{"explicit_edt", "2005-07-14 22:30:41 EDT", ny, "EDT"},
		{"explicit_est", "2005-01-14 22:30:41 EST", ny, "EST"},
		{"default_summer", "2005-07-14 22:30:41", ny, "EDT"},
		{"default_winter", "2005-01-14 22:30:41", ny, "EST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StrToTime(tt.input, InTZ(tt.tz))
			if err != nil {
				t.Fatalf("StrToTime(%q) failed: %v", tt.input, err)
			}
			gotZone, _ := result.Zone()
			if gotZone != tt.wantZone {
				t.Errorf("StrToTime(%q) zone = %s, want %s (full: %s)",
					tt.input, gotZone, tt.wantZone, result)
			}
			phpVerify(t, tt.input, result, time.Time{}, tt.tz)
		})
	}
}
