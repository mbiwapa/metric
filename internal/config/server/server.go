package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds all the server configurations.
type Config struct {
	Addr          string // Server address and port
	StoreInterval int64  // Interval in seconds to save current server metrics to disk
	StoragePath   string // Full path to the file where current values are saved
	Restore       bool   // Whether to load previously saved values from the specified file at server startup
	DatabaseDSN   string // DSN string for connecting to the database
	Key           string // Key for hash computation
}

// MustLoadConfig loads the configuration from command-line flags and environment variables.
// It returns a pointer to the Config struct populated with the loaded values.
func MustLoadConfig() *Config {
	var config Config

	// Define command-line flags and their default values
	flag.StringVar(&config.Addr, "a", "localhost:8080", "Адрес порт сервера")
	flag.Int64Var(&config.StoreInterval, "i", 300, "Интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&config.StoragePath, "f", "/tmp/metrics-db.json", "Полное имя файла, куда сохраняются текущие значения")
	flag.BoolVar(&config.Restore, "r", true, "Загружать или нет ранее сохранённые значения из указанного файла при старте сервера")
	flag.StringVar(&config.DatabaseDSN, "d", "", "DSN строка для соединения с базой данных") //user=postgres password=postgres host=localhost port=5432 database=postgres sslmode=disable
	flag.StringVar(&config.Key, "k", "", "Ключ для вычисления хеша")
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

	return &config
}
