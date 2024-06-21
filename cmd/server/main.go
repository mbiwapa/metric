// Package main is the entry point of the application. It initializes the configuration, logger, storage, and backup mechanisms.
// It also sets up the HTTP server with appropriate routes and middleware, and handles graceful shutdown on receiving termination signals.

// buildVersion, buildDate, and buildCommit are used to store the build version, build date, and build commit
// information, respectively. These variables are set during the build process and can be used to
// identify the specific build of the application.
package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	config "github.com/mbiwapa/metric/internal/config/server"
	"github.com/mbiwapa/metric/internal/logger"
	"github.com/mbiwapa/metric/internal/server/backuper"
	"github.com/mbiwapa/metric/internal/server/decoder"
	"github.com/mbiwapa/metric/internal/server/handlers/home"
	"github.com/mbiwapa/metric/internal/server/handlers/ping"
	"github.com/mbiwapa/metric/internal/server/handlers/update"
	"github.com/mbiwapa/metric/internal/server/handlers/updates"
	"github.com/mbiwapa/metric/internal/server/handlers/value"
	mwDecoder "github.com/mbiwapa/metric/internal/server/middleware/decoder"
	"github.com/mbiwapa/metric/internal/server/middleware/decompressor"
	mwLogger "github.com/mbiwapa/metric/internal/server/middleware/logger"
	signatureCheck "github.com/mbiwapa/metric/internal/server/middleware/signature/check"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
	"github.com/mbiwapa/metric/internal/storage/postgre"
)

var buildVersion string
var buildDate string
var buildCommit string

// main is the entry point of the application. It initializes the configuration, logger, storage, and backup mechanisms.
// It also sets up the HTTP server with appropriate routes and middleware, and handles graceful shutdown on receiving termination signals.
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

	// Create a context that listens for the interrupt signal from the OS.
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load configuration settings.
	conf := config.MustLoadConfig()

	// Initialize the logger.
	logger, err := logger.New("info")
	if err != nil {
		panic("Logger initialization error: " + err.Error())
	}

	logger.Info("Start service!")

	decoder, err := decoder.New(conf.PrivateKeyPath)
	if err != nil {
		logger.Error("Can't create decoder", zap.Error(err))
	}

	// Initialize in-memory storage.
	storage, err := memstorage.New()
	if err != nil {
		logger.Error("Can't create storage", zap.Error(err))
	}

	// Initialize PostgreSQL storage if DatabaseDSN is provided.
	var pgstorage *postgre.Storage
	if conf.DatabaseDSN != "" {
		pgstorage, err = postgre.New(conf.DatabaseDSN)
		if err != nil {
			logger.Error("Can't create postgree storage", zap.Error(err))
		}
		defer pgstorage.Close()
	}

	// Initialize the backup mechanism.
	var backup *backuper.Buckuper
	if conf.DatabaseDSN == "" {
		backup, err = backuper.New(
			storage,
			conf.StoreInterval,
			conf.StoragePath,
			logger)
	} else {
		backup, err = backuper.New(
			pgstorage,
			conf.StoreInterval,
			conf.StoragePath,
			logger)
	}
	if err != nil {
		logger.Error("Can't create saver", zap.Error(err))
	}
	defer backup.SaveToFile()
	if conf.Restore {
		backup.Restore()
	}
	if conf.StoreInterval > 0 {
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
		signatureCheck.New(conf.Key, logger),
		mwDecoder.New(decoder),
	)
	router.Post("/", undefinedType)

	// Set up routes based on the storage type.
	if conf.DatabaseDSN == "" {
		router.Post("/update/{type}/{name}/{value}", update.New(logger, storage, backup))
		router.Post("/update/", update.NewJSON(logger, storage, backup, conf.Key))
		router.Get("/value/{type}/{name}", value.New(logger, storage, conf.Key))
		router.Post("/value/", value.NewJSON(logger, storage, conf.Key))
		router.Get("/", home.New(logger, storage, conf.Key))
		router.Post("/updates/", updates.NewJSON(logger, storage, backup, conf.Key))
	} else {
		router.Post("/{type}/{name}/{value}", update.New(logger, pgstorage, backup))
		router.Post("/update/", update.NewJSON(logger, pgstorage, backup, conf.Key))
		router.Get("/value/{type}/{name}", value.New(logger, pgstorage, conf.Key))
		router.Post("/value/", value.NewJSON(logger, pgstorage, conf.Key))
		router.Get("/", home.New(logger, pgstorage, conf.Key))
		router.Get("/ping", ping.New(logger, pgstorage))
		router.Post("/updates/", updates.NewJSON(logger, pgstorage, backup, conf.Key))
	}

	// Create and start the HTTP server.
	srv := &http.Server{
		Addr:    conf.Addr,
		Handler: router,
		BaseContext: func(_ net.Listener) context.Context {
			return mainCtx
		},
	}

	go func() {
		g, gCtx := errgroup.WithContext(mainCtx)
		g.Go(func() error {
			logger.Info("Starting server: ", zap.String("Addr", srv.Addr))
			return srv.ListenAndServe()
		})
		g.Go(func() error {
			<-gCtx.Done()
			logger.Info("Shutdown server!")
			return srv.Shutdown(context.Background())
		})
		if err := g.Wait(); err != nil {
			logger.Info("Exit reason: ", zap.Error(err))
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
