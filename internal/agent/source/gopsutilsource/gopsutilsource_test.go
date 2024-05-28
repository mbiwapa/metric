package gopsutilsource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkMetricGet(b *testing.B) {
	storage := &MetricsRepo{}
	metricName := "TotalMemory"
	sourceType := "uint"

	// Benchmark for MetricGet function
	b.Run("MetricGet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = storage.MetricGet(metricName, sourceType)
		}
	})
}

func TestNew(t *testing.T) {
	// Test that New function returns a non-nil instance of MetricsRepo
	t.Run("New returns non-nil instance", func(t *testing.T) {
		storage, err := New()
		assert.NoError(t, err)
		assert.NotNil(t, storage)
	})

	// Test that New function returns an instance of the correct type
	t.Run("New returns correct type", func(t *testing.T) {
		storage, err := New()
		assert.NoError(t, err)
		assert.IsType(t, &MetricsRepo{}, storage)
	})

	// Test that New function does not return an error
	t.Run("New does not return error", func(t *testing.T) {
		_, err := New()
		assert.NoError(t, err)
	})

	// Test that New function initializes the MetricsRepo struct correctly
	t.Run("New initializes MetricsRepo correctly", func(t *testing.T) {
		storage, err := New()
		assert.NoError(t, err)
		assert.Equal(t, &MetricsRepo{}, storage)
	})

	// Test that New function can be called multiple times without error
	t.Run("New can be called multiple times", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			storage, err := New()
			assert.NoError(t, err)
			assert.NotNil(t, storage)
		}
	})
}

func TestMetricGet_UndefinedMetricError(t *testing.T) {
	storage := &MetricsRepo{}
	_, err := storage.MetricGet("UndefinedMetric", "memory")
	assert.Error(t, err)
	assert.Equal(t, "undefined metric: UndefinedMetric", err.Error())
}

func TestGetObservableMetrics(t *testing.T) {
	storage := &MetricsRepo{}

	t.Run("Returns memory metrics", func(t *testing.T) {
		metrics, err := storage.GetObservableMetrics()
		assert.NoError(t, err)
		assert.Contains(t, metrics, "TotalMemory")
		assert.Contains(t, metrics, "FreeMemory")
		assert.Equal(t, "memory", metrics["TotalMemory"])
		assert.Equal(t, "memory", metrics["FreeMemory"])
	})
}
