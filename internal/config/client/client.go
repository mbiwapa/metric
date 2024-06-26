// Package config provides functionality for loading server configurations
// from command-line flags, environment variables, and a JSON file.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config holds all the server configurations.
type Config struct {
	Addr           string `json:"address,omitempty"`         // Server address and port for metric collection
	ReportInterval int64  `json:"report_interval,omitempty"` // Frequency of sending metrics to the server (in seconds)
	PollInterval   int64  `json:"poll_interval,omitempty"`   // Frequency of polling metrics from the source (in seconds)
	Key            string `json:"key,omitempty"`             // Key for hash computation
	WorkerCount    int    `json:"worker_count,omitempty"`    // Number of threads for sending metrics
	PublicKeyPath  string `json:"crypto_key,omitempty"`      // Path to the public key file
}

// MustLoadConfig loads the configuration from command-line flags, environment variables, and a JSON file.
// It returns a pointer to a Config struct and an error if any invalid values are encountered.
//
// Returns:
//   - *Config: A pointer to the loaded configuration struct.
//   - error: An error if any invalid values are encountered.
func MustLoadConfig() (*Config, error) {
	var Addr string
	var PollInterval int64
	var ReportInterval int64
	var Key string
	var err error
	var WorkerCount int
	var PublicKeyPath string
	var configFilePath string

	// Define command-line flags
	flag.StringVar(&Addr, "a", "localhost:8080", "Адрес  и порт сервера по сбору метрик")
	flag.Int64Var(&ReportInterval, "r", 10, "Частота отправки метрик на сервер (по умолчанию 10 секунд)")
	flag.Int64Var(&PollInterval, "p", 2, "Частота опроса метрик из источника (по умолчанию 2 секунды)")
	flag.StringVar(&Key, "k", "", "Ключ для вычисления хеша")
	flag.IntVar(&WorkerCount, "l", 1, "Количество потоков для отправки метрик (по умолчанию 1 поток)")
	flag.StringVar(&PublicKeyPath, "crypto-key", "", "Путь к файлу с публичным ключом")
	flag.StringVar(&configFilePath, "c", "", "Путь к файлу конфигурации")
	flag.StringVar(&configFilePath, "config", "", "Путь к файлу конфигурации")
	flag.Parse()

	// Override with environment variables if they are set
	envAddr := os.Getenv("ADDRESS")
	envPollInterval := os.Getenv("REPORT_INTERVAL")
	envReportInterval := os.Getenv("POLL_INTERVAL")
	envWorkerCount := os.Getenv("RATE_LIMIT")
	envKey := os.Getenv("KEY")
	envPublicKeyPath := os.Getenv("CRYPTO_KEY")
	envConfigFilePath := os.Getenv("CONFIG")

	if envAddr != "" {
		Addr = envAddr
	}
	if envPollInterval != "" {
		PollInterval, err = strconv.ParseInt(envPollInterval, 10, 64)
		if err != nil && envPollInterval != "" {
			return nil, fmt.Errorf("invalid env value: %s. %s", envPollInterval, err)
		}
	}
	if envReportInterval != "" {
		ReportInterval, err = strconv.ParseInt(envReportInterval, 10, 64)
		if err != nil && envReportInterval != "" {
			return nil, fmt.Errorf("invalid env value: %s. %s", envReportInterval, err)
		}
	}
	if envKey != "" {
		Key = envKey
	}
	if envWorkerCount != "" {
		WorkerCount, err = strconv.Atoi(envWorkerCount)
		if err != nil && envWorkerCount != "" {
			return nil, fmt.Errorf("invalid env value: %s. %s", envWorkerCount, err)
		}
	}
	if envPublicKeyPath != "" {
		PublicKeyPath = envPublicKeyPath
	}
	if envConfigFilePath != "" {
		configFilePath = envConfigFilePath
	}

	// Load configuration from JSON file if specified
	if configFilePath != "" {
		file, errOpen := os.Open(configFilePath)
		if errOpen == nil {
			decoder := json.NewDecoder(file)
			fileConfig := Config{}
			errDecode := decoder.Decode(&fileConfig)
			if errDecode == nil {
				if Addr == "localhost:8080" && fileConfig.Addr != "" {
					Addr = fileConfig.Addr
				}
				if ReportInterval == 10 && fileConfig.ReportInterval != 0 {
					ReportInterval = fileConfig.ReportInterval
				}
				if PollInterval == 2 && fileConfig.PollInterval != 0 {
					PollInterval = fileConfig.PollInterval
				}
				if PublicKeyPath == "" && fileConfig.PublicKeyPath != "" {
					PublicKeyPath = fileConfig.PublicKeyPath
				}
			}
			fmt.Println(errDecode)
			_ = file.Close()
		} else {
			return nil, errOpen
		}
	}

	if _, err = os.Stat(PublicKeyPath); os.IsNotExist(err) && PublicKeyPath != "" {
		return nil, fmt.Errorf("file not found: %s. %s", PublicKeyPath, err)
	}

	// Create the configuration struct
	cfg := &Config{
		Addr:           "http://" + Addr,
		PollInterval:   PollInterval,
		ReportInterval: ReportInterval,
		Key:            Key,
		WorkerCount:    WorkerCount,
		PublicKeyPath:  PublicKeyPath,
	}

	return cfg, nil
}
