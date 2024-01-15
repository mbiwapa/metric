package sender

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
)

// AllMetricGeter interface for Metric repo
type AllMetricGeter interface {
	GetAllMetrics(ctx context.Context) ([][]string, [][]string, error)
}

// MetricSender interface for sender
type MetricSender interface {
	Send(typ string, name string, value string) error
}

// Start запускает процесс отправки метрик раз в reportInterval секунд
func Start(stor AllMetricGeter, sender MetricSender, reportInterval int64, logger *zap.Logger) {
	logger.Info("Start Sender!")
	ctx := context.Background()
	for {
		gauge, counter, err := stor.GetAllMetrics(ctx)
		if err != nil {
			//TODO error chanel
			logger.Error(
				"Cant get all metrics")
			panic("Stor unavailable!")
		}
		for _, metric := range gauge {

			if metric[0] != "" && metric[1] != "" {
				err = sender.Send(format.Gauge, metric[0], metric[1])
				if err != nil {
					//TODO error chanel
					logger.Error(
						"Cant send metric",
						zap.String("type", format.Gauge),
						zap.String("name", metric[0]),
						zap.String("value", metric[1]),
						zap.Error(err))
					panic(err.Error())

				}
			}
		}
		for _, metric := range counter {

			if metric[0] != "" && metric[1] != "" {
				err = sender.Send(format.Counter, metric[0], metric[1])
				if err != nil {
					//TODO error chanel
					logger.Error(
						"Cant send metric",
						zap.String("type", format.Counter),
						zap.String("name", metric[0]),
						zap.String("value", metric[1]),
						zap.Error(err))
					panic(err.Error())
				}
			}
		}
		sleepSecond := time.Duration(reportInterval) * time.Second
		time.Sleep(sleepSecond)
	}
}
