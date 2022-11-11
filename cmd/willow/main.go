package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/server"
)

var (
	port = flag.String("port", "8080", "willow server port")
)

func main() {
	flag.Parse()

	taskManager := goasync.NewTaskManager()
	taskManager.AddTask("willow", server.NewTCP(*port))

	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng willow cleanly: ", errs)
	}
}
