package strtotime

import (
	"strings"
	"time"
)

// Common timezone abbreviations
var timezoneAbbreviations = map[string]*time.Location{
	// North American time zones
	"est": mustLoadLocation("America/New_York"),    // Eastern Standard Time (UTC-5)
	"edt": mustLoadLocation("America/New_York"),    // Eastern Daylight Time (UTC-4)
	"cst": mustLoadLocation("America/Chicago"),     // Central Standard Time (UTC-6)
	"cdt": mustLoadLocation("America/Chicago"),     // Central Daylight Time (UTC-5)
	"mst": mustLoadLocation("America/Denver"),      // Mountain Standard Time (UTC-7)
	"mdt": mustLoadLocation("America/Denver"),      // Mountain Daylight Time (UTC-6)
	"pst": mustLoadLocation("America/Los_Angeles"), // Pacific Standard Time (UTC-8)
	"pdt": mustLoadLocation("America/Los_Angeles"), // Pacific Daylight Time (UTC-7)
	"akst": mustLoadLocation("America/Anchorage"),  // Alaska Standard Time (UTC-9)
	"akdt": mustLoadLocation("America/Anchorage"),  // Alaska Daylight Time (UTC-8)
	"hst": mustLoadLocation("Pacific/Honolulu"),    // Hawaii Standard Time (UTC-10)

	// European time zones
	"gmt": mustLoadLocation("Europe/London"),  // Greenwich Mean Time (UTC+0)
	"bst": mustLoadLocation("Europe/London"),  // British Summer Time (UTC+1)
	"iet": mustLoadLocation("Europe/Dublin"),  // Irish Standard Time (UTC+1)
	"cet": mustLoadLocation("Europe/Paris"),   // Central European Time (UTC+1)
	"cest": mustLoadLocation("Europe/Paris"),  // Central European Summer Time (UTC+2)
	"eet": mustLoadLocation("Europe/Helsinki"), // Eastern European Time (UTC+2)
	"eest": mustLoadLocation("Europe/Helsinki"), // Eastern European Summer Time (UTC+3)

	// Australian time zones
	"awst": mustLoadLocation("Australia/Perth"),    // Australian Western Standard Time (UTC+8)
	"acst": mustLoadLocation("Australia/Adelaide"), // Australian Central Standard Time (UTC+9:30)
	"aest": mustLoadLocation("Australia/Sydney"),   // Australian Eastern Standard Time (UTC+10)
	"aedt": mustLoadLocation("Australia/Sydney"),   // Australian Eastern Daylight Time (UTC+11)

	// Asian time zones
	"jst": mustLoadLocation("Asia/Tokyo"),   // Japan Standard Time (UTC+9)
	"ct": mustLoadLocation("Asia/Shanghai"), // China Standard Time (UTC+8)
	"ist": mustLoadLocation("Asia/Kolkata"),  // Indian Standard Time (UTC+5:30)

	// Other common time zones
	"utc": time.UTC,      // Universal Coordinated Time
	"z": time.UTC,        // Z (Zulu time) in ISO format
}

// Common full timezone names
var timezoneNames = map[string]string{
	// North America
	"eastern time": "America/New_York",
	"et": "America/New_York",
	"eastern": "America/New_York",
	"central time": "America/Chicago",
	"ct": "America/Chicago",
	"central": "America/Chicago",
	"mountain time": "America/Denver",
	"mt": "America/Denver",
	"mountain": "America/Denver",
	"pacific time": "America/Los_Angeles",
	"pt": "America/Los_Angeles",
	"pacific": "America/Los_Angeles",
	"alaska time": "America/Anchorage",
	"alaska": "America/Anchorage",
	"hawaii time": "Pacific/Honolulu",
	"hawaii": "Pacific/Honolulu",

	// Europe
	"greenwich mean time": "Europe/London",
	"british time": "Europe/London",
	"british": "Europe/London",
	"western european time": "Europe/London",
	"central european time": "Europe/Paris",
	"eastern european time": "Europe/Helsinki",

	// Australia
	"australian western time": "Australia/Perth",
	"australian central time": "Australia/Adelaide",
	"australian eastern time": "Australia/Sydney",

	// Asia
	"japan time": "Asia/Tokyo",
	"china time": "Asia/Shanghai",
	"india time": "Asia/Kolkata",
	"india": "Asia/Kolkata",

	// Universal
	"universal time": "UTC",
	"universal coordinated time": "UTC",
	"zulu time": "UTC",
	"zulu": "UTC",
}

// mustLoadLocation loads a location or panics, used for package initialization
func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		// Fall back to UTC rather than panicking
		return time.UTC
	}
	return loc
}

// tryParseTimezone attempts to parse a timezone from a string
// It handles both abbreviations (PST, EST) and full names (America/New_York, Europe/Paris)
func tryParseTimezone(tzString string) (*time.Location, bool) {
	// Special case for common IANA timezone names that might appear with varied case
	knownTimezones := map[string]string{
		"america/new_york": "America/New_York",
		"europe/london": "Europe/London",
		"europe/paris": "Europe/Paris",
		"asia/tokyo": "Asia/Tokyo",
		"australia/sydney": "Australia/Sydney",
	}

	// Normalize to lowercase for case-insensitive matching
	tzLower := strings.ToLower(tzString)

	// Check if it's a common abbreviation
	if loc, found := timezoneAbbreviations[tzLower]; found {
		return loc, true
	}

	// Check if it's a common full name
	if tzName, found := timezoneNames[tzLower]; found {
		loc, err := time.LoadLocation(tzName)
		if err == nil {
			return loc, true
		}
	}

	// Check if it's a known IANA timezone with different case
	if standardName, found := knownTimezones[tzLower]; found {
		loc, err := time.LoadLocation(standardName)
		if err == nil {
			return loc, true
		}
	}

	// Try to load directly as a timezone with original case
	loc, err := time.LoadLocation(tzString)
	if err == nil {
		return loc, true
	}

	// Try with title case for region/city format (e.g., "europe/paris" -> "Europe/Paris")
	parts := strings.Split(tzString, "/")
	if len(parts) == 2 {
		region := strings.Title(strings.ToLower(parts[0]))
		city := strings.Title(strings.ToLower(parts[1]))
		tzPropCase := region + "/" + city
		
		loc, err := time.LoadLocation(tzPropCase)
		if err == nil {
			return loc, true
		}
	}

	return nil, false
}