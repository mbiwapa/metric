package backoff

import (
	"time"

	"github.com/Rican7/retry/backoff"
)

// Backoff returns a backoff.Algorithm that uses exponential backoff.
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
