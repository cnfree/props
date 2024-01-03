package main

import (
	"github.com/cnfree/props/v3/kvs"
	log "github.com/sirupsen/logrus"
)

func main() {
	//获取程序运行文件所在的路径
	file := kvs.GetCurrentFilePath("api.ini", 2)
	log.Info("config file: ", file)
}
