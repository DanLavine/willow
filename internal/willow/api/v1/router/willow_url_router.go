package router

import (
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/willow/api/v1/handlers"
	"github.com/go-openapi/runtime/middleware"
)

func AddV1WillowRoutes(mux *urlrouter.Router, v1QueueHandler handlers.V1QueueHandler) {
	// OpenAPI endpoints
	// api url to server all the OpenAPI files
	_, currentDir, _, _ := runtime.Caller(0)
	mux.HandleFunc("GET", "/v1/docs/openapi/", func(w http.ResponseWriter, r *http.Request) {
		handle := http.StripPrefix("/v1/docs/openapi/", http.FileServer(http.Dir(filepath.Join(currentDir, "..", "..", "..", "..", "..", "..", "docs", "openapi"))))
		handle.ServeHTTP(w, r)
	})

	// ui that calls the files and knows what to do
	mux.HandleFunc("GET", "/v1/docs", func(w http.ResponseWriter, r *http.Request) {
		middleware.Redoc(middleware.RedocOpts{
			BasePath: "/",
			Path:     "v1/docs",
			SpecURL:  "/v1/docs/openapi/willow/openapi.yaml",
			Title:    "Locker API Documentation",
		}, nil).ServeHTTP(w, r)
	})

	// broke function functions
	mux.HandleFunc("POST", "/v1/brokers/queues/create", v1QueueHandler.Create)

	// message handlers
	mux.HandleFunc("POST", "/v1/brokers/item/enqueue", v1QueueHandler.Enqueue)
	mux.HandleFunc("GET", "/v1/brokers/item/dequeue", v1QueueHandler.Dequeue)
	mux.HandleFunc("POST", "/v1/brokers/item/ack", v1QueueHandler.ACK)
}
