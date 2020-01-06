package main

import (
	//"fmt"
	"log"
	"strconv"
	"sync"

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
		Id: "14",
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

	/*var wg sync.WaitGroup
	wg.Add(5)
	for i := 21; i < 26; i++ {
		go addnode(srv, i, &wg)
	}
	wg.Wait()
	*/
}

func addnode(srv micro.Service, nums int, wg *sync.WaitGroup) {
	defer func() { //匿名函数捕获错误
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("ready config error:%+v\n", err)
		}
	}()

	log.Printf("aaa:%d\n", nums)
	client := pb.NewVpsService(serviceName, srv.Client())
	_, err := client.CreateNode(context.Background(), &pb.Request{
		Id: strconv.Itoa(nums),
	})
	if err != nil {
		log.Printf("new node error: %v-%d", err, nums)
	} else {
		/*var msg interface{}
		if err := json.Unmarshal(resp.Mix, &msg); err != nil {
			log.Println(err)
		}
		log.Println("new node: ", msg)*/
		//log.Println(resp)
		log.Printf("bbb:%d\n", nums)
	}
}
