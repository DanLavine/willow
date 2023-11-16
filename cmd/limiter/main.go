package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/limiter"
	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/internal/server"
	"github.com/DanLavine/willow/internal/server/versions/v1server"
	"github.com/DanLavine/willow/pkg/clients"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Limiter(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewZapLogger(cfg)
	defer logger.Sync()

	// setup shutdown signa
	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)

	// setup locker client
	constructor, err := rules.NewRuleConstructor("memory")
	if err != nil {
		log.Fatal(err)
	}

	clientConfig := &clients.Config{
		URL:           *cfg.LockerURL,
		CAFile:        *cfg.LimiterCA,
		ClientKeyFile: *cfg.LockerClientKey,
		ClientCRTFile: *cfg.LockerClientCRT,
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
				log.Fatal("faile to setup the locker client which is required")
			}

			continue
		}

		break
	}

	// setup async handlers
	//// using strict config ensures that if any process fails, the server will ty and shutdown gracefully
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())

	// v1 api handlers
	//// http2 server to handle all client requests
	taskManager.AddTask("tcp_server", server.NewLimiterTCP(logger, cfg, v1server.NewGroupRuleHandler(logger, limiter.NewRulesManger(constructor, lockerClient))))

	// start all processes
	if errs := taskManager.Run(shutdown); errs != nil {
		logger.Fatal("Failed runnng Limiter cleanly", zap.Any("errors", errs))
	}

	logger.Info("Successfully shutdown")
}
