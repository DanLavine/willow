package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/internal/server"
	"github.com/DanLavine/willow/internal/server/v1server"
	"github.com/DanLavine/willow/internal/v1/queues"
)

func main() {
	cfg := config.Default()
	if err := cfg.Parse(); err != nil {
		log.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	logger := logger.NewZapLogger(cfg)
	defer logger.Sync()

	// setup dead letter queue
	var queueManager queues.QueueManager

	switch cfg.ConfigQueue.StorageType {
	case config.DiskStorage:
		queueManager = queues.NewDiskQueueManager(cfg.ConfigQueue)
	}

	// v1 apis
	v1QueueServer := v1server.NewQueueHandler(logger, queueManager)
	v1MetricsServer := v1server.NewMetricsHandler(logger, queueManager)

	// setup async handlers
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())
	taskManager.AddTask("metrics_server", server.NewMetrics(logger, cfg, v1MetricsServer))
	taskManager.AddTask("tcp_server", server.NewTCP(logger, cfg, v1QueueServer))

	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng willow cleanly: ", errs)
	}

	logger.Info("Successfully shutdown")
}
