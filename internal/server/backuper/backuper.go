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

// AllMetricGeter interface for Metric repo
type AllMetricGeter interface {
	GetAllMetrics(ctx context.Context) ([][]string, [][]string, error)
	UpdateGauge(ctx context.Context, key string, value float64) error
	UpdateCounter(ctx context.Context, key string, value int64) error
}

// metrics struct all metrics
type metrics []format.Metric

// Buckuper Структура для сохранения метрик
type Buckuper struct {
	logger        *zap.Logger
	storage       AllMetricGeter
	storeInterval int64
	storagePath   string
	metrics       metrics
}

// New creates a new instance of Saver
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

// Start запускает процесс сохранения метрик раз в storeInterval секунд
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

// SaveToFile saves a metric to the file
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

// Restore restores the metrics from the file
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
func (s *Buckuper) IsSyncMode() bool {
	if s.storeInterval == 0 {
		return true
	} else {
		return false
	}
}
