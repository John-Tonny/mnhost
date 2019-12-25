package main

import (
	"github.com/John-Tonny/mnhost/system/handler"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/util/log"

	system "github.com/John-Tonny/mnhost/system/proto/system"
)

func main() {
	// New Service
	service := micro.NewService(
		micro.Name("go.mnhosted.srv.system"),
		micro.Version("latest"),
		micro.Address("0.0.0.0:4444"),
	)

	// Initialise service
	service.Init()

	// Register Handler
	system.RegisterSystemHandler(service.Server(), new(handler.System))

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
