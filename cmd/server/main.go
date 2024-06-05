// Package main is the entry point of the application. It initializes the configuration, logger, storage, and backup mechanisms.
// It also sets up the HTTP server with appropriate routes and middleware, and handles graceful shutdown on receiving termination signals.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	config "github.com/mbiwapa/metric/internal/config/server"
	"github.com/mbiwapa/metric/internal/logger"
	"github.com/mbiwapa/metric/internal/server/backuper"
	"github.com/mbiwapa/metric/internal/server/handlers/home"
	"github.com/mbiwapa/metric/internal/server/handlers/ping"
	"github.com/mbiwapa/metric/internal/server/handlers/update"
	"github.com/mbiwapa/metric/internal/server/handlers/updates"
	"github.com/mbiwapa/metric/internal/server/handlers/value"
	"github.com/mbiwapa/metric/internal/server/middleware/decompressor"
	mwLogger "github.com/mbiwapa/metric/internal/server/middleware/logger"
	signatureCheck "github.com/mbiwapa/metric/internal/server/middleware/signature/check"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
	"github.com/mbiwapa/metric/internal/storage/postgre"
)

// main is the entry point of the application. It initializes the configuration, logger, storage, and backup mechanisms.
// It also sets up the HTTP server with appropriate routes and middleware, and handles graceful shutdown on receiving termination signals.
func main() {

	// Create a context that listens for the interrupt signal from the OS.
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load configuration settings.
	config := config.MustLoadConfig()

	// Initialize the logger.
	logger, err := logger.New("info")
	if err != nil {
		panic("Logger initialization error: " + err.Error())
	}

	logger.Info("Start service!")

	// Initialize in-memory storage.
	storage, err := memstorage.New()
	if err != nil {
		logger.Error("Can't create storage", zap.Error(err))
	}

	// Initialize PostgreSQL storage if DatabaseDSN is provided.
	var pgstorage *postgre.Storage
	if config.DatabaseDSN != "" {
		pgstorage, err = postgre.New(config.DatabaseDSN)
		if err != nil {
			logger.Error("Can't create postgree storage", zap.Error(err))
		}
		defer pgstorage.Close()
	}

	// Initialize the backup mechanism.
	var backup *backuper.Buckuper
	if config.DatabaseDSN == "" {
		backup, err = backuper.New(
			storage,
			config.StoreInterval,
			config.StoragePath,
			logger)
	} else {
		backup, err = backuper.New(
			pgstorage,
			config.StoreInterval,
			config.StoragePath,
			logger)
	}
	if err != nil {
		logger.Error("Can't create saver", zap.Error(err))
	}
	defer backup.SaveToFile()
	if config.Restore {
		backup.Restore()
	}
	if config.StoreInterval > 0 {
		go backup.Start()
	}

	// Set up the HTTP router and middleware.
	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		mwLogger.New(logger),
		middleware.URLFormat,
		middleware.Compress(5, "application/json", "text/html"),
		decompressor.New(logger),
		signatureCheck.New(config.Key, logger),
	)
	router.Post("/", undefinedType)

	// Set up routes based on the storage type.
	if config.DatabaseDSN == "" {
		router.Post("/update/{type}/{name}/{value}", update.New(logger, storage, backup))
		router.Post("/update/", update.NewJSON(logger, storage, backup, config.Key))
		router.Get("/value/{type}/{name}", value.New(logger, storage, config.Key))
		router.Post("/value/", value.NewJSON(logger, storage, config.Key))
		router.Get("/", home.New(logger, storage, config.Key))
		router.Post("/updates/", updates.NewJSON(logger, storage, backup, config.Key))
	} else {
		router.Post("/{type}/{name}/{value}", update.New(logger, pgstorage, backup))
		router.Post("/update/", update.NewJSON(logger, pgstorage, backup, config.Key))
		router.Get("/value/{type}/{name}", value.New(logger, pgstorage, config.Key))
		router.Post("/value/", value.NewJSON(logger, pgstorage, config.Key))
		router.Get("/", home.New(logger, pgstorage, config.Key))
		router.Get("/ping", ping.New(logger, pgstorage))
		router.Post("/updates/", updates.NewJSON(logger, pgstorage, backup, config.Key))
	}

	// Create and start the HTTP server.
	srv := &http.Server{
		Addr:    config.Addr,
		Handler: router,
	}

	go func() {
		err = srv.ListenAndServe()
		if err != nil {
			logger.Error("The server did not start!", zap.Error(err))
		}
	}()

	// Wait for the termination signal.
	<-mainCtx.Done()

	// Wait for 3 seconds to allow graceful shutdown of goroutines.
	time.Sleep(3 * time.Second)
	logger.Info("Good bye!")
}

// undefinedType handles requests with undefined metric types by returning a 400 Bad Request status.
func undefinedType(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
