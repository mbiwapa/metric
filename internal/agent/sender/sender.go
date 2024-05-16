package sender

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// AllMetricGeter interface for Metric repo
type AllMetricGeter interface {
	GetAllMetrics(ctx context.Context) ([][]string, [][]string, error)
}

// MetricSender interface for sender
type MetricSender interface {
	Worker(jobs <-chan map[string][][]string, errorChanel chan<- error)
}

// Start запускает процесс отправки метрик раз в reportInterval секунд
func Start(stor AllMetricGeter, sender MetricSender, reportInterval int64, logger *zap.Logger, numWorker int, errorChanel chan<- error) {
	logger.Info("Start Sender!")
	ctx := context.Background()

	jobsChanel := make(chan map[string][][]string)
	for i := 1; i <= numWorker; i++ {
		go sender.Worker(jobsChanel, errorChanel)
	}

	go func(jobs chan<- map[string][][]string) {
		for {
			sleepSecond := time.Duration(reportInterval) * time.Second
			time.Sleep(sleepSecond)
			gauge, counter, err := stor.GetAllMetrics(ctx)
			if err != nil {
				errorChanel <- fmt.Errorf("%s: %w", "Sender:", err)
			}
			jobs <- map[string][][]string{"gauge": gauge, "counter": counter}
		}
	}(jobsChanel)
}
