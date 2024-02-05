package collector

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// MetricGeter interface for Metric repo
type MetricGeter interface {
	MetricGet(string, sourceType string) (float64, error)
}

// MetricUpdater interface for storage
type MetricUpdater interface {
	UpdateCounter(ctx context.Context, key string, value int64) error
	UpdateGauge(ctx context.Context, key string, value float64) error
}

// ObservableMetric структура для метрик за которыми следим
type ObservableMetric struct {
	Name       string
	SourceType string
}

// Start запускает процесс сбора метрик
func Start(repo MetricGeter, stor MetricUpdater, list []ObservableMetric, pollInterval int64, logger *zap.Logger) {
	logger.Info("Start collector!")
	ctx := context.Background()
	for {
		sleepSecond := time.Duration(pollInterval) * time.Second
		time.Sleep(sleepSecond)
		for _, metric := range list {
			value, err := repo.MetricGet(metric.Name, metric.SourceType)
			if err != nil {
				//TODO error chanel
				logger.Error(
					"Cant get metric",
					zap.String("type", metric.SourceType),
					zap.String("name", metric.Name),
					zap.Error(err))
				panic("Metric no longer supported")
			}
			err = stor.UpdateGauge(ctx, metric.Name, value)
			if err != nil {
				//TODO error chanel
				logger.Error(
					"Cant update metric",
					zap.String("type", metric.SourceType),
					zap.String("name", metric.Name),
					zap.Error(err))
				panic("Storage unavailable!")
			}
		}
		err := stor.UpdateGauge(ctx, "RandomValue", rand.Float64())
		if err != nil {
			//TODO error chanel
			logger.Error(
				"Cant update metric",
				zap.String("type", "gauge"),
				zap.String("name", "RandomValue"),
				zap.Error(err))
			panic("Storage unavailable!")
		}

		err = stor.UpdateCounter(ctx, "PollCount", 1)
		if err != nil {
			//TODO error chanel
			logger.Error(
				"Cant update metric",
				zap.String("type", "couner"),
				zap.String("name", "PollCount"),
				zap.Error(err))
			panic("Storage unavailable!")
		}
	}
}
