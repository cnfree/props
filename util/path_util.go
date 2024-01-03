package util

import (
	"os"
	"path/filepath"
)

// GetExecuteFilePath 获取可执行文件所在路径
func GetExecuteFilePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath
}
