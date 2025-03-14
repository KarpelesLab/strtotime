package strtotime

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// BenchmarkStrToTime benchmarks the main StrToTime function with various date formats
func BenchmarkStrToTime(b *testing.B) {
	// Define benchmark cases
	benchmarks := []struct {
		name  string
		input string
	}{
		{"UnixTimestamp", "@1121373041"},
		{"UnixTimestampWithFraction", "@1121373041.123"},
		{"CompactTimestamp", "19970523091528"},
		{"ISODate", "2023-01-15"},
		{"SlashDate", "2023/01/15"},
		{"USDate", "01/15/2023"},
		{"EuropeanDate", "15.01.2023"},
		{"MonthNameDMY", "Jan-15-2006"},
		{"MonthNameYMD", "2006-Jan-15"},
		{"HTTPLogFormat", "10/Oct/2000:13:55:36 +0100"},
		{"NumberedWeekday", "first Monday December 2008"},
		{"RelativeSimple", "now"},
		{"RelativeComplex", "next Monday"},
		{"RelativeOffset", "+1 day"},
		{"CompoundExpression", "next year+4 days"},
		{"DateTimeFormat", "2023-01-15 10:30:45"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Reset the timer to exclude setup time
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				_, err := StrToTime(bm.input, InTZ(time.UTC))
				if err != nil {
					b.Fatalf("Error parsing '%s': %v", bm.input, err)
				}
			}
		})
	}
}

// BenchmarkSpecificParsers benchmarks individual parser functions
func BenchmarkSpecificParsers(b *testing.B) {
	// Define benchmark cases
	benchmarks := []struct {
		name     string
		input    string
		function func(string, *time.Location) (time.Time, bool)
	}{
		{"ISO", "2023-01-15", parseISOFormat},
		{"Slash", "2023/01/15", parseSlashFormat},
		{"US", "01/15/2023", parseUSFormat},
		{"European", "15.01.2023", parseEuropeanFormat},
		{"Compact", "19970523091528", parseCompactTimestamp},
		{"MonthName", "Jan-15-2006", parseMonthNameFormat},
		{"HTTPLog", "10/Oct/2000:13:55:36 +0100", parseHTTPLogFormat},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Reset the timer to exclude setup time
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				_, ok := bm.function(bm.input, time.UTC)
				if !ok {
					b.Fatalf("Failed to parse '%s'", bm.input)
				}
			}
		})
	}
}

// BenchmarkNumberedWeekday benchmarks the parseNumberedWeekday function separately
func BenchmarkNumberedWeekday(b *testing.B) {
	input := "first Monday December 2008"
	reference := time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC)

	// Reset the timer to exclude setup time
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		_, ok := parseNumberedWeekday(input, reference, time.UTC)
		if !ok {
			b.Fatalf("Failed to parse '%s'", input)
		}
	}
}

// BenchmarkRegexCompilation benchmarks the impact of regex compilation in parsing
func BenchmarkRegexCompilation(b *testing.B) {
	b.Run("WithPrecompiledRegex", func(b *testing.B) {
		// Pre-compile the regex outside the benchmark loop
		re := compileNumberedWeekdayRegex()
		input := "first Monday December 2008"
		reference := time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, ok := parseNumberedWeekdayWithRegex(input, reference, time.UTC, re)
			if !ok {
				b.Fatalf("Failed to parse '%s'", input)
			}
		}
	})

	b.Run("WithDynamicRegex", func(b *testing.B) {
		// Regex is compiled within the function for each call
		input := "first Monday December 2008"
		reference := time.Date(2008, 12, 1, 0, 0, 0, 0, time.UTC)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, ok := parseNumberedWeekday(input, reference, time.UTC)
			if !ok {
				b.Fatalf("Failed to parse '%s'", input)
			}
		}
	})
}

// Helper functions for the regex compilation benchmark
func compileNumberedWeekdayRegex() *regexp.Regexp {
	return regexp.MustCompile(`^(?:(\d+)|(?:(first|1st|second|2nd|third|3rd|fourth|4th|fifth|5th|last)))\s+([A-Za-z]+)(?:\s+(?:of\s+)?)?([A-Za-z]+)(?:\s+(\d{4}))?$`)
}

func parseNumberedWeekdayWithRegex(str string, now time.Time, loc *time.Location, re *regexp.Regexp) (time.Time, bool) {
	if matches := re.FindStringSubmatch(str); matches != nil {
		var ordinal int
		
		// Parse the ordinal (numeric or word)
		if matches[1] != "" {
			// Numeric ordinal
			ordinal, _ = strconv.Atoi(matches[1])
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
			year, _ = strconv.Atoi(matches[5])
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