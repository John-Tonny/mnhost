package main

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	json "github.com/json-iterator/go"
	"github.com/micro/go-micro/broker"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/conf"

	"github.com/John-Tonny/mnhost/utils"

	logPB "github.com/John-Tonny/mnhost/interface/out/log"
	mnPB "github.com/John-Tonny/mnhost/interface/out/mnmsg"
	vps "github.com/John-Tonny/mnhost/interface/out/vps"
)

const cservice = "mnmsg"
const oservice = "vps"
const topic_newnode_success = "Vircle.Mnhost.Topic.NodeNew.Success"
const topic_newnode_fail = "Vircle.Mnhost.Topic.NodeNew.Fail"
const topic_newnode_start = "Vircle.Mnhost.Topic.NodeNew.Start"
const topic_newnode_stop = "Vircle.Mnhost.Topic.NodeNew.Stop"

const topic_delnode_success = "Vircle.Mnhost.Topic.NodeDel.Success"
const topic_delnode_fail = "Vircle.Mnhost.Topic.NodeDel.Fail"
const topic_delnode_start = "Vircle.Mnhost.Topic.NodeDel.Start"
const topic_delnode_stop = "Vircle.Mnhost.Topic.NodeDel.Stop"

const topic_expandvolume_success = "Vircle.Mnhost.Topic.ExpandVolume.Success"
const topic_expandvolume_fail = "Vircle.Mnhost.Topic.ExpandVolume.Fail"
const topic_expandvolume_start = "Vircle.Mnhost.Topic.ExpandVolume.Start"
const topic_expandvolume_stop = "Vircle.Mnhost.Topic.ExpandVolume.Stop"

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
	_, err := bk.Subscribe(topic_newnode_start, nodeNewStart)
	if err != nil {
		log.Fatalf("new node start error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_newnode_success, nodeNewSuccess)
	if err != nil {
		log.Fatalf("new node success error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_newnode_fail, nodeNewFail)
	if err != nil {
		log.Fatalf("new node fail error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_newnode_stop, nodeNewStop)
	if err != nil {
		log.Fatalf("new node stop error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_delnode_start, nodeDelStart)
	if err != nil {
		log.Fatalf("del node start error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_delnode_success, nodeDelSuccess)
	if err != nil {
		log.Fatalf("del node success error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_delnode_fail, nodeDelFail)
	if err != nil {
		log.Fatalf("del node fail error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_delnode_stop, nodeDelStop)
	if err != nil {
		log.Fatalf("del node stop error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_expandvolume_start, expandVolumeStart)
	if err != nil {
		log.Fatalf("del node success error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_expandvolume_success, expandVolumeSuccess)
	if err != nil {
		log.Fatalf("del node success error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_expandvolume_success, expandVolumeFail)
	if err != nil {
		log.Fatalf("del node fail error: %v\n", err)
	}

	_, err = bk.Subscribe(topic_expandvolume_stop, expandVolumeStop)
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
		pubErrMsg(userId, method, utils.JSON_DATAERR, "", topic_newnode_fail)
		return err
	}
	id := (*msg).Id

	srv := common.GetMicroClient(oservice)
	// 创建 user-service 微服务的客户端
	client := vps.NewVpsService(oserviceName, srv.Client())

	sId := strconv.FormatInt(id, 10)

	retrys := 0
	for {
		retrys++
		if retrys > 10 {
			pubErrMsg(userId, method, utils.TIMEOUT_VPS, sId, topic_newnode_fail)
			return errors.New("timeout")
		}
		resp, err := client.NewNode(context.Background(), &vps.Request{
			Id: id,
		})
		if err == nil {
			if resp.Errno == utils.RECODE_OK {
				pubMsg(userId, method, topic_newnode_stop, id)
				return nil
			}
		}
		time.Sleep(time.Second)
	}

	pubErrMsg(userId, method, utils.RECODE_SERVERERR, sId, topic_newnode_fail)

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
	orderId := (*msg).Id
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
		pubErrMsg(userId, method, utils.JSON_DATAERR, "", topic_newnode_fail)
		return err
	}
	id := (*msg).Id

	srv := common.GetMicroClient(oservice)

	// 创建 user-service 微服务的客户端
	client := vps.NewVpsService(oserviceName, srv.Client())

	sId := strconv.FormatInt(id, 10)
	resp, err := client.DelNode(context.Background(), &vps.Request{
		Id: id,
	})
	if err != nil {
		pubErrMsg(userId, method, utils.RECODE_SERVERERR, sId, topic_newnode_fail)
		return err
	}

	if resp.Errno == utils.RECODE_OK {
		pubMsg(userId, method, topic_delnode_stop, id)
		log.Println("del node start finish, userId:%v", userId)
		return nil
	}
	pubErrMsg(userId, method, utils.RECODE_SERVERERR, sId, topic_newnode_fail)

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
	nodeId := (*msg).Id
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

func expandVolumeStart(pub broker.Event) error {
	log.Println("expandVolumeStart start")

	log.Println("expandVolumeStart finish")
	return nil
}

func expandVolumeSuccess(pub broker.Event) error {
	log.Println("expandVolumeSuccess")
	var msg *mnPB.MnMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}
	userId := pub.Message().Header["user_id"]
	log.Printf("expandVolumeSuccess finish, userId:%v\n", userId)
	return nil
}

func expandVolumeFail(pub broker.Event) error {
	log.Printf("expandVolumeFail start")
	var msg *mnPB.MnErrMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}
	userId := pub.Message().Header["user_id"]
	log.Printf("expandVolumeFail finish, userId:%v\n", userId)
	return nil
}

func expandVolumeStop(pub broker.Event) error {
	log.Printf("expandVolumeStop start")
	var msg *mnPB.MnMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}
	log.Println("expandVolumetop finish")
	return nil
}

// 发送vps
func pubMsg(userID, method, topic string, msgId int64) error {
	log.Printf("start pub msg, topic:%v, msgId:%d\n", topic, msgId)
	msg := mnPB.MnMsg{
		Id: msgId,
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
