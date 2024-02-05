package sender

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// AllMetricGeter interface for Metric repo
type AllMetricGeter interface {
	GetAllMetrics(ctx context.Context) ([][]string, [][]string, error)
}

// MetricSender interface for sender
type MetricSender interface {
	Send(gauges [][]string, counters [][]string) error
}

// Start запускает процесс отправки метрик раз в reportInterval секунд
func Start(stor AllMetricGeter, sender MetricSender, reportInterval int64, logger *zap.Logger) {
	logger.Info("Start Sender!")
	ctx := context.Background()
	for {
		sleepSecond := time.Duration(reportInterval) * time.Second
		time.Sleep(sleepSecond)
		gauge, counter, err := stor.GetAllMetrics(ctx)
		if err != nil {
			//TODO error chanel
			logger.Error(
				"Cant get all metrics")
			panic("Stor unavailable!")
		}
		err = sender.Send(gauge, counter)
		if err != nil {
			//TODO error chanel
			logger.Error(
				"Cant send metric",
				zap.Any("gauge", gauge),
				zap.Any("counter", counter),
				zap.Error(err))
			panic(err.Error())

		}
	}
}
