package strtotime

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// parseISOFormat tries to parse a ISO format date (YYYY-MM-DD)
func parseISOFormat(str string, loc *time.Location) (time.Time, bool) {
	if matched, _ := regexp.MatchString(`^\d{4}-\d{1,2}-\d{1,2}$`, str); matched {
		parts := strings.Split(str, "-")
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			return time.Time{}, false
		}

		month, err := strconv.Atoi(parts[1])
		if err != nil {
			return time.Time{}, false
		}

		day, err := strconv.Atoi(parts[2])
		if err != nil {
			return time.Time{}, false
		}

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
	}
	return time.Time{}, false
}

// parseSlashFormat tries to parse a slash format date (YYYY/MM/DD)
func parseSlashFormat(str string, loc *time.Location) (time.Time, bool) {
	if matched, _ := regexp.MatchString(`^\d{4}/\d{1,2}/\d{1,2}$`, str); matched {
		parts := strings.Split(str, "/")
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			return time.Time{}, false
		}

		month, err := strconv.Atoi(parts[1])
		if err != nil {
			return time.Time{}, false
		}

		day, err := strconv.Atoi(parts[2])
		if err != nil {
			return time.Time{}, false
		}

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
	}
	return time.Time{}, false
}

// parseUSFormat tries to parse a US format date (MM/DD/YYYY)
func parseUSFormat(str string, loc *time.Location) (time.Time, bool) {
	if matched, _ := regexp.MatchString(`^\d{1,2}/\d{1,2}/\d{4}$`, str); matched {
		parts := strings.Split(str, "/")
		month, err := strconv.Atoi(parts[0])
		if err != nil {
			return time.Time{}, false
		}

		day, err := strconv.Atoi(parts[1])
		if err != nil {
			return time.Time{}, false
		}

		year, err := strconv.Atoi(parts[2])
		if err != nil {
			return time.Time{}, false
		}

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
	}
	return time.Time{}, false
}

// parseEuropeanFormat tries to parse a European format date (DD.MM.YY or DD.MM.YYYY)
func parseEuropeanFormat(str string, loc *time.Location) (time.Time, bool) {
	if matched, _ := regexp.MatchString(`^\d{1,2}\.\d{1,2}\.\d{2,4}$`, str); matched {
		parts := strings.Split(str, ".")
		day, err := strconv.Atoi(parts[0])
		if err != nil {
			return time.Time{}, false
		}

		month, err := strconv.Atoi(parts[1])
		if err != nil {
			return time.Time{}, false
		}

		year, err := strconv.Atoi(parts[2])
		if err != nil {
			return time.Time{}, false
		}

		// Handle 2-digit years (YY)
		if year < 100 {
			// Special case for two-digit years
			if year < 70 {
				year += 2000 // 00-69 -> 2000-2069
			} else {
				year += 1900 // 70-99 -> 1970-1999
			}
		}

		// Validate the date components
		if month < 1 || month > 12 {
			return time.Time{}, false
		}

		// Simple validation for days - better validation would check days per month
		if day < 1 || day > 31 {
			return time.Time{}, false
		}

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
	}
	return time.Time{}, false
}

// parseTwoDigitYear normalizes 2-digit years according to standard practice
func parseTwoDigitYear(year int) int {
	if year < 100 {
		if year < 70 {
			return year + 2000 // 00-69 -> 2000-2069
		}
		return year + 1900 // 70-99 -> 1970-1999
	}
	return year
}
