package common

import "time"

// AppNowUTC returns the current time in UTC with truncate after seconds.
// This should be used for all time-related functions.
func AppNowUTC() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}
