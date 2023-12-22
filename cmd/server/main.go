package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	config "github.com/mbiwapa/metric/internal/config/server"
	"github.com/mbiwapa/metric/internal/http-server/handlers/home"
	"github.com/mbiwapa/metric/internal/http-server/handlers/update"
	"github.com/mbiwapa/metric/internal/http-server/handlers/value"
	mwLogger "github.com/mbiwapa/metric/internal/http-server/middleware/logger"
	"github.com/mbiwapa/metric/internal/lib/logger/sl"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
)

func main() {

	config := config.MustLoadConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logger.Info("Start service!")

	storage, err := memstorage.New()
	if err != nil {
		logger.Error("Can't create storage", sl.Err(err))
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
		logger.Error("The server did not start!", sl.Err(err))
		os.Exit(1)
	}
}

// undefinedType func return error fo undefined metric type request
func undefinedType(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
