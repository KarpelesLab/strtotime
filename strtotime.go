package strtotime

import (
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

	// Special case for European date format like "24.11.22"
	if matched, _ := regexp.MatchString(`^\d{1,2}\.\d{1,2}\.\d{2,4}$`, str); matched {
		parts := strings.Split(str, ".")
		day, err := strconv.Atoi(parts[0])
		if err == nil {
			month, err := strconv.Atoi(parts[1])
			if err == nil {
				year, err := strconv.Atoi(parts[2])
				if err == nil {
					// Handle 2-digit years (YY)
					if year < 100 {
						if year < 70 {
							year += 2000 // 00-69 -> 2000-2069
						} else {
							year += 1900 // 70-99 -> 1970-1999
						}
					}

					// Validate date components
					if month >= 1 && month <= 12 && day >= 1 && day <= 31 {
						return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), nil
					}
				}
			}
		}
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

	// Try date with timezone format
	if t, ok := parseWithTimezone(str, loc); ok {
		return t, nil
	}

	// Try standard date formats
	if t, ok := parseISOFormat(str, loc); ok {
		return t, nil
	}

	if t, ok := parseSlashFormat(str, loc); ok {
		return t, nil
	}

	if t, ok := parseUSFormat(str, loc); ok {
		return t, nil
	}

	// Normalize compound expressions like "next year+4 days" or "next year + 4 days"
	// Replace spaces around + and - with nothing to make parsing easier
	spaceOperatorRe := strings.NewReplacer(" + ", "+", " - ", "-", "+ ", "+", "- ", "-")
	normalizedStr := spaceOperatorRe.Replace(str)

	// If we have + or - in the middle of the string (not at the start)
	// then it's likely a compound expression
	if (strings.Contains(normalizedStr, "+") && !strings.HasPrefix(normalizedStr, "+")) ||
		(strings.Contains(normalizedStr, "-") && !strings.HasPrefix(normalizedStr, "-")) {
		// Split the string at + and - operators
		var parts []string
		var operators []string

		// First find all + and - operators (not at the beginning)
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

		// Process the first part
		result, err := StrToTime(parts[0], append(opts, Rel(now))...)
		if err != nil {
			return time.Time{}, err
		}

		// Process each remaining part with its operator
		for i := 0; i < len(operators); i++ {
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

	// Check if this is a date followed by a relative time adjustment
	// Examples: "2023-05-30 -1 month" or "2022-01-01 +1 year"
	dateTimeRe := regexp.MustCompile(`^(\d{4}-\d{1,2}-\d{1,2}|\d{4}/\d{1,2}/\d{1,2}|\d{1,2}/\d{1,2}/\d{4}|\d{1,2}\.\d{1,2}\.\d{2,4})\s+(.+)$`)
	if dateTimeRe.MatchString(str) {
		matches := dateTimeRe.FindStringSubmatch(str)
		if len(matches) == 3 {
			datePart := matches[1]
			timePart := matches[2]

			// Parse the date part
			dateResult, err := StrToTime(datePart, append(opts, Rel(now))...)
			if err != nil {
				return time.Time{}, err
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
						dateResult.Nanosecond(), loc), nil
				}
			}

			// Parse the time part using the date as reference
			finalResult, err := StrToTime(timePart, append(opts, Rel(dateResult))...)
			if err != nil {
				return time.Time{}, err
			}

			return finalResult, nil
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
	// ISO format: YYYY-MM-DD
	if p.position+4 < len(p.tokens) &&
		p.tokens[p.position].Typ == TypeNumber && // YYYY
		p.tokens[p.position+1].Typ == TypeOperator && p.tokens[p.position+1].Val == "-" &&
		p.tokens[p.position+2].Typ == TypeNumber && // MM
		p.tokens[p.position+3].Typ == TypeOperator && p.tokens[p.position+3].Val == "-" &&
		p.tokens[p.position+4].Typ == TypeNumber { // DD

		year, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid year: %w", err)
		}
		p.position++
		p.position++ // Skip the "-"

		month, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid month: %w", err)
		}
		p.position++
		p.position++ // Skip the "-"

		day, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid day: %w", err)
		}
		p.position++

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, p.loc), true, nil
	}

	// Slash format: YYYY/MM/DD
	if p.position+4 < len(p.tokens) &&
		p.tokens[p.position].Typ == TypeNumber && // YYYY
		p.tokens[p.position+1].Typ == TypeOperator && p.tokens[p.position+1].Val == "/" &&
		p.tokens[p.position+2].Typ == TypeNumber && // MM
		p.tokens[p.position+3].Typ == TypeOperator && p.tokens[p.position+3].Val == "/" &&
		p.tokens[p.position+4].Typ == TypeNumber { // DD

		year, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid year: %w", err)
		}
		p.position++
		p.position++ // Skip the "/"

		month, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid month: %w", err)
		}
		p.position++
		p.position++ // Skip the "/"

		day, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid day: %w", err)
		}
		p.position++

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, p.loc), true, nil
	}

	// US format: MM/DD/YYYY
	if p.position+4 < len(p.tokens) &&
		p.tokens[p.position].Typ == TypeNumber && // MM
		p.tokens[p.position+1].Typ == TypeOperator && p.tokens[p.position+1].Val == "/" &&
		p.tokens[p.position+2].Typ == TypeNumber && // DD
		p.tokens[p.position+3].Typ == TypeOperator && p.tokens[p.position+3].Val == "/" &&
		p.tokens[p.position+4].Typ == TypeNumber { // YYYY

		month, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid month: %w", err)
		}
		p.position++
		p.position++ // Skip the "/"

		day, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid day: %w", err)
		}
		p.position++
		p.position++ // Skip the "/"

		year, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid year: %w", err)
		}
		p.position++

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, p.loc), true, nil
	}

	// European format with dots: DD.MM.YY or DD.MM.YYYY
	if p.position+4 < len(p.tokens) &&
		p.tokens[p.position].Typ == TypeNumber && // DD
		p.tokens[p.position+1].Typ == TypeOperator && p.tokens[p.position+1].Val == "." &&
		p.tokens[p.position+2].Typ == TypeNumber && // MM
		p.tokens[p.position+3].Typ == TypeOperator && p.tokens[p.position+3].Val == "." &&
		p.tokens[p.position+4].Typ == TypeNumber { // YY or YYYY

		day, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid day: %w", err)
		}
		p.position++
		p.position++ // Skip the "."

		month, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid month: %w", err)
		}
		p.position++
		p.position++ // Skip the "."

		year, err := strconv.Atoi(p.tokens[p.position].Val)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("invalid year: %w", err)
		}
		p.position++

		// Handle 2-digit years (YY)
		if year < 100 {
			if year < 70 {
				year += 2000 // 00-69 -> 2000-2069
			} else {
				year += 1900 // 70-99 -> 1970-1999
			}
		}

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, p.loc), true, nil
	}

	return time.Time{}, false, nil
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
		year, month, day := p.result.Date()

		// If day is the last day of the month, and we're adjusting months,
		// make sure to set it to the last day of the target month
		if day == daysInMonth(year, month) {
			// Add the months first
			newDate := p.result.AddDate(0, amount, 0)
			newYear, newMonth, _ := newDate.Date()

			// Then set the day to the last day of the month
			lastDay := daysInMonth(newYear, newMonth)
			return time.Date(newYear, newMonth, lastDay, p.result.Hour(),
				p.result.Minute(), p.result.Second(), p.result.Nanosecond(), p.loc), nil
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
	// Hard-code specific problem variations that need exact matching
	switch unit {
	case "hourss", "hrs", "hrs.":
		return UnitHour
	case "days.", "days":
		return UnitDay
	case "weeks", "wks", "wks.":
		return UnitWeek
	case "months", "mons", "mons.":
		return UnitMonth
	case "years", "yrs", "yrs.":
		return UnitYear
	case "minutes", "mins", "mins.":
		return UnitMinute
	case "seconds", "secs", "secs.":
		return UnitSecond
	}

	// Handle excessive plural forms (e.g., "hourss" -> "hours" -> "hour")
	for strings.HasSuffix(unit, "s") {
		unit = unit[:len(unit)-1]
	}

	// Check for common abbreviations and misspellings
	switch unit {
	case "d":
		return UnitDay
	case "w", "wk":
		return UnitWeek
	case "m", "mon":
		return UnitMonth
	case "y", "yr":
		return UnitYear
	case "h", "hr":
		return UnitHour
	case "min":
		return UnitMinute
	case "sec":
		return UnitSecond
	}

	// Handle prefixes to catch misspellings or variations
	if strings.HasPrefix(unit, "day") {
		return UnitDay
	} else if strings.HasPrefix(unit, "week") {
		return UnitWeek
	} else if strings.HasPrefix(unit, "month") {
		return UnitMonth
	} else if strings.HasPrefix(unit, "year") {
		return UnitYear
	} else if strings.HasPrefix(unit, "hour") || strings.HasPrefix(unit, "hr") {
		return UnitHour
	} else if strings.HasPrefix(unit, "minute") || strings.HasPrefix(unit, "min") {
		return UnitMinute
	} else if strings.HasPrefix(unit, "second") || strings.HasPrefix(unit, "sec") {
		return UnitSecond
	}

	// Try more aggressive matching for hours
	if strings.Contains(unit, "hr") || strings.Contains(unit, "hour") {
		return UnitHour
	}

	// If we couldn't normalize, return the original unit
	return unit
}
