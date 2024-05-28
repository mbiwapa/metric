package updates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/server/handlers/updates/mocks"

	"github.com/mbiwapa/metric/internal/lib/api/format"
	"github.com/mbiwapa/metric/internal/lib/signature"
)

func TestNewJSON_SuccessfulEncodingAndSigning(t *testing.T) {
	tests := []struct {
		name       string
		metrics    []format.Metric
		sha256key  string
		wantStatus int
	}{
		{
			name: "Successful encoding and signing with SHA256 key",
			metrics: []format.Metric{
				{
					ID:    "testGauge",
					MType: format.Gauge,
					Value: float64Ptr(0.5653),
				},
				{
					ID:    "testCounter",
					MType: format.Counter,
					Delta: int64Ptr(10),
				},
			},
			sha256key:  "testkey",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdaterMock := mocks.NewUpdater(t)
			BackuperMock := mocks.NewBackuper(t)

			UpdaterMock.On("UpdateBatch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			BackuperMock.On("IsSyncMode").Return(false)

			logger, err := zap.NewProduction()
			require.NoError(t, err)

			r := chi.NewRouter()
			r.Use(middleware.RequestID)
			r.Post("/updates", NewJSON(logger, UpdaterMock, BackuperMock, tt.sha256key))

			ts := httptest.NewServer(r)
			defer ts.Close()

			body, err := json.Marshal(tt.metrics)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, ts.URL+"/updates", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantStatus == http.StatusOK {
				var responseMetrics []format.Metric
				err = json.NewDecoder(resp.Body).Decode(&responseMetrics)
				require.NoError(t, err)
				require.Equal(t, tt.metrics, responseMetrics)

				hashStr := resp.Header.Get("HashSHA256")
				expectedHash := signature.GetHash(tt.sha256key, string(body), logger)
				require.Equal(t, expectedHash, hashStr)
			}
		})
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func ExampleNewJSON() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	mockUpdater := &mocks.Updater{}
	mockBackuper := &mocks.Backuper{}

	// Set up the mock expectation for UpdateBatch
	mockUpdater.On("UpdateBatch", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// Set up the mock expectation for IsSyncMode
	mockBackuper.On("IsSyncMode").Return(false)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Post("/updates", NewJSON(logger, mockUpdater, mockBackuper, "testkey"))

	reqBody := []format.Metric{
		{
			ID:    "testGauge",
			MType: format.Gauge,
			Value: float64Ptr(0.5653),
		},
		{
			ID:    "testCounter",
			MType: format.Counter,
			Delta: int64Ptr(10),
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Content-Type"))
	fmt.Println(rr.Header().Get("HashSHA256"))
	fmt.Println(rr.Body.String())

	// Output:
	//200
	//application/json
	//1bd2e889d00afc2fcf7b3dba6f9426ae97e29db3fff69a78011b5baf5325d125
	//[{"id":"testGauge","type":"gauge","value":0.5653},{"id":"testCounter","type":"counter","delta":10}]
}
