package strtotime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StrToTime will convert the provided string into a time similarly to how PHP strtotime() works.
func StrToTime(str string, opts ...Option) (time.Time, error) {
	var now time.Time
	loc := time.Local // Default timezone to local
	tzExplicit := false

	for _, opt := range opts {
		switch v := opt.(type) {
		case Rel: // relative to
			now = time.Time(v)
		case tzOption: // timezone
			if v.loc != nil {
				loc = v.loc
				tzExplicit = true
			}
		}
	}

	// If a Rel time was provided but no explicit timezone, inherit from Rel
	if !now.IsZero() && !tzExplicit {
		loc = now.Location()
	}

	if now.IsZero() {
		now = time.Now().In(loc)
	} else if now.Location() != loc {
		now = now.In(loc)
	}

	// Normalize string - trim and lowercase
	str = strings.ToLower(strings.TrimSpace(str))
	if str == "" {
		return time.Time{}, ErrEmptyTimeString
	}

	// Try Unix timestamp format (@timestamp)
	if len(str) > 0 && str[0] == '@' {
		// Parse the Unix timestamp format (e.g., "@1121373041" or "@1121373041.5")
		unixTimeStr := str[1:]

		// Check if there's a timezone specification after the timestamp
		tzParts := strings.SplitN(unixTimeStr, " ", 2)
		timestamp := tzParts[0]

		// Check if timestamp has fractional seconds
		if idx := strings.Index(timestamp, "."); idx != -1 {
			// Parse the whole seconds part
			unixTime, err := strconv.ParseInt(timestamp[:idx], 10, 64)
			if err != nil {
				// If we can't parse the integer part, don't try to handle as Unix timestamp
				goto nextFormat
			}

			// Parse the fractional part as a float
			fracPart, err := strconv.ParseFloat("0."+timestamp[idx+1:], 64)
			if err != nil {
				// If we can't parse the fraction, just use the integer part
				fracPart = 0.0
			}

			// Convert fraction to nanoseconds (range: 0-999999999)
			nanoSec := int64(fracPart * 1e9)

			// Create the time with the proper Unix seconds and nanoseconds
			result := time.Unix(unixTime, nanoSec).In(loc)

			// If there's a timezone specified, try to use it
			if len(tzParts) > 1 && tzParts[1] != "" {
				if tzLoc, found := tryParseTimezone(tzParts[1]); found {
					result = result.In(tzLoc)
				}
			}

			return result, nil
		} else {
			// No fractional part, parse as an integer
			unixTime, err := strconv.ParseInt(timestamp, 10, 64)
			if err == nil {
				result := time.Unix(unixTime, 0).In(loc)

				// If there's a timezone specified, try to use it
				if len(tzParts) > 1 && tzParts[1] != "" {
					if tzLoc, found := tryParseTimezone(tzParts[1]); found {
						result = result.In(tzLoc)
					}
				}

				return result, nil
			}
		}
	}
nextFormat:

	// Try European date format like "24.11.22"
	if t, ok := parseEuropeanFormat(str, loc); ok {
		return t, nil
	}

	// Handle special cases for simple strings
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
	case "midnight":
		year, month, day := now.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, loc), nil
	case "noon":
		year, month, day := now.Date()
		return time.Date(year, month, day, 12, 0, 0, 0, loc), nil
	}

	// Try "front of" / "back of" Scottish time expressions
	if t, ok := parseFrontBackOf(str, now, loc); ok {
		return t, nil
	}

	// Try Roman numeral month date: "20 VI. 2005"
	if t, ok := parseRomanNumeralDate(str, loc); ok {
		return t, nil
	}

	// Try zero date special case
	if t, ok := parseZeroDate(str, loc); ok {
		return t, nil
	}

	// Try negative year format: -YYYY-MM-DD [HH:MM:SS [TZ]]
	if len(str) > 0 && str[0] == '-' {
		if t, ok := parseNegativeYear(str, loc); ok {
			return t, nil
		}
	}

	// Try explicit positive year format: +YYYYY...-MM-DD [HH:MM:SS [TZ]]
	if len(str) > 0 && str[0] == '+' {
		if t, ok := parsePositiveYear(str, loc); ok {
			return t, nil
		}
	}

	// Try ISO 8601 formats (T separator, week dates, timezone offsets)
	if t, ok := parseISO8601(str, loc); ok {
		return t, nil
	}

	// Try to parse datetime format (YYYY-MM-DD HH:MM:SS [TZ])
	if t, ok := parseDateTimeFormat(str, loc); ok {
		return t, nil
	}

	// Try date with timezone format
	if t, ok := parseWithTimezone(str, loc); ok {
		return t, nil
	}

	// Try standard date formats
	if t, ok := parseISOFormat(str, loc); ok {
		return t, nil
	}

	// Try YYYY-MM (year-month only)
	if t, ok := parseYearMonthFormat(str, loc); ok {
		return t, nil
	}

	if t, ok := parseSlashFormat(str, loc); ok {
		return t, nil
	}

	if t, ok := parseUSFormat(str, loc); ok {
		return t, nil
	}

	// Try US date with time (MM/DD/YYYY H:MM AM)
	if t, ok := parseUSDateWithTime(str, loc); ok {
		return t, nil
	}

	// Try short-year US date with military time (MM/DD/YY HHMM)
	if t, ok := parseShortYearUSDateWithMilitaryTime(str, loc); ok {
		return t, nil
	}

	// Try extended date formats
	if t, ok := parseCompactTimestamp(str, loc); ok {
		return t, nil
	}

	// Try compact time formats (hhmmss, year+doy, t-prefix, dotted time)
	if t, ok := parseCompactTimeFormats(str, now, loc); ok {
		return t, nil
	}

	if t, ok := parseMonthNameFormat(str, loc); ok {
		return t, nil
	}

	if t, ok := parseHTTPLogFormat(str, loc); ok {
		return t, nil
	}

	// Try date + timezone + relative: "2004-10-31 EDT +1 hour", "Mon, 08 May 2006 13:06:44 -0400 +30 days"
	// Must come before parseDayMonthYear to avoid swallowing the relative part
	if t, ok := parseDateTimeTZRelative(str, loc); ok {
		return t, nil
	}

	// Check if this is a date followed by a relative time adjustment
	// (must come before compound expression check)
	if result, ok := parseDateWithRelativeTime(str, now, loc, opts); ok {
		return result, nil
	}

	// Try DD Mon YYYY format (RFC 2822, etc.)
	if t, ok := parseDayMonthYear(str, loc); ok {
		return t, nil
	}

	// Try month + year only: "Oct 2001", "2001 Oct"
	if t, ok := parseMonthYearOnly(str, loc); ok {
		return t, nil
	}

	// Try time before date: "19:30 Dec 17 2005"
	if t, ok := parseTimeBeforeDate(str, loc); ok {
		return t, nil
	}

	// Try "Month Day Time Year": "Dec 17 19:30 2005"
	if t, ok := parseMonthDayTimeYear(str, loc); ok {
		return t, nil
	}

	// Try "first/last day of YYYY-MM" or "first/last day of +1 month"
	if t, ok := parseFirstLastDayOfDate(str, now, loc); ok {
		return t, nil
	}

	// Try date + TZ: "2014-01-01 Asia/Tokyo"
	if t, ok := parseDateWithTZ(str, loc); ok {
		return t, nil
	}

	// Try parsing numbered weekday (e.g. "first Monday of December 2008")
	if t, ok := parseNumberedWeekday(str, now, loc); ok {
		return t, nil
	}

	// Check if this is a compound expression (contains + or - in the middle)
	if isCompoundExpression(str) {
		return parseCompoundExpression(str, now, opts)
	}

	// Check if this is an ordinal date: "26th Nov"
	if result, ok := parseOrdinalDate(str, now, loc); ok {
		return result, nil
	}

	// Try stripping leading day-of-week name and reparsing the rest
	// Handles: "Sun 2017-01-01", "Fri Aug 20 1993 23:59:59", etc.
	{
		stripped := false
		var rest string
		// Try full weekday names first (must check before 3-letter to avoid partial match)
		for _, name := range []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"} {
			if strings.HasPrefix(str, name) && len(str) > len(name) {
				r := strings.TrimLeft(str[len(name):], ", ")
				if len(r) > 0 {
					rest = r
					stripped = true
					break
				}
			}
		}
		// Try 3-letter abbreviations
		if !stripped && len(str) > 3 {
			prefix := str[:3]
			if getDayOfWeek(prefix) >= 0 {
				r := strings.TrimLeft(str[3:], ", ")
				if len(r) > 0 {
					rest = r
					stripped = true
				}
			}
		}
		if stripped {
			// Don't strip if rest is "next/last week" — the token parser
			// handles "weekday next week [time]" as a unit (bug72719)
			restTrimmed := strings.TrimSpace(rest)
			if strings.HasPrefix(restTrimmed, "next ") || strings.HasPrefix(restTrimmed, "last ") {
				// Let the token parser handle it
			} else if t, err := StrToTime(rest, opts...); err == nil {
				return t, nil
			}
		}
	}

	// Tokenize the input string
	tokens := Tokenize(str)

	// Create a parser to process the tokens
	parser := &Parser{
		tokens:   tokens,
		position: 0,
		result:   now,
		loc:      loc,
	}

	// Parse tokens
	result, err := parser.Parse()
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse time string: %s: %w", str, err)
	}

	return result, nil
}

