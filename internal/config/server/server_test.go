package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	config "github.com/mbiwapa/metric/internal/config/server"
)

func TestMustLoadConfig_JSONFile(t *testing.T) {
	// Create a temporary JSON config file
	configData := `{
        "address": "127.0.0.1:9090",
        "store_interval": 600,
        "store_file": "/tmp/test-metrics-db.json",
        "restore": false,
        "database_dsn": "user:password@/dbname"
    }`
	tmpFile, err := os.CreateTemp("", "config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(configData))
	require.NoError(t, err)
	tmpFile.Close()

	err = os.Setenv("CONFIG", tmpFile.Name())
	require.NoError(t, err)

	// Load the config
	config := config.MustLoadConfig()

	// Validate the loaded config
	require.Equal(t, "127.0.0.1:9090", config.Addr)
	require.Equal(t, int64(600), config.StoreInterval)
	require.Equal(t, "/tmp/test-metrics-db.json", config.StoragePath)
	require.False(t, config.Restore)
	require.Equal(t, "user:password@/dbname", config.DatabaseDSN)
}
