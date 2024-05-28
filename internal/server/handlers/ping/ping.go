package ping

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// Pinger is an interface that defines a method for checking the availability of a database or service.
// It is intended to be implemented by any storage or service that needs to provide a health check mechanism.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Pinger
type Pinger interface {
	// Ping checks the availability of the database or service.
	// It takes a context.Context as a parameter to allow for timeout and cancellation control.
	// Returns an error if the database or service is unavailable or if there is an issue performing the check.
	Ping(ctx context.Context) error
}

// New returns a new HTTP handler function that checks the availability of a database or service.
// It takes a zap.Logger and a Pinger as parameters. The zap.Logger is used for logging, and the Pinger is the interface that defines the method for checking the availability of the database or service.
// The returned http.HandlerFunc checks the availability of the database or service by calling the Ping method of the Pinger interface.
// It takes a context.Context as a parameter to allow for timeout and cancellation control.
// If the database or service is unavailable or if there is an issue performing the check, it logs an error and returns an HTTP status code of 500 (Internal Server Error).
// If the database or service is available, it logs an info message and returns an HTTP status code of 200 (OK).
// The context.Context is used to set a timeout of 10 seconds for the database check.
// The context.Context is also used to cancel the database check if the context is canceled.
// The zap.Logger is used to log the operation and the request ID.
// The zap.Logger is also used to log an error if the database or service is unavailable, and to log an info message if the database or service is available.
// The http.ResponseWriter is used to write the HTTP status code to the response.
// The http.Request is used to get the context.Context.
// The Pinger interface is used to check the availability of the database or service.
// The returned http.HandlerFunc can be used as a middleware to check the availability of a database or service before processing the request.
func New(log *zap.Logger, storage Pinger) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.ping.New"

		ctx := r.Context()

		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(ctx)),
		)

		databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := storage.Ping(databaseCtx)
		if err != nil {
			log.Error("Database is unvailable", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Info("Database is available")
		w.WriteHeader(http.StatusOK)
	}
}
