package main

import (
	"context"
	"errors"
	"log"

	//"strconv"
	"time"

	json "github.com/json-iterator/go"
	"github.com/micro/go-micro/broker"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/conf"

	"github.com/John-Tonny/mnhost/utils"

	logPB "github.com/John-Tonny/mnhost/interface/out/log"
	mnPB "github.com/John-Tonny/mnhost/interface/out/mnmsg"
	vps "github.com/John-Tonny/mnhost/interface/out/vps"
	mnhostTypes "github.com/John-Tonny/mnhost/types"
)

const cservice = "mnmsg"
const oservice = "vps"

var (
	topic        string
	cserviceName string
	oserviceName string
	gBroker      broker.Broker
)

func init() {
	cserviceName = config.GetServiceName(cservice)
	oserviceName = config.GetServiceName(oservice)
	topic = config.GetBrokerTopic("log")
}

func main() {
	log.Println("MnMsg service start")

	srv := common.GetMicroServer(cservice)

	bk := srv.Server().Options().Broker
	gBroker = bk

	// 这里订阅了 一个 topic, 并提供接口处理
	_, err := bk.Subscribe(mnhostTypes.TOPIC_NEWNODE_START, nodeNewStart)
	if err != nil {
		log.Fatalf("new node start error: %v\n", err)
	}

	_, err = bk.Subscribe(mnhostTypes.TOPIC_NEWNODE_SUCCESS, nodeNewSuccess)
	if err != nil {
		log.Fatalf("new node success error: %v\n", err)
	}

	_, err = bk.Subscribe(mnhostTypes.TOPIC_NEWNODE_FAIL, nodeNewFail)
	if err != nil {
		log.Fatalf("new node fail error: %v\n", err)
	}

	_, err = bk.Subscribe(mnhostTypes.TOPIC_NEWNODE_STOP, nodeNewStop)
	if err != nil {
		log.Fatalf("new node stop error: %v\n", err)
	}

	_, err = bk.Subscribe(mnhostTypes.TOPIC_DELNODE_START, nodeDelStart)
	if err != nil {
		log.Fatalf("del node start error: %v\n", err)
	}

	_, err = bk.Subscribe(mnhostTypes.TOPIC_DELNODE_SUCCESS, nodeDelSuccess)
	if err != nil {
		log.Fatalf("del node success error: %v\n", err)
	}

	_, err = bk.Subscribe(mnhostTypes.TOPIC_DELNODE_FAIL, nodeDelFail)
	if err != nil {
		log.Fatalf("del node fail error: %v\n", err)
	}

	_, err = bk.Subscribe(mnhostTypes.TOPIC_DELNODE_STOP, nodeDelStop)
	if err != nil {
		log.Fatalf("del node stop error: %v\n", err)
	}

	log.Println("MnMsg service runing")
	if err = srv.Run(); err != nil {
		log.Fatalf("srv run error: %v\n", err)
	}
}

func nodeNewStart(pub broker.Event) error {
	log.Println("new node start")
	userId := pub.Message().Header["user_id"]
	method := "nodeNewStart"
	var msg *mnPB.MnMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		pubErrMsg(userId, method, utils.JSON_DATAERR, "", mnhostTypes.TOPIC_NEWNODE_FAIL)
		return err
	}
	sId := (*msg).MsgId

	srv := common.GetMicroClient(oservice)
	// 创建 user-service 微服务的客户端
	client := vps.NewVpsService(oserviceName, srv.Client())

	retrys := 0
	for {
		retrys++
		if retrys > 10 {
			pubErrMsg(userId, method, utils.TIMEOUT_VPS, sId, mnhostTypes.TOPIC_NEWNODE_FAIL)
			return errors.New("timeout")
		}
		resp, err := client.CreateNode(context.Background(), &vps.Request{
			Id: sId,
		})
		if err == nil {
			if resp.Errno == utils.RECODE_OK {
				pubMsg(userId, method, mnhostTypes.TOPIC_NEWNODE_STOP, sId)
				return nil
			}
		}
		time.Sleep(time.Second)
	}

	pubErrMsg(userId, method, utils.RECODE_SERVERERR, sId, mnhostTypes.TOPIC_NEWNODE_FAIL)

	log.Println("new node finish")
	return nil
}

func nodeNewSuccess(pub broker.Event) error {
	log.Println("new nodesuccess start")
	var msg *mnPB.MnMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}
	userId := pub.Message().Header["user_id"]
	orderId := (*msg).MsgId
	log.Printf("new nodesuccess finish, userId:%v,orderId:%d\n", userId, orderId)
	return nil
}

