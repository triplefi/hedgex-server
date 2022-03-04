package main

import (
	"hedgex-public/config"
	"hedgex-public/daemon"
	"hedgex-public/gl"
	"hedgex-public/host"
	"hedgex-public/model"
	"hedgex-public/service"
)

func main() {
	if config.Env == "product" {
		daemon.Background("./out.log", true)
	}

	//create out and err logs in logs dir
	gl.CreateLogFiles()

	//connect to mysql database
	model.ConnectToMysql()

	//init the contracts
	gl.InitContract()

	//start contract service
	service.Start()

	//start http service
	if config.HttpPort != 0 {
		host.StartHttpServer()
	}

	//wait to exit single
	daemon.WaitForKill()
}
