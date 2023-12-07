package collector

import (
	"math/rand"
	"time"
)

// MetricGeter interface for Metric repo
type MetricGeter interface {
	MetricGet(metricName string, sourceType string) (float64, error)
}

// MetricUpdater interface for storage
type MetricUpdater interface {
	CounterUpdate(key string, value int64) error
	GaugeUpdate(key string, value float64) error
}

// ObservableMetric структура для метрик за которыми следим
type ObservableMetric struct {
	Name       string
	SourceType string
}

// Start запускает процесс сбора метрик
func Start(repo MetricGeter, stor MetricUpdater, list []ObservableMetric, pollInterval int64) {
	for {
		for _, metric := range list {
			value, err := repo.MetricGet(metric.Name, metric.SourceType)
			if err != nil {
				//TODO ошибку сделать и можно ли тут паниковать?
				panic("Metric no longer supported")
			}
			err = stor.GaugeUpdate(metric.Name, value)
			if err != nil {
				//TODO ошибку сделать
				panic("Stor unavailable!")
			}
		}
		err := stor.GaugeUpdate("RandomValue", rand.Float64())
		if err != nil {
			//TODO ошибку сделать
			panic("Stor unavailable!")
		}

		err = stor.CounterUpdate("PollCount", 1)
		if err != nil {
			//TODO ошибку сделать
			panic("Stor unavailable!")
		}

		sleepSecond := time.Duration(pollInterval) * time.Second
		time.Sleep(sleepSecond)
	}
}
