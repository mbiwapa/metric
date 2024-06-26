package update

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/logger"
	"github.com/mbiwapa/metric/internal/server/handlers/update/mocks"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
		mockError  error
		url        string
		httpMethod string
		typ        string
	}{
		{
			name:       "Gauge Тест 1, успешный ответ",
			wantStatus: http.StatusOK,
			mockError:  nil,
			url:        "/update/gauge/test1/0.5653",
			httpMethod: http.MethodPost,
			typ:        "gauge",
		},
		{
			name:       "Gauge Тест 2, неверный метод GET",
			wantStatus: http.StatusMethodNotAllowed,
			mockError:  nil,
			url:        "/update/gauge/test1/0.5653",
			httpMethod: http.MethodGet,
			typ:        "gauge",
		},
		{
			name:       "Gauge Тест 3, не работает хранилище",
			wantStatus: http.StatusBadRequest,
			mockError:  fmt.Errorf("Stor unavailable"),
			url:        "/update/gauge/test1/0.5653",
			httpMethod: http.MethodPost,
			typ:        "gauge",
		},
		{
			name:       "Gauge Тест 4, не передана метрика или ее значение",
			wantStatus: http.StatusNotFound,
			mockError:  nil,
			url:        "/update/gauge/test1",
			httpMethod: http.MethodPost,
			typ:        "gauge",
		},
		{
			name:       "Gauge Тест 5, передано неверное значение метрики",
			wantStatus: http.StatusBadRequest,
			mockError:  nil,
			url:        "/update/gauge/test1/test2",
			httpMethod: http.MethodPost,
			typ:        "gauge",
		},
		{
			name:       "Counter Тест 1, успешный ответ",
			wantStatus: http.StatusOK,
			mockError:  nil,
			url:        "/update/counter/testc/1",
			httpMethod: http.MethodPost,
			typ:        "counter",
		},
		{
			name:       "Counter Тест 2, неверный метод GET",
			wantStatus: http.StatusMethodNotAllowed,
			mockError:  nil,
			url:        "/update/counter/test1/0.5653",
			httpMethod: http.MethodGet,
			typ:        "counter",
		},
		{
			name:       "Counter Тест 3, не работает хранилище",
			wantStatus: http.StatusBadRequest,
			mockError:  fmt.Errorf("Stor unavailable"),
			url:        "/update/counter/test1/1",
			httpMethod: http.MethodPost,
			typ:        "counter",
		},
		{
			name:       "Counter Тест 4, не передана метрика или ее значение",
			wantStatus: http.StatusNotFound,
			mockError:  nil,
			url:        "/update/counter",
			httpMethod: http.MethodPost,
			typ:        "counter",
		},
		{
			name:       "Counter Тест 5, передано неверное значение метрики",
			wantStatus: http.StatusBadRequest,
			mockError:  nil,
			url:        "/update/counter/test1/test2",
			httpMethod: http.MethodPost,
			typ:        "counter",
		},
		{
			name:       "Counter Тест 6, update_invalid_type",
			wantStatus: http.StatusBadRequest,
			mockError:  nil,
			url:        "/update/unknown/testCounter/100",
			httpMethod: http.MethodPost,
			typ:        "counter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			UpdaterMock := mocks.NewUpdater(t)

			BackuperMock := mocks.NewBackuper(t)

			if tt.wantStatus == http.StatusOK || tt.mockError != nil {
				if tt.typ == "gauge" {
					UpdaterMock.On("UpdateGauge", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("float64")).
						Return(tt.mockError).
						Once()
				}
				if tt.typ == "counter" {
					UpdaterMock.On("UpdateCounter", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int64")).
						Return(tt.mockError).
						Once()
				}
			}
			if tt.wantStatus == http.StatusOK {
				BackuperMock.On("IsSyncMode").Return(false).Once()
			}

			logger, err := logger.New("info")
			if err != nil {
				panic("Logger initialization error: " + err.Error())
			}

			r := chi.NewRouter()
			r.Use(middleware.URLFormat)
			r.Post("/update/{type}/{name}/{value}", New(logger, UpdaterMock, BackuperMock))
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
	log, _ := zap.NewProduction()
	defer log.Sync()

	mockUpdater := &mocks.Updater{}
	mockBackuper := &mocks.Backuper{}

	// Set up the mock expectation for UpdateGauge
	mockUpdater.On("UpdateGauge", mock.Anything, "test1", 0.5653).Return(nil)
	// Set up the mock expectation for IsSyncMode
	mockBackuper.On("IsSyncMode").Return(false)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", New(log, mockUpdater, mockBackuper))

	req, _ := http.NewRequest(http.MethodPost, "/update/gauge/test1/0.5653", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)
	fmt.Println(rr.Code)

	// Output:
	// 200
}
