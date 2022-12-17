package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/brokers/v1brokers"
	"github.com/DanLavine/willow/pkg/config"
	"github.com/DanLavine/willow/pkg/logger"
	"github.com/DanLavine/willow/pkg/server"
	"github.com/DanLavine/willow/pkg/server/v1server"
)

func main() {
	config := config.Default()
	if err := config.Parse(); err != nil {
		log.Fatal(err)
	}

	loger := logger.NewZapLogger(config)
	defer loger.Sync()

	// v1 message handlers
	v1QueueManager := v1brokers.NewQueueManager()

	// v1 apis
	v1QueueServer := v1server.NewQueueHandler(loger, v1QueueManager)

	// setup async handlers
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())
	taskManager.AddTask("metrics_server", server.NewAdmin(loger, config, v1QueueManager))
	taskManager.AddTask("tcp_server", server.NewTCP(loger, config, v1QueueServer))

	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng willow cleanly: ", errs)
	}

	loger.Info("Successfully shutdown")
}
