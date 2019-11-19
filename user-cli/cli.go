package main

import (
	"log"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/conf"
	pb "github.com/John-Tonny/mnhost/interface/out/user"

	"context"
)

const service = "user"

var (
	serviceName string
)

func init() {
	serviceName = config.GetServiceName(service)
}

func main() {
	srv := common.GetMicroClient(service)

	// 创建 user-service 微服务的客户端
	client := pb.NewUserService(serviceName, srv.Client())

	password := "test1"
	mobile := "1358888999"
	realname := "test one"
	idcard := "1234abcdefghhhh9875"
	name := "test1"
	rewardaddress := "12345678aaa"

	resp, err := client.Create(context.Background(), &pb.User{
		Password:      password,
		Mobile:        mobile,
		Name:          name,
		RealName:      realname,
		IdCard:        idcard,
		RewardAddress: rewardaddress,
	})
	log.Println("start")
	if err != nil {
		log.Printf("call Create error: %v", err)
	} else {
		log.Println("created: ", resp.User.Id)
	}

	allResp, err := client.GetAll(context.Background(), &pb.Request{})
	if err != nil {
		log.Printf("call GetAll error: %v", err)
	} else {
		for i, u := range allResp.Users {
			log.Printf("user_%d: %v\n", i, u)
		}
	}

	authResp, err := client.Auth(context.Background(), &pb.User{
		Mobile:   mobile,
		Password: password,
	})
	if err != nil {
		log.Printf("auth failed: %v\n", err)
	} else {
		log.Println("token: ", authResp.Token)
	}

	aa, err := client.ValidateToken(context.Background(), &pb.Token{
		//Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VyIjp7ImlkIjoiNCIsIm5hbWUiOiJ0ZXN0MSIsInBhc3N3b3JkIjoiJDJhJDEwJHo3cHBVSnBtQjNzNjdSM1NLYURtTk9lMS5IY0F0Z3BidDI2cWxqUzV1d2hQWEtRMERKZ2YyIiwibW9iaWxlIjoiMTM1ODg4ODk5OSIsInJlYWxOYW1lIjoidGVzdCBvbmUiLCJpZENhcmQiOiIxMjM0YWJjZGVmZ2hoaGg5ODc1IiwicmV3YXJkQWRkcmVzcyI6IjEyMzQ1Njc4YWFhIn0sImV4cCI6MTU3NDY1MTQyMSwiaXNzIjoiRXRoYW4uTWljcm9TZXJ2aWNlUHJhY3RpY2UudXNlciJ9.NNmVxNBflGH57uNT3Y_2t9VLUp6qJ31X1muiHADH7vs",
		Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VyIjp7ImlkIjoiMyIsInBhc3N3b3JkIjoiJDJhJDEwJGczaHVOWGwyU0d1OWEwOGxCbmNtTU94R0EzWXdDaGZrSXAyN0hSeUZ3QXZFY3k4U2gyRFpDIiwibW9iaWxlIjoiMTM1ODg4ODk5OTcifSwiZXhwIjoxNTc0NjUxMDQxLCJpc3MiOiJFdGhhbi5NaWNyb1NlcnZpY2VQcmFjdGljZS51c2VyIn0.nQdmdWvwkTzmuwXxVJS7SWN2D0TePa8eEJ-i2YGU-a8",
	})
	log.Println(aa)
	if err != nil {
		log.Printf("valid token failed: %v\n", err)
	} else {
		log.Println("valid token success")
	}
}
