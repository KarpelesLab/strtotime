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
	if err != nil || day < 1 || day > 31 {
		return time.Time{}, false
	}

	// Handle 2-digit years
	if year < 100 {
		year = parseTwoDigitYear(year)
	}

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

// parseISOFormat tries to parse a ISO format date (YYYY-MM-DD)
func parseISOFormat(str string, loc *time.Location) (time.Time, bool) {
	if len(str) >= 8 && len(str) <= 10 && isNumericPattern(str, 4, '-') {
		return parseDateFormat(str, "ymd", loc)
	}
	return time.Time{}, false
}

// parseSlashFormat tries to parse a slash format date (YYYY/MM/DD)
func parseSlashFormat(str string, loc *time.Location) (time.Time, bool) {
	if len(str) >= 8 && len(str) <= 10 && isNumericPattern(str, 4, '/') {
		return parseDateFormat(str, "ymd", loc)
	}
	return time.Time{}, false
}

// parseUSFormat tries to parse a US format date (MM/DD/YYYY)
func parseUSFormat(str string, loc *time.Location) (time.Time, bool) {
	if len(str) >= 8 && len(str) <= 10 && strings.Count(str, "/") == 2 {
		// Check if the last part has 4 digits (for year)
		parts := strings.Split(str, "/")
		if len(parts) == 3 && len(parts[2]) == 4 {
			return parseDateFormat(str, "mdy", loc)
		}
	}
	return time.Time{}, false
}

// parseEuropeanFormat tries to parse a European format date (DD.MM.YY or DD.MM.YYYY)
func parseEuropeanFormat(str string, loc *time.Location) (time.Time, bool) {
	if len(str) >= 6 && len(str) <= 10 && strings.Count(str, ".") == 2 {
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
			return parseDateFormat(str, "dmy", loc)
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
