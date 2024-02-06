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
	MetricGet(string, sourceType string) (float64, error)
	GetObservableMetrics() (map[string]string, error)
}

// MetricUpdater interface for storage
type MetricUpdater interface {
	UpdateCounter(ctx context.Context, key string, value int64) error
	UpdateGauge(ctx context.Context, key string, value float64) error
}

// Start запускает процесс сбора метрик
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
