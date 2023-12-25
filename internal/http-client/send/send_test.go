package send

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_Send(t *testing.T) {

	type args struct {
		typ   string
		name  string
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Clietn Тест 1 - успешный тест",
			args: args{
				typ:   "gauge",
				name:  "test2",
				value: "0.5686",
			},
			wantErr: false,
		},
		{
			name: "Clietn Тест 2 - 404 от сервера",
			args: args{
				typ:   "counter",
				name:  "test2",
				value: "0.5676",
			},
			wantErr: true,
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

			c, err := New(srv.URL)

			require.NoError(t, err)

			err = c.Send(tt.args.typ, tt.args.name, tt.args.value)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}
