package bootstrap

import (
	"fmt"
	"time"
	"github.com/shaco-go/tomato-terminal/pkg"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Logger(level zapcore.Level) {
	var cores []zapcore.Core
	for {
		cores = append(cores, levelCore(level))
		level = level + 1
		if level > zapcore.FatalLevel {
			break
		}
	}
	logger := zap.New(zapcore.NewTee(cores...))
	zap.ReplaceGlobals(logger)
}

func levelCore(level zapcore.Level) zapcore.Core {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)

	config := zapcore.NewConsoleEncoder(encoderConfig)
	ws := zapcore.AddSync(&lumberjack.Logger{
		Filename:   pkg.WorkspaceDir(fmt.Sprintf("%s.log", level.String())),
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   // days
		Compress:   true, // disabled by default
	})
	levelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l == level
	})
	return zapcore.NewCore(config, ws, levelEnabler)
}
