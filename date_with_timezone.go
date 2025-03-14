package strtotime

import (
	"regexp"
	"strconv"
	"time"
)

// parseWithTimezone tries to parse dates with timezone information
// Examples: "January 1 2023 PST", "June 1 1985 16:30:00 Europe/Paris", "2005-07-14 22:30:41 GMT"
func parseWithTimezone(str string, loc *time.Location) (time.Time, bool) {
	// First try the full date + time + timezone format
	if t, ok := parseFullDateTimeWithTimezone(str, loc); ok {
		return t, ok
	}
	
	// Try to parse ISO format date + time + timezone
	dateTimeRe := regexp.MustCompile(`^(\d{4}-\d{1,2}-\d{1,2})\s+(\d{1,2}):(\d{1,2}):(\d{1,2})\s+([a-zA-Z0-9/_.]+)$`)
	if matches := dateTimeRe.FindStringSubmatch(str); matches != nil {
		// Parse the date part
		datePart := matches[1]
		hour, _ := strconv.Atoi(matches[2])
		minute, _ := strconv.Atoi(matches[3])
		second, _ := strconv.Atoi(matches[4])
		tzString := matches[5]
		
		// Parse the date
		t, ok := parseISOFormat(datePart, loc)
		if !ok {
			return time.Time{}, false
		}
		
		// Add the time components
		t = time.Date(t.Year(), t.Month(), t.Day(), hour, minute, second, 0, t.Location())
		
		// Parse timezone
		tzLoc, found := tryParseTimezone(tzString)
		if !found {
			// If we couldn't parse the timezone, keep the original location
			tzLoc = loc
		}
		
		// Adjust to the timezone
		return t.In(tzLoc), true
	}
	
	// Try just time + timezone (e.g., "22:30:41 GMT")
	timeOnlyRe := regexp.MustCompile(`^(\d{1,2}):(\d{1,2})(?::(\d{1,2}))?\s+([a-zA-Z0-9/_.]+)$`)
	if matches := timeOnlyRe.FindStringSubmatch(str); matches != nil {
		hour, _ := strconv.Atoi(matches[1])
		minute, _ := strconv.Atoi(matches[2])
		second := 0
		if matches[3] != "" {
			second, _ = strconv.Atoi(matches[3])
		}
		tzString := matches[4]
		
		// Parse timezone
		tzLoc, found := tryParseTimezone(tzString)
		if !found {
			// If we couldn't parse the timezone, keep the original location
			tzLoc = loc
		}
		
		// Use current date with the specified time
		now := time.Now().In(tzLoc)
		return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, tzLoc), true
	}
	
	return time.Time{}, false
}

// parseFullDateTimeWithTimezone parses the month name + day + year + time + timezone format
func parseFullDateTimeWithTimezone(str string, loc *time.Location) (time.Time, bool) {
	// Regular expression to match month name, day, year, optional time, and timezone
	// The timezone can be a 3-letter code, a full region/city name, or any valid IANA timezone
	re := regexp.MustCompile(`^([a-zA-Z]+)\s+(\d{1,2})(?:st|nd|rd|th)?\s+(\d{4})(?:\s+(\d{1,2}):(\d{1,2})(?::(\d{1,2}))?)?\s+([a-zA-Z0-9/_.]+)$`)
	if matches := re.FindStringSubmatch(str); matches != nil {
		// Extract components
		monthName := matches[1]
		dayStr := matches[2]
		yearStr := matches[3]

		// Parse month
		month, ok := getMonthByName(monthName)
		if !ok {
			return time.Time{}, false
		}

		// Parse day and year
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 1 || day > 31 {
			return time.Time{}, false
		}

		year, err := strconv.Atoi(yearStr)
		if err != nil {
			return time.Time{}, false
		}

		// Default time components
		hour, minute, second := 0, 0, 0

		// Parse time if provided
		if matches[4] != "" { // Hour
			hourVal, err := strconv.Atoi(matches[4])
			if err != nil || hourVal < 0 || hourVal > 23 {
				return time.Time{}, false
			}
			hour = hourVal

			if matches[5] != "" { // Minute
				minuteVal, err := strconv.Atoi(matches[5])
				if err != nil || minuteVal < 0 || minuteVal > 59 {
					return time.Time{}, false
				}
				minute = minuteVal

				if matches[6] != "" { // Second
					secondVal, err := strconv.Atoi(matches[6])
					if err != nil || secondVal < 0 || secondVal > 59 {
						return time.Time{}, false
					}
					second = secondVal
				}
			}
		}

		// Parse timezone
		tzString := matches[7]
		tzLoc, found := tryParseTimezone(tzString)
		if !found {
			// If we couldn't parse the timezone, try with the original location
			tzLoc = loc
		}

		// Create the time with the given components
		return time.Date(year, month, day, hour, minute, second, 0, tzLoc), true
	}

	// No match
	return time.Time{}, false
}