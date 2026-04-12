package strtotime

import (
	"strconv"
	"strings"
	"time"
)

// A componentParser is the Into variant of a format parser: it populates a
// ParsedDate with the components it extracted from the input, and returns
// true if the input matched its format.
type componentParser func(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool

// formatParsers is the ordered pipeline shared by StrToTime and DateParse.
// Each entry is a wrapper around one of the parse* functions in
// date_formats.go / extended_formats.go / iso8601.go / date_with_timezone.go,
// with explicit knowledge of which ParsedDate fields that parser populates.
var formatParsers = []componentParser{
	guardDigit(wrapDateOnly(parseEuropeanFormat)),
	guardPrefix("front of ", "back of ")(parseFrontBackOfInto),
	guardDigit(wrapDateOnly(parseRomanNumeralDate)),
	guardPrefix("0000-00-00")(parseZeroDateInto),
	guardByte('-', '+')(parseSignedYearInto),
	parseISO8601Into,
	parseDateTimeFormatInto,
	parseWithTimezoneInto,
	wrapDateOnly(parseISOFormat),
	guardDigit(parseLargeYearAsTimeInto),
	parseYearMonthFormatInto,
	guardDigit(wrapDateOnly(parseSlashFormat)),
	guardDigit(wrapDateOnly(parseUSFormat)),
	guardDigit(parseUSDateWithTimeInto),
	guardDigit(parseShortYearUSDateWithMilitaryTimeInto),
	guardDigit(parseCompactTimestampInto),
	parseCompactTimeFormatsInto,
	parseMonthNameFormatInto,
	guardDigit(parseHTTPLogFormatInto),
	parseDateTimeTZRelativeInto,
	parseDateWithTZInto,
	parseDayMonthYearInto,
	parseMonthYearOnlyInto,
	guardDigit(parseTimeBeforeDateInto),
	parseMonthDayTimeYearInto,
	parseFirstLastDayOfDateInto,
	parseNumberedWeekdayInto,
}

// --- guards (componentParser flavor) ---

func guardDigit(fn componentParser) componentParser {
	return func(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
		if len(str) == 0 || str[0] < '0' || str[0] > '9' {
			return false
		}
		return fn(str, now, loc, opts, pd)
	}
}

func guardByte(bytes ...byte) func(componentParser) componentParser {
	return func(fn componentParser) componentParser {
		return func(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
			if len(str) == 0 {
				return false
			}
			for _, b := range bytes {
				if str[0] == b {
					return fn(str, now, loc, opts, pd)
				}
			}
			return false
		}
	}
}

func guardPrefix(prefixes ...string) func(componentParser) componentParser {
	return func(fn componentParser) componentParser {
		return func(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
			for _, p := range prefixes {
				if strings.HasPrefix(str, p) {
					return fn(str, now, loc, opts, pd)
				}
			}
			return false
		}
	}
}

// wrapDateOnly wraps a legacy date-only parser (yields only y/m/d with time = 00:00:00)
// as a componentParser that records SetDate only.
func wrapDateOnly(fn func(string, *time.Location) (time.Time, bool)) componentParser {
	return func(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
		t, ok := fn(str, loc)
		if !ok {
			return false
		}
		pd.SetDate(t.Year(), int(t.Month()), t.Day())
		pd.setMaterialized(t)
		return true
	}
}

// --- Into adapters (each wraps a specific legacy parser) ---

func parseZeroDateInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseZeroDate(str, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.setMaterialized(t)
	return true
}

func parseFrontBackOfInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseFrontBackOf(str, now, loc)
	if !ok {
		return false
	}
	pd.setMaterialized(t)
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.SetTime(t.Hour(), t.Minute(), t.Second())
	return true
}

func parseSignedYearInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseSignedYear(str, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	if t.Hour() != 0 || t.Minute() != 0 || t.Second() != 0 {
		pd.SetTime(t.Hour(), t.Minute(), t.Second())
	}
	if t.Location() != loc {
		setTZFromLocation(pd, t.Location(), t)
	}
	pd.setMaterialized(t)
	return true
}

