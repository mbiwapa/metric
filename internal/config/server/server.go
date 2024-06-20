// Package config provides functionality for loading server configurations
// from command-line flags, environment variables, and a JSON file.
package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
)

// Config holds all the server configurations.
type Config struct {
	Addr           string `json:"address,omitempty"`        // Addr Server address and port
	StoreInterval  int64  `json:"store_interval,omitempty"` // StoreInterval Interval in seconds to save current server metrics to disk
	StoragePath    string `json:"store_file,omitempty"`     // StoragePath Full path to the file where current values are saved
	Restore        bool   `json:"restore,omitempty"`        // Restore Whether to load previously saved values from the specified file at server startup
	DatabaseDSN    string `json:"database_dsn,omitempty"`   // DatabaseDSN DSN string for connecting to the database
	Key            string // Key for hash computation
	PrivateKeyPath string `json:"crypto_key,omitempty"` // PrivateKeyPath to the private key file
}

// MustLoadConfig loads the configuration from command-line flags, environment variables, and a JSON file.
// It returns a pointer to the Config struct populated with the loaded values.
func MustLoadConfig() *Config {
	var config Config
	var configFilePath string

	// Define command-line flags and their default values
	flag.StringVar(&config.Addr, "a", "localhost:8080", "Адрес порт сервера")
	flag.Int64Var(&config.StoreInterval, "i", 300, "Интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&config.StoragePath, "f", "/tmp/metrics-db.json", "Полное имя файла, куда сохраняются текущие значения")
	flag.BoolVar(&config.Restore, "r", true, "Загружать или нет ранее сохранённые значения из указанного файла при старте сервера")
	flag.StringVar(&config.DatabaseDSN, "d", "", "DSN строка для соединения с базой данных")
	flag.StringVar(&config.Key, "k", "", "Ключ для вычисления хеша")
	flag.StringVar(&config.PrivateKeyPath, "crypto-key", "", "Путь к файлу с закрытым ключом")
	flag.StringVar(&configFilePath, "c", "", "Путь к файлу конфигурации")
	flag.StringVar(&configFilePath, "config", "", "Путь к файлу конфигурации")
	flag.Parse()

	// Override with environment variables if they are set
	envAddr := os.Getenv("ADDRESS")
	if envAddr != "" {
		config.Addr = envAddr
	}

	storeInterval := os.Getenv("STORE_INTERVAL")
	if storeInterval != "" {
		i, _ := strconv.ParseInt(storeInterval, 10, 64)
		config.StoreInterval = i
	}

	filePath := os.Getenv("FILE_STORAGE_PATH")
	if filePath != "" {
		config.StoragePath = filePath
	}

	restore := os.Getenv("RESTORE")
	if restore != "" {
		b, _ := strconv.ParseBool(restore)
		config.Restore = b
	}

	databaseDSN := os.Getenv("DATABASE_DSN")
	if databaseDSN != "" {
		config.DatabaseDSN = databaseDSN
	}

	envKey := os.Getenv("KEY")
	if envKey != "" {
		config.Key = envKey
	}

	envPrivateKeyPath := os.Getenv("CRYPTO_KEY")
	if envPrivateKeyPath != "" {
		config.PrivateKeyPath = envPrivateKeyPath
	}

	if _, err := os.Stat(config.PrivateKeyPath); os.IsNotExist(err) && config.PrivateKeyPath != "" {
		os.Exit(5)
	}

	envConfigFilePath := os.Getenv("CONFIG")
	if envConfigFilePath != "" {
		configFilePath = envConfigFilePath
	}

	// Load configuration from JSON file if specified
	if configFilePath != "" {
		file, err := os.Open(configFilePath)
		if err == nil {
			decoder := json.NewDecoder(file)
			fileConfig := Config{}
			if errDecode := decoder.Decode(&fileConfig); errDecode == nil {
				if config.Addr == "localhost:8080" && fileConfig.Addr != "" {
					config.Addr = fileConfig.Addr
				}
				if config.StoreInterval == 300 && fileConfig.StoreInterval != 0 {
					config.StoreInterval = fileConfig.StoreInterval
				}
				if config.StoragePath == "/tmp/metrics-db.json" && fileConfig.StoragePath != "" {
					config.StoragePath = fileConfig.StoragePath
				}
				if config.Restore {
					config.Restore = fileConfig.Restore
				}
				if config.DatabaseDSN == "" {
					config.DatabaseDSN = fileConfig.DatabaseDSN
				}
				if config.PrivateKeyPath == "" {
					config.PrivateKeyPath = fileConfig.PrivateKeyPath
				}
			}
			_ = file.Close()
		}
	}

	return &config
}
