package logger

import (
	"log"

	"github.com/po3rin/esreindexer/config"
	"go.uber.org/zap"
)

var L *zap.SugaredLogger

func init() {
	var (
		logger *zap.Logger
		err    error
	)
	if config.Conf.LoggingLevel == "Debug" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		log.Fatal("init logger")
	}

	defer logger.Sync() // flushes buffer, if any
	L = logger.Sugar()
}
