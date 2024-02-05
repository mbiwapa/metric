package signature

import (
	"crypto/sha256"
	"encoding/hex"

	"go.uber.org/zap"
)

// GetHash function get hash from body and key
func GetHash(key string, body string, log *zap.Logger) string {
	log = log.With(
		zap.String("op", "lib.signature.GetHash"),
	)
	hash := sha256.Sum256([]byte(body + key))
	hashStr := hex.EncodeToString(hash[:])
	log.Info("Hash is generated", zap.String("hash", hashStr))
	return hashStr
}
