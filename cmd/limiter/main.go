package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/limiter"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/internal/server"
	"github.com/DanLavine/willow/internal/server/versions/v1server"
)

func main() {
	cfg, err := config.Limiter(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewZapLogger(cfg)
	defer logger.Sync()

	// setup async handlers
	//// using strict config ensures that if any process fails, the server will ty and shutdown gracefully
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())

	// v1 api handlers
	//// http2 server to handle all client requests
	taskManager.AddTask("tcp_server", server.NewLimiterTCP(logger, cfg, v1server.NewGroupRuleHandler(logger, limiter.NewRulesManger())))

	// start all processes
	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng willow cleanly", errs)
	}

	logger.Info("Successfully shutdown")
}
