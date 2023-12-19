package router

import (
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/limiter/api/v1/handlers"
	"github.com/go-openapi/runtime/middleware"
)

func AddV1LimiterRoutes(mux *urlrouter.Router, v1Handler handlers.V1LimiterRuleHandler) {
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
			//RedocURL: "https://cdn.jsdelivr.net/npm/redoc@2.0.0/bundles/redoc.standalone.js",
			BasePath: "/",
			Path:     "v1/docs",
			SpecURL:  "/docs/openapi/limiter/openapi.yaml",
			Title:    "Limiter API Documentation",
		}, nil).ServeHTTP(w, r)
	})

	// crud operations for group rules
	mux.HandleFunc("GET", "/v1/limiter/rules", v1Handler.ListRules)
	mux.HandleFunc("POST", "/v1/limiter/rules", v1Handler.CreateRule)
	mux.HandleFunc("GET", "/v1/limiter/rules/:rule_name", v1Handler.GetRule)
	mux.HandleFunc("PUT", "/v1/limiter/rules/:rule_name", v1Handler.UpdateRule)
	mux.HandleFunc("DELETE", "/v1/limiter/rules/:rule_name", v1Handler.DeleteRule)

	// create an override for a specific rule
	mux.HandleFunc("GET", "/v1/limiter/rules/:rule_name/overrides", v1Handler.ListOverrides)
	mux.HandleFunc("POST", "/v1/limiter/rules/:rule_name/overrides", v1Handler.SetOverride)
	mux.HandleFunc("DELETE", "/v1/limiter/rules/:rule_name/overrides/:override_name", v1Handler.DeleteOverride)

	// operations to check items against arbitrary rules
	mux.HandleFunc("GET", "/v1/limiter/counters", v1Handler.ListCounters)
	mux.HandleFunc("POST", "/v1/limiter/counters", v1Handler.Increment)
	mux.HandleFunc("DELETE", "/v1/limiter/counters", v1Handler.Decrement)

	// operations to setup or clean counters without checking rules
	mux.HandleFunc("POST", "/v1/limiter/counters/set", v1Handler.SetCounters)
}
