package storage

import (
	"errors"
)

var (
	// ErrMetricNotFound is returned when a Metric is not found.
	ErrMetricNotFound = errors.New("Metric not found")
)
