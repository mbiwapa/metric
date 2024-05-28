// Package backuper provides a structure for saving and restoring metrics.
// It contains the necessary components to periodically save and restore metrics from a storage.
package backuper

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
)

// AllMetricGeter is an interface for the Metric repository.
// It defines methods for retrieving and updating metrics.
type AllMetricGeter interface {
	// GetAllMetrics retrieves all metrics from the storage.
	// Parameters:
	// - ctx: a context.Context for managing request-scoped values, cancelation, and deadlines.
	// Returns:
	// - [][]string: a slice of slices containing gauge metrics.
	// - [][]string: a slice of slices containing counter metrics.
	// - error: an error if any occurs during the retrieval process.
	GetAllMetrics(ctx context.Context) ([][]string, [][]string, error)

	// UpdateGauge updates the value of a gauge metric in the storage.
	// Parameters:
	// - ctx: a context.Context for managing request-scoped values, cancelation, and deadlines.
	// - key: a string representing the name of the gauge metric.
	// - value: a float64 representing the new value of the gauge metric.
	// Returns:
	// - error: an error if any occurs during the update process.
	UpdateGauge(ctx context.Context, key string, value float64) error

	// UpdateCounter updates the value of a counter metric in the storage.
	// Parameters:
	// - ctx: a context.Context for managing request-scoped values, cancelation, and deadlines.
	// - key: a string representing the name of the counter metric.
	// - value: an int64 representing the new value of the counter metric.
	// Returns:
	// - error: an error if any occurs during the update process.
	UpdateCounter(ctx context.Context, key string, value int64) error
}

// metrics is a type alias for a slice of format.Metric.
// It is used to hold a collection of metrics.
type metrics []format.Metric

// Buckuper is a structure for saving metrics.
// It contains the necessary components to periodically save and restore metrics from a storage.
type Buckuper struct {
	// logger is a zap.Logger instance used for logging information, warnings, and errors.
	logger *zap.Logger

	// storage is an implementation of the AllMetricGeter interface, which provides methods to get and update metrics.
	storage AllMetricGeter

	// storeInterval is the interval in seconds at which metrics should be saved.
	storeInterval int64

	// storagePath is the file path where metrics will be saved.
	storagePath string

	// metrics is a slice that holds the metrics to be saved.
	metrics metrics
}

// New creates a new instance of Saver
// Parameters:
// - storage: an implementation of the AllMetricGeter interface for metric storage
// - storeInterval: the interval in seconds at which metrics should be saved
// - storagePath: the file path where metrics will be saved
// - logger: a zap.Logger instance for logging
// Returns:
// - a pointer to a new Buckuper instance
// - an error if any occurs during the creation of the Buckuper instance
func New(storage AllMetricGeter, storeInterval int64, storagePath string, logger *zap.Logger) (*Buckuper, error) {
	var metrics metrics

	return &Buckuper{
		logger:        logger,
		storage:       storage,
		storeInterval: storeInterval,
		storagePath:   storagePath,
		metrics:       metrics,
	}, nil
}

// Start initiates the process of periodically saving metrics to a file.
// It runs an infinite loop that sleeps for the duration specified by storeInterval,
// retrieves all metrics from the storage, saves them to the internal metrics slice,
// and then writes the metrics to a file.
//
// The method logs the start of the saver, the sleep duration, and any errors encountered
// during the retrieval, saving, or writing of metrics.
//
// Parameters:
// - None
//
// Returns:
// - None
func (s *Buckuper) Start() {
	const op = "server.saver.Start"
	s.logger.With(zap.String("op", op))
	s.logger.Info("Start Saver!")
	ctx := context.Background()

	for {
		s.logger.Info("Sleep " + strconv.FormatInt(s.storeInterval, 10) + " seconds")
		sleepSecond := time.Duration(s.storeInterval) * time.Second
		time.Sleep(sleepSecond)

		gauge, counter, err := s.storage.GetAllMetrics(ctx)
		if err != nil {
			//TODO error chanel
			s.logger.Error("Cant get all metrics", zap.Error(err))
		}

		for _, metric := range gauge {
			if metric[0] != "" && metric[1] != "" {
				err = s.SaveToStruct(format.Gauge, metric[0], metric[1])
				if err != nil {
					//TODO error chanel
					s.logger.Error("Cant save metric to struct", zap.Error(err))
				}
			}
		}
		for _, metric := range counter {
			if metric[0] != "" && metric[1] != "" {
				s.SaveToStruct(format.Counter, metric[0], metric[1])
				if err != nil {
					//TODO error chanel
					s.logger.Error("Cant save metric to struct", zap.Error(err))
				}
			}
		}
		s.SaveToFile()
	}
}

