package update

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mbiwapa/metric/internal/http-server/handlers/counter/update/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
		mockError  error
		url        string
		httpMethod string
	}{
		{
			name:       "Counter Тест 1, успешный ответ",
			wantStatus: http.StatusOK,
			mockError:  nil,
			url:        "/update/counter/testc/1",
			httpMethod: http.MethodPost,
		},
		{
			name:       "Counter Тест 2, неверный метод GET",
			wantStatus: http.StatusMethodNotAllowed,
			mockError:  nil,
			url:        "/update/counter/test1/0.5653",
			httpMethod: http.MethodGet,
		},
		{
			name:       "Counter Тест 3, не работает хранилище",
			wantStatus: http.StatusBadRequest,
			mockError:  fmt.Errorf("Stor unavailable"),
			url:        "/update/counter/test1/1",
			httpMethod: http.MethodPost,
		},
		{
			name:       "Counter Тест 4, не передана метрика или ее значение",
			wantStatus: http.StatusNotFound,
			mockError:  nil,
			url:        "/update/counter",
			httpMethod: http.MethodPost,
		},
		{
			name:       "Counter Тест 5, передано неверное значение метрики",
			wantStatus: http.StatusBadRequest,
			mockError:  nil,
			url:        "/update/counter/test1/test2",
			httpMethod: http.MethodPost,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			CounterUpdaterMock := mocks.NewCounterUpdater(t)

			if tt.wantStatus == http.StatusOK || tt.mockError != nil {
				CounterUpdaterMock.On("CounterUpdate", mock.AnythingOfType("string"), mock.AnythingOfType("int64")).
					Return(tt.mockError).
					Once()
			}

			r := chi.NewRouter()
			r.Use(middleware.URLFormat)
			r.Post("/update/counter/{name}/{value}", New(CounterUpdaterMock))
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
