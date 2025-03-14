package strtotime

import "time"

type Option interface {
	isOption() bool
}

// Rel represents a relative time to use as base
type Rel time.Time

func (r Rel) isOption() bool {
	return true
}

// TZ represents a timezone to use for parsing
type TZ struct {
	Location *time.Location
}

func (t TZ) isOption() bool {
	return true
}