func parseISO8601Into(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	// Week date: YYYY-Www[-D] — PHP represents this as year-Jan-1 plus a
	// relative day offset equal to (week-1)*7 + (dayOfWeek-1).
	if t, ok := parseISOWeekDate(str, loc); ok {
		year, week, dayOfWeek := extractWeekDateParts(str)
		if year > 0 {
			pd.SetDate(year, 1, 1)
			days := (week-1)*7 + dayOfWeek
			if days != 0 {
				pd.AddRelative(UnitDay, days)
			}
			pd.setMaterialized(t)
			pd.relativeApplied = true
			return true
		}
		pd.SetDate(t.Year(), int(t.Month()), t.Day())
		pd.setMaterialized(t)
		return true
	}
	// Date-time with T separator
	if t, ok := parseISO8601DateTime(str, loc); ok {
		// Populate components via the decomposing variant, but also
		// remember the full materialized time to preserve legacy semantics
		// (24:00 rollover, DST gap handling).
		parseISO8601DateTimeInto(str, loc, pd)
		pd.setMaterialized(t)
		return true
	}
	return false
}

// parseISO8601DateTimeInto reparses an ISO 8601 date+time string, populating
// pd with exactly the components that were present in the input.
func parseISO8601DateTimeInto(str string, loc *time.Location, pd *ParsedDate) bool {
	tIdx := -1
	for i := 1; i < len(str)-1; i++ {
		if str[i] == 't' && str[i-1] >= '0' && str[i-1] <= '9' && str[i+1] >= '0' && str[i+1] <= '9' {
			tIdx = i
			break
		}
	}
	if tIdx < 0 {
		return false
	}

	datePart := str[:tIdx]
	rest := str[tIdx+1:]

	var year, month, day int
	if strings.Contains(datePart, "-") {
		t, ok := parseISOFormat(datePart, loc)
		if !ok {
			return false
		}
		year = t.Year()
		month = int(t.Month())
		day = t.Day()
	} else if len(datePart) >= 8 && isAllDigits(datePart) {
		year, _ = strconv.Atoi(datePart[:len(datePart)-4])
		month, _ = strconv.Atoi(datePart[len(datePart)-4 : len(datePart)-2])
		day, _ = strconv.Atoi(datePart[len(datePart)-2:])
		if !IsValidDate(year, month, day) {
			return false
		}
	} else {
		return false
	}

	hour, minute, second, nanos, timeConsumed, ok := parseISO8601Time(rest)
	if !ok {
		return false
	}

	// Delegate the heavy lifting (validation, 24:00 handling) to the legacy
	// parser, but also populate pd. Rather than reimplement everything, call
	// legacy then overwrite:
	t, ok := parseISO8601DateTime(str, loc)
	if !ok {
		return false
	}

	pd.SetDate(year, month, day)
	pd.SetTime(hour, minute, second)
	if nanos > 0 {
		pd.SetFraction(float64(nanos) / 1e9)
	}

	// Detect and record timezone metadata if the input had one.
	tzRest := strings.TrimLeft(rest[timeConsumed:], " ")
	if len(tzRest) > 0 {
		recordISO8601TZ(tzRest, pd, t)
	}

	return true
}

// recordISO8601TZ parses a trailing timezone suffix from an ISO 8601 input
// and records its metadata in pd.
func recordISO8601TZ(tzStr string, pd *ParsedDate, t time.Time) {
	// PHP treats "Z" as an abbreviation (zone_type 2), not a UTC offset.
	if strings.EqualFold(tzStr, "Z") {
		pd.SetTZAbbreviation(time.UTC, "Z", 0, false)
		return
	}
	// Numeric offset (+HH:MM, +HHMM, +HH)
	if _, consumed, ok := parseNumericTimezoneOffset(tzStr); ok {
		remaining := strings.TrimSpace(tzStr[consumed:])
		if remaining != "" {
			return
		}
		offset := computeOffsetSeconds(tzStr)
		pd.SetTZOffset(t.Location(), offset)
		return
	}
	// Named (abbreviation or IANA)
	if loc, found := tryParseTimezone(tzStr); found {
		setTZFromLocation(pd, loc, t)
	}
}

// computeOffsetSeconds parses "+HH[:]MM" / "-HHMM" / "Z" into an offset in seconds.
func computeOffsetSeconds(s string) int {
	s = strings.TrimSpace(s)
	if s == "z" || s == "Z" {
		return 0
	}
	if len(s) < 2 {
		return 0
	}
	sign := 1
	switch s[0] {
	case '+':
		sign = 1
	case '-':
		sign = -1
	default:
		return 0
	}
	body := s[1:]
	body = strings.ReplaceAll(body, ":", "")
	if len(body) < 2 {
		return 0
	}
	h, _ := strconv.Atoi(body[:2])
	m := 0
	if len(body) >= 4 {
		m, _ = strconv.Atoi(body[2:4])
	}
	return sign * (h*3600 + m*60)
}

func parseDateTimeFormatInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseDateTimeFormat(str, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.SetTime(t.Hour(), t.Minute(), t.Second())
	if t.Nanosecond() != 0 {
		pd.SetFraction(float64(t.Nanosecond()) / 1e9)
	}
	if t.Location() != loc {
		setTZFromLocation(pd, t.Location(), t)
	}
	pd.setMaterialized(t)
	return true
}

func parseWithTimezoneInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseWithTimezone(str, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	if strings.Contains(str, ":") {
		pd.SetTime(t.Hour(), t.Minute(), t.Second())
	} else if t.Hour() != 0 || t.Minute() != 0 || t.Second() != 0 {
		pd.SetTime(t.Hour(), t.Minute(), t.Second())
	}
	if t.Location() != loc {
		setTZFromLocation(pd, t.Location(), t)
	}
	pd.setMaterialized(t)
	return true
}

func parseLargeYearAsTimeInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseLargeYearAsTime(str, now, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.SetTime(t.Hour(), t.Minute(), t.Second())
	if t.Location() != loc {
		setTZFromLocation(pd, t.Location(), t)
	}
	pd.setMaterialized(t)
	return true
}

func parseYearMonthFormatInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	if len(str) == 0 || str[0] < '0' || str[0] > '9' {
		return false
	}
	t, ok := parseYearMonthFormat(str, loc)
	if !ok {
		return false
	}
	// Disambiguate YYYY-MM (month, no day) vs YYYY-DDD (day of year, day set).
	parts := strings.SplitN(str, "-", 2)
	if len(parts) == 2 && len(parts[1]) == 3 {
		pd.SetDate(t.Year(), int(t.Month()), t.Day())
	} else {
		pd.SetYear(t.Year())
		pd.SetMonth(int(t.Month()))
	}
	pd.setMaterialized(t)
	return true
}

func parseUSDateWithTimeInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseUSDateWithTime(str, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.SetTime(t.Hour(), t.Minute(), t.Second())
	pd.setMaterialized(t)
	return true
}

func parseShortYearUSDateWithMilitaryTimeInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseShortYearUSDateWithMilitaryTime(str, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.SetTime(t.Hour(), t.Minute(), t.Second())
	pd.setMaterialized(t)
	return true
}

func parseCompactTimestampInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseCompactTimestamp(str, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	digits := strings.SplitN(str, " ", 2)[0]
	if len(digits) == 14 {
		pd.SetTime(t.Hour(), t.Minute(), t.Second())
	}
	if t.Location() != loc {
		setTZFromLocation(pd, t.Location(), t)
	}
	pd.setMaterialized(t)
	return true
}

func parseCompactTimeFormatsInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseCompactTimeFormats(str, now, loc)
	if !ok {
		return false
	}
	pd.SetTime(t.Hour(), t.Minute(), t.Second())
	if t.Nanosecond() != 0 {
		pd.SetFraction(float64(t.Nanosecond()) / 1e9)
	}
	if len(str) == 7 && isAllDigits(str) {
		pd.SetDate(t.Year(), int(t.Month()), t.Day())
	}
	if t.Location() != loc {
		setTZFromLocation(pd, t.Location(), t)
	}
	pd.setMaterialized(t)
	return true
}

func parseMonthNameFormatInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseMonthNameFormat(str, loc)
	if !ok {
		return false
	}
	// PHP omits the year when the input didn't contain one (e.g. "Dec 25").
	if hasFourDigitYear(str) {
		pd.SetDate(t.Year(), int(t.Month()), t.Day())
	} else {
		pd.SetMonth(int(t.Month()))
		pd.SetDay(t.Day())
	}
	pd.setMaterialized(t)
	return true
}

// hasFourDigitYear reports whether str contains a 4-digit number that could
// serve as an explicit year.
func hasFourDigitYear(str string) bool {
	run := 0
	for i := 0; i <= len(str); i++ {
		if i < len(str) && str[i] >= '0' && str[i] <= '9' {
			run++
			continue
		}
		if run == 4 {
			return true
		}
		run = 0
	}
	return false
}

func parseHTTPLogFormatInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseHTTPLogFormat(str, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.SetTime(t.Hour(), t.Minute(), t.Second())
	if t.Location() != loc {
		setTZFromLocation(pd, t.Location(), t)
	}
	pd.setMaterialized(t)
	return true
}

func parseDateTimeTZRelativeInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	// Detect the relative-expression suffix(es) and extract them into the pd
	// Relative block, then parse the remaining date+tz portion.
	type relExpr struct {
		amount int
		unit   string
	}
	var rels []relExpr
	remaining := str

	for {
		remaining = strings.TrimRight(remaining, " ")
		if len(remaining) == 0 {
			break
		}
		found := false
		for i := len(remaining) - 1; i > 0; i-- {
			if (remaining[i] == '+' || remaining[i] == '-') && remaining[i-1] == ' ' {
				relPart := remaining[i:]
				relFields := strings.Fields(relPart)
				if len(relFields) != 2 {
					continue
				}
				amount, err := strconv.Atoi(relFields[0])
				if err != nil {
					continue
				}
				unit := normalizeTimeUnit(relFields[1])
				switch unit {
				case UnitDay, UnitWeek, UnitWeekDay, UnitMonth, UnitYear,
					UnitHour, UnitMinute, UnitSecond:
					rels = append(rels, relExpr{amount, unit})
					remaining = strings.TrimSpace(remaining[:i])
					found = true
				default:
					continue
				}
				break
			}
		}
		if !found {
			break
		}
	}
	if len(rels) == 0 {
		return false
	}

	datePart := remaining
	sub := newParsedDate()
	ok := parseISO8601Into(datePart, now, loc, opts, sub) ||
		parseDateTimeFormatInto(datePart, now, loc, opts, sub) ||
		parseWithTimezoneInto(datePart, now, loc, opts, sub) ||
		parseDateWithTZInto(datePart, now, loc, opts, sub) ||
		wrapDateOnly(parseISOFormat)(datePart, now, loc, opts, sub) ||
		parseDayMonthYearInto(datePart, now, loc, opts, sub)
	if !ok {
		return false
	}

	copyComponents(pd, sub)
	if sub.hasMaterialized {
		pd.setMaterialized(sub.materialized)
	}
	for i := len(rels) - 1; i >= 0; i-- {
		pd.AddRelative(rels[i].unit, rels[i].amount)
	}
	return true
}

func parseDateWithTZInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	fields := strings.Fields(str)
	if len(fields) != 2 {
		return false
	}
	t, ok := parseISOFormat(fields[0], loc)
	if !ok {
		return false
	}
	tzLoc, found := tryParseTimezone(fields[1])
	if !found {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	setTZFromName(pd, fields[1], tzLoc)
	// Rebuild with target tz since parseISOFormat used the original loc.
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tzLoc)
	pd.setMaterialized(t)
	return true
}

func parseDayMonthYearInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	// Peek weekday prefix so we can record it as relative.weekday.
	_, dayNum, stripped := stripWeekdayPrefix(str)

	t, ok := parseDayMonthYear(str, now, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	// Time is recorded whenever a time portion was present. Detect via ":"
	// in the input (post-weekday-prefix, after stripping the date fields).
	if strings.Contains(str, ":") {
		pd.SetTime(t.Hour(), t.Minute(), t.Second())
		if t.Nanosecond() != 0 {
			pd.SetFraction(float64(t.Nanosecond()) / 1e9)
		}
	}
	// Detect an explicit trailing timezone in the input regardless of whether
	// t.Location() == loc — both may happen to be UTC.
	if tzStr, ok := lastFieldLooksLikeTZ(str); ok {
		if len(tzStr) > 0 && (tzStr[0] == '+' || tzStr[0] == '-') {
			offset := computeOffsetSeconds(tzStr)
			pd.SetTZOffset(t.Location(), offset)
		} else if resolved, found := tryParseTimezone(tzStr); found {
			setTZFromName(pd, tzStr, resolved)
		}
	}
	if stripped && dayNum >= 0 {
		pd.SetRelativeWeekday(dayNum)
	}
	pd.setMaterialized(t)
	return true
}

