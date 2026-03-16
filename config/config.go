package config

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/shaco-go/tomato-terminal/pkg"
)

var Conf = &conf{}
var filename = pkg.WorkspaceDir("config.json")

type conf struct {
	ItemID       string
	Cursor       int
	Line         int
	MaxRuneCount int
	Margin       int
	Cookie       []*http.Cookie
}

// Load 加载
func (c *conf) Load() {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()

	raw, _ := io.ReadAll(file)
	_ = json.Unmarshal(raw, c)
}

// Flush 刷新到磁盘
func (c *conf) Flush() {
	raw, _ := json.Marshal(c)
	_ = os.WriteFile(filename, raw, 0644)
}
