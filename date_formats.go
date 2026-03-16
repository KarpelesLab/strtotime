package strtotime

import (
	"strconv"
	"strings"
	"time"
	"unicode"
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
		if !unicode.IsDigit(r) {
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
	if err != nil || day < 1 {
		return time.Time{}, false
	}
	
	// Handle 2-digit years
	if year < 100 {
		year = parseTwoDigitYear(year)
	}
	
	// Validate the date using our utility function
	if !IsValidDate(year, month, day) {
		return time.Time{}, false
	}
	
	// If we made it here, the date is valid
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
}


// isNumericPattern checks if a string matches a specific pattern of digits with separator
func isNumericPattern(str string, firstPartLen int, separator rune) bool {
	parts := [3]int{0, 0, 0} // Count digits in each part
	partIndex := 0
	
	for _, char := range str {
		if unicode.IsDigit(char) {
			parts[partIndex]++
			continue
		}
		
		// When we hit a separator
		if char == separator {
			// Move to next part
			partIndex++
			// If we already have 3 parts, this format is invalid
			if partIndex > 2 {
				return false
			}
			continue
		}
		
		// Any other character makes this invalid
		return false
	}
	
	// Validate we have exactly 3 parts
	if partIndex != 2 {
		return false
	}
	
	// For YMD format, verify first part is exactly 4 digits (year)
	if firstPartLen > 0 && parts[0] != firstPartLen {
		return false
	}
	
	// Verify all parts have at least 1 digit
	return parts[0] > 0 && parts[1] > 0 && parts[2] > 0
}

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

// parseDateTimeFormat parses "YYYY-MM-DD HH:MM:SS" without regexp
func parseDateTimeFormat(str string, loc *time.Location) (time.Time, bool) {
	// Find the space separating date from time
	spaceIdx := strings.IndexByte(str, ' ')
	if spaceIdx < 0 {
		return time.Time{}, false
	}

	datePart := str[:spaceIdx]
	timePart := strings.TrimSpace(str[spaceIdx+1:])

	// Time part must be exactly H:M:S (no trailing content)
	timeParts := strings.Split(timePart, ":")
	if len(timeParts) != 3 {
		return time.Time{}, false
	}

	// Validate time parts are numeric
	for _, tp := range timeParts {
		if !isAllDigits(tp) || len(tp) == 0 {
			return time.Time{}, false
		}
	}

	hour, _ := strconv.Atoi(timeParts[0])
	minute, _ := strconv.Atoi(timeParts[1])
	second, _ := strconv.Atoi(timeParts[2])

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 || second < 0 || second > 59 {
		return time.Time{}, false
	}

	// Parse the date
	t, ok := parseISOFormat(datePart, loc)
	if !ok {
		return time.Time{}, false
	}

	return time.Date(t.Year(), t.Month(), t.Day(), hour, minute, second, 0, loc), true
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
