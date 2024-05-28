package home

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/logger"
	"github.com/mbiwapa/metric/internal/server/handlers/home/mocks"
)

func TestNew(t *testing.T) {

	metrics := make([][]string, 0, 1)
	metric := []string{"test", "1.56"}
	metrics = append(metrics, metric)

	tests := []struct {
		name        string
		wantStatus  int
		mockError   error
		httpMethod  string
		wantMetrics [][]string
	}{
		{
			name:        "Home Тест 1, успешный ответ",
			wantStatus:  http.StatusOK,
			mockError:   nil,
			httpMethod:  http.MethodGet,
			wantMetrics: metrics,
		},
		{
			name:        "Home Тест 2, хранилище не отвечает",
			wantStatus:  http.StatusBadRequest,
			mockError:   fmt.Errorf("Stor unavailable"),
			httpMethod:  http.MethodGet,
			wantMetrics: metrics,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			AllMetricGeterMock := mocks.NewAllMetricGeter(t)

			if tt.wantStatus == http.StatusOK || tt.mockError != nil {
				AllMetricGeterMock.On("GetAllMetrics", mock.Anything).
					Return(tt.wantMetrics, tt.wantMetrics, tt.mockError).
					Once()
			}

			logger, err := logger.New("info")
			if err != nil {
				panic("Logger initialization error: " + err.Error())
			}

			r := chi.NewRouter()
			r.Use(middleware.URLFormat)
			r.Get("/", New(logger, AllMetricGeterMock, ""))
			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(tt.httpMethod, ts.URL, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			require.Equal(t, resp.StatusCode, tt.wantStatus)
		})
	}
}

// Mock implementation of the AllMetricGeter interface for testing purposes.
type MockMetricStorage struct{}

func (m *MockMetricStorage) GetAllMetrics(ctx context.Context) ([][]string, [][]string, error) {
	gaugeMetrics := [][]string{
		{"metric1", "1.23"},
		{"metric2", "4.56"},
	}
	counterMetrics := [][]string{
		{"metric3", "789"},
	}
	return gaugeMetrics, counterMetrics, nil
}

func ExampleNew() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	storage := &MockMetricStorage{}
	sha256key := "exampleSHA256Key"

	handler := New(logger, storage, sha256key)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Content-Type"))
	fmt.Println(rr.Header().Get("HashSHA256"))
	fmt.Println(rr.Body.String())

	// Output:
	//200
	//text/html
	//bebb681d299074ccd44c0023aaa8d4614bbd7ed8a426092f8da6f92c58e52f6a
	//<!DOCTYPE html><html><head><title>Метрики</title><body><h1>Метрики</h1><ul><li>metric1: 1.23</li><li>metric2: 4.56</li><li>metric3: 789</li></ul></body></html>
}
