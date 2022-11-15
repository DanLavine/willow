package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/brokers"
	"github.com/DanLavine/willow/pkg/server"
	"go.uber.org/zap"
)

var (
	port = flag.String("port", "8080", "willow server port")
)

func main() {
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	queues := brokers.NewQueues(logger)

	taskManager := goasync.NewTaskManager()
	taskManager.AddTask("willow", server.NewTCP(logger, *port, queues))

	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng willow cleanly: ", errs)
	}

	logger.Info("Successfully shutdown")
}
