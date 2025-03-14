package strtotime

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  []Token
	}{
		{
			"hello world",
			[]Token{
				{Val: "hello", Typ: TypeString, Pos: 0},
				{Val: " ", Typ: TypeWhitespace, Pos: 5},
				{Val: "world", Typ: TypeString, Pos: 6},
			},
		},
		{
			"2023-01-15",
			[]Token{
				{Val: "2023", Typ: TypeNumber, Pos: 0},
				{Val: "-", Typ: TypeOperator, Pos: 4},
				{Val: "01", Typ: TypeNumber, Pos: 5},
				{Val: "-", Typ: TypeOperator, Pos: 7},
				{Val: "15", Typ: TypeNumber, Pos: 8},
			},
		},
		{
			"+1 day",
			[]Token{
				{Val: "+", Typ: TypeOperator, Pos: 0},
				{Val: "1", Typ: TypeNumber, Pos: 1},
				{Val: " ", Typ: TypeWhitespace, Pos: 2},
				{Val: "day", Typ: TypeString, Pos: 3},
			},
		},
		{
			"next Friday at 3:30pm",
			[]Token{
				{Val: "next", Typ: TypeString, Pos: 0},
				{Val: " ", Typ: TypeWhitespace, Pos: 4},
				{Val: "Friday", Typ: TypeString, Pos: 5},
				{Val: " ", Typ: TypeWhitespace, Pos: 11},
				{Val: "at", Typ: TypeString, Pos: 12},
				{Val: " ", Typ: TypeWhitespace, Pos: 14},
				{Val: "3", Typ: TypeNumber, Pos: 15},
				{Val: ":", Typ: TypeOperator, Pos: 16},
				{Val: "30", Typ: TypeNumber, Pos: 17},
				{Val: "pm", Typ: TypeString, Pos: 19},
			},
		},
		{
			"Jan 15, 2023",
			[]Token{
				{Val: "Jan", Typ: TypeString, Pos: 0},
				{Val: " ", Typ: TypeWhitespace, Pos: 3},
				{Val: "15", Typ: TypeNumber, Pos: 4},
				{Val: ",", Typ: TypePunctuation, Pos: 6},
				{Val: " ", Typ: TypeWhitespace, Pos: 7},
				{Val: "2023", Typ: TypeNumber, Pos: 8},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Tokenize(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tokenize(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
