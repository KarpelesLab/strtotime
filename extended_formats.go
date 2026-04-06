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

// parseCompactTimestamp parses compact timestamp formats:
// - "19970523091528" (YYYYMMDDhhmmss, exactly 14 digits)
// - "20050620091407 GMT" (14 digits + timezone)
// - "20101212" (YYYYMMDD, exactly 8 digits)
func parseCompactTimestamp(str string, loc *time.Location) (time.Time, bool) {
	// Split on space to handle optional timezone suffix
	parts := strings.SplitN(str, " ", 2)
	digits := parts[0]

	// 8-digit YYYYMMDD format
	if len(digits) == 8 && isAllDigits(digits) {
		year, _ := strconv.Atoi(digits[0:4])
		month, _ := strconv.Atoi(digits[4:6])
		day, _ := strconv.Atoi(digits[6:8])
		if month >= 1 && month <= 12 && day >= 1 && day <= 31 {
			return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), true
		}
		return time.Time{}, false
	}

	// 14-digit YYYYMMDDhhmmss format (with optional timezone)
	if len(digits) != 14 || !isAllDigits(digits) {
		return time.Time{}, false
	}

	year, _ := strconv.Atoi(digits[0:4])
	month, _ := strconv.Atoi(digits[4:6])
	day, _ := strconv.Atoi(digits[6:8])
	hour, _ := strconv.Atoi(digits[8:10])
	minute, _ := strconv.Atoi(digits[10:12])
	second, _ := strconv.Atoi(digits[12:14])

	if month < 1 || month > 12 || day < 1 || day > 31 {
		return time.Time{}, false
	}
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 || second < 0 || second > 59 {
		return time.Time{}, false
	}

	tzLoc := loc
	if len(parts) > 1 && parts[1] != "" {
		if parsed, found := tryParseTimezone(strings.TrimSpace(parts[1])); found {
			tzLoc = parsed
		}
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, tzLoc), true
}

