package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/limiter/api"
	"github.com/DanLavine/willow/internal/limiter/counters"
	"github.com/DanLavine/willow/internal/limiter/overrides"
	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/clients"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	"go.uber.org/zap"

	v1handlers "github.com/DanLavine/willow/internal/limiter/api/v1/handlers"
	"github.com/DanLavine/willow/internal/limiter/api/v1/router"
	commonapi "github.com/DanLavine/willow/pkg/models/api"
)

func main() {
	cfg, err := config.Limiter(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewZapLogger(cfg)
	defer logger.Sync()

	// setup shutdown signal
	shutdown, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	// setup locker client config and validate it
	clientConfig := &clients.Config{
		URL:             *cfg.LockerURL,
		ContentEncoding: commonapi.ContentTypeJSON,
		CAFile:          *cfg.LockerClientCA,
		ClientKeyFile:   *cfg.LockerClientKey,
		ClientCRTFile:   *cfg.LockerClientCRT,
	}
	lockerClient, err := lockerclient.NewLockClient(shutdown, clientConfig, nil)
	if err != nil {
		log.Fatal(err)
	}

	// ensure the locker client can connect to the locker service
	for i := 0; i < 10; i++ {
		if err := lockerClient.Healthy(); err != nil {
			logger.Error("error checking health of locker service", zap.Error(err))
			time.Sleep(10 * time.Second)

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
	v1handler := v1handlers.NewGroupRuleHandler(logger, shutdown, clientConfig, rulesClient, countersClient)

	// add v1 routes
	router.AddV1LimiterRoutes(mux, v1handler)

	// setup async handlers
	//// using strict config ensures that if any process fails, the server will ty and shutdown gracefully
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())
	taskManager.AddTask("tcp_server", api.NewLimiterTCP(logger, cfg, mux)) // tcp server

	// start all processes
	if errs := taskManager.Run(shutdown); errs != nil {
		logger.Fatal("Failed runnng Limiter cleanly", zap.Any("errors", errs))
	}

	logger.Info("Successfully shutdown")
}
