package signature

import (
	"testing"

	"go.uber.org/zap"
)

func BenchmarkGetHash(b *testing.B) {
	key := "test_key"
	body := "test_body"
	log := zap.NewNop() // Replace with your logger instance

	// Benchmark for GetHash function
	b.Run("GetHash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GetHash(key, body, log)
		}
	})
}
