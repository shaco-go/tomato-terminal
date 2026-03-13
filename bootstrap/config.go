package bootstrap

import (
	"github.com/shaco-go/tomato-terminal/config"
	"github.com/shaco-go/tomato-terminal/pkg"
	"github.com/shaco-go/tomato-terminal/types"
)

func Config() {
	config.Conf.Load()
	if err := types.LoadCharMap(pkg.WorkspaceDir("char.json")); err != nil {
		panic(err)
	}
}
