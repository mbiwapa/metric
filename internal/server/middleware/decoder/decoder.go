package decoder

import (
	"bytes"
	"io"
	"net/http"
)

type Decoder interface {
	DecryptData(encryptedData []byte) ([]byte, error)
}

func New(decoder Decoder) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			encryptedData, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}

			decryptedData, err := decoder.DecryptData(encryptedData)
			if err != nil {
				http.Error(w, "failed to decrypt request body", http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decryptedData))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
