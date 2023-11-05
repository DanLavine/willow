package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/server/versions/v1server"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type limiterTCP struct {
	closed bool

	logger *zap.Logger
	config *config.LimiterConfig
	server *http.Server

	v1ruleHandler v1server.LimitRuleHandler
}

func NewLimiterTCP(logger *zap.Logger, config *config.LimiterConfig, v1ruleHandler v1server.LimitRuleHandler) *limiterTCP {
	return &limiterTCP{
		closed:        false,
		logger:        logger.Named("limiterTCP_server"),
		config:        config,
		v1ruleHandler: v1ruleHandler,
	}
}

func (limiter *limiterTCP) Cleanup() error { return nil }
func (limiter *limiterTCP) Initialize() error {
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", *limiter.config.LimiterPort),
	}

	// load the server CRT and Key
	cert, err := tls.LoadX509KeyPair(*limiter.config.LimiterServerCRT, *limiter.config.LimiterServerKey)
	if err != nil {
		return err
	}

	// add them to the server
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// load the CA cert if it exists
	if *limiter.config.LimiterCA != "" {
		CaPEM, err := os.ReadFile(*limiter.config.LimiterCA)
		if err != nil {
			return err
		}
		CAs := x509.NewCertPool()
		if !CAs.AppendCertsFromPEM(CaPEM) {
			return fmt.Errorf("failed to parse LimiterCA")
		}

		server.TLSConfig.RootCAs = CAs
	}

	// enforce using http2
	limiter.server = server

	// certs loaded successfully
	return nil
}

func (limiter *limiterTCP) Execute(ctx context.Context) error {
	logger := limiter.logger
	mux := urlrouter.New()

	// crud operations for group rules
	// These operations seem more like a normal DB that I want to do...
	mux.HandleFunc("POST", "/v1/limiter/rules/create", limiter.v1ruleHandler.Create)
	mux.HandleFunc("PUT", "/v1/limiter/rules/:name/update", limiter.v1ruleHandler.Update)
	mux.HandleFunc("DELETE", "/v1/limiter/rules/:name/delete", limiter.v1ruleHandler.Delete)
	mux.HandleFunc("GET", "/v1/limiter/rules/find", limiter.v1ruleHandler.Find)

	// create an override for a specific rule
	mux.HandleFunc("POST", "/v1/limiter/rules/:name/override", limiter.v1ruleHandler.SetOverride)
	mux.HandleFunc("DELETE", "/v1/limiter/rules/:name/override", limiter.v1ruleHandler.DeleteOverride)

	// operations to check items against arbitrary rules
	mux.HandleFunc("POST", "/v1/limiter/items/increment", limiter.v1ruleHandler.Increment)
	mux.HandleFunc("DELETE", "/v1/limiter/items/decrement", limiter.v1ruleHandler.Decrement)

	// set the server mux
	limiter.server.Handler = mux
	http2.ConfigureServer(limiter.server, &http2.Server{})

	// handle shutdown
	shutdownErr := make(chan error)
	go func() {
		<-ctx.Done()
		shutdownErr <- limiter.server.Shutdown(context.Background())
	}()

	// return any error other than the server closed
	logger.Info("Limiter TCP server running")
	if err := limiter.server.ListenAndServeTLS("", ""); err != nil {
		select {
		case <-ctx.Done():
			if err != http.ErrServerClosed {
				// must be an unexpected error during shutdown
				return err
			}

			// context was closed and server closed error. clean shutdown case
		default:
			// always return the error if the context was not closed
			return err
		}
	}

	// wait for any pending connections to drain
	if err := <-shutdownErr; err != nil {
		logger.Error("server shutdown with error", zap.Error(err))
		return err
	}

	logger.Info("server shutdown successfully")
	return nil
}
