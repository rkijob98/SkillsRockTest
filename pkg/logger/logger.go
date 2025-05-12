package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log  *zap.Logger
	once sync.Once
)

// Init инициализирует логгер (Singleton)
func Init(env string) *zap.Logger {
	once.Do(func() {
		var config zap.Config

		if env == "production" {
			config = zap.NewProductionConfig()
		} else {
			config = zap.NewDevelopmentConfig()
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}

		// Настройка вывода в файл (опционально)
		//config.OutputPaths = []string{"stdout", "./logs/app.log"}
		//config.ErrorOutputPaths = []string{"stderr", "./logs/error.log"}

		var err error
		log, err = config.Build()
		if err != nil {
			panic(err)
		}
	})

	return log
}

// Get возвращает инстанс логгера
func Get() *zap.Logger {
	if log == nil {
		panic("Logger not initialized. Call Init() first.")
	}
	return log
}
