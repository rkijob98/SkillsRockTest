package main

import (
	"go.uber.org/zap"
	"task-manager/internal/app"
	"task-manager/pkg/config"
	"task-manager/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.Init(cfg.Environment)
	defer log.Sync()

	application := app.New(cfg, log)

	if err := application.Run(); err != nil {
		log.Fatal("Application failed", zap.Error(err))
	}
}
