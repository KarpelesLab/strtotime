package strtotime

import (
	"strconv"
	"strings"
	"time"
	"unicode"
)

// isAllDigits checks if a string contains only ASCII digits
func isAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

// isAlpha checks if a string contains only ASCII letters
func isAlpha(s string) bool {
	for _, c := range s {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return len(s) > 0
}

// parseCompactTimestamp parses timestamp formats like "19970523091528" (YYYYMMDDhhmmss)
func parseCompactTimestamp(str string, loc *time.Location) (time.Time, bool) {
	// Must be exactly 14 digits
	if len(str) != 14 || !isAllDigits(str) {
		return time.Time{}, false
	}

	year, _ := strconv.Atoi(str[0:4])
	month, _ := strconv.Atoi(str[4:6])
	day, _ := strconv.Atoi(str[6:8])
	hour, _ := strconv.Atoi(str[8:10])
	minute, _ := strconv.Atoi(str[10:12])
	second, _ := strconv.Atoi(str[12:14])

	// Validate date components
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return time.Time{}, false
	}

	// Validate time components
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 || second < 0 || second > 59 {
		return time.Time{}, false
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, loc), true
}

// parseMonthNameFormat parses formats like "Jan-15-2006" or "2006-Jan-15"
func parseMonthNameFormat(str string, loc *time.Location) (time.Time, bool) {
	parts := strings.Split(str, "-")
	if len(parts) != 3 {
		return time.Time{}, false
	}

	// Try "Jan-15-2006" format: alpha-digits-4digits
	if isAlpha(parts[0]) && len(parts[0]) >= 3 {
		day, dayErr := strconv.Atoi(parts[1])
		year, yearErr := strconv.Atoi(parts[2])

		if dayErr == nil && yearErr == nil {
			month, ok := getMonthByName(parts[0])
			if ok && IsValidDate(year, int(month), day) {
				return time.Date(year, month, day, 0, 0, 0, 0, loc), true
			}
		}
	}

	// Try "2006-Jan-15" format: 4digits-alpha-digits
	if len(parts[0]) == 4 && isAlpha(parts[1]) && len(parts[1]) >= 3 {
		year, yearErr := strconv.Atoi(parts[0])
		day, dayErr := strconv.Atoi(parts[2])

		if yearErr == nil && dayErr == nil {
			month, ok := getMonthByName(parts[1])
			if ok && IsValidDate(year, int(month), day) {
				return time.Date(year, month, day, 0, 0, 0, 0, loc), true
			}
		}
	}

	return time.Time{}, false
}

// parseHTTPLogFormat parses formats like "10/Oct/2000:13:55:36 +0100"
func parseHTTPLogFormat(str string, loc *time.Location) (time.Time, bool) {
	// Find the space separating datetime from timezone offset
	spaceIdx := strings.IndexByte(str, ' ')
	if spaceIdx < 0 {
		return time.Time{}, false
	}

	datePart := str[:spaceIdx]
	tzOffset := strings.TrimSpace(str[spaceIdx+1:])

	// Parse date part: DD/Mon/YYYY:HH:MM:SS
	slash1 := strings.IndexByte(datePart, '/')
	if slash1 < 0 || slash1 < 1 || slash1 > 2 {
		return time.Time{}, false
	}

	slash2 := strings.IndexByte(datePart[slash1+1:], '/')
	if slash2 < 0 {
		return time.Time{}, false
	}
	slash2 += slash1 + 1

	// Month part must be exactly 3 letters
	monthStr := datePart[slash1+1 : slash2]
	if len(monthStr) != 3 || !isAlpha(monthStr) {
		return time.Time{}, false
	}

	// After second slash: YYYY:HH:MM:SS
	rest := datePart[slash2+1:]
	colon1 := strings.IndexByte(rest, ':')
	if colon1 < 0 {
		return time.Time{}, false
	}

	yearStr := rest[:colon1]
	if len(yearStr) != 4 {
		return time.Time{}, false
	}

	timeStr := rest[colon1+1:]
	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 3 {
		return time.Time{}, false
	}

	// Validate all parts are digits with correct lengths
	if len(timeParts[0]) != 2 || len(timeParts[1]) != 2 || len(timeParts[2]) != 2 {
		return time.Time{}, false
	}

	day, dayErr := strconv.Atoi(datePart[:slash1])
	year, yearErr := strconv.Atoi(yearStr)
	hour, hourErr := strconv.Atoi(timeParts[0])
	minute, minErr := strconv.Atoi(timeParts[1])
	second, secErr := strconv.Atoi(timeParts[2])

	if dayErr != nil || yearErr != nil || hourErr != nil || minErr != nil || secErr != nil {
		return time.Time{}, false
	}

	month, ok := getMonthByName(monthStr)
	if !ok {
		return time.Time{}, false
	}

	if !IsValidDate(year, int(month), day) {
		return time.Time{}, false
	}

	if !IsValidTime(hour, minute, second) {
		return time.Time{}, false
	}

	// Parse the timezone offset (format: "+0100" or "-0500")
	if len(tzOffset) != 5 || (tzOffset[0] != '+' && tzOffset[0] != '-') {
		return time.Time{}, false
	}

	tzHour, tzHourErr := strconv.Atoi(tzOffset[1:3])
	tzMin, tzMinErr := strconv.Atoi(tzOffset[3:5])

	if tzHourErr != nil || tzMinErr != nil || tzHour < 0 || tzHour > 23 || tzMin < 0 || tzMin > 59 {
		return time.Time{}, false
	}

	tzOffsetSeconds := tzHour*3600 + tzMin*60
	if tzOffset[0] == '-' {
		tzOffsetSeconds = -tzOffsetSeconds
	}

	tz := time.FixedZone("", tzOffsetSeconds)
	return time.Date(year, month, day, hour, minute, second, 0, tz), true
}

