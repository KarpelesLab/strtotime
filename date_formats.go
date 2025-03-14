package strtotime

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// parseDateFormat tries to parse a date using a common format.
// This is a generic function that can handle various date formats with different separators and component orders.
func parseDateFormat(str string, format string, loc *time.Location) (time.Time, bool) {
	var yearIdx, monthIdx, dayIdx int
	var separator string

	switch format {
	case "ymd":
		yearIdx, monthIdx, dayIdx = 0, 1, 2
	case "mdy":
		monthIdx, dayIdx, yearIdx = 0, 1, 2
	case "dmy":
		dayIdx, monthIdx, yearIdx = 0, 1, 2
	default:
		return time.Time{}, false
	}

	// Determine the separator based on the first non-digit character
	for _, r := range str {
		if r != '0' && r != '1' && r != '2' && r != '3' && r != '4' && r != '5' && r != '6' && r != '7' && r != '8' && r != '9' {
			separator = string(r)
			break
		}
	}

	if separator == "" {
		return time.Time{}, false
	}

	parts := strings.Split(str, separator)
	if len(parts) != 3 {
		return time.Time{}, false
	}

	// Parse components
	year, err := strconv.Atoi(parts[yearIdx])
	if err != nil {
		return time.Time{}, false
	}

	month, err := strconv.Atoi(parts[monthIdx])
	if err != nil || month < 1 || month > 12 {
		return time.Time{}, false
	}

	day, err := strconv.Atoi(parts[dayIdx])
	if err != nil || day < 1 || day > 31 {
		return time.Time{}, false
	}

	// Handle 2-digit years
	if year < 100 {
		year = parseTwoDigitYear(year)
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
}

// parseISOFormat tries to parse a ISO format date (YYYY-MM-DD)
func parseISOFormat(str string, loc *time.Location) (time.Time, bool) {
	if matched, _ := regexp.MatchString(`^\d{4}-\d{1,2}-\d{1,2}$`, str); matched {
		return parseDateFormat(str, "ymd", loc)
	}
	return time.Time{}, false
}

// parseSlashFormat tries to parse a slash format date (YYYY/MM/DD)
func parseSlashFormat(str string, loc *time.Location) (time.Time, bool) {
	if matched, _ := regexp.MatchString(`^\d{4}/\d{1,2}/\d{1,2}$`, str); matched {
		return parseDateFormat(str, "ymd", loc)
	}
	return time.Time{}, false
}

// parseUSFormat tries to parse a US format date (MM/DD/YYYY)
func parseUSFormat(str string, loc *time.Location) (time.Time, bool) {
	if matched, _ := regexp.MatchString(`^\d{1,2}/\d{1,2}/\d{4}$`, str); matched {
		return parseDateFormat(str, "mdy", loc)
	}
	return time.Time{}, false
}

// parseEuropeanFormat tries to parse a European format date (DD.MM.YY or DD.MM.YYYY)
func parseEuropeanFormat(str string, loc *time.Location) (time.Time, bool) {
	if matched, _ := regexp.MatchString(`^\d{1,2}\.\d{1,2}\.\d{2,4}$`, str); matched {
		return parseDateFormat(str, "dmy", loc)
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
