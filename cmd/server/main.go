package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	config "github.com/mbiwapa/metric/internal/config/server"
	"github.com/mbiwapa/metric/internal/http-server/handlers/home"
	"github.com/mbiwapa/metric/internal/http-server/handlers/update"
	"github.com/mbiwapa/metric/internal/http-server/handlers/value"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
)

func main() {
	conf := config.MustLoadConfig()

	storage, err := memstorage.New()
	if err != nil {
		panic("Storage unavailable!")
	}

	router := chi.NewRouter()
	router.Use(middleware.URLFormat)
	router.Route("/update", func(r chi.Router) {
		r.Post("/", undefinedType)
		r.Post("/{type}/{name}/{value}", update.New(storage))
	})
	router.Get("/value/{type}/{name}", value.New(storage))
	router.Get("/", home.New(storage))

	srv := &http.Server{
		Addr:    conf.Addr,
		Handler: router,
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic("The server did not start!")
	}
}

// undefinedType func return error fo undefined metric type request
func undefinedType(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
