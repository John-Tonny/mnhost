package main

import (
	//"log"
	"os"
	"time"

	json "github.com/json-iterator/go"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/conf"
	pb "github.com/John-Tonny/mnhost/interface/out/log"

	"github.com/micro/go-micro/broker"

	log "github.com/sirupsen/logrus"
)

const service = "log"

var (
	topic string
)

func logInit() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	file := time.Now().Format("20060102") + ".txt" //文件名
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if nil != err {
		panic(err)
	}
	log.SetOutput(logFile)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func init() {
	topic = config.GetBrokerTopic(service)

	logInit()
}

func main() {
	log.Info("log service start")
	srv := common.GetMicroServer(service)

	bk := srv.Server().Options().Broker

	// 这里订阅了 一个 topic, 并提供接口处理
	_, err := bk.Subscribe(topic, subLog)
	if err != nil {
		log.Fatalf("sub error: %v\n", err)
	}

	log.Info("log service runing")
	if err = srv.Run(); err != nil {
		log.Fatalf("srv run error: %v\n", err)
	}
}

func subLog(pub broker.Event) error {
	var logPB *pb.Log
	if err := json.Unmarshal(pub.Message().Body, &logPB); err != nil {
		return err
	}
	log.Printf("[Log]: user_id: %s,  Msg: %v\n", pub.Message().Header["user_id"], logPB)
	return nil
}
