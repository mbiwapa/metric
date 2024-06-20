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

	// Load the config
	config := config.MustLoadConfig()

	// Validate the loaded config
	require.Equal(t, "127.0.0.1:9090", config.Addr)
	require.Equal(t, int64(600), config.StoreInterval)
	require.Equal(t, "/tmp/test-metrics-db.json", config.StoragePath)
	require.False(t, config.Restore)
	require.Equal(t, "user:password@/dbname", config.DatabaseDSN)
}

func TestMustLoadConfig_JSONFile_OverrideEnv(t *testing.T) {
	// Create a temporary JSON config file
	configData := `{
        "address": "127.0.0.1:9090",
        "store_interval": 600,
        "store_file": "/tmp/test-metrics-db.json",
        "restore": false,
        "database_dsn": "user:password@/dbname",
    }`
	tmpFile, err := os.CreateTemp("", "config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(configData))
	require.NoError(t, err)
	tmpFile.Close()

	err = os.Setenv("CONFIG", tmpFile.Name())

	// Override some environment variables
	os.Setenv("ADDRESS", "localhost:8082")
	os.Setenv("STORE_INTERVAL", "1800")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/env-metrics-db.json")
	os.Setenv("RESTORE", "true")
	os.Setenv("DATABASE_DSN", "user:password@/env_dbname")

	// Load the config
	config := config.MustLoadConfig()

	// Validate the loaded config
	require.Equal(t, "localhost:8082", config.Addr)
	require.Equal(t, int64(1800), config.StoreInterval)
	require.Equal(t, "/tmp/env-metrics-db.json", config.StoragePath)
	require.True(t, config.Restore)
	require.Equal(t, "user:password@/env_dbname", config.DatabaseDSN)
}

func TestMustLoadConfig_JSONFile_InvalidPath(t *testing.T) {
	_ = os.Setenv("CONFIG", "invalid-path.json")

	// Load the config
	config := config.MustLoadConfig()

	// Validate the default config
	require.Equal(t, "localhost:8080", config.Addr)
	require.Equal(t, int64(300), config.StoreInterval)
	require.Equal(t, "/tmp/metrics-db.json", config.StoragePath)
	require.True(t, config.Restore)
	require.Equal(t, "", config.DatabaseDSN)
	require.Equal(t, "", config.PrivateKeyPath)
}

func TestMustLoadConfig_JSONFile_InvalidJSON(t *testing.T) {
	// Create a temporary invalid JSON config file
	tmpFile, err := os.CreateTemp("", "config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(`{invalid-json}`))
	require.NoError(t, err)
	tmpFile.Close()

	err = os.Setenv("CONFIG", tmpFile.Name())

	// Load the config
	config := config.MustLoadConfig()

	// Validate the default config
	require.Equal(t, "localhost:8080", config.Addr)
	require.Equal(t, int64(300), config.StoreInterval)
	require.Equal(t, "/tmp/metrics-db.json", config.StoragePath)
	require.True(t, config.Restore)
	require.Equal(t, "", config.DatabaseDSN)
	require.Equal(t, "", config.PrivateKeyPath)
}
