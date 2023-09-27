package main

import (
	"log"
	"os"

	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/logger"
)

func main() {
	cfg, err := config.Locker(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewZapLogger(cfg)
	defer logger.Sync()

	logger.Info("Successfully shutdown")
}
