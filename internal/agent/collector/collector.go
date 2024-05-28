// Package collector provides a collector that periodically polls metric sources and updates the storage with the collected metrics.
// It also updates a random gauge value and a poll count counter at each interval.
package collector

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// MetricGeter interface for Metric repo
type MetricGeter interface {
	// MetricGet retrieves the metric value for a given metric name and source type.
	// Returns the metric value as a float64 and an error if the retrieval fails.
	MetricGet(string, sourceType string) (float64, error)

	// GetObservableMetrics returns a map of metric names to their types that can be observed.
	// Returns a map of metric names to types and an error if the retrieval fails.
	GetObservableMetrics() (map[string]string, error)
}

// MetricUpdater interface for storage
type MetricUpdater interface {
	// UpdateCounter updates the counter metric with the given key and value.
	// Returns an error if the update fails.
	UpdateCounter(ctx context.Context, key string, value int64) error

	// UpdateGauge updates the gauge metric with the given key and value.
	// Returns an error if the update fails.
	UpdateGauge(ctx context.Context, key string, value float64) error
}

// Start initializes and starts the metric collection process.
// It periodically polls the provided MetricGeter sources for metrics and updates the storage using MetricUpdater.
// Additionally, it updates a random gauge value and a poll count counter at each interval.
//
// Parameters:
// - stor: MetricUpdater interface for updating metrics in storage.
// - pollInterval: Interval in seconds between each poll.
// - logger: Logger for logging information and errors.
// - errorChanel: Channel for sending errors encountered during metric collection.
// - sources: Variadic parameter of MetricGeter sources to poll for metrics.
func Start(stor MetricUpdater, pollInterval int64, logger *zap.Logger, errorChanel chan<- error, sources ...MetricGeter) {
	logger.Info("Start collector!")
	ctx := context.Background()

	for _, source := range sources {

		list, err := source.GetObservableMetrics()
		if err != nil {
			errorChanel <- err
		}

		go func(source MetricGeter, list map[string]string) {
			for {
				sleepSecond := time.Duration(pollInterval) * time.Second
				time.Sleep(sleepSecond)
				for name, typ := range list {
					value, err := source.MetricGet(name, typ)
					if err != nil {
						errorChanel <- fmt.Errorf("%s: %w", "Collector:", err)
					}
					err = stor.UpdateGauge(ctx, name, value)
					if err != nil {
						errorChanel <- fmt.Errorf("%s: %w", "Collector:", err)
					}
				}
			}
		}(source, list)
	}

	go func() {
		for {
			sleepSecond := time.Duration(pollInterval) * time.Second
			time.Sleep(sleepSecond)

			err := stor.UpdateGauge(ctx, "RandomValue", rand.Float64())
			if err != nil {
				errorChanel <- fmt.Errorf("%s: %w", "Collector:", err)
			}

			err = stor.UpdateCounter(ctx, "PollCount", 1)
			if err != nil {
				errorChanel <- fmt.Errorf("%s: %w", "Collector:", err)
			}
		}

	}()
}
