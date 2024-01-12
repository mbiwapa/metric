package main

import (
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/agent/client"
	"github.com/mbiwapa/metric/internal/agent/collector"
	"github.com/mbiwapa/metric/internal/agent/sender"
	"github.com/mbiwapa/metric/internal/agent/source/memstats"
	config "github.com/mbiwapa/metric/internal/config/client"
	"github.com/mbiwapa/metric/internal/logger"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
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

	client, err := client.New(conf.Addr, logger)
	if err != nil {
		logger.Error("Dont create http client", zap.Error(err))
		os.Exit(1)
	}

	go collector.Start(metricsRepo, storage, conf.ObservableMetrics, conf.PollInterval, logger)

	go sender.Start(storage, client, conf.ReportInterval, logger)

	//TODO переделать
	time.Sleep(10 * time.Minute)

}