// SaveToStruct saves a metric to the metrics slice
// Parameters:
// - typ: the type of the metric (e.g., Gauge or Counter)
// - name: the name of the metric
// - value: the value of the metric as a string
// Returns:
// - an error if any occurs during the saving process
func (s *Buckuper) SaveToStruct(typ string, name string, value string) error {
	const op = "server.saver.SaveToStruct"
	s.logger.With(zap.String("op", op))

	m := format.Metric{
		MType: typ,
		ID:    name,
	}

	switch typ {
	case format.Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			s.logger.Error("Cant parse gauge metric", zap.Error(err))
			return err
		}
		m.Value = &val
	case format.Counter:
		val, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			s.logger.Error("Cant parse counter metric", zap.Error(err))
			return err
		}
		m.Delta = &val
	default:
	}

	changed := false

	for i := 0; i < len(s.metrics); i++ {
		if s.metrics[i].ID == name {
			s.metrics[i] = m
			changed = true
			break
		}
	}

	if !changed {
		s.metrics = append(s.metrics, m)
	}

	return nil
}

// SaveToFile saves the current metrics to a file in JSON format.
// It marshals the metrics slice into a JSON-formatted byte array and writes it to the specified storage path.
// The method logs the start and completion of the save process, as well as any errors encountered during
// the encoding or writing of the JSON data.
//
// Parameters:
// - None
//
// Returns:
// - None
func (s *Buckuper) SaveToFile() {
	const op = "server.saver.SaveToFile"
	s.logger.With(zap.String("op", op))

	s.logger.Info("Start save!")

	data, err := json.MarshalIndent(s.metrics, "", "   ")
	if err != nil {
		s.logger.Error(
			"Cant encoding metric to json", zap.Error(err))
	}
	err = os.WriteFile(s.storagePath, data, 0666)
	if err != nil {
		s.logger.Error(
			"Cant write json to file", zap.Error(err))
	}

	s.logger.Info("Complete save!")
}

// Restore restores the metrics from the file and updates the storage with the restored metrics.
// It reads the metrics from the specified storage file, unmarshals the JSON data into the metrics slice,
// and updates the storage with the restored metrics.
//
// The method logs the start and completion of the restore process, as well as any errors encountered
// during reading, decoding, or updating the metrics.
//
// Parameters:
// - None
//
// Returns:
// - None
func (s *Buckuper) Restore() {
	const op = "server.saver.Restore"
	s.logger.With(zap.String("op", op))
	s.logger.Info("Start Restore!")

	ctx := context.Background()

	data, err := os.ReadFile(s.storagePath)
	if err != nil {
		s.logger.Error(
			"Cant read json from file", zap.Error(err))
	}

	err = json.Unmarshal(data, &s.metrics)
	if err != nil {
		s.logger.Error(
			"Cant decode json", zap.Error(err))
	}

	for _, sourceMetric := range s.metrics {
		databaseCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		switch sourceMetric.MType {
		case format.Gauge:
			err = s.storage.UpdateGauge(databaseCtx, sourceMetric.ID, *sourceMetric.Value)
			if err != nil {
				s.logger.Error("Failed to update gauge value", zap.Error(err))
			}
		case format.Counter:
			err = s.storage.UpdateCounter(databaseCtx, sourceMetric.ID, *sourceMetric.Delta)
			if err != nil {
				s.logger.Error("Failed to update counter value", zap.Error(err))
			}
		default:
		}
	}
	s.logger.Info("Complete Restore!")
}

// IsSyncMode returns true if sync mode is enabled
// Returns:
// - true if storeInterval is 0, indicating sync mode is enabled
// - false otherwise
func (s *Buckuper) IsSyncMode() bool {
	if s.storeInterval == 0 {
		return true
	} else {
		return false
	}
}
