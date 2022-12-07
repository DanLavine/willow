package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/brokers"
	"github.com/DanLavine/willow/pkg/brokers/v1brokers"
	"github.com/DanLavine/willow/pkg/server"
	"go.uber.org/zap"
)

var (
	willowPort  = flag.String("willow-port", "8080", "willow server port")
	metricsPort = flag.String("metrics-port", "8081", "willow server metrics port")
)

func main() {
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// v1 message handlers
	v1QueueManager := v1brokers.NewQueueManager()
	v1BrokerManager := v1brokers.NewBrokerManager(v1QueueManager)

	// setup broker to switch version
	brokerManager := brokers.NewBrokerManager(v1BrokerManager)

	// setup async handlers
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())
	taskManager.AddTask("queue_manager", v1QueueManager)
	taskManager.AddTask("metrics_server", server.NewAdmin(logger, *metricsPort, brokerManager))
	taskManager.AddTask("tcp_server", server.NewTCP(logger, *willowPort, brokerManager))

	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng willow cleanly: ", errs)
	}

	logger.Info("Successfully shutdown")
}
