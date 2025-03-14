package strtotime

import (
	"regexp"
	"strconv"
	"time"
)

// parseWithTimezone tries to parse dates with timezone information
// Examples: "January 1 2023 PST", "June 1 1985 16:30:00 Europe/Paris"
func parseWithTimezone(str string, loc *time.Location) (time.Time, bool) {
	// Check for dates with timezone - FORMAT: Month Day Year [Time] Timezone
	// e.g. "January 1 2023 PST", "June 1 1985 16:30:00 Europe/Paris"

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
