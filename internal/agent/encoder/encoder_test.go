package encoder

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func generateTestPublicKey(t *testing.T) string {
	// Generate a new RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Extract the public key
	publicKey := &privateKey.PublicKey

	// Encode the public key to PEM format
	pubASN1, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	// Write the public key to a temporary file
	tmpFile, err := os.CreateTemp("", "public_key_*.pem")
	require.NoError(t, err)
	defer tmpFile.Close()

	_, err = tmpFile.Write(pubPEM)
	require.NoError(t, err)

	return tmpFile.Name()
}

func TestEncryptData_EmptyData(t *testing.T) {
	// Generate a test public key and get the file path
	pubKeyPath := generateTestPublicKey(t)
	defer os.Remove(pubKeyPath)

	// Initialize the Encoder with the test public key
	encoder, err := New(pubKeyPath)
	require.NoError(t, err)
	require.NotNil(t, encoder)

	// Test encrypting an empty data slice
	encryptedData, err := encoder.EncryptData([]byte{})
	require.NoError(t, err)
	require.NotNil(t, encryptedData)
	require.NotEmpty(t, encryptedData)
}

func TestNew_InvalidPath(t *testing.T) {
	_, err := New("invalid_path.pem")
	require.Error(t, err)
}

func TestNew_InvalidPEM(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid_pem_*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte("invalid pem data"))
	require.NoError(t, err)

	_, err = New(tmpFile.Name())
	require.Error(t, err)
}

func TestNew_NotRSAPublicKey(t *testing.T) {
	// Generate a new ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Extract the public key
	publicKey := &privateKey.PublicKey

	// Encode the public key to PEM format
	pubASN1, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	// Write the public key to a temporary file
	tmpFile, err := os.CreateTemp("", "public_key_*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(pubPEM)
	require.NoError(t, err)

	_, err = New(tmpFile.Name())
	require.Error(t, err)
}
