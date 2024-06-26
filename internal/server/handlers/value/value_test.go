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

	"github.com/mbiwapa/metric/internal/logger"
	"github.com/mbiwapa/metric/internal/server/handlers/value/mocks"
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
				MetricGeterMock.On("GetMetric", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(tt.wantBody, tt.mockError).
					Once()
			}

			logger, err := logger.New("info")
			if err != nil {
				panic("Logger initialization error: " + err.Error())
			}

			r := chi.NewRouter()
			r.Use(middleware.URLFormat)
			r.Get("/value/{type}/{name}", New(logger, MetricGeterMock, ""))
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

func ExampleNew() {
	logger, _ := logger.New("info")

	mockMetricGeter := &mocks.MetricGeter{}
	mockMetricGeter.On("GetMetric", mock.Anything, "gauge", "test1").Return("1", nil)

	r := chi.NewRouter()
	r.Use(middleware.URLFormat)
	r.Get("/value/{type}/{name}", New(logger, mockMetricGeter, ""))

	req, _ := http.NewRequest(http.MethodGet, "/value/gauge/test1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Output:
	// 200
	// 1
}