// parseNumberedWeekday parses formats like "1 Monday December 2008", "second Monday December 2008"
// It handles formats like "first Monday of December 2008" or "3rd Friday of January"
func parseNumberedWeekday(str string, now time.Time, loc *time.Location) (time.Time, bool) {
	fields := strings.Fields(str)
	if len(fields) < 3 {
		return time.Time{}, false
	}

	idx := 0
	var ordinal int

	// Parse the ordinal (numeric or word)
	if n, err := strconv.Atoi(fields[idx]); err == nil {
		if n <= 0 || n > 5 {
			return time.Time{}, false
		}
		ordinal = n
	} else {
		switch strings.ToLower(fields[idx]) {
		case "first", "1st":
			ordinal = 1
		case "second", "2nd":
			ordinal = 2
		case "third", "3rd":
			ordinal = 3
		case "fourth", "4th":
			ordinal = 4
		case "fifth", "5th":
			ordinal = 5
		case "last":
			ordinal = -1
		default:
			return time.Time{}, false
		}
	}
	idx++

	// Parse the day of week
	if idx >= len(fields) {
		return time.Time{}, false
	}
	dayOfWeek := getDayOfWeek(fields[idx])
	if dayOfWeek < 0 {
		return time.Time{}, false
	}
	idx++

	// Skip optional "of"
	if idx < len(fields) && strings.ToLower(fields[idx]) == "of" {
		idx++
	}

	// Parse the month
	if idx >= len(fields) {
		return time.Time{}, false
	}
	month, ok := getMonthByName(fields[idx])
	if !ok {
		return time.Time{}, false
	}
	idx++

	// Parse the optional year
	year := now.Year()
	if idx < len(fields) {
		var err error
		year, err = strconv.Atoi(fields[idx])
		if err != nil || year < 1 || year > 9999 {
			return time.Time{}, false
		}
		idx++
	}

	// Make sure we consumed all fields
	if idx != len(fields) {
		return time.Time{}, false
	}

	// Find the first day of the month
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)

	// Find the first occurrence of the specified day of week
	firstDayOfWeek := int(firstOfMonth.Weekday())
	daysUntilFirst := (dayOfWeek - firstDayOfWeek + 7) % 7

	var resultDay int

	if ordinal > 0 {
		// Calculate the day for the nth occurrence
		resultDay = 1 + daysUntilFirst + (ordinal-1)*7

		// Check if this date exists in the month
		lastDayOfMonth := daysInMonth(year, month)
		if resultDay > lastDayOfMonth {
			return time.Time{}, false
		}
	} else if ordinal == -1 {
		// Handle "last" occurrence
		lastDayOfMonth := daysInMonth(year, month)
		lastOfMonth := time.Date(year, month, lastDayOfMonth, 0, 0, 0, 0, loc)
		lastDayOfWeek := int(lastOfMonth.Weekday())

		if lastDayOfWeek == dayOfWeek {
			resultDay = lastDayOfMonth
		} else {
			daysToSubtract := (lastDayOfWeek - dayOfWeek + 7) % 7
			resultDay = lastDayOfMonth - daysToSubtract
		}
	} else {
		return time.Time{}, false
	}

	return time.Date(year, month, resultDay, 0, 0, 0, 0, loc), true
}
