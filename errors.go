package strtotime

import "errors"

// Common errors
var (
	ErrEmptyTimeString = errors.New("empty time string")
	ErrInvalidTimeUnit = errors.New("invalid time unit")
	ErrInvalidNumber   = errors.New("invalid number")
	ErrExpectedTimeUnit = errors.New("expected time unit")
	ErrMissingAmount = errors.New("missing amount")
	ErrMissingDay = errors.New("expected day after month name")
)