package config

import (
	"flag"

	"github.com/mbiwapa/metric/internal/metrics/collector"
)

// Config Структура со всеми конфигурациями сервера
type Config struct {
	Addr              string
	ReportInterval    int64
	PollInterval      int64
	ObservableMetrics []collector.ObservableMetric
}

// MustLoadConfig загрузка конфигурации
func MustLoadConfig() *Config {
	var config Config
	flag.StringVar(&config.Addr, "a", "localhost:8080", "Адрес  и порт сервера по сбору метрик")
	config.Addr = "http://" + config.Addr
	flag.Int64Var(&config.ReportInterval, "r", 10, "Частота отправки метрик на сервер (по умолчанию 10 секунд)")
	flag.Int64Var(&config.PollInterval, "p", 2, "Частота опроса метрик из источника (по умолчанию 2 секунды)")
	config.ObservableMetrics = getObservableMetrics()
	flag.Parse()

	return &config
}

// getObservableMetrics возвращает список метрик для отслеживание агентом
func getObservableMetrics() []collector.ObservableMetric {
	observableMetrics := []collector.ObservableMetric{
		{Name: "Frees", SourceType: "uint"},
		{Name: "Alloc", SourceType: "uint"},
		{Name: "BuckHashSys", SourceType: "uint"},
		{Name: "GCCPUFraction", SourceType: "float"},
		{Name: "GCSys", SourceType: "uint"},
		{Name: "HeapAlloc", SourceType: "uint"},
		{Name: "HeapIdle", SourceType: "uint"},
		{Name: "HeapInuse", SourceType: "uint"},
		{Name: "HeapObjects", SourceType: "uint"},
		{Name: "HeapReleased", SourceType: "uint"},
		{Name: "HeapSys", SourceType: "uint"},
		{Name: "LastGC", SourceType: "uint"},
		{Name: "Lookups", SourceType: "uint"},
		{Name: "MCacheInuse", SourceType: "uint"},
		{Name: "MCacheSys", SourceType: "uint"},
		{Name: "MSpanInuse", SourceType: "uint"},
		{Name: "MSpanSys", SourceType: "uint"},
		{Name: "Mallocs", SourceType: "uint"},
		{Name: "NextGC", SourceType: "uint"},
		{Name: "NumForcedGC", SourceType: "uint"},
		{Name: "NumGC", SourceType: "uint"},
		{Name: "OtherSys", SourceType: "uint"},
		{Name: "PauseTotalNs", SourceType: "uint"},
		{Name: "StackInuse", SourceType: "uint"},
		{Name: "StackSys", SourceType: "uint"},
		{Name: "Sys", SourceType: "uint"},
		{Name: "TotalAlloc", SourceType: "uint"},
	}
	return observableMetrics
}
