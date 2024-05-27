// Package signature provides a function to generate a SHA-256 hash from the concatenation of the provided key and body strings.
// It also logs the operation and the generated hash using the provided zap.Logger.
package signature

import (
	"crypto/sha256"
	"encoding/hex"

	"go.uber.org/zap"
)

// GetHash generates a SHA-256 hash from the concatenation of the provided key and body strings.
// It logs the operation and the generated hash using the provided zap.Logger.
//
// Parameters:
//   - key: A string that will be concatenated with the body to form the input for the hash function.
//   - body: A string that will be concatenated with the key to form the input for the hash function.
//   - log: A pointer to a zap.Logger used for logging the operation and the generated hash.
//
// Returns:
//   - A string representing the hexadecimal encoding of the SHA-256 hash of the concatenated key and body.
func GetHash(key string, body string, log *zap.Logger) string {
	log = log.With(
		zap.String("op", "lib.signature.GetHash"),
	)
	hash := sha256.Sum256([]byte(body + key))
	hashStr := hex.EncodeToString(hash[:])
	log.Info("Hash is generated", zap.String("hash", hashStr))
	return hashStr
}
