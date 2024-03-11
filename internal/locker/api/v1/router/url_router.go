package router

import (
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/locker/api/v1/handlers"
	"github.com/go-openapi/runtime/middleware"
)

func AddV1LockerRoutes(mux *urlrouter.Router, v1Handler handlers.V1LockerHandler) {
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
			SpecURL:  "/v1/docs/openapi/locker/openapi.yaml",
			Title:    "Locker API Documentation",
		}, nil).ServeHTTP(w, r)
	})

	mux.HandleFunc("POST", "/v1/locks", v1Handler.Create)
	mux.HandleFunc("DELETE", "/v1/locks/:lock_id", v1Handler.Delete)
	mux.HandleFunc("POST", "/v1/locks/:lock_id/heartbeat", v1Handler.Heartbeat)

	// Admin APIs
	// TODO: Need to actual account for auth for this
	mux.HandleFunc("GET", "/v1/locks", v1Handler.List)
}
