package main

import (
	"time"

	"github.com/mbiwapa/metric/internal/http-client/send"
	"github.com/mbiwapa/metric/internal/metrics/collector"
	"github.com/mbiwapa/metric/internal/metrics/sender"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
	"github.com/mbiwapa/metric/internal/storage/metrepo"
)

func main() {

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
	var pollInterval int64
	pollInterval = 2

	var reportInterval int64
	reportInterval = 10

	metricsRepo, err := metrepo.New()
	if err != nil {
		panic("Metrics Repo unavailable!")
	}

	storage, err := memstorage.New()
	if err != nil {
		panic("Stor unavailable!")
	}

	client, err := send.New("http://localhost:8080/update")
	if err != nil {
		panic("Stor unavailable!")
	}

	go collector.Start(metricsRepo, storage, observableMetrics, pollInterval)

	go sender.Start(storage, client, reportInterval)

	//TODO переделать
	time.Sleep(10 * time.Minute)

}
