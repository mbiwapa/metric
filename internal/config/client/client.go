package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config holds all the server configurations.
type Config struct {
	Addr           string // Server address and port for metric collection
	ReportInterval int64  // Frequency of sending metrics to the server (in seconds)
	PollInterval   int64  // Frequency of polling metrics from the source (in seconds)
	Key            string // Key for hash computation
	WorkerCount    int    // Number of threads for sending metrics
}

// MustLoadConfig loads the configuration from command-line flags and environment variables.
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

	// Define command-line flags
	flag.StringVar(&Addr, "a", "localhost:8080", "Адрес  и порт сервера по сбору метрик")
	flag.Int64Var(&ReportInterval, "r", 10, "Частота отправки метрик на сервер (по умолчанию 10 секунд)")
	flag.Int64Var(&PollInterval, "p", 2, "Частота опроса метрик из источника (по умолчанию 2 секунды)")
	flag.StringVar(&Key, "k", "", "Ключ для вычисления хеша")
	flag.IntVar(&WorkerCount, "l", 1, "Количество потоков для отправки метрик (по умолчанию 1 поток)")
	flag.Parse()

	// Override with environment variables if they are set
	envAddr := os.Getenv("ADDRESS")
	envPollInterval := os.Getenv("REPORT_INTERVAL")
	envReportInterval := os.Getenv("POLL_INTERVAL")
	envWorkerCount := os.Getenv("RATE_LIMIT")
	envKey := os.Getenv("KEY")
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

	// Create the configuration struct
	cfg := &Config{
		Addr:           "http://" + Addr,
		PollInterval:   PollInterval,
		ReportInterval: ReportInterval,
		Key:            Key,
		WorkerCount:    WorkerCount,
	}

	return cfg, nil
}
