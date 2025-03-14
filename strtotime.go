package strtotime

import (
	"errors"
	"time"
)

// StrToTime will convert the provided string into a time similarly to how PHP strtotime() works.
func StrToTime(str string, opts ...Option) (time.Time, error) {
	var now time.Time

	for _, opt := range opts {
		switch v := opt.(type) {
		case Rel: // relative to
			now = time.Time(v)
		}
	}

	if now.IsZero() {
		now = time.Now()
	}

	return time.Time{}, errors.New("TODO: do the thing")
}
