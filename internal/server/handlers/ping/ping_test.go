package ping

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/server/handlers/ping/mocks"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		pingReturn error
		wantStatus int
		requestID  string
	}{
		{
			name:       "Success",
			pingReturn: nil,
			wantStatus: http.StatusOK,
			requestID:  "test-request-id",
		},
		{
			name:       "WithoutRequestID",
			pingReturn: nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "WithRequestID",
			pingReturn: nil,
			wantStatus: http.StatusOK,
			requestID:  "test-request-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewProduction()
			defer logger.Sync()

			mockPinger := mocks.NewPinger(t)
			mockPinger.On("Ping", mock.Anything).Return(tt.pingReturn)

			handler := New(logger, mockPinger)

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			require.NoError(t, err)

			if tt.requestID != "" {
				req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, tt.requestID))
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
			mockPinger.AssertExpectations(t)
		})
	}
}

// Mock implementation of the Pinger interface for testing purposes.
type MockPinger struct{}

func (m *MockPinger) Ping(ctx context.Context) error {
	return nil
}

func ExampleNew() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	mockPinger := &MockPinger{}

	handler := New(logger, mockPinger)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Content-Type"))
	fmt.Println(rr.Body.String())

	// Output:
	// 200
}
