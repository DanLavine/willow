package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/internal/brokers"
	"github.com/DanLavine/willow/internal/brokers/queues"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/internal/server"
	"github.com/DanLavine/willow/internal/server/versions/v1willow"
)

func main() {
	cfg, err := config.Willow(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewZapLogger(cfg)
	defer logger.Sync()

	queueConstructor := queues.NewQueueConstructor(cfg)
	queueManager := brokers.NewBrokerManager(queueConstructor)

	// setup async handlers
	//// using strict config ensures that if any process fails, the server will ty and shutdown gracefully
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())
	//// queue manager
	taskManager.AddTask("queue manager", queueManager)

	// v1 api handlers
	//// http2 server to handle all client requests
	taskManager.AddTask("tcp_server", server.NewWillowTCP(logger, cfg, queueManager, v1willow.NewQueueHandler(logger, queueManager)))
	//// http server to handle a dashboard's metrics requests
	taskManager.AddTask("metrics_server", server.NewMetrics(logger, cfg, v1willow.NewMetricsHandler(logger, queueManager)))

	// start all processes
	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng willow cleanly", errs)
	}

	logger.Info("Successfully shutdown")
}
