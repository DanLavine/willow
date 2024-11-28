package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DanLavine/goasync/v2"
	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/limiter/api"
	"github.com/DanLavine/willow/internal/limiter/counters"
	"github.com/DanLavine/willow/internal/limiter/overrides"
	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/pkg/clients"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	"go.uber.org/zap"

	v1handlers "github.com/DanLavine/willow/internal/limiter/api/v1/handlers"
	"github.com/DanLavine/willow/internal/limiter/api/v1/router"
)

func main() {
	cfg, err := config.Limiter(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := reporting.NewZapLogger(cfg)
	defer logger.Sync()

	// setup shutdown signal
	shutdown, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	// setup locker client config and validate it
	clientConfig := &clients.Config{
		URL:           *cfg.LockerURL,
		CAFile:        *cfg.LockerClientCA,
		ClientKeyFile: *cfg.LockerClientKey,
		ClientCRTFile: *cfg.LockerClientCRT,
	}
	lockerClient, err := lockerclient.NewLockClient(clientConfig)
	if err != nil {
		log.Fatal(err)
	}

	// ensure the locker client can connect to the locker service
	for i := 0; i < 10; i++ {
		if err := lockerClient.Healthy(); err != nil {
			logger.Error("error checking health of locker service. Will try again in 10 seconds", zap.Error(err))

			select {
			case <-time.After(10 * time.Second):
			case <-shutdown.Done():
				os.Exit(1)
			}

			if i == 9 {
				log.Fatal("failed to setup the locker client which is required")
			}

			continue
		}

		break
	}

	// overrides client
	overridesConstructor, err := overrides.NewOverrideConstructor("memory")
	if err != nil {
		logger.Fatal("failed to setup override constructor", zap.Error(err))
	}
	overridesClient := overrides.NewDefaultOverridesClientLocal(overridesConstructor)

	// rules client
	rulesConstructor, err := rules.NewRuleConstructor("memory")
	if err != nil {
		logger.Fatal("failed to setup rule constructor", zap.Error(err))
	}
	rulesClient := rules.NewLocalRulesClient(rulesConstructor, overridesClient)

	// counters client
	countersConstructor, err := counters.NewCountersConstructor("memory")
	if err != nil {
		logger.Fatal("failed to setup counters constructor", zap.Error(err))
	}
	countersClient := counters.NewCountersClientLocal(countersConstructor, rulesClient)

	// setup server mux that is passed to all handlers
	mux := urlrouter.New()
	v1handler := v1handlers.NewGroupRuleHandler(shutdown, clientConfig, rulesClient, countersClient)

	// add v1 routes
	router.AddV1LimiterRoutes(logger, mux, v1handler)

	// setup async handlers
	//// using strict config ensures that if any process fails, the server will ty and shutdown gracefully
	taskManager := goasync.NewTaskManager()
	taskManager.AddTask("tcp_server", api.NewLimiterTCP(logger, cfg, mux), goasync.EXECUTE_TASK_TYPE_STRICT) // tcp server

	// start all processes
	if errs := taskManager.Run(shutdown); errs != nil {
		logger.Fatal("Failed runnng Limiter cleanly", zap.Any("errors", errs))
	}

	logger.Info("Successfully shutdown")
}
