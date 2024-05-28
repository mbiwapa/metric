// Package backoff provides a backoff algorithm that uses exponential backoff.
// It provides a function to determine the duration to wait before retrying an operation based on the number of attempts made so far.
package backoff

import (
	"time"

	"github.com/Rican7/retry/backoff"
)

// Backoff returns a backoff.Algorithm that uses exponential backoff.
// The algorithm determines the duration to wait before retrying an operation
// based on the number of attempts made so far.
//
// The backoff durations are as follows:
// - 1st attempt: 1 second
// - 2nd attempt: 3 seconds
// - 3rd attempt: 5 seconds
// - Any subsequent attempts: 1 second
//
// Returns:
// - A function that takes the attempt number (uint) and returns the duration (time.Duration) to wait before the next retry.
func Backoff() backoff.Algorithm {
	return func(attempt uint) time.Duration {
		var duration time.Duration
		switch attempt {
		case 1:
			duration = 1 * time.Second
		case 2:
			duration = 3 * time.Second
		case 3:
			duration = 5 * time.Second
		default:
			duration = 1 * time.Second
		}
		return duration
	}
}