// parseCompactTimeFormats handles PHP-specific compact time/date formats:
// - "022233": 6-digit hhmmss time-only
// - "2006167": 7-digit year+day-of-year (pgydotd)
// - "t0222": 't' prefix + hhmm time
// - "22.49.12.42GMT": dotted time with optional fractional seconds and timezone
func parseCompactTimeFormats(str string, now time.Time, loc *time.Location) (time.Time, bool) {
	// "t" prefix + 4 digits = tHHMM time format
	if len(str) >= 5 && str[0] == 't' && isAllDigits(str[1:5]) {
		hour, _ := strconv.Atoi(str[1:3])
		minute, _ := strconv.Atoi(str[3:5])
		if IsValidTime(hour, minute, 0) {
			y, m, d := now.Date()
			return time.Date(y, m, d, hour, minute, 0, 0, loc), true
		}
	}

	// Dotted time format: "HH.MM.SS[.frac][TZ]"
	if len(str) >= 8 && str[2] == '.' && str[5] == '.' {
		if isAllDigits(str[0:2]) && isAllDigits(str[3:5]) && str[6] >= '0' && str[6] <= '9' {
			hour, _ := strconv.Atoi(str[0:2])
			minute, _ := strconv.Atoi(str[3:5])
			// Parse seconds (may be followed by .frac or tz)
			pos := 6
			for pos < len(str) && str[pos] >= '0' && str[pos] <= '9' {
				pos++
			}
			second, _ := strconv.Atoi(str[6:pos])
			if !IsValidTime(hour, minute, second) {
				return time.Time{}, false
			}
			// Skip optional fractional part .NN
			if pos < len(str) && str[pos] == '.' {
				pos++
				for pos < len(str) && str[pos] >= '0' && str[pos] <= '9' {
					pos++
				}
			}
			// Parse optional timezone suffix (no space)
			tzLoc := loc
			if pos < len(str) {
				tzStr := str[pos:]
				// Strip leading space if present
				tzStr = strings.TrimSpace(tzStr)
				if parsed, found := tryParseTimezone(tzStr); found {
					tzLoc = parsed
				} else {
					return time.Time{}, false
				}
			}
			y, m, d := now.Date()
			return time.Date(y, m, d, hour, minute, second, 0, tzLoc), true
		}
	}

	// All-digit formats
	if !isAllDigits(str) {
		return time.Time{}, false
	}

	// 6-digit hhmmss: compact time-only format
	if len(str) == 6 {
		hour, _ := strconv.Atoi(str[0:2])
		minute, _ := strconv.Atoi(str[2:4])
		second, _ := strconv.Atoi(str[4:6])
		if IsValidTime(hour, minute, second) {
			y, m, d := now.Date()
			return time.Date(y, m, d, hour, minute, second, 0, loc), true
		}
	}

	// 7-digit pgydotd: year + day-of-year (e.g., "2006167" = June 16, 2006)
	if len(str) == 7 {
		year, _ := strconv.Atoi(str[0:4])
		doy, _ := strconv.Atoi(str[4:7])
		if year >= 1 && doy >= 1 && doy <= 366 {
			t := time.Date(year, 1, 1, 0, 0, 0, 0, loc).AddDate(0, 0, doy-1)
			if t.Year() == year { // ensure doy didn't overflow into next year
				return t, true
			}
		}
	}

	return time.Time{}, false
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

	// Try "15-Jan-2006" format: digits-alpha-4digits (DD-Mon-YYYY)
	if isAllDigits(parts[0]) && isAlpha(parts[1]) && len(parts[1]) >= 3 && isAllDigits(parts[2]) {
		day, dayErr := strconv.Atoi(parts[0])
		year, yearErr := strconv.Atoi(parts[2])

		if dayErr == nil && yearErr == nil {
			if year < 100 {
				year = parseTwoDigitYear(year)
			}
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

// stripOrdinalSuffix removes ordinal suffixes: "26th" → "26", "1st" → "1"
func stripOrdinalSuffix(s string) string {
	lower := strings.ToLower(s)
	for _, suffix := range []string{"st", "nd", "rd", "th"} {
		if strings.HasSuffix(lower, suffix) && len(s) > len(suffix) {
			prefix := s[:len(s)-len(suffix)]
			if isAllDigits(prefix) {
				return prefix
			}
		}
	}
	return s
}

// stripMonthPeriod removes trailing period from month abbreviations: "dec." → "dec"
func stripMonthPeriod(s string) string {
	return strings.TrimSuffix(s, ".")
}

// getMonthByNameFlex tries to match a month name, handling trailing periods
func getMonthByNameFlex(name string) (time.Month, bool) {
	m, ok := getMonthByName(name)
	if ok {
		return m, true
	}
	return getMonthByName(stripMonthPeriod(name))
}

// parseDayMonthYear parses formats like "DD Mon YYYY", "DD-Mon-YYYY", "DDMonYYYY"
// with optional time and timezone. Also handles day-of-week prefix and ordinal suffix.
// Examples: "11 Oct 2005", "11-MAY-1988 12:00:00AM", "11Oct2005",
//
//	"Sat 26th Nov 2005 18:18", "Thu, 20 Nov 2003 16:20:42 +0000"
func parseDayMonthYear(str string, loc *time.Location) (time.Time, bool) {
	s := str

	// Skip leading day-of-week like "Sat " or "Thursday, "
	// Try full weekday names first, then 3-letter abbreviations
	stripped := false
	for _, wdLen := range []int{9, 8, 7, 6, 3} { // wednesday=9, thursday=8, saturday=8, tuesday=7, monday=6, etc.
		if len(s) >= wdLen {
			prefix := s[:wdLen]
			if getDayOfWeek(prefix) >= 0 {
				s = s[wdLen:]
				s = strings.TrimLeft(s, ", ")
				stripped = true
				break
			}
		}
	}
	if !stripped && len(s) >= 3 {
		prefix := s[:3]
		if getDayOfWeek(prefix) >= 0 {
			s = s[3:]
			s = strings.TrimLeft(s, ", ")
		}
	}

	// Try to extract day, month, year from the remaining string
	fields := strings.Fields(s)
	if len(fields) < 2 {
		// Try compact DDMonYYYY (no spaces)
		return parseDayMonthYearCompact(s, loc)
	}

	idx := 0

	// Handle hyphen-separated date: "24-Jan-2019" or "24-Jan-19"
	// (used by DATE_COOKIE and DATE_RFC850 formats)
	if strings.Contains(fields[0], "-") {
		dateParts := strings.SplitN(fields[0], "-", 3)
		if len(dateParts) == 3 {
			// Replace the single hyphenated field with 3 separate fields
			rest := fields[1:]
			fields = append(dateParts, rest...)
		}
	}

	// Parse day (may include ordinal suffix)
	dayStr := stripOrdinalSuffix(fields[idx])
	day, err := strconv.Atoi(dayStr)
	if err != nil || day < 1 || day > 31 {
		// Day might be fused with month: "11Oct" in "11Oct 2005"
		return parseDayMonthYearCompact(s, loc)
	}
	idx++

	// Parse month name
	if idx >= len(fields) {
		return time.Time{}, false
	}
	month, ok := getMonthByNameFlex(fields[idx])
	if !ok {
		return time.Time{}, false
	}
	idx++

	// Parse year
	if idx >= len(fields) {
		return time.Time{}, false
	}
	year, err := strconv.Atoi(fields[idx])
	if err != nil || year < 0 {
		return time.Time{}, false
	}
	if year < 100 {
		year = parseTwoDigitYear(year)
	}
	idx++

	hour, minute, second, nanos := 0, 0, 0, 0

	// Parse optional time
	if idx < len(fields) {
		timeStr := fields[idx]
		if strings.Contains(timeStr, ":") {
			h, m, sec, consumed, ok := parseFlexTime(timeStr)
			if ok {
				hour, minute, second = h, m, sec
				idx++
				// Check for AM/PM attached to time or as next field
				remaining := timeStr[consumed:]
				if ampm := strings.ToLower(remaining); ampm == "am" || ampm == "pm" {
					hour = applyAMPM(hour, ampm)
				} else if idx < len(fields) {
					ampm = strings.ToLower(fields[idx])
					if ampm == "am" || ampm == "pm" {
						hour = applyAMPM(hour, ampm)
						idx++
					}
				}
				// Handle fractional seconds
				if consumed < len(timeStr) && timeStr[consumed] == '.' {
					fracStart := consumed + 1
					fracEnd := fracStart
					for fracEnd < len(timeStr) && timeStr[fracEnd] >= '0' && timeStr[fracEnd] <= '9' {
						fracEnd++
					}
					if fracEnd > fracStart {
						fracStr := timeStr[fracStart:fracEnd]
						for len(fracStr) < 9 {
							fracStr += "0"
						}
						if len(fracStr) > 9 {
							fracStr = fracStr[:9]
						}
						nanos, _ = strconv.Atoi(fracStr)
					}
				}
			}
		}
	}

	// Parse optional timezone
	tzLoc := loc
	if idx < len(fields) {
		tzStr := strings.Join(fields[idx:], " ")
		if parsed, _, ok := parseNumericTimezoneOffset(tzStr); ok {
			tzLoc = parsed
		} else if parsed, found := tryParseTimezone(tzStr); found {
			tzLoc = parsed
		}
	}

	if month > 0 && day > 0 {
		return time.Date(year, month, day, hour, minute, second, nanos, tzLoc), true
	}
	return time.Time{}, false
}

// parseDayMonthYearCompact parses "DDMonYYYY", "DDMon YYYY", or "DDMon" (no year = current year)
func parseDayMonthYearCompact(str string, loc *time.Location) (time.Time, bool) {
	s := strings.TrimSpace(str)
	if len(s) < 4 { // at least DMon
		return time.Time{}, false
	}

	// Extract leading digits as day
	dayEnd := 0
	for dayEnd < len(s) && s[dayEnd] >= '0' && s[dayEnd] <= '9' {
		dayEnd++
	}
	if dayEnd == 0 || dayEnd > 2 {
		return time.Time{}, false
	}
	day, _ := strconv.Atoi(s[:dayEnd])

	// Extract month name (3+ letters)
	monthStart := dayEnd
	monthEnd := monthStart
	for monthEnd < len(s) && ((s[monthEnd] >= 'a' && s[monthEnd] <= 'z') || (s[monthEnd] >= 'A' && s[monthEnd] <= 'Z')) {
		monthEnd++
	}
	if monthEnd-monthStart < 3 {
		return time.Time{}, false
	}
	month, ok := getMonthByNameFlex(s[monthStart:monthEnd])
	if !ok {
		return time.Time{}, false
	}

	// Rest may have optional space then year, then optional time
	rest := strings.TrimSpace(s[monthEnd:])
	if rest == "" {
		// No year — default to current year (e.g., "11Oct")
		if day < 1 || day > 31 {
			return time.Time{}, false
		}
		return time.Date(time.Now().Year(), month, day, 0, 0, 0, 0, loc), true
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return time.Time{}, false
	}

	year, err := strconv.Atoi(fields[0])
	if err != nil {
		return time.Time{}, false
	}
	if year < 100 {
		year = parseTwoDigitYear(year)
	}

	hour, minute, second, nanos := 0, 0, 0, 0
	fidx := 1

	// Parse optional time
	if fidx < len(fields) && strings.Contains(fields[fidx], ":") {
		h, m, sec, consumed, ok := parseFlexTime(fields[fidx])
		if ok {
			hour, minute, second = h, m, sec
			// Check for AM/PM
			remaining := fields[fidx][consumed:]
			if ampm := strings.ToLower(remaining); ampm == "am" || ampm == "pm" {
				hour = applyAMPM(hour, ampm)
			} else if fidx+1 < len(fields) {
				ampm = strings.ToLower(fields[fidx+1])
				if ampm == "am" || ampm == "pm" {
					hour = applyAMPM(hour, ampm)
				}
			}
			// Handle fractional seconds
			if consumed < len(fields[fidx]) && fields[fidx][consumed] == '.' {
				fracStart := consumed + 1
				fracEnd := fracStart
				for fracEnd < len(fields[fidx]) && fields[fidx][fracEnd] >= '0' && fields[fidx][fracEnd] <= '9' {
					fracEnd++
				}
				if fracEnd > fracStart {
					fracStr := fields[fidx][fracStart:fracEnd]
					for len(fracStr) < 9 {
						fracStr += "0"
					}
					if len(fracStr) > 9 {
						fracStr = fracStr[:9]
					}
					nanos, _ = strconv.Atoi(fracStr)
				}
			}
		}
	}

	if day < 1 || day > 31 {
		return time.Time{}, false
	}

	return time.Date(year, month, day, hour, minute, second, nanos, loc), true
}

// applyAMPM converts 12-hour time to 24-hour format
func applyAMPM(hour int, ampm string) int {
	if ampm == "am" {
		if hour == 12 {
			return 0
		}
		return hour
	}
	// pm
	if hour == 12 {
		return 12
	}
	return hour + 12
}

// parseMonthYearOnly parses "Oct 2001" or "2001 Oct" (month + year, day defaults to 1)
func parseMonthYearOnly(str string, loc *time.Location) (time.Time, bool) {
	fields := strings.Fields(str)
	if len(fields) != 2 {
		return time.Time{}, false
	}

	// Try "Month Year" — year must be >= 100 or have 4+ digits (e.g., "0099") to avoid
	// "April 4" being treated as year 4
	if month, ok := getMonthByNameFlex(fields[0]); ok {
		year, err := strconv.Atoi(fields[1])
		if err == nil && (year >= 100 || len(fields[1]) >= 4) {
			return time.Date(year, month, 1, 0, 0, 0, 0, loc), true
		}
	}

	// Try "Year Month"
	if year, err := strconv.Atoi(fields[0]); err == nil && (year >= 100 || len(fields[0]) >= 4) {
		if month, ok := getMonthByNameFlex(fields[1]); ok {
			return time.Date(year, month, 1, 0, 0, 0, 0, loc), true
		}
	}

	return time.Time{}, false
}

// parseTimeBeforeDate parses formats where time precedes the date:
// "19:30 Dec 17 2005", "17:00 2004-01-01", "1pm Aug 1 GMT 2007"
func parseTimeBeforeDate(str string, loc *time.Location) (time.Time, bool) {
	fields := strings.Fields(str)
	if len(fields) < 2 {
		return time.Time{}, false
	}

	var hour, minute, second int

	// Try colon-based time (19:30, 17:00:00, etc.)
	if strings.Contains(fields[0], ":") {
		var ok bool
		hour, minute, second, _, ok = parseFlexTime(fields[0])
		if !ok || !IsValidTime(hour, minute, second) {
			return time.Time{}, false
		}
	} else {
		// Try bare AM/PM time: "1pm", "12am", "3am"
		f := strings.ToLower(fields[0])
		var ampm string
		if strings.HasSuffix(f, "pm") {
			ampm = "pm"
			f = f[:len(f)-2]
		} else if strings.HasSuffix(f, "am") {
			ampm = "am"
			f = f[:len(f)-2]
		}
		if ampm == "" {
			return time.Time{}, false
		}
		h, err := strconv.Atoi(f)
		if err != nil || h < 1 || h > 12 {
			return time.Time{}, false
		}
		hour = applyAMPM(h, ampm)
	}

	// Try to parse the rest as a date
	dateStr := strings.Join(fields[1:], " ")

	// Try ISO date
	if t, ok := parseISOFormat(dateStr, loc); ok {
		return time.Date(t.Year(), t.Month(), t.Day(), hour, minute, second, 0, loc), true
	}

	// Try month name date with optional timezone: "Aug 1 2007", "Aug 1 GMT 2007"
	dateFields := strings.Fields(dateStr)
	if len(dateFields) >= 2 {
		if month, ok := getMonthByNameFlex(dateFields[0]); ok {
			dayStr := stripOrdinalSuffix(strings.TrimSuffix(dateFields[1], ","))
			day, err := strconv.Atoi(dayStr)
			if err == nil && day >= 1 && day <= 31 {
				year := time.Now().Year()
				tzLoc := loc
				fidx := 2

				// Parse optional timezone and year from remaining fields
				for fidx < len(dateFields) {
					// Try as year
					if y, err := strconv.Atoi(dateFields[fidx]); err == nil && y > 0 {
						year = y
						fidx++
						continue
					}
					// Try as timezone
					if tz, found := tryParseTimezone(dateFields[fidx]); found {
						tzLoc = tz
						fidx++
						continue
					}
					break
				}

				return time.Date(year, month, day, hour, minute, second, 0, tzLoc), true
			}
		}
	}

	return time.Time{}, false
}

// parseUSDateWithTime parses "MM/DD/YYYY H:MM AM" format (US date with 12-hour time)
func parseUSDateWithTime(str string, loc *time.Location) (time.Time, bool) {
	fields := strings.Fields(str)
	if len(fields) < 2 {
		return time.Time{}, false
	}

	// First field should be the date
	t, ok := parseUSFormat(fields[0], loc)
	if !ok {
		return time.Time{}, false
	}

	// Parse time
	if len(fields) >= 2 && strings.Contains(fields[1], ":") {
		hour, minute, second, consumed, ok := parseFlexTime(fields[1])
		if !ok {
			return time.Time{}, false
		}
		// Check for AM/PM
		remaining := fields[1][consumed:]
		if ampm := strings.ToLower(remaining); ampm == "am" || ampm == "pm" {
			hour = applyAMPM(hour, ampm)
		} else if len(fields) >= 3 {
			ampm = strings.ToLower(fields[2])
			if ampm == "am" || ampm == "pm" {
				hour = applyAMPM(hour, ampm)
			}
		}
		return time.Date(t.Year(), t.Month(), t.Day(), hour, minute, second, 0, loc), true
	}

	return time.Time{}, false
}

// parseFirstLastDayOfDate parses "first day of YYYY-MM" and "last day of YYYY-MM" patterns
func parseFirstLastDayOfDate(str string, now time.Time, loc *time.Location) (time.Time, bool) {
	lower := strings.ToLower(strings.TrimSpace(str))

	var isFirst, isLast bool
	var rest string

	if strings.HasPrefix(lower, "first day of ") {
		isFirst = true
		rest = strings.TrimSpace(lower[13:])
	} else if strings.HasPrefix(lower, "last day of ") {
		isLast = true
		rest = strings.TrimSpace(lower[12:])
	} else {
		return time.Time{}, false
	}

	// Try to parse rest as a relative expression: "+1 month", "-2 months"
	if len(rest) > 0 && (rest[0] == '+' || rest[0] == '-') {
		fields := strings.Fields(rest)
		if len(fields) == 2 {
			amount, err := strconv.Atoi(fields[0])
			if err == nil {
				unit := normalizeTimeUnit(fields[1])
				var refTime time.Time
				switch unit {
				case UnitMonth:
					refTime = now.AddDate(0, amount, 0)
				case UnitYear:
					refTime = now.AddDate(amount, 0, 0)
				default:
					return time.Time{}, false
				}
				year, month, _ := refTime.Date()
				if isFirst {
					return time.Date(year, month, 1, now.Hour(), now.Minute(), now.Second(), 0, loc), true
				}
				return time.Date(year, month, daysInMonth(year, month), now.Hour(), now.Minute(), now.Second(), 0, loc), true
			}
		}
	}

	// Try to parse rest as YYYY-MM or YYYY-M
	if t, ok := parseYearMonthFormat(rest, loc); ok {
		year, month, _ := t.Date()
		if isFirst {
			return time.Date(year, month, 1, 0, 0, 0, 0, loc), true
		}
		if isLast {
			return time.Date(year, month, daysInMonth(year, month), 0, 0, 0, 0, loc), true
		}
	}

	// Try to parse rest as a month name with optional year
	fields := strings.Fields(rest)
	if len(fields) >= 1 {
		if month, ok := getMonthByNameFlex(fields[0]); ok {
			year := now.Year()
			if len(fields) >= 2 {
				if y, err := strconv.Atoi(fields[1]); err == nil {
					year = y
				}
			}
			if isFirst {
				return time.Date(year, month, 1, 0, 0, 0, 0, loc), true
			}
			return time.Date(year, month, daysInMonth(year, month), 0, 0, 0, 0, loc), true
		}
	}

	return time.Time{}, false
}

// parseOrdinalDate parses "26th Nov" or "December 4th, 2005" etc.
// with optional time. Handles month name followed by ordinal day or ordinal day followed by month.
func parseOrdinalDate(str string, now time.Time, loc *time.Location) (time.Time, bool) {
	fields := strings.Fields(str)
	if len(fields) < 2 {
		return time.Time{}, false
	}

	// Try "DDth Mon [YYYY] [time]" format
	dayStr := stripOrdinalSuffix(fields[0])
	if day, err := strconv.Atoi(dayStr); err == nil && day >= 1 && day <= 31 {
		if month, ok := getMonthByNameFlex(fields[1]); ok {
			year := now.Year()
			fidx := 2
			if fidx < len(fields) {
				if y, err := strconv.Atoi(fields[fidx]); err == nil {
					year = y
					fidx++
				}
			}
			hour, minute, second := 0, 0, 0
			if fidx < len(fields) && strings.Contains(fields[fidx], ":") {
				h, m, s, _, ok := parseFlexTime(fields[fidx])
				if ok {
					hour, minute, second = h, m, s
				}
			}
			return time.Date(year, month, day, hour, minute, second, 0, loc), true
		}
	}

	return time.Time{}, false
}

// parseMonthDayTimeYear parses "Dec 17 19:30 2005" (month day time year)
func parseMonthDayTimeYear(str string, loc *time.Location) (time.Time, bool) {
	fields := strings.Fields(str)
	if len(fields) != 4 {
		return time.Time{}, false
	}

	month, ok := getMonthByNameFlex(fields[0])
	if !ok {
		return time.Time{}, false
	}

	dayStr := stripOrdinalSuffix(strings.TrimSuffix(fields[1], ","))
	day, err := strconv.Atoi(dayStr)
	if err != nil || day < 1 || day > 31 {
		return time.Time{}, false
	}

	// Third field should be time
	if !strings.Contains(fields[2], ":") {
		return time.Time{}, false
	}
	hour, minute, second, _, ok := parseFlexTime(fields[2])
	if !ok {
		return time.Time{}, false
	}

	// Fourth field should be year
	year, err := strconv.Atoi(fields[3])
	if err != nil {
		return time.Time{}, false
	}

	return time.Date(year, month, day, hour, minute, second, 0, loc), true
}

// parseDateTimeTZRelative parses "YYYY-MM-DD TZ +N unit" or "YYYY-MM-DDThh:mm:ss+ZZZZ +N unit"
// Handles date with timezone followed by relative time adjustment.
func parseDateTimeTZRelative(str string, loc *time.Location) (time.Time, bool) {
	// Look for patterns like "2004-10-31 EDT +1 hour" or "2008-07-01T22:35:17+0200 +7 days"
	// Strategy: find the last +/- that starts a relative expression
	lower := str

	// Try to find a relative expression at the end: +N unit or -N unit
	// Look backward for " +N " or " -N " pattern
	for i := len(lower) - 1; i > 0; i-- {
		if lower[i] == '+' || lower[i] == '-' {
			// Check if preceded by space
			if i > 0 && lower[i-1] == ' ' {
				relPart := lower[i:]
				datePart := strings.TrimSpace(lower[:i])

				// Validate the relative part looks like "+N unit"
				relFields := strings.Fields(relPart)
				if len(relFields) != 2 {
					continue
				}
				amount, err := strconv.Atoi(relFields[0])
				if err != nil {
					continue
				}
				unit := normalizeTimeUnit(relFields[1])
				if unit == relFields[1] { // normalizeTimeUnit returns input if unrecognized
					// Check more carefully
					if unit != UnitDay && unit != UnitWeek && unit != UnitMonth && unit != UnitYear &&
						unit != UnitHour && unit != UnitMinute && unit != UnitSecond {
						continue
					}
				}

				// Try to parse the date part
				// Try ISO 8601
				if t, ok := parseISO8601(datePart, loc); ok {
					return applyRelativeUnit(t, amount, unit), true
				}
				// Try datetime with tz
				if t, ok := parseDateTimeFormat(datePart, loc); ok {
					return applyRelativeUnit(t, amount, unit), true
				}
				// Try date + time + tz: "2005-07-14 22:30:41 GMT"
				if t, ok := parseISODateTimeWithTimezone(datePart, loc); ok {
					return applyRelativeUnit(t, amount, unit), true
				}
				// Try date + tz (no time): "2004-10-31 EDT"
				if t, ok := parseDateWithTZ(datePart, loc); ok {
					return applyRelativeUnit(t, amount, unit), true
				}
				// Try plain date
				if t, ok := parseISOFormat(datePart, loc); ok {
					return applyRelativeUnit(t, amount, unit), true
				}
				// Try RFC 2822 / DD Mon YYYY format: "Mon, 08 May 2006 13:06:44 -0400"
				if t, ok := parseDayMonthYear(datePart, loc); ok {
					return applyRelativeUnit(t, amount, unit), true
				}
			}
		}
	}

	return time.Time{}, false
}

// applyRelativeUnit applies a relative time offset
func applyRelativeUnit(t time.Time, amount int, unit string) time.Time {
	switch unit {
	case UnitDay:
		return t.AddDate(0, 0, amount)
	case UnitWeek:
		return t.AddDate(0, 0, amount*7)
	case UnitMonth:
		return t.AddDate(0, amount, 0)
	case UnitYear:
		return t.AddDate(amount, 0, 0)
	case UnitHour:
		return t.Add(time.Duration(amount) * time.Hour)
	case UnitMinute:
		return t.Add(time.Duration(amount) * time.Minute)
	case UnitSecond:
		return t.Add(time.Duration(amount) * time.Second)
	}
	return t
}

// parseDateWithTZ parses "YYYY-MM-DD TZname" (date + timezone without time)
func parseDateWithTZ(str string, loc *time.Location) (time.Time, bool) {
	fields := strings.Fields(str)
	if len(fields) != 2 {
		return time.Time{}, false
	}
	t, ok := parseISOFormat(fields[0], loc)
	if !ok {
		return time.Time{}, false
	}
	tzLoc, found := tryParseTimezone(fields[1])
	if !found {
		return time.Time{}, false
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tzLoc), true
}

// parseFrontBackOf parses Scottish time expressions:
// "front of 7" = 6:45 (15 minutes before the hour)
// "back of 7" = 7:15 (15 minutes after the hour)
// Also supports AM/PM: "front of 12pm" = 11:45, "back of 3am" = 3:15
func parseFrontBackOf(str string, now time.Time, loc *time.Location) (time.Time, bool) {
	lower := strings.ToLower(strings.TrimSpace(str))

	var isFront bool
	var rest string

	if strings.HasPrefix(lower, "front of ") {
		isFront = true
		rest = strings.TrimSpace(lower[9:])
	} else if strings.HasPrefix(lower, "back of ") {
		isFront = false
		rest = strings.TrimSpace(lower[8:])
	} else {
		return time.Time{}, false
	}

	// Parse the hour, possibly with am/pm suffix
	ampm := ""
	if strings.HasSuffix(rest, "am") {
		ampm = "am"
		rest = strings.TrimSpace(rest[:len(rest)-2])
	} else if strings.HasSuffix(rest, "pm") {
		ampm = "pm"
		rest = strings.TrimSpace(rest[:len(rest)-2])
	}

	hour, err := strconv.Atoi(rest)
	if err != nil || hour < 0 || hour > 24 {
		return time.Time{}, false
	}

	if ampm != "" {
		hour = applyAMPM(hour, ampm)
	}

	year, month, day := now.Date()
	if isFront {
		// "front of N" = (N-1):45
		return time.Date(year, month, day, hour-1, 45, 0, 0, loc), true
	}
	// "back of N" = N:15
	return time.Date(year, month, day, hour, 15, 0, 0, loc), true
}

// romanNumeralMonths maps Roman numerals to month numbers
var romanNumeralMonths = map[string]time.Month{
	"i": time.January, "ii": time.February, "iii": time.March,
	"iv": time.April, "v": time.May, "vi": time.June,
	"vii": time.July, "viii": time.August, "ix": time.September,
	"x": time.October, "xi": time.November, "xii": time.December,
}

// parseRomanNumeralDate parses dates with Roman numeral months: "20 VI. 2005", "1 III 2010"
func parseRomanNumeralDate(str string, loc *time.Location) (time.Time, bool) {
	fields := strings.Fields(str)
	if len(fields) < 3 {
		return time.Time{}, false
	}

	// Parse day
	day, err := strconv.Atoi(fields[0])
	if err != nil || day < 1 || day > 31 {
		return time.Time{}, false
	}

	// Parse Roman numeral month (strip trailing period)
	monthStr := strings.ToLower(strings.TrimSuffix(fields[1], "."))
	month, ok := romanNumeralMonths[monthStr]
	if !ok {
		return time.Time{}, false
	}

	// Parse year
	year, err := strconv.Atoi(fields[2])
	if err != nil {
		return time.Time{}, false
	}

	if !IsValidDate(year, int(month), day) {
		return time.Time{}, false
	}

	return time.Date(year, month, day, 0, 0, 0, 0, loc), true
}

// parseNumberedWeekday parses formats like "1 Monday December 2008", "second Monday December 2008"
// It handles formats like "first Monday of December 2008" or "3rd Friday of January"
// Also handles "+N week Thursday Nov 2007" (week offset + weekday + month + year)
func parseNumberedWeekday(str string, now time.Time, loc *time.Location) (time.Time, bool) {
	fields := strings.Fields(str)
	if len(fields) < 3 {
		return time.Time{}, false
	}

	idx := 0
	var ordinal int
	isWordOrdinal := false // tracks "first/second/third" vs numeric ordinals

	// Parse the ordinal (numeric or word)
	// Also handle bare weekday name as first field: "Thursday Nov 2007" = first Thursday
	if n, err := strconv.Atoi(fields[idx]); err == nil {
		if n <= 0 || n > 5 {
			return time.Time{}, false
		}
		ordinal = n
		idx++
	} else {
		switch strings.ToLower(fields[idx]) {
		case "first", "1st":
			ordinal = 1
			isWordOrdinal = true
			idx++
		case "second", "2nd":
			ordinal = 2
			isWordOrdinal = true
			idx++
		case "third", "3rd":
			ordinal = 3
			isWordOrdinal = true
			idx++
		case "fourth", "4th":
			ordinal = 4
			isWordOrdinal = true
			idx++
		case "fifth", "5th":
			ordinal = 5
			isWordOrdinal = true
			idx++
		case "last":
			ordinal = -1
			idx++
		default:
			// Check if first field is a weekday name (implies ordinal=1)
			if getDayOfWeek(fields[idx]) >= 0 {
				ordinal = 1
				// Don't advance idx — the weekday will be parsed next
			} else {
				return time.Time{}, false
			}
		}
	}

	// Handle "+N week(s) Weekday Month Year" — skip "week(s)" after numeric ordinal
	if idx < len(fields) {
		unit := normalizeTimeUnit(fields[idx])
		if unit == UnitWeek {
			isWordOrdinal = true // week-based ordinals use same semantics as word ordinals
			idx++
		}
	}

	// Parse the day of week or "day" keyword
	if idx >= len(fields) {
		return time.Time{}, false
	}
	isDayOfMonth := false
	dayOfWeek := getDayOfWeek(fields[idx])
	if dayOfWeek < 0 {
		if strings.ToLower(fields[idx]) == "day" {
			isDayOfMonth = true
		} else {
			return time.Time{}, false
		}
	}
	idx++

	// Check for optional "of" — its presence changes ordinal semantics
	hasOf := false
	if idx < len(fields) && strings.ToLower(fields[idx]) == "of" {
		hasOf = true
		idx++
	}

	// Parse the month context: either a literal month name or "next/last month/year"
	if idx >= len(fields) {
		return time.Time{}, false
	}

	var month time.Month
	var year int

	direction := strings.ToLower(fields[idx])
	if direction == DirectionNext || direction == DirectionLast {
		// Relative month/year: "next month", "last year", etc.
		idx++
		if idx >= len(fields) {
			return time.Time{}, false
		}
		unit := normalizeTimeUnit(fields[idx])
		idx++

		switch unit {
		case UnitMonth:
			if direction == DirectionNext {
				ref := now.AddDate(0, 1, 0)
				month = ref.Month()
				year = ref.Year()
			} else {
				ref := now.AddDate(0, -1, 0)
				month = ref.Month()
				year = ref.Year()
			}
		case UnitYear:
			if direction == DirectionNext {
				month = time.January
				year = now.Year() + 1
			} else {
				month = time.December
				year = now.Year() - 1
			}
		default:
			return time.Time{}, false
		}
	} else {
		// Literal month name
		var ok bool
		month, ok = getMonthByName(fields[idx])
		if !ok {
			return time.Time{}, false
		}
		idx++

		// Parse the optional year
		year = now.Year()
		if idx < len(fields) {
			var err error
			year, err = strconv.Atoi(fields[idx])
			if err != nil || year < 1 || year > 9999 {
				return time.Time{}, false
			}
			idx++
		}
	}

	// Make sure we consumed all fields
	if idx != len(fields) {
		return time.Time{}, false
	}

	var resultDay int

	if isDayOfMonth {
		// "first/last day of" — use ordinal directly as the day number
		lastDay := daysInMonth(year, month)
		if ordinal > 0 {
			resultDay = ordinal
			if resultDay > lastDay {
				return time.Time{}, false
			}
		} else if ordinal == -1 {
			resultDay = lastDay
		} else {
			return time.Time{}, false
		}
	} else {
		// Weekday-based: find the nth occurrence of the specified day of week
		firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)
		firstDayOfWeek := int(firstOfMonth.Weekday())
		daysUntilFirst := (dayOfWeek - firstDayOfWeek + 7) % 7

		if ordinal > 0 {
			if isWordOrdinal && !hasOf {
				// PHP semantics: "first Thursday Nov" = +1 week from first occurrence
				// "first/second/third" without "of" skips forward by ordinal weeks
				resultDay = 1 + daysUntilFirst + ordinal*7
			} else {
				// With "of" or numeric ordinals: Nth occurrence in the month
				resultDay = 1 + daysUntilFirst + (ordinal-1)*7
			}

			lastDayOfMonth := daysInMonth(year, month)
			if resultDay > lastDayOfMonth {
				return time.Time{}, false
			}
		} else if ordinal == -1 {
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
	}

	return time.Date(year, month, resultDay, 0, 0, 0, 0, loc), true
}
