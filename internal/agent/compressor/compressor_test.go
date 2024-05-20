package compressor

import (
	"testing"

	"go.uber.org/zap"
)

func BenchmarkGetCompressedData(b *testing.B) {
	compressor := New(zap.NewNop())
	data := []byte("This is a test data for compression.")

	// Benchmark for GetCompressedData function
	b.Run("GetCompressedData", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = compressor.GetCompressedData(data)
		}
	})
}
