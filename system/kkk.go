package main

import (
	//"fmt"
	//"time"
	"context"

	"log"

	system "github.com/John-Tonny/mnhost/system/proto/system"
	//"github.com/micro/go-micro"
	client "github.com/micro/go-micro/client"
)

func main() {
	//var c client.Client

	client := client.NewClient(client.Option{address: "192.168.246.182:4444"})
	//client := system.NewSystemService("go.mnhosted.srv.system", c)
	resp, err := client.GetSysStatus(context.Background(), &system.Request{})
	if err != nil {
		log.Printf("new node error: %v", err)
	} else {
		log.Printf("cpu:%f, mem:%f\n", resp.CpuPercent, resp.MemPercent)
	}

}
