package home

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

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
