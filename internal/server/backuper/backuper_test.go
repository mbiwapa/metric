package backuper

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
)

// Mock implementation of the AllMetricGeter interface for testing purposes.
type MockAllMetricGeter struct {
	mock.Mock
}

func (m *MockAllMetricGeter) GetAllMetrics(ctx context.Context) ([][]string, [][]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([][]string), args.Get(1).([][]string), args.Error(2)
}

func (m *MockAllMetricGeter) UpdateGauge(ctx context.Context, key string, value float64) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockAllMetricGeter) UpdateCounter(ctx context.Context, key string, value int64) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func TestNew(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	mockStorage := new(MockAllMetricGeter)
	storeInterval := int64(10)
	storagePath := "test_metrics.json"

	buckuper, err := New(mockStorage, storeInterval, storagePath, logger)
	require.NoError(t, err)
	require.NotNil(t, buckuper)
	require.Equal(t, storeInterval, buckuper.storeInterval)
	require.Equal(t, storagePath, buckuper.storagePath)
}

func TestSaveToStruct(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	mockStorage := new(MockAllMetricGeter)
	buckuper, _ := New(mockStorage, 10, "test_metrics.json", logger)

	err := buckuper.SaveToStruct(format.Gauge, "testGauge", "123.45")
	require.NoError(t, err)
	require.Len(t, buckuper.metrics, 1)
	require.Equal(t, "testGauge", buckuper.metrics[0].ID)
	require.Equal(t, format.Gauge, buckuper.metrics[0].MType)
	require.Equal(t, 123.45, *buckuper.metrics[0].Value)

	err = buckuper.SaveToStruct(format.Counter, "testCounter", "678")
	require.NoError(t, err)
	require.Len(t, buckuper.metrics, 2)
	require.Equal(t, "testCounter", buckuper.metrics[1].ID)
	require.Equal(t, format.Counter, buckuper.metrics[1].MType)
	require.Equal(t, int64(678), *buckuper.metrics[1].Delta)
}

func TestSaveToFile(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	mockStorage := new(MockAllMetricGeter)
	buckuper, _ := New(mockStorage, 10, "test_metrics.json", logger)

	buckuper.SaveToStruct(format.Gauge, "testGauge", "123.45")
	buckuper.SaveToFile()

	data, err := os.ReadFile("test_metrics.json")
	require.NoError(t, err)

	var metrics []format.Metric
	err = json.Unmarshal(data, &metrics)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "testGauge", metrics[0].ID)
	require.Equal(t, format.Gauge, metrics[0].MType)
	require.Equal(t, 123.45, *metrics[0].Value)

	// Clean up
	os.Remove("test_metrics.json")
}

func TestRestore(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	mockStorage := new(MockAllMetricGeter)
	buckuper, _ := New(mockStorage, 10, "test_metrics.json", logger)

	// Prepare test data
	metrics := []format.Metric{
		{MType: format.Gauge, ID: "testGauge", Value: new(float64)},
		{MType: format.Counter, ID: "testCounter", Delta: new(int64)},
	}
	*metrics[0].Value = 123.45
	*metrics[1].Delta = 678

	data, _ := json.Marshal(metrics)
	os.WriteFile("test_metrics.json", data, 0666)

	mockStorage.On("UpdateGauge", mock.Anything, "testGauge", 123.45).Return(nil)
	mockStorage.On("UpdateCounter", mock.Anything, "testCounter", int64(678)).Return(nil)

	buckuper.Restore()

	mockStorage.AssertExpectations(t)

	// Clean up
	os.Remove("test_metrics.json")
}

func TestStart(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	mockStorage := new(MockAllMetricGeter)
	buckuper, _ := New(mockStorage, 1, "test_metrics.json", logger)

	mockStorage.On("GetAllMetrics", mock.Anything).Return([][]string{{"testGauge", "123.45"}}, [][]string{{"testCounter", "678"}}, nil)

	go buckuper.Start()

	time.Sleep(2 * time.Second)

	data, err := os.ReadFile("test_metrics.json")
	require.NoError(t, err)

	var metrics []format.Metric
	err = json.Unmarshal(data, &metrics)
	require.NoError(t, err)
	require.Len(t, metrics, 2)
	require.Equal(t, "testGauge", metrics[0].ID)
	require.Equal(t, format.Gauge, metrics[0].MType)
	require.Equal(t, 123.45, *metrics[0].Value)
	require.Equal(t, "testCounter", metrics[1].ID)
	require.Equal(t, format.Counter, metrics[1].MType)
	require.Equal(t, int64(678), *metrics[1].Delta)

	// Clean up
	os.Remove("test_metrics.json")
}
