package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mbiwapa/metric/internal/http-server/handlers/update"

	"github.com/mbiwapa/metric/internal/http-server/handlers/home"
	"github.com/mbiwapa/metric/internal/http-server/handlers/value"
	"github.com/mbiwapa/metric/internal/storage/memstorage"
)

func main() {

	stor, err := memstorage.New()

	if err != nil {
		panic("Storage unavailable!")
	}

	router := chi.NewRouter()

	router.Use(middleware.URLFormat)

	router.Route("/update", func(r chi.Router) {
		r.Post("/", undefinedType)
		// TODO без / на конце обрезает по первой .
		r.Post("/{type}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
			return
		})
		r.Post("/{type}/{name}/{value}/", update.New(stor))
	})
	router.Get("/value/{type}/{name}", value.New(stor))
	router.Get("/", home.New(stor))

	srv := &http.Server{
		Addr:    "localhost:8080",
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
