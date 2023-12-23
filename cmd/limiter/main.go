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
	"github.com/DanLavine/willow/internal/limiter"
	"github.com/DanLavine/willow/internal/limiter/api"
	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/clients"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	"go.uber.org/zap"

	v1handlers "github.com/DanLavine/willow/internal/limiter/api/v1/handlers"
	v1router "github.com/DanLavine/willow/internal/limiter/api/v1/router"
	commonapi "github.com/DanLavine/willow/pkg/models/api"
)

func main() {
	cfg, err := config.Limiter(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewZapLogger(cfg)
	defer logger.Sync()

	// setup shutdown signa
	shutdown, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	// setup locker client
	clientConfig := &clients.Config{
		URL:             *cfg.LockerURL,
		ContentEncoding: commonapi.ContentTypeJSON,
		CAFile:          *cfg.LimiterCA,
		ClientKeyFile:   *cfg.LockerClientKey,
		ClientCRTFile:   *cfg.LockerClientCRT,
	}
	lockerClient, err := lockerclient.NewLockerClient(shutdown, clientConfig, nil)
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

	// setup server mux that is passed to all handlers
	mux := urlrouter.New()
	//// add the versioned apis to the server mux
	constructor, err := rules.NewRuleConstructor("memory")
	if err != nil {
		log.Fatal(err)
	}
	generalLocker := limiter.NewRulesManger(constructor)
	v1router.AddV1LimiterRoutes(mux, v1handlers.NewGroupRuleHandler(logger, shutdown, clientConfig, generalLocker))

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
