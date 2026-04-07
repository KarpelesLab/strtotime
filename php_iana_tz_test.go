package strtotime

import (
	"testing"
	"time"
)

// TestIANATimezoneNamesInDateStrings tests that all IANA timezone names
// can be used in date strings like "2008-01-01 13:00:00 America/New_York".
// These fail in PHP's bug46111 test.
func TestIANATimezoneNamesInDateStrings(t *testing.T) {
	base := time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC)

	// Timezone names that currently fail due to case normalization or
	// multi-level path parsing issues
	zones := []string{
		// Multi-level paths (3 segments)
		"America/Argentina/Buenos_Aires",
		"America/Argentina/Catamarca",
		"America/Argentina/Cordoba",
		"America/Argentina/Jujuy",
		"America/Argentina/La_Rioja",
		"America/Argentina/Mendoza",
		"America/Argentina/Rio_Gallegos",
		"America/Argentina/Salta",
		"America/Argentina/San_Juan",
		"America/Argentina/San_Luis",
		"America/Argentina/Tucuman",
		"America/Argentina/Ushuaia",
		"America/Indiana/Indianapolis",
		"America/Indiana/Knox",
		"America/Indiana/Marengo",
		"America/Indiana/Petersburg",
		"America/Indiana/Tell_City",
		"America/Indiana/Vevay",
		"America/Indiana/Vincennes",
		"America/Indiana/Winamac",
		"America/Kentucky/Louisville",
		"America/Kentucky/Monticello",
		"America/North_Dakota/Beulah",
		"America/North_Dakota/Center",
		"America/North_Dakota/New_Salem",

		// Underscore names that get mangled by case normalization
		"Africa/Dar_es_Salaam",
		"America/Port_of_Spain",

		// Hyphenated names
		"Africa/Porto-Novo",
		"America/Blanc-Sablon",
		"America/Port-au-Prince",
		"Asia/Ust-Nera",

		// Short/backwards-compat names
		"Australia/ACT",
		"Australia/LHI",
		"Australia/NSW",
		"America/Knox_IN",

		// Special names
		"Antarctica/DumontDUrville",
		"Antarctica/McMurdo",
		"Europe/Isle_of_Man",
		"America/Argentina/ComodRivadavia",
	}

	for _, zone := range zones {
		t.Run(zone, func(t *testing.T) {
			s := "2008-01-01 13:00:00 " + zone
			result, err := StrToTime(s, InTZ(time.UTC), Rel(base))
			if err != nil {
				t.Errorf("StrToTime(%q) failed: %v", s, err)
				return
			}
			if result.IsZero() {
				t.Errorf("StrToTime(%q) returned zero time", s)
			}
			phpVerify(t, s, result, base, time.UTC)
		})
	}
}
