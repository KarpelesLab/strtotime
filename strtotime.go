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

	// Normalize string - trim and lowercase
	str = strings.ToLower(strings.TrimSpace(str))
	if str == "" {
		return time.Time{}, errors.New("empty time string")
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

	// Direct pattern matching for date formats
	// ISO format: YYYY-MM-DD
	if matched, _ := regexp.MatchString(`^\d{4}-\d{1,2}-\d{1,2}$`, str); matched {
		parts := strings.Split(str, "-")
		year, _ := strconv.Atoi(parts[0])
		month, _ := strconv.Atoi(parts[1])
		day, _ := strconv.Atoi(parts[2])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), nil
	}

	// Slash format: YYYY/MM/DD
	if matched, _ := regexp.MatchString(`^\d{4}/\d{1,2}/\d{1,2}$`, str); matched {
		parts := strings.Split(str, "/")
		year, _ := strconv.Atoi(parts[0])
		month, _ := strconv.Atoi(parts[1])
		day, _ := strconv.Atoi(parts[2])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), nil
	}

	// US format: MM/DD/YYYY
	if matched, _ := regexp.MatchString(`^\d{1,2}/\d{1,2}/\d{4}$`, str); matched {
		parts := strings.Split(str, "/")
		month, _ := strconv.Atoi(parts[0])
		day, _ := strconv.Atoi(parts[1])
		year, _ := strconv.Atoi(parts[2])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc), nil
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

		// Try "next/last" expressions
		if t, ok, err := p.tryParseNextLastExpression(); ok {
			if err != nil {
				return time.Time{}, err
			}
			p.result = t
			parsed = true
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

// tryParseStandardDate attempts to parse standard date formats like ISO dates
func (p *Parser) tryParseStandardDate() (time.Time, bool, error) {
	// ISO format: YYYY-MM-DD
	if p.position+4 < len(p.tokens) &&
		p.tokens[p.position].Typ == TypeNumber && // YYYY
		p.tokens[p.position+1].Typ == TypeOperator && p.tokens[p.position+1].Val == "-" &&
		p.tokens[p.position+2].Typ == TypeNumber && // MM
		p.tokens[p.position+3].Typ == TypeOperator && p.tokens[p.position+3].Val == "-" &&
		p.tokens[p.position+4].Typ == TypeNumber { // DD
		
		year, _ := strconv.Atoi(p.tokens[p.position].Val)
		p.position++
		p.position++ // Skip the "-"
		month, _ := strconv.Atoi(p.tokens[p.position].Val)
		p.position++
		p.position++ // Skip the "-"
		day, _ := strconv.Atoi(p.tokens[p.position].Val)
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
		
		year, _ := strconv.Atoi(p.tokens[p.position].Val)
		p.position++
		p.position++ // Skip the "/"
		month, _ := strconv.Atoi(p.tokens[p.position].Val)
		p.position++
		p.position++ // Skip the "/"
		day, _ := strconv.Atoi(p.tokens[p.position].Val)
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
		
		month, _ := strconv.Atoi(p.tokens[p.position].Val)
		p.position++
		p.position++ // Skip the "/"
		day, _ := strconv.Atoi(p.tokens[p.position].Val)
		p.position++
		p.position++ // Skip the "/"
		year, _ := strconv.Atoi(p.tokens[p.position].Val)
		p.position++

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
	if token.Typ != TypeString || (token.Val != "next" && token.Val != "last") {
		return time.Time{}, false, nil
	}

	isNext := token.Val == "next"
	p.position++
	p.skipWhitespace()

	// Check for the unit token
	if p.position >= len(p.tokens) {
		return time.Time{}, false, fmt.Errorf("expected time unit after %s", token.Val)
	}

	unitToken := p.tokens[p.position]
	if unitToken.Typ != TypeString {
		return time.Time{}, false, fmt.Errorf("expected time unit after %s, got %s", token.Val, unitToken.Val)
	}

	p.position++

	// Handle special case: "next week" and "last week"
	if unitToken.Val == "week" {
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
	case "month":
		if isNext {
			return p.result.AddDate(0, 1, 0), true, nil
		} else {
			return p.result.AddDate(0, -1, 0), true, nil
		}
	case "year":
		if isNext {
			return p.result.AddDate(1, 0, 0), true, nil
		} else {
			return p.result.AddDate(-1, 0, 0), true, nil
		}
	default:
		return time.Time{}, false, fmt.Errorf("unknown time unit: %s", unitToken.Val)
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
		return time.Time{}, false, fmt.Errorf("expected amount after %s", token.Val)
	}

	amountToken := p.tokens[p.position]
	if amountToken.Typ != TypeNumber {
		return time.Time{}, false, fmt.Errorf("expected number after %s, got %s", token.Val, amountToken.Val)
	}

	amount, err := strconv.Atoi(amountToken.Val)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("invalid number: %s", amountToken.Val)
	}
	amount *= sign
	p.position++

	// Skip whitespace
	p.skipWhitespace()

	// Check for the unit
	if p.position >= len(p.tokens) {
		return time.Time{}, false, fmt.Errorf("expected time unit after %d", amount)
	}

	unitToken := p.tokens[p.position]
	if unitToken.Typ != TypeString {
		return time.Time{}, false, fmt.Errorf("expected time unit after %d, got %s", amount, unitToken.Val)
	}

	p.position++

	// Process the unit
	unit := unitToken.Val
	// Handle plural forms by removing trailing 's'
	if strings.HasSuffix(unit, "s") {
		unit = unit[:len(unit)-1]
	}

	switch unit {
	case "day":
		return p.result.AddDate(0, 0, amount), true, nil
	case "week":
		return p.result.AddDate(0, 0, amount*7), true, nil
	case "month":
		return p.result.AddDate(0, amount, 0), true, nil
	case "year":
		return p.result.AddDate(amount, 0, 0), true, nil
	case "hour":
		return p.result.Add(time.Duration(amount) * time.Hour), true, nil
	case "minute":
		return p.result.Add(time.Duration(amount) * time.Minute), true, nil
	case "second":
		return p.result.Add(time.Duration(amount) * time.Second), true, nil
	default:
		return time.Time{}, false, fmt.Errorf("unknown time unit: %s", unit)
	}
}

// tryParseMonthNameFormat attempts to parse expressions like "January 15 2023" or "Jan 15, 2023"
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

	// Skip optional punctuation (comma) and whitespace
	if p.position < len(p.tokens) && p.tokens[p.position].Typ == TypePunctuation {
		p.position++
	}
	p.skipWhitespace()

	// Check for year
	if p.position >= len(p.tokens) {
		return time.Time{}, false, fmt.Errorf("expected year after day number")
	}

	yearToken := p.tokens[p.position]
	if yearToken.Typ != TypeNumber {
		return time.Time{}, false, fmt.Errorf("expected year after day number, got %s", yearToken.Val)
	}

	year, err := strconv.Atoi(yearToken.Val)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("invalid year: %s", yearToken.Val)
	}
	p.position++

	return time.Date(year, month, day, 0, 0, 0, 0, p.loc), true, nil
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