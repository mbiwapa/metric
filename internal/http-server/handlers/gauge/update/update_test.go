package update

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mbiwapa/metric/internal/http-server/handlers/gauge/update/mocks"
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
			name:       "Gauge Тест 1, успешный ответ",
			wantStatus: http.StatusOK,
			mockError:  nil,
			url:        "/update/gauge/test1/0.5653",
			httpMethod: http.MethodPost,
		},
		{
			name:       "Gauge Тест 2, неверный метод GET",
			wantStatus: http.StatusMethodNotAllowed,
			mockError:  nil,
			url:        "/update/gauge/test1/0.5653",
			httpMethod: http.MethodGet,
		},
		{
			name:       "Gauge Тест 3, не работает хранилище",
			wantStatus: http.StatusBadRequest,
			mockError:  fmt.Errorf("Stor unavailable"),
			url:        "/update/gauge/test1/0.5653",
			httpMethod: http.MethodPost,
		},
		{
			name:       "Gauge Тест 4, не передана метрика или ее значение",
			wantStatus: http.StatusNotFound,
			mockError:  nil,
			url:        "/update/gauge/test1",
			httpMethod: http.MethodPost,
		},
		{
			name:       "Gauge Тест 5, передано неверное значение метрики",
			wantStatus: http.StatusBadRequest,
			mockError:  nil,
			url:        "/update/gauge/test1/test2",
			httpMethod: http.MethodPost,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			GaugeUpdaterMock := mocks.NewGaugeUpdater(t)

			if tt.wantStatus == http.StatusOK || tt.mockError != nil {
				GaugeUpdaterMock.On("GaugeUpdate", mock.AnythingOfType("string"), mock.AnythingOfType("float64")).
					Return(tt.mockError).
					Once()
			}

			handler := New(GaugeUpdaterMock)

			req, err := http.NewRequest(tt.httpMethod, tt.url, nil)

			require.NoError(t, err)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, tt.wantStatus)
		})
	}
}