// Parser represents a token stream parser for time expressions
type Parser struct {
	tokens   []Token
	position int
	result   time.Time
	loc      *time.Location
	tzFound  bool // Flag to indicate if a timezone was parsed from the input
}

// Parse processes the token stream and returns a time.Time result
func (p *Parser) Parse() (time.Time, error) {
	// Skip any leading whitespace
	p.skipWhitespace()

	// Try standard date formats first
	if t, ok, err := p.tryParseStandardDate(); ok {
		return t, err
	}

	// Try relative expressions
	for p.position < len(p.tokens) {
		// Skip whitespace between expressions
		p.skipWhitespace()

		if p.position >= len(p.tokens) {
			break
		}

		// Try to parse each expression type
		parsed := false

		// Try to parse timezone
		if !p.tzFound {
			if ok := p.tryParseTimezone(); ok {
				parsed = true
			}
		}

		// Try "first/last day of this/next/last month/year"
		if !parsed {
			if t, ok, err := p.tryParseFirstLastDayOfExpression(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try "next/last" expressions
		if !parsed {
			if t, ok, err := p.tryParseNextLastExpression(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try bare weekday or "weekday next/last week [time]"
		if !parsed {
			if t, ok, err := p.tryParseBareWeekday(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try +/- relative time
		if !parsed {
			if t, ok, err := p.tryParseRelativeTime(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try implicit positive relative time (e.g., "4 days" without explicit +)
		if !parsed {
			if t, ok, err := p.tryParseImplicitRelativeTime(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try "N weekday ago"
		if !parsed {
			if t, ok, err := p.tryParseWeekdayAgo(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try ordinal word + unit (e.g., "eighth day")
		if !parsed {
			if t, ok, err := p.tryParseOrdinalRelativeTime(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try standalone time expression "HH:MM[:SS]"
		if !parsed {
			if t, ok, err := p.tryParseTimeExpression(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try time keywords "midnight" / "noon"
		if !parsed {
			if t, ok, err := p.tryParseTimeKeyword(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try month only format (e.g., "January" or "Feb")
		if !parsed {
			if t, ok, err := p.tryParseMonthOnlyFormat(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try month name format
		if !parsed {
			if t, ok, err := p.tryParseMonthNameFormat(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Try bare 4-digit year (must be last number check)
		if !parsed {
			if t, ok, err := p.tryParseYearOnly(); ok {
				if err != nil {
					return time.Time{}, err
				}
				p.result = t
				parsed = true
			}
		}

		// Handle unrecognized token
		if !parsed && p.position < len(p.tokens) {
			currentToken := p.tokens[p.position]
			p.position++
			if currentToken.Typ != TypeWhitespace {
				return time.Time{}, fmt.Errorf("unexpected token: %s", currentToken.Val)
			}
		}

		// Skip whitespace after expressions
		p.skipWhitespace()
	}

	return p.result, nil
}

// skipWhitespace advances the position past any whitespace tokens
func (p *Parser) skipWhitespace() {
	for p.position < len(p.tokens) && p.tokens[p.position].Typ == TypeWhitespace {
		p.position++
	}
}

// tryParseTimezone attempts to parse a timezone from the token stream
// This handles abbreviations (PST, EST), slash-separated paths (America/New_York,
// America/Argentina/Buenos_Aires), hyphenated names (America/Port-au-Prince),
// and multi-word names (Eastern Time).
func (p *Parser) tryParseTimezone() bool {
	if p.position >= len(p.tokens) {
		return false
	}

	// Must start with a string token
	if p.tokens[p.position].Typ != TypeString {
		return false
	}

	startPos := p.position

	// Try single token timezone first (EST, GMT, etc.)
	tzString := p.tokens[p.position].Val
	if loc, found := tryParseTimezone(tzString); found {
		p.loc = loc
		p.tzFound = true
		p.position++
		p.result = p.result.In(p.loc)
		return true
	}

	// Try extending with / and - to build timezone paths like
	// America/New_York, America/Argentina/Buenos_Aires, America/Port-au-Prince
	var bestLoc *time.Location
	bestPos := -1

	pos := p.position + 1
	for pos+1 < len(p.tokens) {
		sep := p.tokens[pos]
		if sep.Typ != TypeOperator || (sep.Val != "/" && sep.Val != "-") {
			break
		}
		next := p.tokens[pos+1]
		if next.Typ != TypeString {
			break
		}
		tzString = tzString + sep.Val + next.Val
		pos += 2

		if loc, found := tryParseTimezone(tzString); found {
			bestLoc = loc
			bestPos = pos
		}
	}

	if bestLoc != nil {
		p.loc = bestLoc
		p.tzFound = true
		p.position = bestPos
		p.result = p.result.In(p.loc)
		return true
	}

	// Try parsing multi-word timezone names (like "Eastern Time")
	if p.position+2 < len(p.tokens) &&
		p.tokens[p.position+1].Typ == TypeWhitespace &&
		p.tokens[p.position+2].Typ == TypeString {

		tzString = p.tokens[p.position].Val + " " + p.tokens[p.position+2].Val

		if loc, found := tryParseTimezone(tzString); found {
			p.loc = loc
			p.tzFound = true
			p.position += 3
			p.result = p.result.In(p.loc)
			return true
		}
	}

	p.position = startPos
	return false
}

// tryParseStandardDate attempts to parse standard date formats like ISO dates
func (p *Parser) tryParseStandardDate() (time.Time, bool, error) {
	// Check if we have enough tokens for a date format (at least 5 tokens: num op num op num)
	if p.position+4 >= len(p.tokens) {
		return time.Time{}, false, nil
	}

	// First make sure we have potential date format tokens
	if p.tokens[p.position].Typ != TypeNumber ||
		p.tokens[p.position+2].Typ != TypeNumber ||
		p.tokens[p.position+4].Typ != TypeNumber {
		return time.Time{}, false, nil
	}

	// Get the first three numbers (potential year, month, day in some order)
	firstNum, err1 := strconv.Atoi(p.tokens[p.position].Val)
	if err1 != nil {
		return time.Time{}, false, fmt.Errorf("invalid number in date: %w", err1)
	}

	secondNum, err2 := strconv.Atoi(p.tokens[p.position+2].Val)
	if err2 != nil {
		return time.Time{}, false, fmt.Errorf("invalid number in date: %w", err2)
	}

	thirdNum, err3 := strconv.Atoi(p.tokens[p.position+4].Val)
	if err3 != nil {
		return time.Time{}, false, fmt.Errorf("invalid number in date: %w", err3)
	}

	// Check the separators
	if p.tokens[p.position+1].Typ != TypeOperator || p.tokens[p.position+3].Typ != TypeOperator {
		return time.Time{}, false, nil
	}

	separator := p.tokens[p.position+1].Val
	if separator != p.tokens[p.position+3].Val {
		return time.Time{}, false, nil
	}

	// Determine the format based on the separators and numbers
	var year, month, day int

	switch separator {
	case "-":
		// ISO format: YYYY-MM-DD or D-M-YYYY
		if len(p.tokens[p.position].Val) >= 4 {
			year, month, day = firstNum, secondNum, thirdNum
		} else if len(p.tokens[p.position+4].Val) >= 4 {
			// D-M-YYYY (European style with dashes)
			day, month, year = firstNum, secondNum, thirdNum
		} else {
			// Short year, try as Y-M-D
			year, month, day = firstNum, secondNum, thirdNum
			if year < 100 {
				year = parseTwoDigitYear(year)
			}
		}
	case "/":
		// Could be YYYY/MM/DD or MM/DD/YYYY
		if len(p.tokens[p.position].Val) >= 4 {
			year, month, day = firstNum, secondNum, thirdNum
		} else if len(p.tokens[p.position+4].Val) >= 4 {
			month, day, year = firstNum, secondNum, thirdNum
		} else {
			return time.Time{}, false, nil
		}
	case ".":
		// European format: DD.MM.YY or DD.MM.YYYY
		day, month, year = firstNum, secondNum, thirdNum
		// Handle 2-digit years
		if year < 100 {
			year = parseTwoDigitYear(year)
		}
	default:
		return time.Time{}, false, nil
	}

	// Validate date components using our utility function
	if !IsValidDate(year, month, day) {
		return time.Time{}, false, NewInvalidDateError(year, month, day)
	}

	// Advance position past the parsed date
	p.position += 5

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, p.loc), true, nil
}

// tryParseNextLastExpression attempts to parse expressions like "next Monday" or "last year"
func (p *Parser) tryParseNextLastExpression() (time.Time, bool, error) {
	if p.position >= len(p.tokens) {
		return time.Time{}, false, nil
	}

	// Check for "next", "last", or "this"
	token := p.tokens[p.position]
	if token.Typ != TypeString || (token.Val != DirectionNext && token.Val != DirectionLast && token.Val != "this") {
		return time.Time{}, false, nil
	}

	isNext := token.Val == DirectionNext
	isThis := token.Val == "this"
	p.position++
	p.skipWhitespace()

	// Check for the unit token
	if p.position >= len(p.tokens) {
		return time.Time{}, false, fmt.Errorf("%w after %s", ErrExpectedTimeUnit, token.Val)
	}

	unitToken := p.tokens[p.position]
	if unitToken.Typ != TypeString {
		return time.Time{}, false, fmt.Errorf("%w after %s, got %s", ErrExpectedTimeUnit, token.Val, unitToken.Val)
	}

	p.position++

	// Handle special case: "next week" and "last week"
	// PHP treats Monday as the first day of the week.
	if unitToken.Val == UnitWeek {
		dayOfWeek := int(p.result.Weekday())
		// Days since this week's Monday (Go weekday: 0=Sun,1=Mon,...,6=Sat)
		daysSinceMonday := (dayOfWeek + 6) % 7

		if isNext {
			// Next week = next Monday (always 1-7 days ahead)
			return p.result.AddDate(0, 0, 7-daysSinceMonday), true, nil
		} else if isThis {
			// This week = this Monday
			return p.result.AddDate(0, 0, -daysSinceMonday), true, nil
		} else {
			// Last week = previous week's Monday (always 7-13 days back)
			return p.result.AddDate(0, 0, -(daysSinceMonday + 7)), true, nil
		}
	}

	// Check if it's a day of the week
	dayNum := getDayOfWeek(unitToken.Val)
	if dayNum >= 0 {
		// Handle day of week
		currentDay := int(p.result.Weekday())
		if isThis {
			// "this X" = the X of the current week (can be past or future)
			daysUntil := (dayNum - currentDay + 7) % 7
			targetDay := p.result.AddDate(0, 0, daysUntil)
			year, month, day := targetDay.Date()
			return time.Date(year, month, day, 0, 0, 0, 0, p.loc), true, nil
		} else if isNext {
			// "next X" = the upcoming occurrence of that day
			daysUntil := (dayNum - currentDay + 7) % 7
			if daysUntil == 0 {
				daysUntil = 7 // If today is the target day, go to next week
			}
			nextDay := p.result.AddDate(0, 0, daysUntil)
			year, month, day := nextDay.Date()
			return time.Date(year, month, day, 0, 0, 0, 0, p.loc), true, nil
		} else {
			// Calculate days since the last occurrence
			daysSince := (currentDay - dayNum + 7) % 7
			if daysSince == 0 {
				daysSince = 7 // If today is the target day, go to last week
			}
			lastDay := p.result.AddDate(0, 0, -daysSince)
			year, month, day := lastDay.Date()
			return time.Date(year, month, day, 0, 0, 0, 0, p.loc), true, nil
		}
	}

	// Handle other time units
	switch unitToken.Val {
	case UnitMonth:
		if isNext {
			return p.result.AddDate(0, 1, 0), true, nil
		} else {
			return p.result.AddDate(0, -1, 0), true, nil
		}
	case UnitYear:
		if isNext {
			return p.result.AddDate(1, 0, 0), true, nil
		} else {
			return p.result.AddDate(-1, 0, 0), true, nil
		}
	default:
		return time.Time{}, false, fmt.Errorf("%w: %s", ErrInvalidTimeUnit, unitToken.Val)
	}
}

// daysInMonth returns the number of days in a given month and year
func daysInMonth(year int, month time.Month) int {
	// Create a date for the first day of the next month, then subtract one day
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	lastDay := nextMonth.AddDate(0, 0, -1)
	return lastDay.Day()
}

// isCompoundExpression checks if a string is a compound time expression (contains + or - in the middle)
func isCompoundExpression(str string) bool {
	// Normalize spaces around operators
	spaceOperatorRe := strings.NewReplacer(" + ", "+", " - ", "-", "+ ", "+", "- ", "-")
	normalizedStr := spaceOperatorRe.Replace(str)

	// Check if we have + or - in the middle of the string (not at the start)
	return (strings.Contains(normalizedStr, "+") && !strings.HasPrefix(normalizedStr, "+")) ||
		(strings.Contains(normalizedStr, "-") && !strings.HasPrefix(normalizedStr, "-"))
}

// parseDateWithRelativeTime parses a date followed by a relative time adjustment
// Examples: "2023-05-30 -1 month" or "2022-01-01 +1 year"
func parseDateWithRelativeTime(str string, now time.Time, loc *time.Location, opts []Option) (time.Time, bool) {
	// Split on first whitespace to get date part and rest
	datePart, timePart, ok := splitDateAndRest(str)
	if !ok {
		return time.Time{}, false
	}

	// Parse the date part
	dateResult, err := StrToTime(datePart, append(opts, Rel(now))...)
	if err != nil {
		return time.Time{}, false
	}

	// Handle special case for month end dates when subtracting months
	if timePart == "-1 month" {
		year, month, day := dateResult.Date()

		// Check if it's the last day of the month
		if day == daysInMonth(year, month) {
			// Create a date for the first day of the current month
			firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)
			// Subtract one day to get the last day of the previous month
			prevMonth := firstOfMonth.AddDate(0, -1, 0)
			// Get the last day of the previous month
			lastDay := daysInMonth(prevMonth.Year(), prevMonth.Month())

			// Create the final date with the last day of the previous month,
			// preserving hour, minute, second from the original date
			return time.Date(prevMonth.Year(), prevMonth.Month(), lastDay,
				dateResult.Hour(), dateResult.Minute(), dateResult.Second(),
				dateResult.Nanosecond(), loc), true
		}
	}

	// Parse the time part using the date as reference
	finalResult, err := StrToTime(timePart, append(opts, Rel(dateResult))...)
	if err != nil {
		return time.Time{}, false
	}

	return finalResult, true
}

// parseCompoundExpression parses a compound time expression like "next year+4 days"
func parseCompoundExpression(str string, now time.Time, opts []Option) (time.Time, error) {
	// Normalize compound expressions like "next year+4 days" or "next year + 4 days"
	// Replace spaces around + and - with nothing to make parsing easier
	spaceOperatorRe := strings.NewReplacer(" + ", "+", " - ", "-", "+ ", "+", "- ", "-")
	normalizedStr := spaceOperatorRe.Replace(str)

	// Split the string at + and - operators
	var parts []string
	var operators []string

	// Find all + and - operators (not at the beginning)
	currentPart := ""
	for i := 0; i < len(normalizedStr); i++ {
		if (normalizedStr[i] == '+' || normalizedStr[i] == '-') && i > 0 {
			parts = append(parts, currentPart)
			operators = append(operators, string(normalizedStr[i]))
			currentPart = ""
		} else {
			currentPart += string(normalizedStr[i])
		}
	}

	// Add the last part
	if currentPart != "" {
		parts = append(parts, currentPart)
	}

	// Validate that we have at least one part and one operator
	if len(parts) < 2 || len(operators) < 1 {
		return time.Time{}, errors.New("invalid compound expression format")
	}

	// Process the first part
	result, err := StrToTime(parts[0], append(opts, Rel(now))...)
	if err != nil {
		return time.Time{}, err
	}

	// Process each remaining part with its operator
	for i := 0; i < len(operators); i++ {
		// Check if we have a corresponding part for this operator
		if i+1 >= len(parts) {
			return time.Time{}, errors.New("missing operand after operator in compound expression")
		}

		// Apply the operator to the part
		opPart := operators[i] + parts[i+1]
		nextResult, err := StrToTime(opPart, append(opts, Rel(result))...)
		if err != nil {
			return time.Time{}, err
		}
		result = nextResult
	}

	return result, nil
}

// applyTimeUnitOffset applies a time unit offset to the base time
func (p *Parser) applyTimeUnitOffset(amount int, unitStr string) (time.Time, error) {
	unit := normalizeTimeUnit(unitStr)

	switch unit {
	case UnitDay:
		return p.result.AddDate(0, 0, amount), nil
	case UnitWeek:
		return p.result.AddDate(0, 0, amount*7), nil
	case UnitWeekDay:
		return addWeekdays(p.result, amount), nil
	case UnitMonth:
		return p.result.AddDate(0, amount, 0), nil
	case UnitYear:
		return p.result.AddDate(amount, 0, 0), nil
	case UnitHour:
		return p.result.Add(time.Duration(amount) * time.Hour), nil
	case UnitMinute:
		return p.result.Add(time.Duration(amount) * time.Minute), nil
	case UnitSecond:
		return p.result.Add(time.Duration(amount) * time.Second), nil
	default:
		return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidTimeUnit, unitStr)
	}
}

// addWeekdays adds N business days (Mon-Fri) to the given time.
// PHP behavior: if starting on Sat/Sun, snap to Monday first (counts as 1 weekday),
// then continue adding remaining weekdays.
func addWeekdays(t time.Time, n int) time.Time {
	if n == 0 {
		return t
	}

	step := 1
	if n < 0 {
		step = -1
		n = -n
	}

	result := t
	for i := 0; i < n; i++ {
		result = result.AddDate(0, 0, step)
		// Skip weekends
		for result.Weekday() == time.Saturday || result.Weekday() == time.Sunday {
			result = result.AddDate(0, 0, step)
		}
	}
	return result
}

// tryParseRelativeTime attempts to parse expressions like "+1 day" or "-3 weeks"
func (p *Parser) tryParseRelativeTime() (time.Time, bool, error) {
	if p.position >= len(p.tokens) {
		return time.Time{}, false, nil
	}

	// Check for +/- operator (may be multi-character: "+-" means negative, "--" means positive)
	// PHP rejects "++" but allows "--" and "+-"
	token := p.tokens[p.position]
	if token.Typ != TypeOperator {
		return time.Time{}, false, nil
	}
	// Multi-character sign operators must contain at least one '-'
	if len(token.Val) > 1 && !strings.Contains(token.Val, "-") {
		return time.Time{}, false, nil
	}
	// Count minus signs to determine final sign
	sign := 1
	for _, c := range token.Val {
		if c == '-' {
			sign = -sign
		} else if c != '+' {
			return time.Time{}, false, nil
		}
	}
	p.position++

	// Check for the amount
	if p.position >= len(p.tokens) {
		return time.Time{}, false, fmt.Errorf("%w after %s", ErrMissingAmount, token.Val)
	}

	amountToken := p.tokens[p.position]
	if amountToken.Typ != TypeNumber {
		return time.Time{}, false, fmt.Errorf("expected number after %s, got %s", token.Val, amountToken.Val)
	}

	amount, err := strconv.Atoi(amountToken.Val)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("%w: %s", ErrInvalidNumber, amountToken.Val)
	}
	amount *= sign
	p.position++

	// Skip whitespace
	p.skipWhitespace()

	// Check for the unit
	if p.position >= len(p.tokens) {
		return time.Time{}, false, fmt.Errorf("%w after %d", ErrExpectedTimeUnit, amount)
	}

	unitToken := p.tokens[p.position]
	if unitToken.Typ != TypeString {
		return time.Time{}, false, fmt.Errorf("%w after %d, got %s", ErrExpectedTimeUnit, amount, unitToken.Val)
	}

	p.position++

	// Process the unit by calling the common helper function
	result, err := p.applyTimeUnitOffset(amount, unitToken.Val)
	if err != nil {
		return time.Time{}, false, err
	}

	return result, true, nil
}

// tryParseImplicitRelativeTime attempts to parse expressions like "4 days" or "10 minutes" (without explicit + operator)
func (p *Parser) tryParseImplicitRelativeTime() (time.Time, bool, error) {
	if p.position >= len(p.tokens) {
		return time.Time{}, false, nil
	}

	// Save position for rollback on failure
	startPos := p.position

	// Check for a number (the amount)
	token := p.tokens[p.position]
	if token.Typ != TypeNumber {
		return time.Time{}, false, nil
	}

	amount, err := strconv.Atoi(token.Val)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("%w: %s", ErrInvalidNumber, token.Val)
	}

	// Always treat as positive (implicit +)
	p.position++

	// Skip whitespace
	p.skipWhitespace()

	// Check for the unit
	if p.position >= len(p.tokens) {
		p.position = startPos
		return time.Time{}, false, nil
	}

	unitToken := p.tokens[p.position]
	if unitToken.Typ != TypeString {
		p.position = startPos
		return time.Time{}, false, nil
	}

	p.position++

	// Process the unit by calling the common helper function
	result, err := p.applyTimeUnitOffset(amount, unitToken.Val)
	if err != nil {
		p.position = startPos
		return time.Time{}, false, nil
	}

	return result, true, nil
}

// tryParseMonthOnlyFormat attempts to parse just a month name like "January" or "Feb"
func (p *Parser) tryParseMonthOnlyFormat() (time.Time, bool, error) {
	if p.position >= len(p.tokens) {
		return time.Time{}, false, nil
	}

	// Check if this is actually a month name followed by a day
	// If it is, we should let tryParseMonthNameFormat handle it
	if p.position+1 < len(p.tokens) {
		nextToken := p.tokens[p.position+1]
		// If next token is a number, this might be "Month Day" format
		if nextToken.Typ == TypeNumber {
			return time.Time{}, false, nil
		}
		// If next token is "." (period after abbreviation like "Dec."), defer to month name format
		if nextToken.Typ == TypeOperator && nextToken.Val == "." {
			return time.Time{}, false, nil
		}
		// Or if it's whitespace followed by a number, it might be "Month Day" with space
		if nextToken.Typ == TypeWhitespace && p.position+2 < len(p.tokens) &&
			p.tokens[p.position+2].Typ == TypeNumber {
			return time.Time{}, false, nil
		}
	}

	// Check for a month name
	monthToken := p.tokens[p.position]
	if monthToken.Typ != TypeString {
		return time.Time{}, false, nil
	}

	month, ok := getMonthByName(monthToken.Val)
	if !ok {
		return time.Time{}, false, nil
	}

	// Consume the month token
	p.position++

	// Use the current year and preserve the day from the base time (PHP behavior)
	year := p.result.Year()
	day := p.result.Day()

	// Clamp day to the max days in the target month
	maxDays := daysInMonth(year, month)
	if day > maxDays {
		day = maxDays
	}

	return time.Date(year, month, day, 0, 0, 0, 0, p.loc), true, nil
}

// tryParseMonthNameFormat attempts to parse expressions like "January 15 2023", "Jan 15, 2023", "April 4th", or "June 1 1985 16:30:00 Europe/Paris"
func (p *Parser) tryParseMonthNameFormat() (time.Time, bool, error) {
	if p.position >= len(p.tokens) {
		return time.Time{}, false, nil
	}

	// Check for a month name
	monthToken := p.tokens[p.position]
	if monthToken.Typ != TypeString {
		return time.Time{}, false, nil
	}

	month, ok := getMonthByName(monthToken.Val)
	if !ok {
		return time.Time{}, false, nil
	}
	p.position++

	// Skip optional period after month abbreviation (e.g., "Dec.")
	if p.position < len(p.tokens) && p.tokens[p.position].Typ == TypeOperator && p.tokens[p.position].Val == "." {
		p.position++
	}
	p.skipWhitespace()

	// Check for day number
	if p.position >= len(p.tokens) {
		return time.Time{}, false, fmt.Errorf("expected day after month name")
	}

	dayToken := p.tokens[p.position]
	if dayToken.Typ != TypeNumber {
		return time.Time{}, false, fmt.Errorf("expected day number after month name, got %s", dayToken.Val)
	}

	day, err := strconv.Atoi(dayToken.Val)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("invalid day number: %s", dayToken.Val)
	}
	p.position++

	// Check for ordinal suffix (like "th", "st", "nd", "rd")
	if p.position < len(p.tokens) && p.tokens[p.position].Typ == TypeString {
		suffix := strings.ToLower(p.tokens[p.position].Val)
		if suffix == "st" || suffix == "nd" || suffix == "rd" || suffix == "th" {
			// Skip the ordinal suffix
			p.position++
		}
	}

	// Skip optional punctuation (comma) and whitespace
	if p.position < len(p.tokens) && p.tokens[p.position].Typ == TypePunctuation {
		p.position++
	}
	p.skipWhitespace()

	// Check for year (optional - if not present, use current year)
	year := p.result.Year() // Default to current year

	if p.position < len(p.tokens) {
		yearToken := p.tokens[p.position]
		if yearToken.Typ == TypeNumber {
			// Check if this number is actually a time (followed by ':')
			// If so, don't consume it as a year
			isTime := p.position+1 < len(p.tokens) &&
				p.tokens[p.position+1].Typ == TypeOperator &&
				p.tokens[p.position+1].Val == ":"
			if !isTime {
				yearVal, err := strconv.Atoi(yearToken.Val)
				if err != nil {
					return time.Time{}, false, fmt.Errorf("invalid year: %s", yearToken.Val)
				}
				year = yearVal
				p.position++
			}
		}
	}

	// Validate date components before returning
	if !IsValidDate(year, int(month), day) {
		return time.Time{}, false, fmt.Errorf("invalid date: %s %d, %d", month, day, year)
	}

	// Default time components
	hour, minute, second := 0, 0, 0

	// Check for time (optional)
	// Format: HH:MM:SS
	p.skipWhitespace()
	if p.position+2 < len(p.tokens) &&
		p.tokens[p.position].Typ == TypeNumber && // HH
		p.tokens[p.position+1].Typ == TypeOperator && p.tokens[p.position+1].Val == ":" &&
		p.tokens[p.position+2].Typ == TypeNumber { // MM

		// Parse hour
		hourVal, err := strconv.Atoi(p.tokens[p.position].Val)
		if err == nil && hourVal >= 0 && hourVal <= 23 {
			hour = hourVal
			p.position += 2 // Skip HH:

			// Parse minute
			minuteVal, err := strconv.Atoi(p.tokens[p.position].Val)
			if err == nil && minuteVal >= 0 && minuteVal <= 59 {
				minute = minuteVal
				p.position++

				// Check for seconds (optional)
				if p.position+1 < len(p.tokens) &&
					p.tokens[p.position].Typ == TypeOperator && p.tokens[p.position].Val == ":" &&
					p.tokens[p.position+1].Typ == TypeNumber {

					p.position++ // Skip :
					secondVal, err := strconv.Atoi(p.tokens[p.position].Val)
					if err == nil && secondVal >= 0 && secondVal <= 59 {
						second = secondVal
						p.position++
					}
				}

				// Check for AM/PM (attached or space-separated)
				p.skipWhitespace()
				if p.position < len(p.tokens) && p.tokens[p.position].Typ == TypeString {
					ampm := strings.ToLower(p.tokens[p.position].Val)
					if ampm == "am" || ampm == "pm" {
						hour = applyAMPM(hour, ampm)
						p.position++
					}
				}
			}
		}
	}

	// Check for timezone (optional)
	p.skipWhitespace()
	tzStartPos := p.position
	if ok := p.tryParseTimezone(); ok {
		// Timezone was successfully parsed and p.loc has been updated
	} else {
		// No timezone found, restore position
		p.position = tzStartPos
	}

	return time.Date(year, month, day, hour, minute, second, 0, p.loc), true, nil
}

// getMonthByName converts a month name to its number
func getMonthByName(name string) (time.Month, bool) {
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

	lower := strings.ToLower(name)
	month, ok := monthNames[lower]
	if ok {
		return month, true
	}
	// Try without trailing period (e.g. "dec." → "dec")
	month, ok = monthNames[strings.TrimSuffix(lower, ".")]
	return month, ok
}

// getDayOfWeek converts day name to day number (0 = Sunday, 6 = Saturday)
func getDayOfWeek(day string) int {
	switch strings.ToLower(day) {
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

// normalizeTimeUnit converts various time unit notations to a canonical form
func normalizeTimeUnit(unit string) string {
	// Map of common time unit variations to their canonical forms
	unitMap := map[string]string{
		// Day variations
		"d":     UnitDay,
		"day":   UnitDay,
		"days":  UnitDay,
		"days.": UnitDay,

		// Week variations
		"w":     UnitWeek,
		"wk":    UnitWeek,
		"wks":   UnitWeek,
		"wks.":  UnitWeek,
		"week":  UnitWeek,
		"weeks": UnitWeek,

		// Weekday (business day) variations
		"weekday":  UnitWeekDay,
		"weekdays": UnitWeekDay,

		// Month variations
		"m":      UnitMonth,
		"mon":    UnitMonth,
		"mons":   UnitMonth,
		"mons.":  UnitMonth,
		"month":  UnitMonth,
		"months": UnitMonth,

		// Year variations
		"y":     UnitYear,
		"yr":    UnitYear,
		"yrs":   UnitYear,
		"yrs.":  UnitYear,
		"year":  UnitYear,
		"years": UnitYear,

		// Hour variations
		"h":      UnitHour,
		"hr":     UnitHour,
		"hrs":    UnitHour,
		"hrs.":   UnitHour,
		"hour":   UnitHour,
		"hours":  UnitHour,
		"hourss": UnitHour,

		// Minute variations
		"min":     UnitMinute,
		"mins":    UnitMinute,
		"mins.":   UnitMinute,
		"minute":  UnitMinute,
		"minutes": UnitMinute,

		// Second variations
		"sec":     UnitSecond,
		"secs":    UnitSecond,
		"secs.":   UnitSecond,
		"second":  UnitSecond,
		"seconds": UnitSecond,
	}

	// Try exact match first
	if canonical, found := unitMap[strings.ToLower(unit)]; found {
		return canonical
	}

	// Remove trailing 's' if present for plurals not in the map
	trimmed := strings.TrimSuffix(strings.ToLower(unit), "s")
	if canonical, found := unitMap[trimmed]; found {
		return canonical
	}

	// Handle prefixes for longer variations
	lowerUnit := strings.ToLower(unit)
	if strings.HasPrefix(lowerUnit, "day") {
		return UnitDay
	} else if strings.HasPrefix(lowerUnit, "weekday") {
		return UnitWeekDay
	} else if strings.HasPrefix(lowerUnit, "week") {
		return UnitWeek
	} else if strings.HasPrefix(lowerUnit, "month") {
		return UnitMonth
	} else if strings.HasPrefix(lowerUnit, "year") {
		return UnitYear
	} else if strings.HasPrefix(lowerUnit, "hour") || strings.HasPrefix(lowerUnit, "hr") {
		return UnitHour
	} else if strings.HasPrefix(lowerUnit, "min") {
		return UnitMinute
	} else if strings.HasPrefix(lowerUnit, "sec") {
		return UnitSecond
	}

	// If we couldn't normalize, return the original unit
	return unit
}

// tryParseBareWeekday handles:
// - Bare weekday name: "tuesday" (next occurrence)
// - Weekday + "next/last week" with optional time: "monday next week 13:00"
// - Weekday + month [year]: "thursday nov 2007" (first occurrence in that month)
func (p *Parser) tryParseBareWeekday() (time.Time, bool, error) {
	if p.position >= len(p.tokens) {
		return time.Time{}, false, nil
	}

	token := p.tokens[p.position]
	if token.Typ != TypeString {
		return time.Time{}, false, nil
	}

	dayNum := getDayOfWeek(token.Val)
	if dayNum < 0 {
		return time.Time{}, false, nil
	}

	p.position++
	p.skipWhitespace()

	// Check for "next/last week" after weekday
	if p.position < len(p.tokens) && p.tokens[p.position].Typ == TypeString {
		direction := p.tokens[p.position].Val
		if direction == DirectionNext || direction == DirectionLast {
			savedPos := p.position
			p.position++
			p.skipWhitespace()

			if p.position < len(p.tokens) && p.tokens[p.position].Typ == TypeString &&
				p.tokens[p.position].Val == UnitWeek {
				p.position++

				// Calculate: find the Monday of next/last week, then offset to target weekday
				currentDay := int(p.result.Weekday())
				daysSinceMonday := (currentDay + 6) % 7

				var monday time.Time
				if direction == DirectionNext {
					monday = p.result.AddDate(0, 0, 7-daysSinceMonday)
				} else {
					monday = p.result.AddDate(0, 0, -(daysSinceMonday + 7))
				}

				// Offset from Monday to target weekday (Mon=0, Tue=1, ..., Sun=6)
				targetOffset := (dayNum + 6) % 7
				result := monday.AddDate(0, 0, targetOffset)

				year, month, day := result.Date()
				hour, minute, second := 0, 0, 0

				// Check for optional time
				p.skipWhitespace()
				if p.position+2 < len(p.tokens) &&
					p.tokens[p.position].Typ == TypeNumber &&
					p.tokens[p.position+1].Typ == TypeOperator && p.tokens[p.position+1].Val == ":" &&
					p.tokens[p.position+2].Typ == TypeNumber {

					h, err := strconv.Atoi(p.tokens[p.position].Val)
					if err == nil && h >= 0 && h <= 23 {
						hour = h
						p.position += 2 // Skip HH:
						m, err := strconv.Atoi(p.tokens[p.position].Val)
						if err == nil && m >= 0 && m <= 59 {
							minute = m
							p.position++
						}
					}
				}

				return time.Date(year, month, day, hour, minute, second, 0, p.loc), true, nil
			}

			// Not "next/last week", restore position
			p.position = savedPos
		}

		// Check for weekday + month [year]: "thursday nov 2007"
		if m, ok := getMonthByName(p.tokens[p.position].Val); ok {
			p.position++
			p.skipWhitespace()

			year := p.result.Year()
			if p.position < len(p.tokens) && p.tokens[p.position].Typ == TypeNumber {
				if y, err := strconv.Atoi(p.tokens[p.position].Val); err == nil && y > 0 {
					year = y
					p.position++
				}
			}

			// Find 1st occurrence of the weekday in that month
			firstOfMonth := time.Date(year, m, 1, 0, 0, 0, 0, p.loc)
			firstDayOfWeek := int(firstOfMonth.Weekday())
			daysUntilFirst := (dayNum - firstDayOfWeek + 7) % 7
			resultDay := 1 + daysUntilFirst

			return time.Date(year, m, resultDay, 0, 0, 0, 0, p.loc), true, nil
		}
	}

	// Bare weekday name - find next occurrence (PHP behavior)
	currentDay := int(p.result.Weekday())
	daysUntil := (dayNum - currentDay + 7) % 7
	if daysUntil == 0 {
		daysUntil = 7
	}
	nextDay := p.result.AddDate(0, 0, daysUntil)
	year, month, day := nextDay.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, p.loc), true, nil
}

// tryParseFirstLastDayOfExpression handles "first/last day of this/next/last month/year"
func (p *Parser) tryParseFirstLastDayOfExpression() (time.Time, bool, error) {
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeString {
		return time.Time{}, false, nil
	}

	startPos := p.position
	token := p.tokens[p.position]

	var isFirst bool
	switch token.Val {
	case "first":
		isFirst = true
	case "last":
		isFirst = false
	default:
		return time.Time{}, false, nil
	}
	p.position++
	p.skipWhitespace()

	// Must be followed by "day"
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeString || p.tokens[p.position].Val != "day" {
		p.position = startPos
		return time.Time{}, false, nil
	}
	p.position++
	p.skipWhitespace()

	// Must be followed by "of"
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeString || p.tokens[p.position].Val != "of" {
		p.position = startPos
		return time.Time{}, false, nil
	}
	p.position++
	p.skipWhitespace()

	// Must be followed by "this"/"next"/"last"
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeString {
		p.position = startPos
		return time.Time{}, false, nil
	}

	direction := p.tokens[p.position].Val
	if direction != "this" && direction != DirectionNext && direction != DirectionLast {
		p.position = startPos
		return time.Time{}, false, nil
	}
	p.position++
	p.skipWhitespace()

	// Must be followed by "month" or "year"
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeString {
		p.position = startPos
		return time.Time{}, false, nil
	}

	unit := normalizeTimeUnit(p.tokens[p.position].Val)
	p.position++

	year := p.result.Year()
	month := p.result.Month()

	switch unit {
	case UnitMonth:
		if direction == DirectionNext {
			ref := p.result.AddDate(0, 1, 0)
			year, month, _ = ref.Date()
		} else if direction == DirectionLast {
			ref := p.result.AddDate(0, -1, 0)
			year, month, _ = ref.Date()
		}
	case UnitYear:
		if direction == DirectionNext {
			year = p.result.Year() + 1
			month = time.January
		} else if direction == DirectionLast {
			year = p.result.Year() - 1
			month = time.December
		}
	default:
		p.position = startPos
		return time.Time{}, false, nil
	}

	var day int
	if isFirst {
		day = 1
	} else {
		day = daysInMonth(year, month)
	}

	return time.Date(year, month, day, p.result.Hour(), p.result.Minute(), p.result.Second(), 0, p.loc), true, nil
}

// tryParseTimeKeyword handles "midnight" and "noon" keywords in token stream
func (p *Parser) tryParseTimeKeyword() (time.Time, bool, error) {
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeString {
		return time.Time{}, false, nil
	}

	year, month, day := p.result.Date()
	switch p.tokens[p.position].Val {
	case "midnight":
		p.position++
		return time.Date(year, month, day, 0, 0, 0, 0, p.loc), true, nil
	case "noon":
		p.position++
		return time.Date(year, month, day, 12, 0, 0, 0, p.loc), true, nil
	default:
		return time.Time{}, false, nil
	}
}

// tryParseWeekdayAgo handles "N weekday ago" or "N weekdays ago"
func (p *Parser) tryParseWeekdayAgo() (time.Time, bool, error) {
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeNumber {
		return time.Time{}, false, nil
	}

	startPos := p.position
	amount, err := strconv.Atoi(p.tokens[p.position].Val)
	if err != nil || amount <= 0 {
		return time.Time{}, false, nil
	}
	p.position++
	p.skipWhitespace()

	// Check for weekday name (singular or plural)
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeString {
		p.position = startPos
		return time.Time{}, false, nil
	}

	dayName := p.tokens[p.position].Val
	singularName := strings.TrimSuffix(dayName, "s")
	dayNum := getDayOfWeek(singularName)
	if dayNum < 0 {
		dayNum = getDayOfWeek(dayName)
	}
	if dayNum < 0 {
		p.position = startPos
		return time.Time{}, false, nil
	}
	p.position++
	p.skipWhitespace()

	// Check for "ago"
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeString || p.tokens[p.position].Val != "ago" {
		p.position = startPos
		return time.Time{}, false, nil
	}
	p.position++

	// Go back N occurrences of that weekday
	currentDay := int(p.result.Weekday())
	daysSince := (currentDay - dayNum + 7) % 7
	if daysSince == 0 {
		daysSince = 7
	}
	totalDays := daysSince + (amount-1)*7
	result := p.result.AddDate(0, 0, -totalDays)

	// Preserve time from p.result
	year, month, day := result.Date()
	return time.Date(year, month, day, p.result.Hour(), p.result.Minute(), p.result.Second(), p.result.Nanosecond(), p.loc), true, nil
}

// tryParseTimeExpression handles standalone time like "HH:MM" or "HH:MM:SS"
func (p *Parser) tryParseTimeExpression() (time.Time, bool, error) {
	if p.position+2 >= len(p.tokens) {
		return time.Time{}, false, nil
	}

	if p.tokens[p.position].Typ != TypeNumber ||
		p.tokens[p.position+1].Typ != TypeOperator || p.tokens[p.position+1].Val != ":" ||
		p.tokens[p.position+2].Typ != TypeNumber {
		return time.Time{}, false, nil
	}

	hour, err := strconv.Atoi(p.tokens[p.position].Val)
	if err != nil || hour < 0 || hour > 23 {
		return time.Time{}, false, nil
	}
	p.position += 2 // Skip HH:

	minute, err := strconv.Atoi(p.tokens[p.position].Val)
	if err != nil || minute < 0 || minute > 59 {
		return time.Time{}, false, nil
	}
	p.position++

	second := 0
	if p.position+1 < len(p.tokens) &&
		p.tokens[p.position].Typ == TypeOperator && p.tokens[p.position].Val == ":" &&
		p.tokens[p.position+1].Typ == TypeNumber {
		p.position++ // Skip :
		s, err := strconv.Atoi(p.tokens[p.position].Val)
		if err == nil && s >= 0 && s <= 59 {
			second = s
			p.position++
		}
	}

	year, month, day := p.result.Date()
	return time.Date(year, month, day, hour, minute, second, 0, p.loc), true, nil
}

// tryParseOrdinalRelativeTime handles ordinal words as implicit relative time
// e.g., "eighth day" = +8 days
func (p *Parser) tryParseOrdinalRelativeTime() (time.Time, bool, error) {
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeString {
		return time.Time{}, false, nil
	}

	startPos := p.position
	amount := ordinalWordToNumber(p.tokens[p.position].Val)
	if amount <= 0 {
		return time.Time{}, false, nil
	}
	p.position++
	p.skipWhitespace()

	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeString {
		p.position = startPos
		return time.Time{}, false, nil
	}

	result, err := p.applyTimeUnitOffset(amount, p.tokens[p.position].Val)
	if err != nil {
		p.position = startPos
		return time.Time{}, false, nil
	}
	p.position++
	return result, true, nil
}

// tryParseYearOnly handles a bare 4-digit year number, setting the year on current result
func (p *Parser) tryParseYearOnly() (time.Time, bool, error) {
	if p.position >= len(p.tokens) || p.tokens[p.position].Typ != TypeNumber {
		return time.Time{}, false, nil
	}
	val := p.tokens[p.position].Val
	if len(val) != 4 {
		return time.Time{}, false, nil
	}
	year, err := strconv.Atoi(val)
	if err != nil || year < 1 {
		return time.Time{}, false, nil
	}

	// Only treat as year if it's the last token (or followed only by whitespace)
	nextPos := p.position + 1
	for nextPos < len(p.tokens) && p.tokens[nextPos].Typ == TypeWhitespace {
		nextPos++
	}
	if nextPos != len(p.tokens) {
		return time.Time{}, false, nil
	}

	p.position++
	return time.Date(year, p.result.Month(), p.result.Day(), p.result.Hour(), p.result.Minute(), p.result.Second(), p.result.Nanosecond(), p.loc), true, nil
}

// ordinalWordToNumber converts ordinal words to numbers
func ordinalWordToNumber(word string) int {
	switch strings.ToLower(word) {
	case "first":
		return 1
	case "second":
		return 2
	case "third":
		return 3
	case "fourth":
		return 4
	case "fifth":
		return 5
	case "sixth":
		return 6
	case "seventh":
		return 7
	case "eighth":
		return 8
	case "ninth":
		return 9
	case "tenth":
		return 10
	case "eleventh":
		return 11
	case "twelfth":
		return 12
	}
	return 0
}
