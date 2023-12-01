package main

import (
	"net/http"

	cupdate "github.com/mbiwapa/metric/internal/http-server/handlers/counter/update"
	gupdate "github.com/mbiwapa/metric/internal/http-server/handlers/gauge/update"
	"github.com/mbiwapa/metric/internal/storage"
)

func main() {

	stor, err := storage.New()

	if err != nil {
		panic("Storage unavailable!")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/update/gauge/", gupdate.New(stor))
	mux.HandleFunc("/update/counter/", cupdate.New(stor))
	mux.HandleFunc("/update/", undefinedType)

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		panic("The server did not start!")
	}
}

// undefinedType func return error fo undefined metric type request
func undefinedType(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
