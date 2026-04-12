package strtotime

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestDateParse_PHPShapes asserts that DateParse's JSON output matches the
// structure produced by PHP's date_parse() for a representative set of inputs.
// We compare parsed JSON trees (not raw bytes) so the assertion is resilient
// to whitespace, but every expected key/value is verified.
func TestDateParse_PHPShapes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want map[string]any
	}{
		{
			name: "absolute datetime with fractional seconds and UTC offset",
			in:   "2006-12-12T10:00:00.5+01:00",
			want: map[string]any{
				"year": 2006.0, "month": 12.0, "day": 12.0,
				"hour": 10.0, "minute": 0.0, "second": 0.0,
				"fraction":      0.5,
				"warning_count": 0.0, "warnings": []any{},
				"error_count": 0.0, "errors": []any{},
				"is_localtime": true,
				"zone_type":    1.0,
				"zone":         3600.0,
				"is_dst":       false,
			},
		},
		{
			name: "month and year only",
			in:   "May 2020",
			want: map[string]any{
				"year": 2020.0, "month": 5.0, "day": false,
				"hour": false, "minute": false, "second": false,
				"fraction":      false,
				"warning_count": 0.0, "warnings": []any{},
				"error_count": 0.0, "errors": []any{},
				"is_localtime": false,
			},
		},
		{
			name: "IANA timezone identifier (zone_type 3)",
			in:   "2023-01-01 Asia/Tokyo",
			want: map[string]any{
				"year": 2023.0, "month": 1.0, "day": 1.0,
				"hour": false, "minute": false, "second": false,
				"fraction":      false,
				"warning_count": 0.0, "warnings": []any{},
				"error_count": 0.0, "errors": []any{},
				"is_localtime": true,
				"zone_type":    3.0,
				"tz_id":        "Asia/Tokyo",
			},
		},
		{
			name: "empty string returns error",
			in:   "",
			want: map[string]any{
				"year": false, "month": false, "day": false,
				"hour": false, "minute": false, "second": false,
				"fraction":      false,
				"warning_count": 0.0, "warnings": map[string]any{},
				"error_count":  1.0,
				"is_localtime": false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pd := DateParse(tc.in)
			raw, err := json.Marshal(pd)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal: %v\nraw: %s", err, raw)
			}

			for k, v := range tc.want {
				gv, ok := got[k]
				if !ok {
					t.Errorf("missing key %q in output %s", k, raw)
					continue
				}
				if !jsonEq(gv, v) {
					t.Errorf("field %q: got %v (%T), want %v (%T)\nfull: %s",
						k, gv, gv, v, v, raw)
				}
			}
		})
	}
}

// TestDateParse_Errors ensures that unparseable inputs surface a non-zero
// error_count and populate the errors map.
func TestDateParse_Errors(t *testing.T) {
	for _, in := range []string{
		"garbage",
		"not-a-date",
		"abcdef",
		"!@#$%^&*()",
	} {
		t.Run(in, func(t *testing.T) {
			pd := DateParse(in)
			if pd.ErrorCount == 0 {
				t.Errorf("DateParse(%q): expected ErrorCount > 0", in)
			}
			if len(pd.Errors) == 0 {
				t.Errorf("DateParse(%q): expected non-empty Errors map", in)
			}
		})
	}
}

// TestDateParse_Relative asserts the "relative" block contents for inputs
// containing relative offsets.
func TestDateParse_Relative(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		wantRel     map[string]any
		wantAbsYear any // the "year" field: numeric if set, false otherwise
	}{
		{
			name: "date + positive relative days",
			in:   "2020-01-15 +3 days",
			wantRel: map[string]any{
				"year": 0.0, "month": 0.0, "day": 3.0,
				"hour": 0.0, "minute": 0.0, "second": 0.0,
			},
			wantAbsYear: 2020.0,
		},
		{
			name: "compound relative",
			in:   "2023-01-01 +1 year -2 months",
			wantRel: map[string]any{
				"year": 1.0, "month": -2.0, "day": 0.0,
				"hour": 0.0, "minute": 0.0, "second": 0.0,
			},
			wantAbsYear: 2023.0,
		},
		{
			name: "bare weekday (relative)",
			in:   "next monday",
			wantRel: map[string]any{
				"year": 0.0, "month": 0.0, "day": 0.0,
				"hour": 0.0, "minute": 0.0, "second": 0.0,
				"weekday": 1.0,
			},
			wantAbsYear: false,
		},
		{
			name: "first day of next month (relative)",
			in:   "first day of next month",
			wantRel: map[string]any{
				"year": 0.0, "month": 1.0, "day": 0.0,
				"hour": 0.0, "minute": 0.0, "second": 0.0,
			},
			wantAbsYear: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pd := DateParse(tc.in)
			raw, err := json.Marshal(pd)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if !jsonEq(got["year"], tc.wantAbsYear) {
				t.Errorf("year: got %v, want %v\nfull: %s", got["year"], tc.wantAbsYear, raw)
			}
			relRaw, ok := got["relative"]
			if !ok {
				t.Fatalf("missing relative block\nfull: %s", raw)
			}
			rel, _ := relRaw.(map[string]any)
			for k, v := range tc.wantRel {
				if !jsonEq(rel[k], v) {
					t.Errorf("relative.%s: got %v, want %v\nfull: %s", k, rel[k], v, raw)
				}
			}
		})
	}
}

// TestDateParse_FieldOrder verifies PHP-compatible field order for a canonical
// absolute datetime.
func TestDateParse_FieldOrder(t *testing.T) {
	pd := DateParse("2006-12-12T10:00:00.5+01:00")
	raw, err := json.Marshal(pd)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	expected := []string{
		`"year"`, `"month"`, `"day"`,
		`"hour"`, `"minute"`, `"second"`,
		`"fraction"`,
		`"warning_count"`, `"warnings"`,
		`"error_count"`, `"errors"`,
		`"is_localtime"`,
		`"zone_type"`, `"zone"`, `"is_dst"`,
	}
	s := string(raw)
	pos := 0
	for _, key := range expected {
		idx := strings.Index(s[pos:], key)
		if idx < 0 {
			t.Fatalf("key %s not found (or out of order) in %s", key, s)
		}
		pos += idx + len(key)
	}
}

// jsonEq compares two values decoded from JSON, treating numbers as float64
// and any two empty collections ([]any / map[string]any) as equivalent.
func jsonEq(a, b any) bool {
	if isEmptyCollection(a) && isEmptyCollection(b) {
		return true
	}
	af, aok := a.(float64)
	bf, bok := b.(float64)
	if aok && bok {
		return af == bf
	}
	return a == b
}

func isEmptyCollection(v any) bool {
	switch x := v.(type) {
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	}
	return false
}
