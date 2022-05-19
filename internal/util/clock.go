package util

import (
	"time"
)

// Clock provides a type for fetching the current time.
type Clock struct{}

// Now fetches the current time using the standard time library.
func (c *Clock) Now() time.Time {
	return time.Now()
}
