package main

import (
	"time"

	config "github.com/mbiwapa/metric/internal/config/client"
	"github.com/mbiwapa/metric/internal/http-client/send"
	"github.com/mbiwapa/metric/internal/metrics/collector"
	"github.com/mbiwapa/metric/internal/metrics/sender"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
	"github.com/mbiwapa/metric/internal/storage/metrepo"
)

func main() {

	conf, err := config.MustLoadConfig()
	if err != nil {
		panic(err)
	}

	metricsRepo, err := metrepo.New()
	if err != nil {
		panic("Metrics Repo unavailable!")
	}

	storage, err := memstorage.New()
	if err != nil {
		panic("Stor unavailable!")
	}

	client, err := send.New(conf.Addr)
	if err != nil {
		panic("Stor unavailable!")
	}

	go collector.Start(metricsRepo, storage, conf.ObservableMetrics, conf.PollInterval)

	go sender.Start(storage, client, conf.ReportInterval)

	//TODO переделать
	time.Sleep(10 * time.Minute)

}