// lastFieldLooksLikeTZ returns the last whitespace-separated field if it
// looks like a timezone token (numeric offset, IANA name with "/", or a
// resolvable abbreviation).
func lastFieldLooksLikeTZ(str string) (string, bool) {
	fields := strings.Fields(str)
	if len(fields) == 0 {
		return "", false
	}
	last := fields[len(fields)-1]
	if len(last) == 0 {
		return "", false
	}
	if last[0] == '+' || last[0] == '-' {
		if _, _, ok := parseNumericTimezoneOffset(last); ok {
			return last, true
		}
	}
	if _, ok := tryParseTimezone(last); ok {
		return last, true
	}
	return "", false
}

func parseMonthYearOnlyInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseMonthYearOnly(str, loc)
	if !ok {
		return false
	}
	// PHP date_parse fills in day=1 for "Month Year" inputs.
	pd.SetDate(t.Year(), int(t.Month()), 1)
	pd.setMaterialized(t)
	return true
}

func parseTimeBeforeDateInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseTimeBeforeDate(str, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.SetTime(t.Hour(), t.Minute(), t.Second())
	pd.setMaterialized(t)
	return true
}

func parseMonthDayTimeYearInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseMonthDayTimeYear(str, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.SetTime(t.Hour(), t.Minute(), t.Second())
	pd.setMaterialized(t)
	return true
}

func parseFirstLastDayOfDateInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseFirstLastDayOfDate(str, now, loc)
	if !ok {
		return false
	}
	// PHP reports day=1 for "first/last day of <Month> <Year>" and records
	// the first_day_of_month / last_day_of_month flag on the relative block.
	if ordinal, ok := detectFirstLastDayOfMonthYear(str); ok {
		pd.SetDate(t.Year(), int(t.Month()), 1)
		if ordinal == 1 {
			pd.SetFirstLastDayOf(1)
		} else {
			pd.SetFirstLastDayOf(2)
		}
		pd.setMaterialized(t)
		pd.relativeApplied = true
		return true
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.setMaterialized(t) // preserves base-time hour/minute/second for +/- forms
	return true
}

func parseNumberedWeekdayInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseNumberedWeekday(str, now, loc)
	if !ok {
		return false
	}
	// "first day of next/last/this month/year" is purely relative; report it
	// as such rather than as an absolute date derived from a zero base time.
	if ordinal, unit, direction, isRelFirstLast := detectRelativeFirstLastDay(str); isRelFirstLast {
		if ordinal == 1 {
			pd.SetFirstLastDayOf(1)
		} else {
			pd.SetFirstLastDayOf(2)
		}
		var amount int
		switch direction {
		case DirectionNext:
			amount = 1
		case DirectionLast:
			amount = -1
		}
		if amount != 0 {
			pd.AddRelative(unit, amount)
		}
		pd.setMaterialized(t)
		pd.relativeApplied = true
		return true
	}
	// "first/last day of <Month> [<Year>]" → PHP reports year/month with
	// day=1 and flags the relative first/last_day_of_month, regardless of
	// whether it's actually the last day.
	if ordinal, ok := detectFirstLastDayOfMonthYear(str); ok {
		if ordinal == 1 {
			pd.SetFirstLastDayOf(1)
		} else {
			pd.SetFirstLastDayOf(2)
		}
		pd.SetDate(t.Year(), int(t.Month()), 1)
		// The materialized t from parseNumberedWeekday already resolved the
		// correct day; use it so StrToTime keeps working.
		pd.setMaterialized(t)
		pd.relativeApplied = true
		return true
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.setMaterialized(t)
	return true
}

