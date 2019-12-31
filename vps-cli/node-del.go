package main

import (
	//"fmt"
	"log"
	"strconv"

	//"time"

	//"strings"
	"sync"

	//json "github.com/json-iterator/go"

	"github.com/micro/go-micro"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/conf"
	pb "github.com/John-Tonny/mnhost/interface/out/vps"

	//mnhostTypes "github.com/John-Tonny/mnhost/types"

	//"github.com/docker/docker/api/types"
	//"github.com/docker/docker/api/types/filters"
	//"github.com/aws/aws-sdk-go/aws"

	"context"
	//uec2 "github.com/John-Tonny/micro/vps/amazon"
)

const service = "vps"

var (
	serviceName string
)

func init() {
	serviceName = config.GetServiceName(service)
}

func main() {
	/*c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		log.Fatalf("init: client error:%+v!\n", err)
	}

	results, err := c.GetDescribeInstance([]string{})
	if err != nil {
		log.Fatalf("init: desc instance error:%+v!\n", err)
	}
	for _, result := range results.Reservations {
		fmt.Printf("%s\n", aws.StringValue(result.Instances[0].InstanceId))
		fmt.Printf("%s\n", aws.StringValue(result.Instances[0].Placement.AvailabilityZone))
		fmt.Printf("%s\n", aws.StringValue(result.Instances[0].PrivateIpAddress))
		fmt.Printf("%s\n", aws.StringValue(result.Instances[0].PublicIpAddress))
	}

	publicIp, privateIp, err := common.GetVpsIp("cluster1")
	if err != nil {
		fmt.Println(err)
	}
	log.Println(publicIp)
	mc, _, err := common.DockerNewClient(publicIp, privateIp)
	if err != nil {
		fmt.Println(err)
	}
	defer mc.Close()

	//for {
	resp, _, err := mc.ServiceInspectWithRaw(context.Background(), "dash10000")
	if err != nil {
		fmt.Printf("err:+v\n", err)
		//continue
	}
	fmt.Printf("resp:%+v\n", resp)

	//time.Sleep(time.Second * 20)
	//}*/

	/*err = mc.ServiceRemoveA("dash", 10002)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("finish remove service %s\n", "*****")*/

	/*resp1, err := mc.Info(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", resp1)

	aa, err := mc.DescribeInstanceStatusA("i-05659b3bf477114ee")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("aa:%+v\n", aa)

	client := common.SshNewClient(publicIp, privateIp, "vpub$999000")
	if client == nil {
		fmt.Println(client)
	}
	defer client.Close()

	cmd := fmt.Sprintf("docker service ps %s | awk '{if($5==\"Running\"){print $4;}}'", "dash10000")
	log.Println(cmd)
	result, err := client.Execute(cmd)
	if err != nil {
		fmt.Println(err)
	}
	log.Println(result)
	hostName := fmt.Sprintf("%s", result.Stdout())
	log.Printf("len1:%d", len(hostName))
	hostName = strings.Replace(hostName, " ", "", -1)
	log.Printf("len2:%d", len(hostName))
	hostName = strings.Replace(hostName, "\n", "", -1)
	log.Printf("len3:%d", len(hostName))
	hostName = strings.Replace(hostName, "\r", "", -1)
	log.Printf("len4:%d", len(hostName))
	log.Printf("hostname:%s", hostName)

	log.Printf("%s", "55")
	log.Printf("%s", "66")

	//nodeName := fmt.Sprintf("%s%d", "dash", 8000)
	//service1, err := mc.ServiceInspectA(nodeName)
	f := filters.NewArgs()
	f.Add("name", hostName)
	result1, err := mc.NodeListA(types.NodeListOptions{
		Filters: f,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("nums:%d\n", len(result1))
	fmt.Printf("%+v\n", result)
	//fmt.Println(service1.ID)
	*/

	srv := common.GetMicroClient(service)

	// 创建 user-service 微服务的客户端
	client1 := pb.NewVpsService(serviceName, srv.Client())

	start := 223
	//stop := 220
	starts := strconv.Itoa(start)
	resp, err := client1.RemoveNode(context.Background(), &pb.Request{
		Id: starts,
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
	/*
		var wg sync.WaitGroup
		wg.Add(9)
		for i := start + 1; i <= stop; i++ {
			id := strconv.Itoa(i)
			go delnode(id, srv, &wg)
		}
		wg.Wait()
	*/
}

func delnode(id string, srv micro.Service, wg *sync.WaitGroup) {
	defer func() { //匿名函数捕获错误
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("ready config error:%+v\n", err)
		}
	}()

	log.Printf("start:%s\n", id)
	// 创建 user-service 微服务的客户端
	client1 := pb.NewVpsService(serviceName, srv.Client())

	_, err := client1.RemoveNode(context.Background(), &pb.Request{
		Id: id,
	})

	if err != nil {
		log.Printf("new node error:%d---%v", id, err)
	} else {
		/*var msg interface{}
		if err := json.Unmarshal(resp.Mix, &msg); err != nil {
			log.Println(err)
		}
		log.Println("new node: ", msg)*/
		//log.Println(resp)
		log.Printf("stop:%s\n", id)
	}
}
