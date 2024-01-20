package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mbiwapa/metric/internal/logger"
)

func TestClient_Send(t *testing.T) {

	type args struct {
		gauge   [][]string
		counter [][]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Clietn Тест 1 - успешный тест",
			args: args{
				gauge:   [][]string{{"test", "0.567"}},
				counter: [][]string{{"test2", "1"}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				if tt.wantErr {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer srv.Close()

			logger, err := logger.New("info")
			if err != nil {
				panic("Logger initialization error: " + err.Error())
			}

			c, err := New(srv.URL, logger)

			require.NoError(t, err)

			err = c.Send(tt.args.gauge, tt.args.counter)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}
