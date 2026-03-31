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

	// Handle AM/PM — check if rest ends with AM or PM (possibly attached to time)
	ampm := ""
	restLower := strings.ToLower(rest)
	if strings.HasSuffix(restLower, "am") || strings.HasSuffix(restLower, "pm") {
		ampm = restLower[len(restLower)-2:]
		rest = strings.TrimSpace(rest[:len(rest)-2])
	} else {
		// Check for " AM" or " PM" as separate word
		upperRest := strings.ToUpper(rest)
		if strings.HasSuffix(upperRest, " AM") || strings.HasSuffix(upperRest, " PM") {
			ampm = strings.ToLower(upperRest[len(upperRest)-2:])
			rest = strings.TrimSpace(rest[:len(rest)-3])
		}
	}

	// Parse time using the ISO 8601 time parser (handles HH:MM:SS and fractional seconds)
	hour, minute, second, nanos, consumed, ok := parseISO8601Time(rest)
	if !ok {
		return time.Time{}, false
	}

	// Apply AM/PM
	if ampm != "" {
		hour = applyAMPM(hour, ampm)
	}

	// Parse the date — try ISO format first, then month-name format
	t, dateOk := parseISOFormat(datePart, loc)
	if !dateOk {
		t, dateOk = parseMonthNameFormat(datePart, loc)
		if !dateOk {
			return time.Time{}, false
		}
	}

	// Check for timezone offset after the time
	tzLoc := loc
	tzRest := rest[consumed:]
	if len(tzRest) > 0 {
		tzStr := strings.TrimSpace(tzRest)
		if parsed, _, ok := parseNumericTimezoneOffset(tzStr); ok {
			tzLoc = parsed
		} else if len(tzStr) > 0 {
			// Try named timezone (abbreviation or full name)
			if parsed, found := tryParseTimezone(tzStr); found {
				tzLoc = parsed
			} else {
				return time.Time{}, false
			}
		}
	}

	return time.Date(t.Year(), t.Month(), t.Day(), hour, minute, second, nanos, tzLoc), true
}

// parseYearMonthFormat parses "YYYY-MM" or "YYYY-M" as year-month (day defaults to 1)
func parseYearMonthFormat(str string, loc *time.Location) (time.Time, bool) {
	if strings.Count(str, "-") != 1 {
		return time.Time{}, false
	}

	parts := strings.SplitN(str, "-", 2)
	if len(parts) != 2 || !isAllDigits(parts[0]) || !isAllDigits(parts[1]) {
		return time.Time{}, false
	}
	if len(parts[0]) < 4 {
		return time.Time{}, false
	}

	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])

	if month < 1 || month > 12 {
		return time.Time{}, false
	}

	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc), true
}

// parseNegativeYear parses "-YYYY-MM-DD" format (negative year)
func parseNegativeYear(str string, loc *time.Location) (time.Time, bool) {
	if len(str) < 2 || str[0] != '-' {
		return time.Time{}, false
	}
	// Must have format -YYYY-MM-DD (at least -Y-M-D)
	rest := str[1:]
	if strings.Count(rest, "-") != 2 {
		return time.Time{}, false
	}
	parts := strings.Split(rest, "-")
	if len(parts) != 3 {
		return time.Time{}, false
	}
	if !isAllDigits(parts[0]) || !isAllDigits(parts[1]) || !isAllDigits(parts[2]) {
		return time.Time{}, false
	}
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return time.Time{}, false
	}
	return time.Date(-year, time.Month(month), day, 0, 0, 0, 0, loc), true
}

// parseZeroDate handles the special case "0000-00-00 ..." which PHP maps to -0001-11-30
func parseZeroDate(str string, loc *time.Location) (time.Time, bool) {
	trimmed := strings.TrimSpace(str)
	if !strings.HasPrefix(trimmed, "0000-00-00") {
		return time.Time{}, false
	}
	// Return year 0, month 1, day 1 (Go's zero-ish date)
	return time.Date(0, 1, 1, 0, 0, 0, 0, loc), true
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
