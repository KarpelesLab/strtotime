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
		case tzOption: // timezone
			if v.loc != nil {
				loc = v.loc
			}
		}
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
	}

	// Try to parse datetime format (YYYY-MM-DD HH:MM:SS)
	dateTimeRe := regexp.MustCompile(`^(\d{4}-\d{1,2}-\d{1,2})\s+(\d{1,2}):(\d{1,2}):(\d{1,2})$`)
	if matches := dateTimeRe.FindStringSubmatch(str); matches != nil {
		// Parse the date part
		datePart := matches[1]
		hour, errH := strconv.Atoi(matches[2])
		minute, errM := strconv.Atoi(matches[3])
		second, errS := strconv.Atoi(matches[4])
		
		// Validate time components
		if errH != nil || hour < 0 || hour > 23 || 
		   errM != nil || minute < 0 || minute > 59 || 
		   errS != nil || second < 0 || second > 59 {
			return time.Time{}, fmt.Errorf("invalid time components in datetime: %s", str)
		}
		
		// Parse the date
		t, ok := parseISOFormat(datePart, loc)
		if !ok {
			return time.Time{}, fmt.Errorf("invalid date format in datetime: %s", str)
		}
		
		// Add the time components
		return time.Date(t.Year(), t.Month(), t.Day(), hour, minute, second, 0, loc), nil
	}
	
	// Try date with timezone format
	if t, ok := parseWithTimezone(str, loc); ok {
		return t, nil
	}

	// Try standard date formats - the string should be directly validated by these functions
	// Certain irregular date formats like "2023-13" will just fall through
	if t, ok := parseISOFormat(str, loc); ok {
		return t, nil
	}

	if t, ok := parseSlashFormat(str, loc); ok {
		return t, nil
	}

	if t, ok := parseUSFormat(str, loc); ok {
		return t, nil
	}
	
	// Try extended date formats
	if t, ok := parseCompactTimestamp(str, loc); ok {
		return t, nil
	}
	
	if t, ok := parseMonthNameFormat(str, loc); ok {
		return t, nil
	}
	
	if t, ok := parseHTTPLogFormat(str, loc); ok {
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

	// Check if this is a date followed by a relative time adjustment
	if result, ok := parseDateWithRelativeTime(str, now, loc, opts); ok {
		return result, nil
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

// peek returns the next token without advancing the position
func (p *Parser) peek() *Token {
	if p.position >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.position]
}

// consume advances the position and returns the current token
func (p *Parser) consume() *Token {
	if p.position >= len(p.tokens) {
		return nil
	}
	token := &p.tokens[p.position]
	p.position++
	return token
}

// tryParseTimezone attempts to parse a timezone from the token stream
// This handles both abbreviations (PST, EST) and full names (America/New_York, Europe/Paris)
func (p *Parser) tryParseTimezone() bool {
	if p.position >= len(p.tokens) {
		return false
	}

	// Save the current position in case we need to backtrack
	startPos := p.position

	// Try parsing a single token timezone (like "EST", "GMT", etc.)
	if p.position < len(p.tokens) && p.tokens[p.position].Typ == TypeString {
		tzString := p.tokens[p.position].Val
		if loc, found := tryParseTimezone(tzString); found {
			p.loc = loc
			p.tzFound = true
			p.position++

			// Update result to be in the new timezone
			p.result = p.result.In(p.loc)
			return true
		}
	}

	// Try parsing a full timezone name with slashes (like "America/New_York")
	// This could be multiple tokens like "Europe" "/" "Paris"
	if p.position+2 < len(p.tokens) &&
		p.tokens[p.position].Typ == TypeString &&
		p.tokens[p.position+1].Typ == TypeOperator &&
		p.tokens[p.position+1].Val == "/" &&
		p.tokens[p.position+2].Typ == TypeString {

		// Construct the timezone string with slash
		tzString := p.tokens[p.position].Val + "/" + p.tokens[p.position+2].Val

		if loc, found := tryParseTimezone(tzString); found {
			p.loc = loc
			p.tzFound = true
			p.position += 3 // Skip all three tokens

			// Update result to be in the new timezone
			p.result = p.result.In(p.loc)
			return true
		}
	}

	// Try parsing multi-word timezone names (like "Eastern Time")
	if p.position+2 < len(p.tokens) &&
		p.tokens[p.position].Typ == TypeString &&
		p.tokens[p.position+1].Typ == TypeWhitespace &&
		p.tokens[p.position+2].Typ == TypeString {

		// Try to combine the tokens to form a full name
		tzString := p.tokens[p.position].Val + " " + p.tokens[p.position+2].Val

		if loc, found := tryParseTimezone(tzString); found {
			p.loc = loc
			p.tzFound = true
			p.position += 3 // Skip all three tokens

			// Update result to be in the new timezone
			p.result = p.result.In(p.loc)
			return true
		}
	}

	// Restore the position if we couldn't parse a timezone
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
		// ISO format: YYYY-MM-DD
		if len(p.tokens[p.position].Val) == 4 { // YYYY format
			year, month, day = firstNum, secondNum, thirdNum
		} else {
			return time.Time{}, false, nil
		}
	case "/":
		// Could be YYYY/MM/DD or MM/DD/YYYY
		if len(p.tokens[p.position].Val) == 4 { // YYYY/MM/DD
			year, month, day = firstNum, secondNum, thirdNum
		} else if len(p.tokens[p.position+4].Val) == 4 { // MM/DD/YYYY
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

	// Check for "next" or "last"
	token := p.tokens[p.position]
	if token.Typ != TypeString || (token.Val != DirectionNext && token.Val != DirectionLast) {
		return time.Time{}, false, nil
	}

	isNext := token.Val == DirectionNext
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
	if unitToken.Val == UnitWeek {
		if isNext {
			// Next week means the Monday of next week
			dayOfWeek := int(p.result.Weekday())
			var daysToAdd int
			switch dayOfWeek {
			case 0: // Sunday
				daysToAdd = 1 // Next Monday is 1 day away
			case 1: // Monday
				daysToAdd = 0 // This is already Monday
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
			return p.result.AddDate(0, 0, daysToAdd), true, nil
		} else {
			// Last week means the Monday of the previous week
			dayOfWeek := int(p.result.Weekday())
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
			return p.result.AddDate(0, 0, -daysToSubtract), true, nil
		}
	}

	// Check if it's a day of the week
	dayNum := getDayOfWeek(unitToken.Val)
	if dayNum >= 0 {
		// Handle day of week
		currentDay := int(p.result.Weekday())
		if isNext {
			// Calculate days until the next occurrence
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
	dateTimeRe := regexp.MustCompile(`^(\d{4}-\d{1,2}-\d{1,2}|\d{4}/\d{1,2}/\d{1,2}|\d{1,2}/\d{1,2}/\d{4}|\d{1,2}\.\d{1,2}\.\d{2,4})\s+(.+)$`)
	if !dateTimeRe.MatchString(str) {
		return time.Time{}, false
	}
	
	matches := dateTimeRe.FindStringSubmatch(str)
	if len(matches) != 3 {
		return time.Time{}, false
	}
	
	datePart := matches[1]
	timePart := matches[2]
	
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

// handleMonthEndDates handles the special case for month-end dates when adding/subtracting months
func handleMonthEndDates(t time.Time, amount int, loc *time.Location) (time.Time, bool) {
	year, month, day := t.Date()
	
	// Check if the current day is the last day of the month
	if day == daysInMonth(year, month) {
		// Add the months first
		newDate := t.AddDate(0, amount, 0)
		newYear, newMonth, _ := newDate.Date()
		
		// Then set the day to the last day of the target month
		lastDay := daysInMonth(newYear, newMonth)
		
		return time.Date(
			newYear, 
			newMonth, 
			lastDay, 
			t.Hour(),
			t.Minute(), 
			t.Second(), 
			t.Nanosecond(), 
			loc,
		), true
	}
	
	return t, false
}

// applyTimeUnitOffset applies a time unit offset to the base time
func (p *Parser) applyTimeUnitOffset(amount int, unitStr string) (time.Time, error) {
	unit := normalizeTimeUnit(unitStr)

	switch unit {
	case UnitDay:
		return p.result.AddDate(0, 0, amount), nil
	case UnitWeek:
		return p.result.AddDate(0, 0, amount*7), nil
	case UnitMonth:
		// Special handling for end-of-month dates
		if adjustedTime, handled := handleMonthEndDates(p.result, amount, p.loc); handled {
			return adjustedTime, nil
		}
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

// tryParseRelativeTime attempts to parse expressions like "+1 day" or "-3 weeks"
func (p *Parser) tryParseRelativeTime() (time.Time, bool, error) {
	if p.position >= len(p.tokens) {
		return time.Time{}, false, nil
	}

	// Check for +/- operator
	token := p.tokens[p.position]
	if token.Typ != TypeOperator || (token.Val != "+" && token.Val != "-") {
		return time.Time{}, false, nil
	}

	sign := 1
	if token.Val == "-" {
		sign = -1
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

	// Use the current year and day 1 of the given month
	year := p.result.Year()
	day := 1

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
			yearVal, err := strconv.Atoi(yearToken.Val)
			if err != nil {
				return time.Time{}, false, fmt.Errorf("invalid year: %s", yearToken.Val)
			}
			year = yearVal
			p.position++
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
	if p.position+4 < len(p.tokens) &&
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
				if p.position+2 < len(p.tokens) &&
					p.tokens[p.position].Typ == TypeOperator && p.tokens[p.position].Val == ":" &&
					p.tokens[p.position+1].Typ == TypeNumber {

					p.position++ // Skip :
					secondVal, err := strconv.Atoi(p.tokens[p.position].Val)
					if err == nil && secondVal >= 0 && secondVal <= 59 {
						second = secondVal
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

// tryMatch checks if the tokens at the current position match a pattern
func (p *Parser) tryMatch(matcher func([]Token, int) bool) bool {
	if p.position >= len(p.tokens) {
		return false
	}
	return matcher(p.tokens, p.position)
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

	month, ok := monthNames[strings.ToLower(name)]
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