// extractWeekDateParts parses "YYYY-Www[-D]" or "YYYYWww[D]" and returns
// (year, week, dayOfWeek). Returns year=0 when the input isn't a week date.
func extractWeekDateParts(str string) (year, week, dayOfWeek int) {
	dayOfWeek = 1
	wIdx := -1
	for i := 1; i < len(str); i++ {
		if str[i] == 'w' && str[i-1] >= '0' && str[i-1] <= '9' {
			wIdx = i
			break
		}
		if str[i] == 'w' && str[i-1] == '-' && i >= 2 && str[i-2] >= '0' && str[i-2] <= '9' {
			wIdx = i
			break
		}
	}
	if wIdx < 0 {
		return 0, 0, 0
	}
	yearPart := strings.TrimSuffix(str[:wIdx], "-")
	if !isAllDigits(yearPart) || len(yearPart) == 0 {
		return 0, 0, 0
	}
	y, err := strconv.Atoi(yearPart)
	if err != nil {
		return 0, 0, 0
	}
	year = y
	rest := str[wIdx+1:]
	if len(rest) < 2 || !isAllDigits(rest[:2]) {
		return 0, 0, 0
	}
	w, _ := strconv.Atoi(rest[:2])
	week = w
	rest = rest[2:]
	if len(rest) > 0 && rest[0] == '-' {
		rest = rest[1:]
	}
	if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
		dow, _ := strconv.Atoi(rest[:1])
		if dow >= 1 && dow <= 7 {
			dayOfWeek = dow
		}
	}
	return year, week, dayOfWeek
}

// detectFirstLastDayOfMonthYear matches "first/last day of <MonthName> [<Year>]".
func detectFirstLastDayOfMonthYear(str string) (ordinal int, ok bool) {
	fields := strings.Fields(str)
	if len(fields) < 4 || len(fields) > 5 {
		return 0, false
	}
	switch strings.ToLower(fields[0]) {
	case "first":
		ordinal = 1
	case "last":
		ordinal = -1
	default:
		return 0, false
	}
	if strings.ToLower(fields[1]) != "day" || strings.ToLower(fields[2]) != "of" {
		return 0, false
	}
	if _, isMonth := getMonthByNameFlex(fields[3]); !isMonth {
		return 0, false
	}
	return ordinal, true
}

// detectRelativeFirstLastDay recognises "first/last day of next/last/this
// month/year" inputs and returns the ordinal (1 or -1), unit (month/year),
// direction, and true. Otherwise returns false.
func detectRelativeFirstLastDay(str string) (ordinal int, unit string, direction string, ok bool) {
	fields := strings.Fields(str)
	if len(fields) != 5 {
		return 0, "", "", false
	}
	switch strings.ToLower(fields[0]) {
	case "first":
		ordinal = 1
	case "last":
		ordinal = -1
	default:
		return 0, "", "", false
	}
	if strings.ToLower(fields[1]) != "day" || strings.ToLower(fields[2]) != "of" {
		return 0, "", "", false
	}
	direction = strings.ToLower(fields[3])
	if direction != DirectionNext && direction != DirectionLast && direction != "this" {
		return 0, "", "", false
	}
	unit = normalizeTimeUnit(fields[4])
	if unit != UnitMonth && unit != UnitYear {
		return 0, "", "", false
	}
	return ordinal, unit, direction, true
}

// --- post-pipeline Into adapters ---

func parseUnixTimestampInto(str string, loc *time.Location, pd *ParsedDate) bool {
	t, ok := tryParseUnixTimestamp(str, loc)
	if !ok {
		return false
	}
	// PHP reports @N as an absolute 1970-01-01 00:00:00 UTC plus a relative
	// offset of N seconds. Fractional seconds are parsed but discarded.
	seconds := int64(0)
	if str[0] == '@' {
		body := str[1:]
		if idx := strings.Index(body, "."); idx >= 0 {
			body = body[:idx]
		}
		if space := strings.Index(body, " "); space >= 0 {
			body = body[:space]
		}
		seconds, _ = strconv.ParseInt(body, 10, 64)
	}
	pd.SetDate(1970, 1, 1)
	pd.SetTime(0, 0, 0)
	pd.SetFraction(0)
	pd.IsLocaltime = true
	pd.ZoneType = 1
	pd.Zone = 0
	pd.IsDST = false
	pd.sourceLoc = time.UTC
	pd.relative().Second = int(seconds)
	pd.setMaterialized(t)
	pd.relativeApplied = true
	return true
}

