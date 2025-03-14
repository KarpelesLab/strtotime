package strtotime

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Pre-compiled regular expressions for better performance
var (
	// Compact timestamp regex: "19970523091528" (YYYYMMDDhhmmss)
	compactTimestampRegex = regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})$`)
	
	// Month name formats
	monthNameMDYRegex = regexp.MustCompile(`^([A-Za-z]{3,})-(\d{1,2})-(\d{4})$`)
	monthNameYMDRegex = regexp.MustCompile(`^(\d{4})-([A-Za-z]{3,})-(\d{1,2})$`)
	
	// HTTP log format: "10/Oct/2000:13:55:36 +0100"
	httpLogRegex = regexp.MustCompile(`^(\d{1,2})/([A-Za-z]{3})/(\d{4}):(\d{2}):(\d{2}):(\d{2})\s+([+-]\d{4})$`)
	
	// Numbered weekday regex
	numberedWeekdayRegex = regexp.MustCompile(`^(?:(\d+)|(?:(first|1st|second|2nd|third|3rd|fourth|4th|fifth|5th|last)))\s+([A-Za-z]+)(?:\s+(?:of\s+)?)?([A-Za-z]+)(?:\s+(\d{4}))?$`)
)

// parseCompactTimestamp parses timestamp formats like "19970523091528" (YYYYMMDDhhmmss)
func parseCompactTimestamp(str string, loc *time.Location) (time.Time, bool) {
	if matches := compactTimestampRegex.FindStringSubmatch(str); matches != nil {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		hour, _ := strconv.Atoi(matches[4])
		minute, _ := strconv.Atoi(matches[5])
		second, _ := strconv.Atoi(matches[6])
		
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
	
	return time.Time{}, false
}

// parseMonthNameFormat parses formats like "Jan-15-2006" or "2006-Jan-15"
func parseMonthNameFormat(str string, loc *time.Location) (time.Time, bool) {
	// Handle "Jan-15-2006" format
	if matches := monthNameMDYRegex.FindStringSubmatch(str); matches != nil {
		monthName := matches[1]
		day, dayErr := strconv.Atoi(matches[2])
		year, yearErr := strconv.Atoi(matches[3])
		
		// Check for parsing errors
		if dayErr != nil || yearErr != nil {
			return time.Time{}, false
		}
		
		month, ok := getMonthByName(monthName)
		if !ok {
			return time.Time{}, false
		}
		
		// Validate date using our helper function
		if !IsValidDate(year, int(month), day) {
			return time.Time{}, false
		}
		
		return time.Date(year, month, day, 0, 0, 0, 0, loc), true
	}
	
	// Handle "2006-Jan-15" format
	if matches := monthNameYMDRegex.FindStringSubmatch(str); matches != nil {
		year, yearErr := strconv.Atoi(matches[1])
		monthName := matches[2]
		day, dayErr := strconv.Atoi(matches[3])
		
		// Check for parsing errors
		if dayErr != nil || yearErr != nil {
			return time.Time{}, false
		}
		
		month, ok := getMonthByName(monthName)
		if !ok {
			return time.Time{}, false
		}
		
		// Validate date using our helper function
		if !IsValidDate(year, int(month), day) {
			return time.Time{}, false
		}
		
		return time.Date(year, month, day, 0, 0, 0, 0, loc), true
	}
	
	return time.Time{}, false
}

// parseHTTPLogFormat parses formats like "10/Oct/2000:13:55:36 +0100"
func parseHTTPLogFormat(str string, loc *time.Location) (time.Time, bool) {
	if matches := httpLogRegex.FindStringSubmatch(str); matches != nil {
		day, dayErr := strconv.Atoi(matches[1])
		monthStr := matches[2]
		year, yearErr := strconv.Atoi(matches[3])
		hour, hourErr := strconv.Atoi(matches[4])
		minute, minErr := strconv.Atoi(matches[5])
		second, secErr := strconv.Atoi(matches[6])
		tzOffset := matches[7]
		
		// Check for parsing errors
		if dayErr != nil || yearErr != nil || hourErr != nil || minErr != nil || secErr != nil {
			return time.Time{}, false
		}
		
		// Get month from month string
		month, ok := getMonthByName(monthStr)
		if !ok {
			return time.Time{}, false
		}
		
		// Validate date components
		if !IsValidDate(year, int(month), day) {
			return time.Time{}, false
		}
		
		// Validate time components
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
		
		// Create a fixed timezone with the given offset
		tz := time.FixedZone("", tzOffsetSeconds)
		
		return time.Date(year, month, day, hour, minute, second, 0, tz), true
	}
	
	return time.Time{}, false
}

// parseNumberedWeekday parses formats like "1 Monday December 2008", "second Monday December 2008"
// It handles formats like "first Monday of December 2008" or "3rd Friday of January"
func parseNumberedWeekday(str string, now time.Time, loc *time.Location) (time.Time, bool) {
	if matches := numberedWeekdayRegex.FindStringSubmatch(str); matches != nil {
		var ordinal int
		
		// Parse the ordinal (numeric or word)
		if matches[1] != "" {
			// Numeric ordinal
			var err error
			ordinal, err = strconv.Atoi(matches[1])
			if err != nil || ordinal <= 0 || ordinal > 5 {
				return time.Time{}, false // PHP strtotime() only supports 1-5
			}
		} else {
			// Word ordinal
			switch strings.ToLower(matches[2]) {
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
				ordinal = -1 // Special case for last occurrence
			default:
				return time.Time{}, false
			}
		}
		
		// Parse the day of week
		dayOfWeek := getDayOfWeek(matches[3])
		if dayOfWeek < 0 {
			return time.Time{}, false
		}
		
		// Parse the month
		month, ok := getMonthByName(matches[4])
		if !ok {
			return time.Time{}, false
		}
		
		// Parse the year (optional, default to current year)
		year := now.Year()
		if matches[5] != "" {
			var err error
			year, err = strconv.Atoi(matches[5])
			if err != nil || year < 1 || year > 9999 {
				return time.Time{}, false
			}
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
				return time.Time{}, false // The specified occurrence doesn't exist in this month
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
			return time.Time{}, false // Invalid ordinal (should be > 0 or -1)
		}
		
		return time.Date(year, month, resultDay, 0, 0, 0, 0, loc), true
	}
	
	return time.Time{}, false
}