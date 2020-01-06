package main

import (
	"log"

	//json "github.com/json-iterator/go"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/conf"
	pb "github.com/John-Tonny/mnhost/interface/out/monitor"

	"context"
)

const service = "monitor"

var (
	serviceName string
)

func init() {
	serviceName = config.GetServiceName(service)
}

func main() {
	srv := common.GetMicroClient(service)

	// 创建 user-service 微服务的客户端
	client := pb.NewMonitorService(serviceName, srv.Client())

	resp, err := client.UpdateService(context.Background(), &pb.UpdateRequest{
		CoinName:   "snowreg",
		DockerName: "johntonny2019/snowreg:v1.01",
	})
	if err != nil {
		log.Printf("new node error: %v", err)
	} else {
		//var msg interface{}
		//if err := json.Unmarshal(resp.Mix, &msg); err != nil {
		//	log.Println(err)
		//}
		log.Println("update: ", resp)
	}
}
