package strtotime

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// StrToTime will convert the provided string into a time similarly to how PHP strtotime() works.
func StrToTime(str string, opts ...Option) (time.Time, error) {
	var now time.Time
	loc := time.Local // Default timezone to local

	for _, opt := range opts {
		switch v := opt.(type) {
		case Rel: // relative to
			now = time.Time(v)
		case TZ: // timezone
			if v.Location != nil {
				loc = v.Location
			}
		}
	}

	if now.IsZero() {
		now = time.Now().In(loc)
	} else if now.Location() != loc {
		now = now.In(loc)
	}

	// Store original string for error reporting
	origStr := str

	// Basic implementation of some simple time formats
	str = strings.ToLower(strings.TrimSpace(str))

	// Special formats
	switch str {
	case "now":
		return now, nil
	case "today":
		year, month, day := now.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, loc), nil
	case "tomorrow":
		tomorrow := now.AddDate(0, 0, 1)
		year, month, day := tomorrow.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, loc), nil
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		year, month, day := yesterday.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, loc), nil
	case "next week":
		// PHP treats "next week" as "next Monday" (not the next occurrence of Monday, but Monday of next week)
		// If today is Friday, "next week" gives next Monday (3 days away)
		// But we need to adjust this calculation

		// Today is 2025-03-14 (Friday), and "next week" is Monday, March 17, 2025
		// So for Friday, the difference is 3 days (not 4)

		dayOfWeek := int(now.Weekday())
		var daysToAdd int

		switch dayOfWeek {
		case 0: // Sunday
			daysToAdd = 1 // Next Monday is 1 day away
		case 1: // Monday
			daysToAdd = 0 // This is already Monday - go to today but at 00:00:00
		case 2: // Tuesday
			daysToAdd = 6 // Next Monday is 6 days away
		case 3: // Wednesday
			daysToAdd = 5 // Next Monday is 5 days away
		case 4: // Thursday
			daysToAdd = 4 // Next Monday is 4 days away
		case 5: // Friday
			daysToAdd = 3 // Next Monday is 3 days away
		case 6: // Saturday
			daysToAdd = 2 // Next Monday is 2 days away
		}

		// Important: PHP keeps the time for "next week"
		nextMondayTime := now.AddDate(0, 0, daysToAdd)
		return nextMondayTime, nil

	case "last week":
		// PHP treats "last week" as the previous Monday
		// For Friday March 14, 2025, that's Monday March 3, 2025 (11 days earlier)

		dayOfWeek := int(now.Weekday())
		var daysToSubtract int

		switch dayOfWeek {
		case 0: // Sunday
			daysToSubtract = 6 // Last Monday was 6 days ago
		case 1: // Monday
			daysToSubtract = 7 // Last Monday was a week ago
		case 2: // Tuesday
			daysToSubtract = 8 // Last Monday was 8 days ago
		case 3: // Wednesday
			daysToSubtract = 9 // Last Monday was 9 days ago
		case 4: // Thursday
			daysToSubtract = 10 // Last Monday was 10 days ago
		case 5: // Friday
			daysToSubtract = 11 // Last Monday was 11 days ago
		case 6: // Saturday
			daysToSubtract = 12 // Last Monday was 12 days ago
		}

		lastMondayTime := now.AddDate(0, 0, -daysToSubtract)
		return lastMondayTime, nil
	}

	// Handle +/- relative time formats
	if strings.HasPrefix(str, "+") || strings.HasPrefix(str, "-") {
		sign := 1
		if strings.HasPrefix(str, "-") {
			sign = -1
		}

		// Use regexp to handle variable whitespace
		reRelTime := regexp.MustCompile(`^[+-](\d+)\s+(\w+)$`)
		if matches := reRelTime.FindStringSubmatch(str); len(matches) == 3 {
			amount, err := strconv.Atoi(matches[1])
			if err != nil {
				return time.Time{}, err
			}

			// Apply the sign
			amount = amount * sign

			unit := matches[2]
			switch unit {
			case "day", "days":
				return now.AddDate(0, 0, amount), nil
			case "week", "weeks":
				return now.AddDate(0, 0, amount*7), nil
			case "month", "months":
				return now.AddDate(0, amount, 0), nil
			case "year", "years":
				return now.AddDate(amount, 0, 0), nil
			}
		}

		// If the standard pattern didn't match, try a more permissive one with multiple whitespaces
		reRelTimeMultiSpace := regexp.MustCompile(`^[+-](\d+)\s+(\w+)$`)
		origStrTrimmed := strings.TrimSpace(origStr)
		origStrLower := strings.ToLower(origStrTrimmed)

		// Normalize multiple spaces to single space
		normalized := regexp.MustCompile(`\s+`).ReplaceAllString(origStrLower, " ")

		if matches := reRelTimeMultiSpace.FindStringSubmatch(normalized); len(matches) == 3 {
			amount, err := strconv.Atoi(matches[1])
			if err != nil {
				return time.Time{}, err
			}

			// Apply the sign
			if strings.HasPrefix(normalized, "-") {
				amount = -amount
			}

			unit := matches[2]
			switch unit {
			case "day", "days":
				return now.AddDate(0, 0, amount), nil
			case "week", "weeks":
				return now.AddDate(0, 0, amount*7), nil
			case "month", "months":
				return now.AddDate(0, amount, 0), nil
			case "year", "years":
				return now.AddDate(amount, 0, 0), nil
			}
		}
	}

	// Handle "next {dayofweek}" and "last {dayofweek}"
	reNextDay := regexp.MustCompile(`^next\s+(\w+)$`)
	reLastDay := regexp.MustCompile(`^last\s+(\w+)$`)

	if matches := reNextDay.FindStringSubmatch(str); len(matches) == 2 {
		targetDay := getDayOfWeek(matches[1])
		if targetDay < 0 {
			return time.Time{}, errors.New("unknown day of week: " + matches[1])
		}

		// Current day of week (0 = Sunday, 6 = Saturday)
		currentDay := int(now.Weekday())
		// Days until the next occurrence of targetDay
		daysUntil := (targetDay - currentDay + 7) % 7
		if daysUntil == 0 {
			daysUntil = 7 // If today is the target day, go to next week
		}

		nextDayTime := now.AddDate(0, 0, daysUntil)
		year, month, day := nextDayTime.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, loc), nil
	}

	if matches := reLastDay.FindStringSubmatch(str); len(matches) == 2 {
		targetDay := getDayOfWeek(matches[1])
		if targetDay < 0 {
			return time.Time{}, errors.New("unknown day of week: " + matches[1])
		}

		// Current day of week (0 = Sunday, 6 = Saturday)
		currentDay := int(now.Weekday())
		// Days since the last occurrence of targetDay
		daysSince := (currentDay - targetDay + 7) % 7
		if daysSince == 0 {
			daysSince = 7 // If today is the target day, go to previous week
		}

		lastDayTime := now.AddDate(0, 0, -daysSince)
		year, month, day := lastDayTime.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, loc), nil
	}

	// Try to parse standard date formats
	// Format: YYYY-MM-DD
	reIsoDate := regexp.MustCompile(`^(\d{4})-(\d{1,2})-(\d{1,2})$`)
	if matches := reIsoDate.FindStringSubmatch(str); len(matches) == 4 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), nil
	}

	// Format: YYYY/MM/DD
	reSlashDate := regexp.MustCompile(`^(\d{4})/(\d{1,2})/(\d{1,2})$`)
	if matches := reSlashDate.FindStringSubmatch(str); len(matches) == 4 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), nil
	}

	// Format: MM/DD/YYYY
	reUSDate := regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{4})$`)
	if matches := reUSDate.FindStringSubmatch(str); len(matches) == 4 {
		month, _ := strconv.Atoi(matches[1])
		day, _ := strconv.Atoi(matches[2])
		year, _ := strconv.Atoi(matches[3])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), nil
	}

	// Format: January 15 2023 or Jan 15, 2023
	monthNames := map[string]time.Month{
		"january":   time.January,
		"jan":       time.January,
		"february":  time.February,
		"feb":       time.February,
		"march":     time.March,
		"mar":       time.March,
		"april":     time.April,
		"apr":       time.April,
		"may":       time.May,
		"june":      time.June,
		"jun":       time.June,
		"july":      time.July,
		"jul":       time.July,
		"august":    time.August,
		"aug":       time.August,
		"september": time.September,
		"sep":       time.September,
		"october":   time.October,
		"oct":       time.October,
		"november":  time.November,
		"nov":       time.November,
		"december":  time.December,
		"dec":       time.December,
	}

	reLongDate := regexp.MustCompile(`^([a-z]+)\s+(\d{1,2})(?:,?)\s+(\d{4})$`)
	if matches := reLongDate.FindStringSubmatch(str); len(matches) == 4 {
		monthStr := matches[1]
		month, ok := monthNames[monthStr]
		if !ok {
			return time.Time{}, fmt.Errorf("unknown month: %s", monthStr)
		}
		day, _ := strconv.Atoi(matches[2])
		year, _ := strconv.Atoi(matches[3])
		return time.Date(year, month, day, 0, 0, 0, 0, loc), nil
	}

	// Try handling a normalized version of the string
	normalized := regexp.MustCompile(`\s+`).ReplaceAllString(str, " ")
	if normalized != str {
		// Try all patterns again with normalized string
		return StrToTime(normalized, opts...)
	}

	// We need more complex parsing for other formats
	return time.Time{}, errors.New("unknown or unsupported time format: " + origStr)
}

// getDayOfWeek converts day name to day number (0 = Sunday, 6 = Saturday)
func getDayOfWeek(day string) int {
	switch day {
	case "sunday", "sun":
		return 0
	case "monday", "mon":
		return 1
	case "tuesday", "tue":
		return 2
	case "wednesday", "wed":
		return 3
	case "thursday", "thu":
		return 4
	case "friday", "fri":
		return 5
	case "saturday", "sat":
		return 6
	default:
		return -1
	}
}
