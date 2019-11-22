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

	resp, err := client.ExpandVolume(context.Background(), &pb.VolumeRequest{
		VolumeId: "vol-0179f173d400d2629",
		Size:     2,
	})
	if err != nil {
		log.Printf("expand volume error: %v\n", err)
	} else {
		log.Printf("expand volume success %v\n", resp)
	}
}
