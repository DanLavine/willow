package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/logger"

	"github.com/DanLavine/willow/internal/locker/api"
	v1handlers "github.com/DanLavine/willow/internal/locker/api/v1/handlers"
	v1router "github.com/DanLavine/willow/internal/locker/api/v1/router"
	lockmanager "github.com/DanLavine/willow/internal/locker/lock_manager"
)

func main() {
	cfg, err := config.Locker(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewZapLogger(cfg)
	defer logger.Sync()

	// setup server mux that is passed to all handlers
	mux := urlrouter.New()
	// add the versioned apis to the server mux
	exclusiveLocker := lockmanager.NewExclusiveLocker()
	v1router.AddV1LockerRoutes(mux, v1handlers.NewLockHandler(logger, cfg, exclusiveLocker))

	// setup async handlers
	//// using strict config ensures that if any process fails, the server will ty and shutdown gracefully
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())

	// general locker
	taskManager.AddTask("exclusive Locker", exclusiveLocker)

	// v1 api handlers
	//// http2 server to handle all client requests
	taskManager.AddTask("locker_tcp_server", api.NewLockerTCP(logger, cfg, mux))

	// start all processes
	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng locker cleanly", errs)
	}

	logger.Info("Successfully shutdown")
}
