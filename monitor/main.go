package main

import (
	"log"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/monitor/handler"

	monitor "github.com/John-Tonny/mnhost/interface/out/monitor"
)

const serviceName = "monitor"

func main() {
	log.Println("monitor start")

	handler.Init()

	srv := common.GetMicroServer(serviceName)

	bk := srv.Server().Options().Broker

	// 将实现服务端的 API 注册到服务端
	monitor.RegisterMonitorHandler(srv.Server(), handler.GetHandler(bk))

	// Run service
	log.Println("monitor running")
	if err := srv.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
