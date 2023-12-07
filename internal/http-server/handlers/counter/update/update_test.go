package update

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
			url:        "/update/counter/test1/0.5653",
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
			url:        "/update/counter/test1/0.5653",
			httpMethod: http.MethodPost,
		},
		{
			name:       "Counter Тест 4, не передана метрика или ее значение",
			wantStatus: http.StatusNotFound,
			mockError:  nil,
			url:        "/update/counter/test1",
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
				CounterUpdaterMock.On("GaugeUpdate", mock.AnythingOfType("string"), mock.AnythingOfType("float64")).
					Return(tt.mockError).
					Once()
			}

			handler := New(CounterUpdaterMock)

			req, err := http.NewRequest(tt.httpMethod, tt.url, nil)

			require.NoError(t, err)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, tt.wantStatus)
		})
	}
}