func parseKeywordInto(str string, now time.Time, loc *time.Location, pd *ParsedDate) bool {
	switch str {
	case "now":
		// PHP leaves every component false for "now".
		pd.setMaterialized(now)
		return true
	case "today":
		pd.SetTime(0, 0, 0)
		pd.SetFraction(0)
		y, m, d := now.Date()
		pd.setMaterialized(time.Date(y, m, d, 0, 0, 0, 0, loc))
		return true
	case "midnight":
		pd.SetTime(0, 0, 0)
		pd.SetFraction(0)
		y, m, d := now.Date()
		pd.setMaterialized(time.Date(y, m, d, 0, 0, 0, 0, loc))
		return true
	case "noon":
		pd.SetTime(12, 0, 0)
		pd.SetFraction(0)
		y, m, d := now.Date()
		pd.setMaterialized(time.Date(y, m, d, 12, 0, 0, 0, loc))
		return true
	case "tomorrow":
		pd.SetTime(0, 0, 0)
		pd.SetFraction(0)
		pd.AddRelative(UnitDay, 1)
		t := now.AddDate(0, 0, 1)
		y, m, d := t.Date()
		pd.setMaterialized(time.Date(y, m, d, 0, 0, 0, 0, loc))
		pd.relativeApplied = true
		return true
	case "yesterday":
		pd.SetTime(0, 0, 0)
		pd.SetFraction(0)
		pd.AddRelative(UnitDay, -1)
		t := now.AddDate(0, 0, -1)
		y, m, d := t.Date()
		pd.setMaterialized(time.Date(y, m, d, 0, 0, 0, 0, loc))
		pd.relativeApplied = true
		return true
	}
	return false
}

func parseDateWithRelativeTimeInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	t, ok := parseDateWithRelativeTime(str, now, loc, opts)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	if t.Hour() != 0 || t.Minute() != 0 || t.Second() != 0 {
		pd.SetTime(t.Hour(), t.Minute(), t.Second())
	}
	pd.setMaterialized(t)
	return true
}

func tryWeekdayPrefixReparseInto(str string, now time.Time, loc *time.Location, opts []Option, pd *ParsedDate) bool {
	// Peek at the weekday prefix before delegating, so we can record the
	// matching weekday in the relative block (PHP reports Thu/Fri/Sun as
	// weekday indexes on RFC2822-ish inputs).
	_, dayNum, stripped := stripWeekdayPrefix(str)

	t, ok := tryWeekdayPrefixReparse(str, now, loc, opts)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	// Always record time when a weekday prefix was present (PHP includes
	// hour/minute/second=0 even when the input omitted the time).
	if stripped {
		pd.SetTime(t.Hour(), t.Minute(), t.Second())
	} else if t.Hour() != 0 || t.Minute() != 0 || t.Second() != 0 {
		pd.SetTime(t.Hour(), t.Minute(), t.Second())
	}
	if stripped && dayNum >= 0 {
		pd.SetRelativeWeekday(dayNum)
	}
	if t.Location() != loc {
		setTZFromLocation(pd, t.Location(), t)
	}
	pd.setMaterialized(t)
	return true
}

