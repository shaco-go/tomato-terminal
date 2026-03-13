package main

import (
	"fmt"

	"github.com/shaco-go/tomato-terminal/bootstrap"
	"github.com/shaco-go/tomato-terminal/config"
	"github.com/shaco-go/tomato-terminal/page"

	tea "charm.land/bubbletea/v2"
	"go.uber.org/zap"
)

func main() {
	bootstrap.Init()
	defer func() {
		if err := recover(); err != nil {
			zap.L().Fatal("panic", zap.Any("err", err))
		}
		config.Conf.Flush()
		// 清除并重置光标
		fmt.Print("\033[3A") // 向上3行
		fmt.Print("\033[J")  // 清除从光标到屏幕底部的所有内容
	}()

	_, err := tea.NewProgram(page.New()).Run()
	if err != nil {
		panic(err)
	}
}
