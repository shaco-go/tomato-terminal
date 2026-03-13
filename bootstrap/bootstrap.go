package bootstrap

import "go.uber.org/zap"

func Init() {
	Config()
	Logger(zap.DebugLevel)
}
