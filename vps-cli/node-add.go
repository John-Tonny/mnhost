package main

import (
	"log"

	//json "github.com/json-iterator/go"

	"github.com/micro/go-micro"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/conf"
	pb "github.com/John-Tonny/mnhost/interface/out/vps"

	"context"
)

const service = "vps"

var (
	serviceName string
)

func init() {
	serviceName = config.GetServiceName(service)
}

func main() {
	srv := common.GetMicroClient(service)

	// 创建 user-service 微服务的客户端

	client := pb.NewVpsService(serviceName, srv.Client())
	resp, err := client.CreateNode(context.Background(), &pb.Request{
		Id: "4",
	})
	if err != nil {
		log.Printf("new node error: %v", err)
	} else {
		/*var msg interface{}
		if err := json.Unmarshal(resp.Mix, &msg); err != nil {
			log.Println(err)
		}
		log.Println("new node: ", msg)*/
		log.Println(resp)
	}

	/*for i := 0; i < 1; i++ {
		go addnode(srv)
	}

	for {

	}*/

}

func addnode(srv micro.Service) {
	log.Println("aaa")
	client := pb.NewVpsService(serviceName, srv.Client())
	resp, err := client.CreateNode(context.Background(), &pb.Request{
		Id: "1",
	})
	if err != nil {
		log.Printf("new node error: %v", err)
	} else {
		/*var msg interface{}
		if err := json.Unmarshal(resp.Mix, &msg); err != nil {
			log.Println(err)
		}
		log.Println("new node: ", msg)*/
		log.Println(resp)
	}
	log.Println("bbb")

}
