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
	config := config.Default()
	if err := config.Parse(); err != nil {
		log.Fatal(err)
	}

	loger := logger.NewZapLogger(config)
	defer loger.Sync()

	// setup dead letter queue
	var deadLetterQueue queues.Queue

	switch config.StorageType {
	case config.StorageType:
		deadLetterQueue = queues.NewDiskQueueManager(config.DiskStorageDir)
	}

	// v1 apis
	v1QueueServer := v1server.NewQueueHandler(loger, deadLetterQueue)

	// setup async handlers
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())
	taskManager.AddTask("metrics_server", server.NewAdmin(loger, config, deadLetterQueue))
	taskManager.AddTask("tcp_server", server.NewTCP(loger, config, v1QueueServer))

	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng willow cleanly: ", errs)
	}

	loger.Info("Successfully shutdown")
}
