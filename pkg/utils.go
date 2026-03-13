package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/shaco-go/tomato-terminal/types"
)

func WorkspaceDir(filename string) string {
	dir, _ := os.UserHomeDir()
	workspace := fmt.Sprintf(".%s", types.ProjectName)
	_ = os.MkdirAll(filepath.Join(dir, workspace), 0755)
	return filepath.Join(dir, workspace, filename)
}
