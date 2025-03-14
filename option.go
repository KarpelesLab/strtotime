package strtotime

import "time"

type Option interface {
	isOption() bool
}

type Rel time.Time

func (r Rel) isOption() bool {
	return true
}
