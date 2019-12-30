package main

import (
	"log"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/vps/handler"

	vps "github.com/John-Tonny/mnhost/interface/out/vps"
)

const serviceName = "vps"

func main() {
	log.Println("vps start")

	//devicename, _ := handler.GetDeviceName("172.31.46.50")
	//log.Printf("dev:%s\n", devicename)

	handler.Init("cluster1")

	srv := common.GetMicroServer(serviceName)

	bk := srv.Server().Options().Broker

	// 将实现服务端的 API 注册到服务端
	vps.RegisterVpsHandler(srv.Server(), handler.GetHandler(bk))

	// Run service
	log.Println("vps running")
	if err := srv.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
