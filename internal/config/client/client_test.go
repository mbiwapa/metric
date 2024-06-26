package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	config "github.com/mbiwapa/metric/internal/config/client"
)

func TestMustLoadConfig_JSONConfig(t *testing.T) {
	// Create a temporary JSON config file
	configContent := `{
        "address": "127.0.0.1:9090",
        "report_interval": 15,
        "poll_interval": 5
    }`
	tmpFile, err := os.CreateTemp("", "config*.json")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Remove(tmpFile.Name()))
	}()

	_, err = tmpFile.Write([]byte(configContent))
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	err = os.Setenv("CONFIG", tmpFile.Name())
	require.NoError(t, err)

	// Load the configuration
	cfg, err := config.MustLoadConfig()
	require.NoError(t, err)

	// Validate the loaded configuration
	require.Equal(t, "http://127.0.0.1:9090", cfg.Addr)
	require.Equal(t, int64(15), cfg.ReportInterval)
	require.Equal(t, int64(5), cfg.PollInterval)
	//require.Equal(t, "/path/to/public/key", cfg.PublicKeyPath)
}
