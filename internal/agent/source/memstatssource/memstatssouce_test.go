package memstatssource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkMetricGet(b *testing.B) {
	storage := &MetricsRepo{}
	metricName := "HeapInuse"
	sourceType := "uint"

	// Benchmark for MetricGet function
	b.Run("MetricGet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = storage.MetricGet(metricName, sourceType)
		}
	})
}

func TestGetObservableMetrics(t *testing.T) {
	repo, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	metrics, err := repo.GetObservableMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	expectedMetrics := map[string]string{
		"Frees":         "uint",
		"Alloc":         "uint",
		"BuckHashSys":   "uint",
		"GCCPUFraction": "float",
		"GCSys":         "uint",
		"HeapAlloc":     "uint",
		"HeapIdle":      "uint",
		"HeapInuse":     "uint",
		"HeapObjects":   "uint",
		"HeapReleased":  "uint",
		"HeapSys":       "uint",
		"LastGC":        "uint",
		"Lookups":       "uint",
		"MCacheInuse":   "uint",
		"MCacheSys":     "uint",
		"MSpanInuse":    "uint",
		"MSpanSys":      "uint",
		"Mallocs":       "uint",
		"NextGC":        "uint",
		"NumForcedGC":   "uint",
		"NumGC":         "uint",
		"OtherSys":      "uint",
		"PauseTotalNs":  "uint",
		"StackInuse":    "uint",
		"StackSys":      "uint",
		"Sys":           "uint",
		"TotalAlloc":    "uint",
	}

	assert.Equal(t, expectedMetrics, metrics)
}

func TestGetObservableMetricsContainsGCCPUFraction(t *testing.T) {
	repo, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	metrics, err := repo.GetObservableMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	value, exists := metrics["GCCPUFraction"]
	assert.True(t, exists)
	assert.Equal(t, "float", value)
}

func TestGetObservableMetricsContainsHeapAlloc(t *testing.T) {
	repo, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	metrics, err := repo.GetObservableMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	value, exists := metrics["HeapAlloc"]
	assert.True(t, exists)
	assert.Equal(t, "uint", value)
}

func TestGetObservableMetricsContainsNumGC(t *testing.T) {
	repo, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	metrics, err := repo.GetObservableMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	value, exists := metrics["NumGC"]
	assert.True(t, exists)
	assert.Equal(t, "uint", value)
}

func TestGetObservableMetricsContainsTotalAlloc(t *testing.T) {
	repo, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	metrics, err := repo.GetObservableMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	value, exists := metrics["TotalAlloc"]
	assert.True(t, exists)
	assert.Equal(t, "uint", value)
}

func TestNew(t *testing.T) {
	t.Run("Test New returns non-nil MetricsRepo", func(t *testing.T) {
		repo, err := New()
		require.NoError(t, err)
		require.NotNil(t, repo)
	})

	t.Run("Test New returns MetricsRepo instance", func(t *testing.T) {
		repo, err := New()
		require.NoError(t, err)
		require.IsType(t, &MetricsRepo{}, repo)
	})

	t.Run("Test New returns no error", func(t *testing.T) {
		_, err := New()
		require.NoError(t, err)
	})

	t.Run("Test New initializes MetricsRepo correctly", func(t *testing.T) {
		repo, err := New()
		require.NoError(t, err)
		require.Equal(t, &MetricsRepo{}, repo)
	})

	t.Run("Test New multiple calls return independent instances", func(t *testing.T) {
		repo1, err1 := New()
		repo2, err2 := New()
		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NotSame(t, repo1, repo2)
	})
}
