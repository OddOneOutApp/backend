package utils

import "go.uber.org/zap"

var Logger *zap.SugaredLogger
var RawLogger *zap.Logger

func InitializeLogger() {
	RawLogger, _ = zap.NewDevelopment()
	Logger = RawLogger.Sugar()
	defer Logger.Sync()
}