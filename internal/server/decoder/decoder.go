// Package decoder provides functionality for RSA decryption using a private key.
package decoder

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// Decoder is a struct that holds an RSA private key for decryption.
type Decoder struct {
	privateKey *rsa.PrivateKey
	disable    bool
}

// New loads an RSA private key from a file and returns a Decoder instance.
// The path parameter specifies the file path to the PEM-encoded private key.
// It returns an error if the key cannot be read or parsed.
func New(path string) (*Decoder, error) {
	if path == "" {
		return &Decoder{disable: true}, nil
	}
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	privateKey, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return &Decoder{privateKey: privateKey, disable: false}, nil
}

// DecryptData decrypts the given encrypted data using the RSA private key and OAEP padding.
// It returns the decrypted data or an error if the decryption fails.
func (d *Decoder) DecryptData(encryptedData []byte) ([]byte, error) {
	if d.disable {
		return encryptedData, nil
	}
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, d.privateKey, encryptedData, nil)
}
