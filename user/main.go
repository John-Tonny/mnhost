package main

import (
	"log"

	"github.com/John-Tonny/mnhost/common"

	//"github.com/micro/go-micro/util/log"

	pb "github.com/John-Tonny/mnhost/interface/out/user"
	"github.com/John-Tonny/mnhost/user/handler"
)

func main() {
	log.Println("user service start")

	srv := common.GetMicroServer("user")

	bk := srv.Server().Options().Broker
	pb.RegisterUserServiceHandler(srv.Server(), handler.GetHandler(bk))

	log.Println("user service runing")
	if err := srv.Run(); err != nil {
		log.Fatalf("user service error: %v\n", err)
	}
}
