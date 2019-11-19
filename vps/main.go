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
	srv := common.GetMicroServer(serviceName)

	bk := srv.Server().Options().Broker

	// 将实现服务端的 API 注册到服务端
	vps.RegisterVpsHandler(srv.Server(), handler.GetHandler(bk))

	handler.VolumeMount("52.14.4.149", "vpub$999000")

	// Run service
	log.Println("vps running")
	if err := srv.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
