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
	"github.com/DanLavine/willow/internal/willow/api"
	"github.com/DanLavine/willow/internal/willow/api/v1/handlers"
	v1router "github.com/DanLavine/willow/internal/willow/api/v1/router"
	"github.com/DanLavine/willow/internal/willow/brokers"
	"github.com/DanLavine/willow/internal/willow/brokers/queues"
)

func main() {
	cfg, err := config.Willow(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewZapLogger(cfg)
	defer logger.Sync()

	// setup willow server
	willowMux := urlrouter.New()

	queueConstructor := queues.NewQueueConstructor(cfg)
	queueManager := brokers.NewBrokerManager(queueConstructor)

	v1router.AddV1WillowRoutes(willowMux, handlers.NewV1QueueHandler(logger, queueManager))

	// setup metrics server
	metricsMux := urlrouter.New()
	v1router.AddV1WillowMetricsRoutes(metricsMux, handlers.NewV1MetricsHandler(logger, queueManager))

	// setup async handlers
	//// using strict config ensures that if any process fails, the server will ty and shutdown gracefully
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())

	// v1 api handlers
	//// queue manager
	taskManager.AddTask("queue manager", queueManager)
	//// http2 server to handle all client requests
	taskManager.AddTask("tcp_server", api.NewWillowTCP(logger, cfg, willowMux))
	//// http server to handle a dashboard's metrics requests
	taskManager.AddTask("metrics_server", api.NewMetrics(logger, cfg, metricsMux))

	// start all processes
	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng willow cleanly", errs)
	}

	logger.Info("Successfully shutdown")
}
