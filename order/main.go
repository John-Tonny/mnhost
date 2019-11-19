package main

import (
	"github.com/micro/go-micro/util/log"
	"github.com/micro/go-micro"
	"github.com/John-Tonny/mnhost/order/handler"
	"github.com/John-Tonny/mnhost/order/subscriber"

	order "github.com/John-Tonny/mnhost/order/proto/order"
)

func main() {
	// New Service
	service := micro.NewService(
		micro.Name("go.micro.srv.order"),
		micro.Version("latest"),
	)

	// Initialise service
	service.Init()

	// Register Handler
	order.RegisterOrderHandler(service.Server(), new(handler.Order))

	// Register Struct as Subscriber
	micro.RegisterSubscriber("go.micro.srv.order", service.Server(), new(subscriber.Order))

	// Register Function as Subscriber
	micro.RegisterSubscriber("go.micro.srv.order", service.Server(), subscriber.Handler)

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
