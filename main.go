package main

import (
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
	}()
	// reader := fq.NewReader()
	// var content []string
	// content = append(content, reader.Prev()...)
	// content = append(content, reader.Next()...)
	// content = append(content, reader.Next()...)
	// fmt.Println(content)
	_, err := tea.NewProgram(page.New()).Run()
	if err != nil {
		panic(err)
	}
}
