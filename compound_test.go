package strtotime

import (
	"testing"
)

func TestComplexCompoundExpressions(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"+1 week 4 days -3 days +4 hours 10 minutes"},
		{"+1 week +4 days -3 days +4 hours +10 minutes"},
		{"next week 4 days -3 days +4 hours 10 minutes"},
		{"+1 week 4 days"},
		{"4 days +10 hours"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := StrToTime(test.input)
			if err != nil {
				t.Fatalf("Failed to parse %q: %v", test.input, err)
			}

			t.Logf("Successfully parsed %q => %s", test.input, result.Format("2006-01-02 15:04:05"))

			// Check results against PHP equivalent (for manual verification)
			// The test doesn't validate specific values since we're just checking if parsing works
		})
	}
}
