package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	config "github.com/mbiwapa/metric/internal/config/server"
	"github.com/mbiwapa/metric/internal/http-server/handlers/home"
	"github.com/mbiwapa/metric/internal/http-server/handlers/update"
	"github.com/mbiwapa/metric/internal/http-server/handlers/value"
	mwLogger "github.com/mbiwapa/metric/internal/http-server/middleware/logger"
	"github.com/mbiwapa/metric/internal/logger"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
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

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(logger))
	router.Use(middleware.URLFormat)

	router.Route("/update", func(r chi.Router) {
		r.Post("/", undefinedType)
		r.Post("/{type}/{name}/{value}", update.New(logger, storage))
	})
	router.Get("/value/{type}/{name}", value.New(logger, storage))
	router.Get("/", home.New(logger, storage))

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
