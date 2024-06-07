// Package main is the entry point of the application. It initializes the configuration, logger,
// metrics sources, storage, and client. It also starts the collector and sender routines
// and handles graceful shutdown on receiving termination signals.

// buildVersion, buildDate, and buildCommit are used to store the build version, build date, and build commit
// information, respectively. These variables are set during the build process and can be used to
// identify the specific build of the application.
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

var buildVersion string
var buildDate string
var buildCommit string

// main is the entry point of the application. It initializes the configuration, logger,
// metrics sources, storage, and client. It also starts the collector and sender routines
// and handles graceful shutdown on receiving termination signals.
func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	println("Build version:", buildVersion)
	println("Build date:", buildDate)
	println("Build commit:", buildCommit)

	// Create a context that is canceled on receiving an interrupt or SIGTERM signal.
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load configuration.
	conf, err := config.MustLoadConfig()
	if err != nil {
		panic("Logger initialization error: " + err.Error())
	}

	// Initialize logger.
	logger, err := logger.New("info")
	if err != nil {
		panic("Logger initialization error: " + err.Error())
	}

	logger.Info("Start service!")

	// Initialize memory statistics source.
	memSource, err := memstatssource.New()
	if err != nil {
		logger.Error("Metrics source unavailable!", zap.Error(err))
	}

	// Initialize psutil source.
	psutilSource, err := gopsutilsource.New()
	if err != nil {
		logger.Error("Metrics source unavailable!", zap.Error(err))
	}

	// Initialize in-memory storage.
	storage, err := memstorage.New()
	if err != nil {
		logger.Error("Storage unavailable!", zap.Error(err))
	}

	// Initialize HTTP client.
	client, err := client.New(conf.Addr, conf.Key, logger)
	if err != nil {
		logger.Error("Failed to create HTTP client", zap.Error(err))
	}

	// Channel to capture errors from collector and sender.
	errorChanel := make(chan error)

	// Start the collector routine.
	collector.Start(storage, conf.PollInterval, logger, errorChanel, memSource, psutilSource)

	// Start the sender routine.
	sender.Start(mainCtx, storage, client, conf.ReportInterval, logger, conf.WorkerCount, errorChanel)

	// Goroutine to log errors from the error channel.
	go func() {
		for err = range errorChanel {
			logger.Error("Error:", zap.Error(err))
		}
	}()

	// Wait for termination signal.
	<-mainCtx.Done()

	// Wait for 3 seconds to allow all goroutines to finish and log shutdown message.
	time.Sleep(3 * time.Second)
	logger.Info("Good bye!")
}
