package types

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/forbole/callisto/v4/modules/actions/logging"

	"github.com/rs/zerolog/log"
)

// ActionsWorker represents the worker that is used to handle Hasura actions queries
type ActionsWorker struct {
	mux         *http.ServeMux
	context     *Context
	corsEnabled bool
}

// NewActionsWorker returns a new ActionsWorker instance
func NewActionsWorker(context *Context) *ActionsWorker {
	return &ActionsWorker{
		mux:         http.NewServeMux(),
		context:     context,
		corsEnabled: false,
	}
}

func (w *ActionsWorker) SetCORSAllowAll() {
	w.corsEnabled = true
	log.Debug().Msg("CORS Allow-All enabled")
}

// RegisterHandler registers the provided handler to be used on each call to the provided path
func (w *ActionsWorker) RegisterHandler(path string, handler ActionHandler) {
	log.Debug().Str("action", path).Msg("registering actions handler")
	w.mux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()

		// Set the content type
		writer.Header().Set("Content-Type", "application/json")

		// Add CORS headers if enabled
		if w.corsEnabled {
			writer.Header().Set("Access-Control-Allow-Origin", "*")
			writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}

		// Handle preflight OPTIONS requests
		if w.corsEnabled && request.Method == "OPTIONS" {
			writer.WriteHeader(http.StatusOK)
			return
		}

		// Read the body
		reqBody, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, "invalid payload", http.StatusBadRequest)
			return
		}
		defer request.Body.Close()

		// Get the actions payload
		var payload Payload
		err = json.Unmarshal(reqBody, &payload)
		if err != nil {
			http.Error(writer, "invalid payload: failed to unmarshal json", http.StatusInternalServerError)
			return
		}

		// Handle the request
		res, err := handler(w.context, &payload)
		if err != nil {
			logging.ErrorCounter(path)
			w.handleError(writer, path, err)
			return
		}

		// Marshal the response
		data, err := json.Marshal(res)
		if err != nil {
			logging.ErrorCounter(path)
			w.handleError(writer, path, err)
			return
		}

		// Prometheus
		logging.SuccessCounter(path)
		logging.ReponseTimeBuckets(path, start)

		// Write the response
		writer.Write(data)
	})
}

// RegisterHandler registers the provided handler to be used on each call to the provided path
func (w *ActionsWorker) RegisterGetHandler(path string, handler ActionGetHandler) {
	log.Debug().Str("action", path).Msg("registering actions handler")
	w.mux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()

		// Set the content type
		writer.Header().Set("Content-Type", "application/json")

		if w.corsEnabled {
			writer.Header().Set("Access-Control-Allow-Origin", "*")
			writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}

		// Handle preflight OPTIONS requests
		if w.corsEnabled && request.Method == "OPTIONS" {
			writer.WriteHeader(http.StatusOK)
			return
		}

		// Handle the request
		res, err := handler(w.context)
		if err != nil {
			logging.ErrorCounter(path)
			w.handleError(writer, path, err)
			return
		}

		// Marshal the response
		data, err := json.Marshal(res)
		if err != nil {
			logging.ErrorCounter(path)
			w.handleError(writer, path, err)
			return
		}

		// Prometheus
		logging.SuccessCounter(path)
		logging.ReponseTimeBuckets(path, start)

		// Write the response
		writer.Write(data)
	})
}

// handleError allows to handle the given error by writing it to the provided writer
func (w *ActionsWorker) handleError(writer http.ResponseWriter, path string, err error) {
	log.Error().Str("action", path).
		Err(err).Msg("error while executing action")

	errorObject := GraphQLError{
		Message: err.Error(),
	}
	errorBody, err := json.Marshal(errorObject)
	if err != nil {
		panic(err)
	}

	writer.WriteHeader(http.StatusBadRequest)
	writer.Write(errorBody)
}

// Start starts the worker
func (w *ActionsWorker) Start(host string, port uint) {
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           w.mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	err := server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}
