package strtotime

import (
	"strconv"
	"strings"
	"time"
	"unicode"
)



// parseISOFormat tries to parse a ISO format date (YYYY-MM-DD or D-M-YYYY)
func parseISOFormat(str string, loc *time.Location) (time.Time, bool) {
	if strings.Count(str, "-") != 2 {
		return time.Time{}, false
	}

	parts := strings.Split(str, "-")
	if len(parts) != 3 {
		return time.Time{}, false
	}

	// All parts must be numeric
	for _, p := range parts {
		if !isAllDigits(p) || len(p) == 0 {
			return time.Time{}, false
		}
	}

	first, _ := strconv.Atoi(parts[0])
	second, _ := strconv.Atoi(parts[1])
	third, _ := strconv.Atoi(parts[2])

	var year, month, day int

	if len(parts[0]) >= 4 {
		// YYYY-MM-DD (ISO format)
		year, month, day = first, second, third
	} else if len(parts[2]) >= 4 {
		// D-M-YYYY (European style with dashes)
		day, month, year = first, second, third
	} else {
		// Short year: try as YYYY-MM-DD with small year
		year, month, day = first, second, third
		// Handle 2-digit years
		if year < 100 {
			year = parseTwoDigitYear(year)
		}
	}

	if !IsValidDate(year, month, day) {
		return time.Time{}, false
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
}

// parseSlashFormat tries to parse a slash format date (YYYY/MM/DD)
func parseSlashFormat(str string, loc *time.Location) (time.Time, bool) {
	if strings.Count(str, "/") != 2 {
		return time.Time{}, false
	}

	parts := strings.Split(str, "/")
	if len(parts) != 3 || len(parts[0]) < 4 {
		return time.Time{}, false
	}

	// All parts must be numeric
	for _, p := range parts {
		if !isAllDigits(p) || len(p) == 0 {
			return time.Time{}, false
		}
	}

	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])

	if !IsValidDate(year, month, day) {
		return time.Time{}, false
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
}

// parseUSFormat tries to parse a US format date (MM/DD/YYYY)
func parseUSFormat(str string, loc *time.Location) (time.Time, bool) {
	if strings.Count(str, "/") != 2 {
		return time.Time{}, false
	}

	parts := strings.Split(str, "/")
	if len(parts) != 3 || len(parts[2]) < 4 {
		return time.Time{}, false
	}

	// All parts must be numeric
	for _, p := range parts {
		if !isAllDigits(p) || len(p) == 0 {
			return time.Time{}, false
		}
	}

	month, _ := strconv.Atoi(parts[0])
	day, _ := strconv.Atoi(parts[1])
	year, _ := strconv.Atoi(parts[2])

	if !IsValidDate(year, month, day) {
		return time.Time{}, false
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
}

// parseEuropeanFormat tries to parse a European format date (DD.MM.YY or DD.MM.YYYY)
func parseEuropeanFormat(str string, loc *time.Location) (time.Time, bool) {
	if strings.Count(str, ".") == 2 {
		parts := strings.Split(str, ".")
		if len(parts) == 3 {
			// Validate each part contains only digits
			for _, part := range parts {
				for _, char := range part {
					if !unicode.IsDigit(char) {
						return time.Time{}, false
					}
				}
			}
			
			// Parse the components
			day, dayErr := strconv.Atoi(parts[0])
			month, monthErr := strconv.Atoi(parts[1])
			year, yearErr := strconv.Atoi(parts[2])
			
			// Check for parsing errors
			if yearErr != nil || monthErr != nil || dayErr != nil {
				return time.Time{}, false
			}
			
			// Handle 2-digit years
			if year < 100 {
				year = parseTwoDigitYear(year)
			}
			
			// Validate date components
			if !IsValidDate(year, month, day) {
				return time.Time{}, false
			}
			
			// Valid European format date
			return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
		}
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

// parseDateTimeFormat parses "YYYY-MM-DD HH:MM:SS" and optionally a timezone offset
func parseDateTimeFormat(str string, loc *time.Location) (time.Time, bool) {
	// Find the space separating date from time
	spaceIdx := strings.IndexByte(str, ' ')
	if spaceIdx < 0 {
		return time.Time{}, false
	}

	datePart := str[:spaceIdx]
	rest := strings.TrimSpace(str[spaceIdx+1:])

	// Parse time using the ISO 8601 time parser (handles HH:MM:SS and fractional seconds)
	hour, minute, second, nanos, consumed, ok := parseISO8601Time(rest)
	if !ok {
		return time.Time{}, false
	}

	// Parse the date
	t, dateOk := parseISOFormat(datePart, loc)
	if !dateOk {
		return time.Time{}, false
	}

	// Check for timezone offset after the time
	tzLoc := loc
	tzRest := rest[consumed:]
	if len(tzRest) > 0 {
		if parsed, _, ok := parseNumericTimezoneOffset(tzRest); ok {
			tzLoc = parsed
		} else {
			// Not a numeric offset — might be handled by parseWithTimezone later
			// Only match if there's no trailing content
			if len(strings.TrimSpace(tzRest)) > 0 {
				return time.Time{}, false
			}
		}
	}

	return time.Date(t.Year(), t.Month(), t.Day(), hour, minute, second, nanos, tzLoc), true
}

// splitDateAndRest splits a string into a date portion and the rest after whitespace.
// It validates the date portion looks like a recognized date format.
func splitDateAndRest(str string) (string, string, bool) {
	// Find the first whitespace
	spaceIdx := strings.IndexByte(str, ' ')
	if spaceIdx < 0 {
		return "", "", false
	}

	datePart := str[:spaceIdx]
	rest := strings.TrimSpace(str[spaceIdx+1:])
	if rest == "" {
		return "", "", false
	}

	// Validate the date part looks like one of our recognized formats
	if looksLikeDateFormat(datePart) {
		return datePart, rest, true
	}

	return "", "", false
}

// looksLikeDateFormat checks if a string looks like a date format we recognize
func looksLikeDateFormat(s string) bool {
	// YYYY-M-D
	if strings.Count(s, "-") == 2 {
		parts := strings.Split(s, "-")
		if len(parts) == 3 && isAllDigits(parts[0]) && isAllDigits(parts[1]) && isAllDigits(parts[2]) {
			return true
		}
	}

	// YYYY/M/D or M/D/YYYY
	if strings.Count(s, "/") == 2 {
		parts := strings.Split(s, "/")
		if len(parts) == 3 && isAllDigits(parts[0]) && isAllDigits(parts[1]) && isAllDigits(parts[2]) {
			return true
		}
	}

	// DD.MM.YY or DD.MM.YYYY
	if strings.Count(s, ".") == 2 {
		parts := strings.Split(s, ".")
		if len(parts) == 3 && isAllDigits(parts[0]) && isAllDigits(parts[1]) && isAllDigits(parts[2]) {
			return true
		}
	}

	return false
}
