package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/agent/client"
	"github.com/mbiwapa/metric/internal/agent/collector"
	"github.com/mbiwapa/metric/internal/agent/sender"
	"github.com/mbiwapa/metric/internal/agent/source/gopsutilsource"
	"github.com/mbiwapa/metric/internal/agent/source/memstatssource"
	config "github.com/mbiwapa/metric/internal/config/client"
	"github.com/mbiwapa/metric/internal/logger"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
)

func main() {

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conf, err := config.MustLoadConfig()
	if err != nil {
		panic("Logger initialization error: " + err.Error())
	}

	logger, err := logger.New("info")
	if err != nil {
		panic("Logger initialization error: " + err.Error())
	}

	logger.Info("Start service!")

	memSource, err := memstatssource.New()
	if err != nil {
		logger.Error("Metrics source unavailable!", zap.Error(err))
		os.Exit(1)
	}

	psutilSource, err := gopsutilsource.New()
	if err != nil {
		logger.Error("Metrics source unavailable!", zap.Error(err))
		os.Exit(1)
	}

	storage, err := memstorage.New()
	if err != nil {
		logger.Error("Stor unavailable!", zap.Error(err))
		os.Exit(1)
	}

	client, err := client.New(conf.Addr, conf.Key, logger)
	if err != nil {
		logger.Error("Dont create http client", zap.Error(err))
		os.Exit(1)
	}
	errorChanel := make(chan error)

	collector.Start(storage, conf.PollInterval, logger, errorChanel, memSource, psutilSource)

	sender.Start(storage, client, conf.ReportInterval, logger, conf.WorkerCount, errorChanel)

	go func() {
		// Перехватываем ошибки у воркеров
		for err = range errorChanel {
			logger.Error("Error:", zap.Error(err))
		}
	}()

	<-mainCtx.Done()

	// Если придёт сигнал остановки в контекст, ждем 3 секунды завершения всех горутин и прощаемся
	time.Sleep(3 * time.Second)
	logger.Info("Good bye!")

}
