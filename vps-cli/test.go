package main

import (
	//"fmt"
	"errors"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	json "github.com/json-iterator/go"

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

	args_len := len(os.Args)
	var starts, stops string
	var start, stop int
	var err error
	if args_len == 3 {
		starts = os.Args[1]
		stops = os.Args[2]
		start, err = strconv.Atoi(starts)
		if err != nil {
			panic(errors.New("start format err"))
		}
		stop, err = strconv.Atoi(stops)
		if err != nil {
			panic(errors.New("stop format err"))
		}
	} else if args_len == 2 {
		starts = os.Args[1]
		start, err = strconv.Atoi(starts)
		if err != nil {
			panic(errors.New("start format err"))
		}
	} else {
		panic(errors.New("add node params error"))
	}

	client := pb.NewVpsService(serviceName, srv.Client())
	for i := 0; i < 5; i++ {
		resp, err := client.CreateNode(context.Background(), &pb.Request{
			Id: starts,
		})
		if err != nil {
			log.Printf("new node error:%s-%v", starts, err)
			time.Sleep(time.Second * 2)
		} else {
			//log.Println(resp)
			var msg interface{}
			if err := json.Unmarshal(resp.Mix, &msg); err != nil {
				log.Println(err)
			}
			log.Printf("new node:%s-%+v", starts, msg)
			break
		}
	}

	if len(stops) > 0 {
		var wg sync.WaitGroup
		wg.Add(stop - start)
		for i := start + 1; i <= stop; i++ {
			go addnode(srv, i, &wg)
		}
		wg.Wait()
	}
}

func addnode(srv micro.Service, nums int, wg *sync.WaitGroup) {
	defer func() { //匿名函数捕获错误
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("ready config error:%+v\n", err)
		}
	}()

	for i := 0; i < 5; i++ {
		log.Printf("start:%d\n", nums)
		client := pb.NewVpsService(serviceName, srv.Client())
		resp, err := client.CreateNode(context.Background(), &pb.Request{
			Id: strconv.Itoa(nums),
		})
		if err != nil {
			log.Printf("new node error: %d-%v", nums, err)
			time.Sleep(time.Second * 2)
		} else {
			//log.Println(resp)
			var msg interface{}
			if err := json.Unmarshal(resp.Mix, &msg); err != nil {
				log.Println(err)
			}
			log.Printf("success new node:%d-%+v", nums, msg)
			break
		}
	}
}
