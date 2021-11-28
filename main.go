package main

import (
	"context"
	"hedgex-server/config"
	"hedgex-server/gl"
	"hedgex-server/host"
	"hedgex-server/service"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	service.Start()

	if config.Service == 0 {
		//start host server
		go host.StartHttpServer()
	}

	waitForKill()

	if gl.HttpServer != nil {
		gl.HttpServer.Shutdown(context.Background())
	}

	service.Stop()

	gl.OutLogger.Close()
}

func waitForKill() {
	if pid := syscall.Getpid(); pid != 1 {
		ioutil.WriteFile("process.pid", []byte(strconv.Itoa(pid)), 0777)
		ioutil.WriteFile("stop.sh", []byte("kill `cat process.pid`"), 0777)
		defer os.Remove("process.pid")
		defer os.Remove("stop.sh")
	}
	ch := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	s := <-ch
	gl.OutLogger.Info("process stop. %d", s)
}
