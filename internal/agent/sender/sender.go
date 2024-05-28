// Package sender provides a sender that processes and sends metrics.
// It also provides a function to start the process of sending metrics every reportInterval seconds.
package sender

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// AllMetricGeter interface for Metric repo
type AllMetricGeter interface {
	// GetAllMetrics retrieves all metrics from the storage.
	// It returns two slices of string slices representing gauge and counter metrics, respectively, and an error if any occurs.
	GetAllMetrics(ctx context.Context) ([][]string, [][]string, error)
}

// MetricSender interface for sender
type MetricSender interface {
	// Worker processes jobs from the jobs channel and sends errors to the error channel.
	// jobs is a channel that provides maps of metrics to be processed.
	// errorChanel is a channel to send errors encountered during processing.
	Worker(jobs <-chan map[string][][]string, errorChanel chan<- error)
}

// Start initiates the process of sending metrics every reportInterval seconds.
// stor is an implementation of the AllMetricGeter interface to retrieve metrics.
// sender is an implementation of the MetricSender interface to process and send metrics.
// reportInterval is the interval in seconds between each metrics retrieval and sending.
// logger is used for logging information and errors.
// numWorker is the number of worker goroutines to process the metrics.
// errorChanel is a channel to send errors encountered during the process.
// ctx is the context to control the lifecycle of the function.
func Start(ctx context.Context, stor AllMetricGeter, sender MetricSender, reportInterval int64, logger *zap.Logger, numWorker int, errorChanel chan<- error) {
	logger.Info("Start Sender!")

	jobsChanel := make(chan map[string][][]string)
	for i := 1; i <= numWorker; i++ {
		go sender.Worker(jobsChanel, errorChanel)
	}

	go func(jobs chan<- map[string][][]string) {
		for {
			select {
			case <-ctx.Done():
				logger.Info("Stopping Sender!")
				return
			case <-time.After(time.Duration(reportInterval) * time.Second):
				gauge, counter, err := stor.GetAllMetrics(ctx)
				if err != nil {
					errorChanel <- fmt.Errorf("%s: %w", "Sender:", err)
				}
				jobs <- map[string][][]string{"gauge": gauge, "counter": counter}
			}
		}
	}(jobsChanel)
}
