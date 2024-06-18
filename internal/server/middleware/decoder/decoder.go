// Package decoder provides middleware for decrypting HTTP request bodies.
package decoder

import (
	"bytes"
	"io"
	"net/http"
)

// Decoder is an interface that defines a method for decrypting data.
type Decoder interface {
	// DecryptData takes encrypted data as input and returns the decrypted data or an error.
	DecryptData(encryptedData []byte) ([]byte, error)
}

// New creates a new HTTP middleware that decrypts the request body using the provided Decoder.
// It returns a function that takes an http.Handler and returns an http.Handler.
//
// The middleware reads the encrypted request body, decrypts it using the provided Decoder,
// and replaces the request body with the decrypted data before passing the request to the next handler.
//
// If reading the request body or decrypting the data fails, it responds with an appropriate HTTP error.
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
