package main

import (
	"os"
	"time"

	config "github.com/mbiwapa/metric/internal/config/client"
	"github.com/mbiwapa/metric/internal/http-client/send"
	"github.com/mbiwapa/metric/internal/logger"
	"github.com/mbiwapa/metric/internal/metrics/collector"
	"github.com/mbiwapa/metric/internal/metrics/sender"
	"github.com/mbiwapa/metric/internal/metrics/source/memstats"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
	"go.uber.org/zap"
)

func main() {

	conf, err := config.MustLoadConfig()
	if err != nil {
		panic("Logger initialization error: " + err.Error())
	}

	logger, err := logger.New("info")
	if err != nil {
		panic("Logger initialization error: " + err.Error())
	}

	logger.Info("Start service!")

	metricsRepo, err := memstats.New()
	if err != nil {
		logger.Error("Metrics source unavailable!", zap.Error(err))
		os.Exit(1)
	}

	storage, err := memstorage.New()
	if err != nil {
		logger.Error("Stor unavailable!", zap.Error(err))
		os.Exit(1)
	}

	client, err := send.New(conf.Addr, logger)
	if err != nil {
		logger.Error("Dont create http client", zap.Error(err))
		os.Exit(1)
	}

	time.Sleep(10 * time.Second)

	go collector.Start(metricsRepo, storage, conf.ObservableMetrics, conf.PollInterval, logger)

	go sender.Start(storage, client, conf.ReportInterval, logger)

	//TODO переделать
	time.Sleep(10 * time.Minute)

}
