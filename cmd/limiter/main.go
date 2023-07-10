package main

import (
	"log"
	"os"

	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/config"
)

func main() {
	cfg, err := config.Limiter(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewZapLogger(cfg)
	defer logger.Sync()
}
