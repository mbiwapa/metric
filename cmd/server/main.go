package main

import (
	"net/http"
	"os"

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
	"github.com/mbiwapa/metric/internal/storage/memstorage"
	"github.com/mbiwapa/metric/internal/storage/postgre"
)

func main() {

	config := config.MustLoadConfig()

	logger, err := logger.New("info")

	if err != nil {
		panic("Logger initialization error: " + err.Error())
	}

	logger.Info("Start service!")

	storage, err := memstorage.New()
	if err != nil {
		logger.Error("Can't create storage", zap.Error(err))
		os.Exit(1)
	}

	var pgstorage *postgre.Storage
	if config.DatabaseDSN != "" {
		pgstorage, err = postgre.New(config.DatabaseDSN)
		if err != nil {
			logger.Error("Can't create postgree storage", zap.Error(err))
			os.Exit(1)
		}
		defer pgstorage.Close()
	}

	//FIXME спросить у ментора вариант получше. Выглядит так себе полное повторение кода.
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
		os.Exit(1)
	}
	defer backup.SaveToFile()
	if config.Restore {
		backup.Restore()
	}
	if config.StoreInterval > 0 {
		go backup.Start()
	}

	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		mwLogger.New(logger),
		middleware.URLFormat,
		middleware.Compress(5, "application/json", "text/html"),
		decompressor.New(logger),
	)
	router.Post("/", undefinedType)
	//FIXME спросить у ментора вариант получше. Выглядит так себе полное повторение кода.
	if config.DatabaseDSN == "" {
		router.Post("/update/{type}/{name}/{value}", update.New(logger, storage, backup))
		router.Post("/update/", update.NewJSON(logger, storage, backup))
		router.Get("/value/{type}/{name}", value.New(logger, storage))
		router.Post("/value/", value.NewJSON(logger, storage))
		router.Get("/", home.New(logger, storage))
		router.Post("/updates/", updates.NewJSON(logger, storage, backup))
	} else {
		router.Post("/{type}/{name}/{value}", update.New(logger, pgstorage, backup))
		router.Post("/update/", update.NewJSON(logger, pgstorage, backup))
		router.Get("/value/{type}/{name}", value.New(logger, pgstorage))
		router.Post("/value/", value.NewJSON(logger, pgstorage))
		router.Get("/", home.New(logger, pgstorage))
		router.Get("/ping", ping.New(logger, pgstorage))
		router.Post("/updates/", updates.NewJSON(logger, pgstorage, backup))
	}

	srv := &http.Server{
		Addr:    config.Addr,
		Handler: router,
	}

	err = srv.ListenAndServe()
	if err != nil {
		logger.Error("The server did not start!", zap.Error(err))
		os.Exit(1)
	}
}

// undefinedType func return error fo undefined metric type request
func undefinedType(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
