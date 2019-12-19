package main

import (
	"log"

	//json "github.com/json-iterator/go"

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

	resp, err := client.RemoveVps(context.Background(), &pb.Request{
		Id: "i-054d68aab892cc420",
	})
	if err != nil {
		log.Printf("del node error: %v", err)
	} else {
		/*var msg interface{}
		if err := json.Unmarshal(resp.Mix, &msg); err != nil {
			log.Println(err)
		}
		log.Println("del node: ", msg)*/
		log.Printf("del vps:%+v\n", resp)
	}
}
