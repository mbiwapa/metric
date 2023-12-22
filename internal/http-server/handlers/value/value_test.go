package value

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mbiwapa/metric/internal/http-server/handlers/value/mocks"
	"github.com/mbiwapa/metric/internal/logger"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
		mockError  error
		url        string
		httpMethod string
		wantBody   string
	}{
		{
			name:       "Value Тест 1, успешный ответ",
			wantStatus: http.StatusOK,
			mockError:  nil,
			url:        "/value/gauge/test1",
			httpMethod: http.MethodGet,
			wantBody:   "1",
		},
		{
			name:       "Value Тест 2, неверный метод POST",
			wantStatus: http.StatusMethodNotAllowed,
			mockError:  nil,
			url:        "/value/gauge/test1",
			httpMethod: http.MethodPost,
		},
		{
			name:       "Value Тест 3, не передана метрика или ее тип",
			wantStatus: http.StatusNotFound,
			mockError:  nil,
			url:        "/value/gauge",
			httpMethod: http.MethodGet,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			MetricGeterMock := mocks.NewMetricGeter(t)

			if tt.wantStatus == http.StatusOK || tt.mockError != nil {
				MetricGeterMock.On("GetMetric", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(tt.wantBody, tt.mockError).
					Once()
			}

			logger, err := logger.New("info")
			if err != nil {
				fmt.Errorf(err.Error())
			}

			r := chi.NewRouter()
			r.Use(middleware.URLFormat)
			r.Get("/value/{type}/{name}", New(logger, MetricGeterMock))
			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(tt.httpMethod, ts.URL+tt.url, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			require.Equal(t, resp.StatusCode, tt.wantStatus)
		})
	}
}
