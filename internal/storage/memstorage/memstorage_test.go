package memstorage

import (
	"context"
	"testing"

	"github.com/mbiwapa/metric/internal/lib/api/format"
)

func BenchmarkGetMetricMemory(b *testing.B) {
	storage, _ := New()

	_ = storage.UpdateGauge(context.Background(), "test_gauge", 1.0)
	_ = storage.UpdateCounter(context.Background(), "test_counter", 1)

	// Benchmark for Gauge type metric
	b.Run("Gauge", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = storage.GetMetric(context.Background(), format.Gauge, "test_gauge")
		}
	})

	// Benchmark for Counter type metric
	b.Run("Counter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = storage.GetMetric(context.Background(), format.Counter, "test_counter")
		}
	})
}