func nodeNewFail(pub broker.Event) error {
	log.Printf("new nodefail start")
	var msg *mnPB.MnErrMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}
	userId := pub.Message().Header["user_id"]
	log.Printf("new nodefail finish, userId:%v,failmsg:%v\n", userId, msg)
	return nil
}

func nodeNewStop(pub broker.Event) error {
	log.Printf("new nodestop start")
	var msg *mnPB.MnMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}
	userId := pub.Message().Header["user_id"]
	log.Printf("new nodestop finish, userId:%v\n", userId)
	return nil
}

func nodeDelStart(pub broker.Event) error {
	log.Println("del node start")
	userId := pub.Message().Header["user_id"]
	method := "nodeDelStart"
	var msg *mnPB.MnMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		pubErrMsg(userId, method, utils.JSON_DATAERR, "", mnhostTypes.TOPIC_DELNODE_FAIL)
		return err
	}
	sId := (*msg).MsgId

	srv := common.GetMicroClient(oservice)

	// 创建 user-service 微服务的客户端
	client := vps.NewVpsService(oserviceName, srv.Client())

	resp, err := client.RemoveNode(context.Background(), &vps.Request{
		Id: sId,
	})
	if err != nil {
		pubErrMsg(userId, method, utils.RECODE_SERVERERR, sId, mnhostTypes.TOPIC_DELNODE_FAIL)
		return err
	}

	if resp.Errno == utils.RECODE_OK {
		pubMsg(userId, method, mnhostTypes.TOPIC_DELNODE_STOP, sId)
		log.Println("del node start finish, userId:%v", userId)
		return nil
	}
	pubErrMsg(userId, method, utils.RECODE_SERVERERR, sId, mnhostTypes.TOPIC_DELNODE_FAIL)

	log.Println("del node finish")
	return nil
}

func nodeDelSuccess(pub broker.Event) error {
	log.Println("del nodesuccess")
	var msg *mnPB.MnMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}
	userId := pub.Message().Header["user_id"]
	nodeId := (*msg).MsgId
	log.Printf("del node success finish, userId:%v,nodeId:%v\n", userId, nodeId)
	return nil
}

func nodeDelFail(pub broker.Event) error {
	log.Printf("del nodefail start")
	var msg *mnPB.MnErrMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}
	userId := pub.Message().Header["user_id"]
	log.Printf("del nodefail finish, userId:%v\n", userId)
	return nil
}

func nodeDelStop(pub broker.Event) error {
	log.Printf("del nodestop start")
	var msg *mnPB.MnMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}

	userId := pub.Message().Header["user_id"]
	log.Printf("del nodestop finish, userId:%v\n", userId)
	return nil
}

// 发送vps
func pubMsg(userID, method, topic string, msgId string) error {
	log.Printf("start pub msg, topic:%v, msgId:%d\n", topic, msgId)
	msg := mnPB.MnMsg{
		MsgId: msgId,
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	data := &broker.Message{
		Header: map[string]string{
			"user_id": userID,
		},
		Body: body,
	}

	if err := gBroker.Publish(topic, data); err != nil {
		log.Printf("finish pub msg, topic:%v, errmsg:%d\n", topic, err)
		return err
	}
	log.Printf("finish pub msg, topic:%v, msgId:%d\n", topic, msgId)
	return nil
}

func pubErrMsg(userID, method, errno, msg, topic string) error {
	log.Printf("start pub err msg, topic:%v, errmsg:%v\n", msg)
	errmsg := mnPB.MnErrMsg{
		Method: method,
		Origin: oserviceName,
		Errno:  errno,
		Errmsg: utils.RecodeText(errno),
		Msg:    msg,
	}
	body, err := json.Marshal(errmsg)
	if err != nil {
		return err
	}

	data := &broker.Message{
		Header: map[string]string{
			"user_id": userID,
		},
		Body: body,
	}

	if err := gBroker.Publish(topic, data); err != nil {
		log.Printf("fail pub msg, topic:%v, errmsg:%v\n", err)
		return err
	}
	log.Printf("finsih pub err msg, topic:%v, errmsg:%v\n", msg)
	return nil
}

func pubLog(userID, method, msg string) error {
	log.Printf("start pub log msg, method %v, msg:%v\n", method, msg)
	logPB := logPB.Log{
		Method: method,
		Origin: oserviceName,
		Msg:    msg,
	}
	body, err := json.Marshal(logPB)
	if err != nil {
		return err
	}

	data := &broker.Message{
		Header: map[string]string{
			"user_id": userID,
		},
		Body: body,
	}

	if err := gBroker.Publish(topic, data); err != nil {
		log.Printf("fail pub log, method:%v, errmsg:%v\n", method, err)
		return err
	}
	log.Printf("finsih pub log msg, method:%v, msg:%v\n", method, msg)
	return nil
}
