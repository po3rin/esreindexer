package logger

import (
	"go.uber.org/zap"
)

var L *zap.SugaredLogger

func init() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	L = logger.Sugar()
}
