package gopsutilsource

import (
	"testing"
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
