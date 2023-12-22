package main

import (
	"log/slog"
	"os"
	"time"

	config "github.com/mbiwapa/metric/internal/config/client"
	"github.com/mbiwapa/metric/internal/http-client/send"
	"github.com/mbiwapa/metric/internal/lib/logger/sl"
	"github.com/mbiwapa/metric/internal/metrics/collector"
	"github.com/mbiwapa/metric/internal/metrics/sender"
	"github.com/mbiwapa/metric/internal/metrics/source/memstats"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
)

func main() {

	conf, err := config.MustLoadConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logger.Info("Start service!")
	if err != nil {
		logger.Error("Can't load config", sl.Err(err))
		os.Exit(1)
	}

	metricsRepo, err := memstats.New()
	if err != nil {
		logger.Error("Metrics source unavailable!", sl.Err(err))
		os.Exit(1)
	}

	storage, err := memstorage.New()
	if err != nil {
		logger.Error("Stor unavailable!", sl.Err(err))
		os.Exit(1)
	}

	client, err := send.New(conf.Addr)
	if err != nil {
		logger.Error("Dont create http client", sl.Err(err))
		os.Exit(1)
	}

	go collector.Start(metricsRepo, storage, conf.ObservableMetrics, conf.PollInterval)

	go sender.Start(storage, client, conf.ReportInterval)

	//TODO переделать
	time.Sleep(10 * time.Minute)

}
