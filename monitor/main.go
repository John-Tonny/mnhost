package main

import (
	"log"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/monitor/handler"

	monitor "github.com/John-Tonny/mnhost/interface/out/monitor"
	mnhostTypes "github.com/John-Tonny/mnhost/types"
)

const serviceName = "monitor"

func main() {
	log.Println("monitor start")

	handler.Init()

	srv := common.GetMicroServer(serviceName)

	bk := srv.Server().Options().Broker

	// 这里订阅了 一个 topic, 并提供接口处理
	_, err := bk.Subscribe(mnhostTypes.TOPIC_NEWVPS_FAIL, handler.VpsNewFail)
	if err != nil {
		log.Fatalf("new vps fail subscribe error: %v\n", err)
	}

	_, err = bk.Subscribe(mnhostTypes.TOPIC_NEWVPS_SUCCESS, handler.VpsNewSuccess)
	if err != nil {
		log.Fatalf("new vps success subscribe error: %v\n", err)
	}

	//handler.VpsNew("manager")

	// 将实现服务端的 API 注册到服务端
	monitor.RegisterMonitorHandler(srv.Server(), handler.GetHandler(bk))

	// Run service
	log.Println("monitor running")
	if err := srv.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
