package config

import (
	"flag"
	"os"
	"strconv"
)

// Config Структура со всеми конфигурациями сервера
type Config struct {
	Addr          string
	StoreInterval int64
	StoragePath   string
	Restore       bool
}

// MustLoadConfig загрузка конфигурации
func MustLoadConfig() *Config {
	var config Config
	flag.StringVar(&config.Addr, "a", "localhost:8080", "Адрес порт сервера")
	flag.Int64Var(&config.StoreInterval, "i", 300, "Интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&config.StoragePath, "f", "/tmp/metrics-db.json", "Полное имя файла, куда сохраняются текущие значения")
	flag.BoolVar(&config.Restore, "r", true, "Загружать или нет ранее сохранённые значения из указанного файла при старте сервера")
	flag.Parse()

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

	return &config
}
