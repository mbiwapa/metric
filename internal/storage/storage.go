package storage

import (
	"errors"
)

var (
	// ErrMetricNotFound is returned when a Metric is not found.
	ErrMetricNotFound = errors.New("metric not found")
)
