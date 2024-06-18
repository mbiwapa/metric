// Package encoder provides functionality for RSA encryption using a public key.
package encoder

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// Encoder is a struct that holds an RSA public key for encryption.
type Encoder struct {
	publicKey *rsa.PublicKey
	disable   bool
}

// New loads an RSA public key from a file and returns an Encoder instance.
// The path parameter specifies the file path to the PEM-encoded public key.
// It returns an error if the key cannot be read or parsed.
func New(path string) (*Encoder, error) {
	if path == "" {
		return &Encoder{disable: true}, nil
	}
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return &Encoder{publicKey: publicKey, disable: false}, nil
}

// EncryptData encrypts the given data using the RSA public key and OAEP padding.
// It returns the encrypted data or an error if the encryption fails.
func (e *Encoder) EncryptData(data []byte) ([]byte, error) {
	if e.disable {
		return data, nil
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, e.publicKey, data, nil)
}
