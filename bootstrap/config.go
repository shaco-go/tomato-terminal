package bootstrap

import "github.com/shaco-go/tomato-terminal/config"

func Config() {
	config.Conf.Load()

}