func parseCompoundExpressionInto(str string, now time.Time, opts []Option, pd *ParsedDate) bool {
	t, err := parseCompoundExpression(str, now, opts)
	if err != nil {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	pd.SetTime(t.Hour(), t.Minute(), t.Second())
	pd.setMaterialized(t)
	return true
}

func parseOrdinalDateInto(str string, now time.Time, loc *time.Location, pd *ParsedDate) bool {
	t, ok := parseOrdinalDate(str, now, loc)
	if !ok {
		return false
	}
	pd.SetDate(t.Year(), int(t.Month()), t.Day())
	return true
}

// --- helpers ---

// copyComponents copies populated fields from src into dst, preserving any
// fields already set on dst. Used by parsers that dispatch to a subparser.
func copyComponents(dst, src *ParsedDate) {
	if src.Year.Set && !dst.Year.Set {
		dst.Year = src.Year
	}
	if src.Month.Set && !dst.Month.Set {
		dst.Month = src.Month
	}
	if src.Day.Set && !dst.Day.Set {
		dst.Day = src.Day
	}
	if src.Hour.Set && !dst.Hour.Set {
		dst.Hour = src.Hour
	}
	if src.Minute.Set && !dst.Minute.Set {
		dst.Minute = src.Minute
	}
	if src.Second.Set && !dst.Second.Set {
		dst.Second = src.Second
	}
	if src.Fraction.Set && !dst.Fraction.Set {
		dst.Fraction = src.Fraction
	}
	if src.IsLocaltime && !dst.IsLocaltime {
		dst.IsLocaltime = src.IsLocaltime
		dst.ZoneType = src.ZoneType
		dst.Zone = src.Zone
		dst.IsDST = src.IsDST
		dst.TzAbbr = src.TzAbbr
		dst.TzID = src.TzID
		dst.sourceLoc = src.sourceLoc
	}
}

// setTZFromLocation inspects a resolved *time.Location and populates pd's
// timezone metadata. t is the reference time used for DST detection.
func setTZFromLocation(pd *ParsedDate, loc *time.Location, t time.Time) {
	name, offset := t.In(loc).Zone()
	id := loc.String()

	// Numeric-offset locations built by fixedZone have names like "+09:00".
	if len(name) > 0 && (name[0] == '+' || name[0] == '-') {
		pd.SetTZOffset(loc, offset)
		return
	}

	// IANA identifiers contain a "/".
	if strings.Contains(id, "/") {
		pd.SetTZIdentifier(loc, id)
		return
	}

	applyAbbreviationTZ(pd, loc, name, offset)
}

// setTZFromName records timezone metadata given the original name string
// that matched (e.g. "Asia/Tokyo", "EST", "+09:00") and the resolved location.
func setTZFromName(pd *ParsedDate, name string, loc *time.Location) {
	if strings.Contains(name, "/") {
		pd.SetTZIdentifier(loc, canonicalTZName(name, loc))
		return
	}
	// Numeric offset?
	if len(name) > 0 && (name[0] == '+' || name[0] == '-') {
		offset := computeOffsetSeconds(name)
		pd.SetTZOffset(loc, offset)
		return
	}
	// PHP treats "Z" as an abbreviation.
	if strings.EqualFold(name, "Z") {
		pd.SetTZAbbreviation(time.UTC, "Z", 0, false)
		return
	}
	// Abbreviation (EST, GMT, PDT, etc.) — use the caller-supplied location
	// to read the current offset.
	now := time.Now()
	_, offset := now.In(loc).Zone()
	applyAbbreviationTZ(pd, loc, name, offset)
}

// applyAbbreviationTZ records the timezone metadata for an abbreviation like
// EST/GMT/PDT with PHP-specific quirks:
//   - UTC is zone_type 3 and includes both tz_abbr and tz_id.
//   - DST abbreviations (e.g. PDT, EDT) report the *standard* offset in zone
//     and set is_dst:true, matching PHP's timelib.
func applyAbbreviationTZ(pd *ParsedDate, loc *time.Location, abbr string, offset int) {
	upper := strings.ToUpper(abbr)

	// UTC: zone_type 3 with both tz_id and tz_abbr.
	if upper == "UTC" {
		pd.IsLocaltime = true
		pd.ZoneType = 3
		pd.Zone = 0
		pd.IsDST = false
		pd.TzAbbr = "UTC"
		pd.TzID = "UTC"
		pd.sourceLoc = loc
		return
	}

	// Check whether this abbreviation is a DST form. PHP emits the standard
	// offset for DST abbreviations and sets is_dst:true.
	if stdOffset, isDST, ok := dstAbbreviationOffset(upper); ok {
		pd.SetTZAbbreviation(loc, upper, stdOffset, isDST)
		return
	}

	pd.SetTZAbbreviation(loc, upper, offset, false)
}

// dstAbbreviationOffset returns the standard-time offset (in seconds) and
// is_dst flag for a PHP-recognised timezone abbreviation. Returns ok=false
// for unknown abbreviations so the caller falls back to the parsed offset.
func dstAbbreviationOffset(upper string) (offset int, isDST, ok bool) {
	switch upper {
	case "EDT":
		return -5 * 3600, true, true
	case "CDT":
		return -6 * 3600, true, true
	case "MDT":
		return -7 * 3600, true, true
	case "PDT":
		return -8 * 3600, true, true
	case "AKDT":
		return -9 * 3600, true, true
	case "BST":
		return 0, true, true
	case "CEST":
		return 1 * 3600, true, true
	case "EEST":
		return 2 * 3600, true, true
	case "AEDT":
		return 10 * 3600, true, true
	}
	return 0, false, false
}

// canonicalTZName returns the canonical IANA identifier for a matched name.
// The input may be lowercase ("asia/tokyo"); prefer loc.String() when it
// carries a proper ID.
func canonicalTZName(name string, loc *time.Location) string {
	s := loc.String()
	if strings.Contains(s, "/") {
		return s
	}
	// Fall back to title-casing the input.
	parts := strings.Split(name, "/")
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "/")
}
