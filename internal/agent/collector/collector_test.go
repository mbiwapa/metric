package collector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockMetricGeter is a mock implementation of the MetricGeter interface
type MockMetricGeter struct {
	mock.Mock
}

func (m *MockMetricGeter) MetricGet(name string, sourceType string) (float64, error) {
	args := m.Called(name, sourceType)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMetricGeter) GetObservableMetrics() (map[string]string, error) {
	args := m.Called()
	return args.Get(0).(map[string]string), args.Error(1)
}

// MockMetricUpdater is a mock implementation of the MetricUpdater interface
type MockMetricUpdater struct {
	mock.Mock
}

func (m *MockMetricUpdater) UpdateCounter(ctx context.Context, key string, value int64) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockMetricUpdater) UpdateGauge(ctx context.Context, key string, value float64) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func TestCollector_PollIntervals(t *testing.T) {
	tests := []struct {
		name         string
		pollInterval int64
	}{
		{
			name:         "Poll interval 1 second",
			pollInterval: 1,
		},
		{
			name:         "Poll interval 2 seconds",
			pollInterval: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			errorCh := make(chan error, 1)

			mockMetricGeter := new(MockMetricGeter)
			mockMetricUpdater := new(MockMetricUpdater)

			metrics := map[string]string{"testMetric": "gauge"}
			mockMetricGeter.On("GetObservableMetrics").Return(metrics, nil)
			mockMetricGeter.On("MetricGet", "testMetric", "gauge").Return(1.23, nil)
			mockMetricUpdater.On("UpdateGauge", mock.Anything, "testMetric", 1.23).Return(nil)
			mockMetricUpdater.On("UpdateGauge", mock.Anything, "RandomValue", mock.AnythingOfType("float64")).Return(nil)
			mockMetricUpdater.On("UpdateCounter", mock.Anything, "PollCount", int64(1)).Return(nil)

			go Start(context.Background(), mockMetricUpdater, tt.pollInterval, logger, errorCh, mockMetricGeter)

			// Wait for two poll intervals to ensure the collector has run at least once
			time.Sleep(time.Duration(tt.pollInterval*2) * time.Second)

			mockMetricGeter.AssertExpectations(t)
			mockMetricUpdater.AssertExpectations(t)
		})
	}
}
