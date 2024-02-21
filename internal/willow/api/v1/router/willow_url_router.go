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

	// broker functions
	//// all queues operations
	mux.HandleFunc("POST", "/v1/queues", v1QueueHandler.Create)
	mux.HandleFunc("GET", "/v1/queues", v1QueueHandler.List) // just list all queues, don't think a query makes sense
	//// queue's specific operations
	mux.HandleFunc("GET", "/v1/queues/:queue_name", v1QueueHandler.Get)       // this I think can take a query for the queues and return channel details
	mux.HandleFunc("PUT", "/v1/queues/:queue_name", v1QueueHandler.Update)    // update how many items can be saved in a queue
	mux.HandleFunc("DELETE", "/v1/queues/:queue_name", v1QueueHandler.Delete) // delete a queue and all channels

	// message channels
	//// queues
	mux.HandleFunc("POST", "/v1/queues/:queue_name/channels", v1QueueHandler.ChannelEnqueue)
	mux.HandleFunc("GET", "/v1/queues/:queue_name/channels", v1QueueHandler.ChannelDequeue)
	mux.HandleFunc("DELETE", "/v1/queues/:queue_name/channels", v1QueueHandler.ChannelDelete) // Delete a channel by key values

	// item handlers
	//// queues
	mux.HandleFunc("POST", "/v1/queues/:queue_name/channels/items/ack", v1QueueHandler.ItemACK)             // think this is ok? enqueue and dequeue will drive this out
	mux.HandleFunc("POST", "/v1/queues/:queue_name/channels/items/heartbeat", v1QueueHandler.ItemHeartbeat) // think this is ok? enqueue and dequeue will drive this out
}
