package decoder

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func generateTestPrivateKey(t *testing.T) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey
}

func saveTestPrivateKeyToFile(t *testing.T, privateKey *rsa.PrivateKey, path string) {
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	err = os.WriteFile(path, privPem, 0600)
	require.NoError(t, err)
}

func TestDecryptDataWithInvalidData(t *testing.T) {
	privateKey := generateTestPrivateKey(t)
	keyPath := "test_private_key.pem"
	saveTestPrivateKeyToFile(t, privateKey, keyPath)
	defer os.Remove(keyPath)

	decoder, err := New(keyPath)
	require.NoError(t, err)

	tests := []struct {
		name          string
		encryptedData []byte
		expectedError string
	}{
		{
			name:          "Empty data",
			encryptedData: []byte{},
			expectedError: "crypto/rsa: decryption error",
		},
		{
			name:          "Invalid data",
			encryptedData: []byte("invalid data"),
			expectedError: "crypto/rsa: decryption error",
		},
		{
			name:          "Corrupted encrypted data",
			encryptedData: []byte{0x00, 0x01, 0x02, 0x03},
			expectedError: "crypto/rsa: decryption error",
		},
		{
			name:          "Partially valid data",
			encryptedData: append([]byte{0x00, 0x01}, privateKey.PublicKey.N.Bytes()...),
			expectedError: "crypto/rsa: decryption error",
		},
		{
			name:          "Random bytes",
			encryptedData: make([]byte, 256),
			expectedError: "crypto/rsa: decryption error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decoder.DecryptData(tt.encryptedData)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedError)
		})
	}
}
