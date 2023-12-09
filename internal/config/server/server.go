package config

import (
	"flag"
)

// Config Структура со всеми конфигурациями сервера
type Config struct {
	Addr string
}

// MustLoadConfig загрузка конфигурации
func MustLoadConfig() *Config {
	var config Config
	flag.StringVar(&config.Addr, "a", "localhost:8080", "Адрес порт сервера")
	flag.Parse()

	return &config
}
