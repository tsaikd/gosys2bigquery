package applog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var gLogger *zap.Logger
var gDebug = true

// Logger return logger instance
func Logger() *zap.Logger {
	if gLogger == nil {
		if err := initLogger(gDebug); err != nil {
			panic(err)
		}
	}
	return gLogger
}

// Debug return debug flag
func Debug() bool {
	return gDebug
}

// Trace if err != nil
func Trace(err error) {
	if err != nil {
		Logger().Error(err.Error())
	}
}

// GoCmdMode use console encoding even in zap production mode
func GoCmdMode() (err error) {
	return initLogger(true)
}

func initLogger(debug bool) (err error) {
	if debug {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		if gLogger, err = config.Build(); err != nil {
			return
		}
	} else {
		if gLogger, err = zap.NewProduction(); err != nil {
			return
		}
	}
	return nil
}
