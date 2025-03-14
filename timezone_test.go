package strtotime

import (
	"testing"
)

func TestTimezoneAbbreviations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"January 1 2023 EST",
			"2023-01-01 00:00:00 -0500 EST",
		},
		{
			"January 1 2023 PST",
			"2023-01-01 00:00:00 -0800 PST",
		},
		{
			"January 1 2023 GMT",
			"2023-01-01 00:00:00 +0000 GMT",
		},
		{
			"January 1 2023 UTC",
			"2023-01-01 00:00:00 +0000 UTC",
		},
		{
			"January 1 2023 CET",
			"2023-01-01 00:00:00 +0100 CET",
		},
		{
			"January 1 2023 JST",
			"2023-01-01 00:00:00 +0900 JST",
		},
	}

	for _, test := range tests {
		result, err := StrToTime(test.input)
		if err != nil {
			t.Errorf("Error parsing '%s': %v", test.input, err)
			continue
		}

		expected := test.expected
		got := result.Format("2006-01-02 15:04:05 -0700 MST")

		if got != expected {
			t.Errorf("For input '%s': expected %s, got %s", test.input, expected, got)
		}
	}
}

func TestTimezoneWithTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"January 1 2023 12:30:45 EST",
			"2023-01-01 12:30:45 -0500 EST",
		},
		{
			"January 1 2023 08:15 PST",
			"2023-01-01 08:15:00 -0800 PST",
		},
		{
			"June 1 1985 16:30:00 Europe/Paris", 
			"1985-06-01 16:30:00 +0200 CEST",
		},
	}

	for _, test := range tests {
		result, err := StrToTime(test.input)
		if err != nil {
			t.Errorf("Error parsing '%s': %v", test.input, err)
			continue
		}

		expected := test.expected
		got := result.Format("2006-01-02 15:04:05 -0700 MST")

		if got != expected {
			t.Errorf("For input '%s': expected %s, got %s", test.input, expected, got)
		}
	}
}

func TestFullTimezoneNames(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"January 1 2023 America/New_York",
			"2023-01-01 00:00:00 -0500 EST",
		},
		{
			"January 1 2023 Europe/London",
			"2023-01-01 00:00:00 +0000 GMT",
		},
		{
			"January 1 2023 Europe/Paris",
			"2023-01-01 00:00:00 +0100 CET",
		},
		{
			"January 1 2023 Asia/Tokyo",
			"2023-01-01 00:00:00 +0900 JST",
		},
	}

	for _, test := range tests {
		result, err := StrToTime(test.input)
		if err != nil {
			t.Errorf("Error parsing '%s': %v", test.input, err)
			continue
		}

		expected := test.expected
		got := result.Format("2006-01-02 15:04:05 -0700 MST")

		if got != expected {
			t.Errorf("For input '%s': expected %s, got %s", test.input, expected, got)
		}
	}
}