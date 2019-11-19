package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dynport/gossh"
	"github.com/go-ini/ini"
	"github.com/pytool/ssh"

	"github.com/astaxie/beego/orm"
	"github.com/micro/go-micro/broker"

	uec2 "github.com/John-Tonny/micro/vps/amazon"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	logPB "github.com/John-Tonny/mnhost/interface/out/log"
	mnPB "github.com/John-Tonny/mnhost/interface/out/mnmsg"
	vps "github.com/John-Tonny/mnhost/interface/out/vps"

	"github.com/John-Tonny/mnhost/conf"
	"github.com/John-Tonny/mnhost/model"
	"github.com/John-Tonny/mnhost/utils"
)

type Vps struct {
	Broker broker.Broker
}

type VpsInfo struct {
	instanceId      string
	regionName      string
	allocationId    string
	allocationState bool
	publicIp        string
	volumeId        string
	volumeState     bool
}

type Volume struct {
	Mountpoint string `json:"Mountpoint"`
	Name       string `json:"Name"`
}

const service = "vps"
const topic_newnode_success = "Vircle.Mnhost.Topic.NodeNew.Success"
const topic_newnode_fail = "Vircle.Mnhost.Topic.NodeNew.Fail"
const topic_newnode_start = "Vircle.Mnhost.Topic.NodeNew.Start"

const topic_delnode_success = "Vircle.Mnhost.Topic.NodeDel.Success"
const topic_delnode_fail = "Vircle.Mnhost.Topic.NodeDel.Fail"
const topic_delnode_start = "Vircle.Mnhost.Topic.NodeDel.Start"

const topic_expandvolume_success = "Vircle.Mnhost.Topic.ExpandVolume.Success"
const topic_expandvolume_fail = "Vircle.Mnhost.Topic.ExpandVolume.Fail"
const topic_expandvolume_start = "Vircle.Mnhost.Topic.ExpandVolume.Start"

const ssh_password = "vpub$999000"
const rpc_user = "vpub"
const rpc_password = "vpub999000"
const port_from = 19900
const port_to = 20000

const order_tablename = "order_node"
const order_id = "Id"
const test_volume_size = 1

const provider_name = "amazon"
const group_name = "vcl-mngroup"
const group_desc = "basic masternode group"
const key_pair_name = "vcl-keypair"
const device_name = "xvdk"
const device_name1 = "xvdk1"
const core_nums = 1
const memory_size = 1
const masternode_max_nums = 3

const mount_point = "/var/lib/docker/volumes"

const vps_retrys = 3
const node_retrys = 3
const instance_retrys = 30
const volume_retrys = 10

var (
	topic       string
	serviceName string
	version     string
)

func init() {
	topic = config.GetBrokerTopic("log")
	serviceName = config.GetServiceName(service)
	version = config.GetVersion(service)
	if version == "" {
		version = "latest"
	}
}

func GetHandler(bk broker.Broker) *Vps {
	return &Vps{
		Broker: bk,
	}
}

// 发送vps
func (e *Vps) pubMsg(userID, topic string, msgId int64) error {
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

	if err := e.Broker.Publish(topic, data); err != nil {
		log.Printf("pub msg failed:%v\n", err)
		return err
	}
	return nil
}

func (e *Vps) pubErrMsg(userID, method, errno, msg, topic string) error {
	errmsg := mnPB.MnErrMsg{
		Method: method,
		Origin: serviceName,
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

	if err := e.Broker.Publish(topic, data); err != nil {
		log.Printf("pub err msg failed:%v\n", err)
		return err
	}
	return nil
}

func (e *Vps) pubLog(userID, method, msg string) error {
	logPB := logPB.Log{
		Method: method,
		Origin: serviceName,
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

	if err := e.Broker.Publish(topic, data); err != nil {
		log.Printf("pub log msg failed:%v\n", err)
		return err
	}
	return nil
}

func (e *Vps) NewNode(ctx context.Context, req *vps.Request, rsp *vps.Response) error {
	log.Println("new node request")
	orderId := req.Id

	var order models.OrderNode
	o := orm.NewOrm()
	qs := o.QueryTable(order_tablename)
	err := qs.Filter(order_id, orderId).One(&order)
	if err != nil {
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	go e.processNewNode(int64(order.Id))

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	jorder, err := json.Marshal(order)
	rsp.Mix = jorder

	log.Println("new node request success")
	return nil
}

func (e *Vps) DelNode(ctx context.Context, req *vps.Request, rsp *vps.Response) error {
	log.Println("del node request")
	nodeId := req.Id

	var node models.Node
	o := orm.NewOrm()
	qs := o.QueryTable("node")
	err := qs.Filter("id", nodeId).One(&node)
	if err != nil {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}

	go e.processDelNode(int64(node.Id))

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	jnode, err := json.Marshal(node)
	rsp.Mix = jnode

	return nil
}

func (e *Vps) ExpandVolume(ctx context.Context, req *vps.VolumeRequest, rsp *vps.Response) error {
	log.Println("expand volume request")
	volumeId := req.VolumeId
	size := req.Size

	var vps models.Vps
	o := orm.NewOrm()
	qs := o.QueryTable("vps")
	err := qs.Filter("volumeId", volumeId).One(&vps)
	if err != nil {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}

	if size == 0 {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}

	go e.processExpandVolume("", volumeId, size, vps.IpAddress, ssh_password)

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)
	log.Println("success expand volume request")
	return nil
}

func (e *Vps) processNewNode(orderId int64) error {
	log.Printf("process new node from orderid %d\n", orderId)

	var order models.OrderNode
	o := orm.NewOrm()
	qs := o.QueryTable(order_tablename)
	err := qs.Filter(order_id, orderId).One(&order)
	if err != nil {
		e.pubErrMsg("", "newnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
		return err
	}

	userId := strconv.Itoa(order.User.Id)

	var nvpsInfo VpsInfo
	var vpsInfo *VpsInfo
	var vps models.Vps
	nvps := models.Vps{}
	o = orm.NewOrm()
	qs = o.QueryTable("vps")
	err = qs.Filter("usable_nodes__gt", 0).One(&vps)
	retrys := 0
	if err != nil {
		for { //循环
			retrys++
			if retrys > vps_retrys {
				e.pubErrMsg(userId, "newnode", utils.TIMEOUT_VPS, err.Error(), topic_newnode_fail)
				return nil
			}
			vpsInfo, err = newVps("ami-0b0426f6bc13cbfe4", "us-east-2", "", test_volume_size, nvpsInfo)
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			nvpsInfo.allocationId = vpsInfo.allocationId
			nvpsInfo.allocationState = vpsInfo.allocationState
			nvpsInfo.instanceId = vpsInfo.instanceId
			nvpsInfo.publicIp = vpsInfo.publicIp
			nvpsInfo.regionName = vpsInfo.regionName
			nvpsInfo.volumeId = vpsInfo.volumeId
			nvpsInfo.volumeState = vpsInfo.volumeState
			break
		}

		nvps.AllocateId = vpsInfo.allocationId
		nvps.InstanceId = vpsInfo.instanceId
		nvps.VolumeId = vpsInfo.volumeId
		nvps.ProviderName = provider_name
		nvps.Cores = core_nums
		nvps.Memory = memory_size
		nvps.KeyPairName = key_pair_name
		nvps.MaxNodes = masternode_max_nums
		nvps.UsableNodes = masternode_max_nums
		nvps.SecurityGroupName = group_name
		nvps.RegionName = vpsInfo.regionName
		nvps.IpAddress = vpsInfo.publicIp
		o = orm.NewOrm()
		_, err = o.Insert(&nvps)
		if err != nil {
			e.pubErrMsg(userId, "newnode", utils.RECODE_INSERTERR, err.Error(), topic_newnode_fail)
			return nil
		}
		o = orm.NewOrm()
		qs = o.QueryTable("vps")
		err = qs.Filter("usable_nodes__gt", 0).One(&vps)
		if err != nil {
			e.pubErrMsg(userId, "newnode", utils.RECORD_SYSTEMERR, err.Error(), topic_newnode_fail)
			return err
		}
	}

	retrys = 0
	for { //循环
		retrys++
		if retrys > node_retrys {
			e.pubErrMsg(userId, "newnode", utils.TIMEOUT_VOLUME, "", topic_newnode_fail)
			return nil
		}
		err = newNode(vps.Id, order.CoinName, vps.IpAddress, ssh_password)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	rpcPort, _, err := getRpcPort(vps.Id)
	if err != nil {
		e.pubErrMsg(userId, "newnode", utils.RECODE_INSERTERR, err.Error(), topic_newnode_fail)
		return nil
	}

	node := models.Node{}
	node.CoinName = order.CoinName
	node.User = order.User
	node.Vps = &vps
	node.OrderNode = &order
	node.Port = rpcPort
	o = orm.NewOrm()
	_, err = o.Insert(&node)
	if err != nil {
		e.pubErrMsg(userId, "newnode", utils.RECODE_INSERTERR, err.Error(), topic_newnode_fail)
		return nil
	}

	o = orm.NewOrm()
	vps.UsableNodes = vps.UsableNodes - 1
	_, err = o.Update(&vps)
	if err != nil {
		e.pubErrMsg(userId, "newnode", utils.RECODE_UPDATEERR, err.Error(), topic_newnode_fail)
		return nil
	}

	o = orm.NewOrm()
	order.Status = models.ORDER_STATUS_COMPLETE
	_, err = o.Update(&order)
	if err != nil {
		e.pubErrMsg(userId, "newnode", utils.RECODE_UPDATEERR, err.Error(), topic_newnode_fail)
		return nil
	}

	e.pubMsg(userId, topic_newnode_success, orderId)
	log.Printf("process new node %v\n", orderId)
	return nil
}

func (e *Vps) processDelNode(nodeId int64) error {
	log.Printf("process del node from order %v\n", nodeId)

	var node models.Node
	o := orm.NewOrm()
	qs := o.QueryTable("node")
	err := qs.Filter("id", nodeId).One(&node)
	if err != nil {
		e.pubErrMsg("", "delnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
		return err
	}

	userId := strconv.Itoa(node.User.Id)

	var vps models.Vps
	o = orm.NewOrm()
	qs = o.QueryTable("vps")
	err = qs.Filter("id", node.Vps.Id).One(&vps)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
		return err
	}

	var order models.OrderNode
	o = orm.NewOrm()
	qs = o.QueryTable(order_tablename)
	err = qs.Filter(order_id, node.OrderNode.Id).One(&order)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
		return err
	}

	client := gossh.New(vps.IpAddress, "root")
	if client == nil {
		return errors.New("client no connect")
	}

	client.SetPassword(ssh_password)

	defer client.Close()

	port := node.Port
	volumeName := "mn" + strconv.Itoa(port)

	cmd := "docker stop  `docker ps -aq --filter name=" + volumeName + "`"
	sshCmd(client, cmd, 5)
	cmd = "docker rm  `docker ps -aq --filter name=" + volumeName + "`"
	sshCmd(client, cmd, 5)
	cmd = "docker volume rm " + volumeName
	sshCmd(client, cmd, 5)

	o = orm.NewOrm()
	_, err = o.Delete(&node)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_DELETEERR, "", topic_delnode_fail)
		return nil
	}

	o = orm.NewOrm()
	vps.UsableNodes = vps.UsableNodes + 1
	_, err = o.Update(&vps)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_UPDATEERR, "", topic_delnode_fail)
		return nil
	}

	o = orm.NewOrm()
	order.Status = models.ORDER_STATUS_EXPIRED
	_, err = o.Update(&order)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_UPDATEERR, "", topic_delnode_fail)
		return nil
	}

	isDel, err := e.delVps(userId, vps.Id)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_DELETEERR, "", topic_delnode_fail)
	}

	if isDel == true {
		o = orm.NewOrm()
		_, err = o.Delete(&vps)
		if err != nil {
			e.pubErrMsg(userId, "delnode", utils.RECODE_DELETEERR, "", topic_newnode_fail)
			return nil
		}
	}

	e.pubMsg(userId, topic_delnode_success, nodeId)
	log.Printf("success process del order from %v\n", nodeId)

	return nil
}

func (e *Vps) processExpandVolume(userId, volumeId string, size int64, ipAddress, password string) error {
	log.Printf("process expandvolume, volume:%s, size:%d\n", volumeId, size)

	c, err := uec2.NewEc2Client("us-east-2", "test-account")
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.CONNECT_ERR, err.Error(), topic_expandvolume_fail)
		return err
	}

	result, err := c.GetDescribeVolumes([]string{volumeId})
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.CONNECT_ERR, err.Error(), topic_expandvolume_fail)
		return err
	}

	osize := aws.Int64Value(result.Volumes[0].Size)
	if osize >= size {
		e.pubErrMsg(userId, "expandvolume", utils.RECODE_DATAERR, err.Error(), topic_expandvolume_fail)
		return err
	}

	_, err = c.ModifyVolumes(volumeId, size)
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.CONNECT_ERR, err.Error(), topic_expandvolume_fail)
		return err
	}

	client := gossh.New(ipAddress, "root")
	if client == nil {
		return errors.New("client no connect")
	}

	// my default agent authentication is used. use
	client.SetPassword(ssh_password)
	defer client.Close()

	var rsp *gossh.Result
	cmd := fmt.Sprintf("growpart /dev/%s %d", device_name, size)
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.CONNECT_ERR, err.Error(), topic_expandvolume_fail)
		return err
	}
	log.Printf("cmd:%v\n", rsp)

	cmd = fmt.Sprintf("resize2fs /dev/%s", device_name)
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.CONNECT_ERR, err.Error(), topic_expandvolume_fail)
		return err
	}

	log.Printf("success process expandvolume, volume:%v, size:%d\n", volumeId, size)
	return nil
}

func (e *Vps) delVps(userId string, vpsId int) (bool, error) {
	log.Println("start del vps ", vpsId)
	var nodes []models.Node
	o := orm.NewOrm()
	qs := o.QueryTable("node")
	nums, err := qs.Filter("vps_id", vpsId).All(&nodes)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
		return false, err
	}
	if nums == 0 {
		var vps models.Vps
		o = orm.NewOrm()
		qs = o.QueryTable("vps")
		err = qs.Filter("id", vpsId).One(&vps)
		if err != nil {
			e.pubErrMsg(userId, "delnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
			return false, err
		}

		c, err := uec2.NewEc2Client("us-east-2", "test-account")
		if err != nil {
			e.pubErrMsg(userId, "delnode", utils.CONNECT_ERR, err.Error(), topic_newnode_fail)
			return false, err
		}

		_, err = c.TerminateInstance(vps.InstanceId)
		if err != nil {
			e.pubErrMsg(userId, "delnode", utils.TERMINATE_INSTANCE_ERR, err.Error(), topic_newnode_fail)
			return false, err
		}

		retrys := 0
		for { //循环
			retrys++
			if retrys > instance_retrys {
				e.pubErrMsg(userId, "delnode", utils.TIMEOUT_VPS, err.Error(), topic_newnode_fail)
				return false, err
			}
			result, err := c.GetDescribeInstance([]string{vps.InstanceId})
			if err == nil {
				state := aws.StringValue(result.Reservations[0].Instances[0].State.Name)
				if state == "terminated" {
					break
				}
			}
			time.Sleep(time.Second)
		}

		_, err = c.DeleteVolumes(vps.VolumeId)
		if err != nil {
			e.pubErrMsg(userId, "delnode", utils.DELETE_VOLUME_ERR, err.Error(), topic_newnode_fail)
			return false, err
		}

		_, err = c.ReleaseAddresss(vps.AllocateId)
		if err != nil {
			e.pubErrMsg(userId, "delnode", utils.RELEASE_ADDRESS_ERR, err.Error(), topic_newnode_fail)
			return false, err
		}

		return true, nil
	}
	return false, nil
}

func newVps(imageId, zone, instanceType string, volumeSize int64, nvpsInfo VpsInfo) (*VpsInfo, error) {
	log.Printf("new vps start")

	if len(instanceType) == 0 {
		instanceType = "t2.micro"
	}

	if volumeSize == 0 {
		volumeSize = 20
	}

	var vpsInfo VpsInfo
	vpsInfo.allocationState = false
	vpsInfo.volumeState = false

	c, err := uec2.NewEc2Client(zone, "test-account")
	if err != nil {
		return &vpsInfo, err
	}

	var securityGroupId string
	groupResult, err := c.GetDescribeSecurityGroupsFromName([]string{group_name})
	if err != nil {
		ipPermissions := GetIpPermission()
		securityGroupId, err = c.CreateSecurityGroups(group_desc, group_desc, ipPermissions)
		if err != nil {
			log.Print("error***1")
			return &vpsInfo, err
		}
	} else {
		securityGroupId = aws.StringValue(groupResult.SecurityGroups[0].GroupId)
	}

	_, err = c.GetDescribeKeyPairs([]string{key_pair_name})
	if err != nil {
		_, err := c.CreateKeyPairs(key_pair_name)
		if err != nil {
			return &vpsInfo, err
		}
	}

	retrys := 0
	regionName := ""
	instanceId := nvpsInfo.instanceId
	for { //循环
		retrys++
		if retrys > instance_retrys {
			err = errors.New("instance timeout")
			return &vpsInfo, err
		}
		result, err := c.GetDescribeInstance([]string{instanceId})
		if err != nil {
			instanceId, err = c.CreateInstances(imageId, instanceType, key_pair_name, securityGroupId)
			continue
		}
		regionName = aws.StringValue(result.Reservations[0].Instances[0].Placement.AvailabilityZone)
		state := aws.StringValue(result.Reservations[0].Instances[0].State.Name)
		if state == "running" {
			break
		} else {
			c.StartInstance(instanceId)
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(time.Second)
	}
	vpsInfo.instanceId = instanceId
	vpsInfo.regionName = regionName

	allocationId := ""
	publicIp := ""
	if len(nvpsInfo.allocationId) == 0 {
		publicIp, allocationId, err = c.AllocateAddresss(instanceId)
		if err != nil {
			return &vpsInfo, err
		}
	}
	vpsInfo.allocationId = allocationId
	vpsInfo.publicIp = publicIp

	if nvpsInfo.allocationState == false {
		_, err = c.AssociateAddresss(instanceId, allocationId)
		if err != nil {
			return &vpsInfo, err
		}
	}
	vpsInfo.allocationState = true

	volumeId := ""
	if len(nvpsInfo.volumeId) == 0 {
		volumeId, err = c.CreateVolumes(regionName, volumeSize)
		if err != nil {
			return &vpsInfo, err
		}
	} else {
		volumeId = nvpsInfo.volumeId
	}
	vpsInfo.volumeId = volumeId

	if nvpsInfo.volumeState == false {
		retrys = 0
		for { //循环
			retrys++
			if retrys > volume_retrys {
				err = errors.New("volume timeout")
				return &vpsInfo, err
			}
			vResult, err := c.GetDescribeVolumes([]string{volumeId})
			if err == nil {
				vState := aws.StringValue(vResult.Volumes[0].State)
				if vState == "available" {
					break
				}
			}
			time.Sleep(time.Second)
		}
		_, err = c.AttachVolumes(instanceId, volumeId, device_name)
		if err != nil {
			return &vpsInfo, err
		}
	}

	err = VolumeMount(vpsInfo.publicIp, ssh_password)
	if err != nil {
		return &vpsInfo, err
	}

	vpsInfo.volumeState = true
	log.Println("new vps success")

	return &vpsInfo, nil
}

func newNode(vpsId int, coinName string, ipAddress string, password string) error {
	log.Printf("new node from vpsid %d\n", vpsId)
	var coin models.Coin
	o := orm.NewOrm()
	qs := o.QueryTable("coin")
	err := qs.Filter("name", coinName).One(&coin)
	if err != nil {
		return err
	}

	client := gossh.New(ipAddress, "root")
	if client == nil {
		return errors.New("client no connect")
	}

	// my default agent authentication is used. use
	client.SetPassword(ssh_password)

	defer client.Close()

	rpcPort, port, err := getRpcPort(vpsId)
	if err != nil {
		return err
	}

	volumeName := "mn" + strconv.Itoa(rpcPort)
	var rsp *gossh.Result
	cmd := "docker volume inspect " + volumeName
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		cmd = "docker volume create --name=" + volumeName
		rsp, err = sshCmd(client, cmd, 5)
		if err != nil {
			return err
		}
		time.Sleep(10 * time.Second)
		cmd = "docker volume inspect " + volumeName
		rsp, err = sshCmd(client, cmd, 5)
		if err != nil {
			return err
		}
	}

	var part []Volume
	if err = json.Unmarshal([]byte(rsp.Stdout()), &part); err != nil {
		return err
	}
	mountPoint := part[0].Mountpoint

	coinPath := mountPoint + "/" + coin.Path
	cmd = "cd " + coinPath
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		cmd = "mkdir " + coinPath
		_, err = sshCmd(client, cmd, 5)
		if err != nil {
			return err
		}
	}

	localFile := "/tmp/" + coin.Conf
	err = WriteConf(localFile, port)
	if err != nil {
		return err
	}

	err = UploadFile(ipAddress, "root", password, localFile, "/tmp/")
	if err != nil {
		return err
	}

	cmd = "mv /tmp/" + coin.Conf + " " + coinPath
	_, err = sshCmd(client, cmd, 5)
	if err != nil {
		return err
	}

	cmd = "docker images | grep " + coinName
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		localFile = "/root/vpub-vircle-0.1.tar"
		err = UploadFile(ipAddress, "root", password, localFile, "/tmp/")
		if err != nil {
			return err
		}

		cmd = "docker load  --input /tmp/vpub-vircle-0.1.tar"
		rsp, err = sshCmd(client, cmd, 5)
		if err != nil {
			return err
		}
	}

	srpcPort := strconv.Itoa(rpcPort)
	sport := strconv.Itoa(port)
	cmd = "docker run -v " + volumeName + ":/" + coinName + " --name=" + volumeName + " -d -p " + sport + ":" + sport + " -p " + srpcPort + ":" + srpcPort + " " + coin.Docker
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		return err
	}
	log.Printf("new node from vpsid %d success\n", vpsId)
	return nil
}

func GetIpPermission() []*ec2.IpPermission {
	ipPermissions := []*ec2.IpPermission{
		(&ec2.IpPermission{}).
			SetIpProtocol("tcp").
			SetFromPort(80).
			SetToPort(80).
			SetIpRanges([]*ec2.IpRange{
				{CidrIp: aws.String("0.0.0.0/0")},
			}),
		(&ec2.IpPermission{}).
			SetIpProtocol("tcp").
			SetFromPort(22).
			SetToPort(22).
			SetIpRanges([]*ec2.IpRange{
				(&ec2.IpRange{}).
					SetCidrIp("0.0.0.0/0"),
			}),
		(&ec2.IpPermission{}).
			SetIpProtocol("tcp").
			SetFromPort(port_from).
			SetToPort(port_to).
			SetIpRanges([]*ec2.IpRange{
				(&ec2.IpRange{}).
					SetCidrIp("0.0.0.0/0"),
			}),
	}
	return ipPermissions
}

func SecurityGroupIsExist(groupName string, groupResult *ec2.DescribeSecurityGroupsOutput) string {
	for _, group := range groupResult.SecurityGroups {
		if aws.StringValue(group.GroupName) == groupName {
			return aws.StringValue(group.GroupId)
		}
	}
	return ""
}

func UploadFile(ipAddress string, username string, password string, localFile string, remoteDir string) error {
	client, err := ssh.NewClient(ipAddress, "22", username, password)
	if err != nil {
		return err
	}
	defer client.Close()
	err = client.Upload(localFile, remoteDir)
	if err != nil {
		return err
	}
	return nil
}

func WriteConf(confName string, port int) error {
	cfg := ini.Empty()

	cfg.Section("").Key("rpcuser").SetValue(rpc_user)
	cfg.Section("").Key("rpcpassword").SetValue(rpc_password)
	cfg.Section("").Key("rpcallowip").SetValue("1.2.3.4/0.0.0.0")
	cfg.Section("").Key("rpcbind").SetValue("0.0.0.0")
	cfg.Section("").Key("rpcport").SetValue(strconv.Itoa(port))
	cfg.Section("").Key("port").SetValue(strconv.Itoa(port + 1))

	cfg.SaveTo(confName)

	return nil
}

func getRpcPort(vpsId int) (int, int, error) {
	rpcport := port_from
	port := rpcport + 1

	var nodes []models.Node
	o := orm.NewOrm()
	qs := o.QueryTable("node")
	nums, err := qs.Filter("vps_id", vpsId).All(&nodes)
	if err != nil {
		return 0, 0, err
	}
	if nums == 0 {
		return rpcport, port, nil
	}

	for i := port_from; i < port_to; i = i + 2 {
		if portExist(i, &nodes) == false {
			return i, i + 1, nil
		}
	}
	return 0, 0, errors.New("port is full")
}

func portExist(port int, nodes *[]models.Node) bool {
	for _, node := range *nodes {
		if node.Port == port {
			return true
		}
	}
	return false
}

func sshCmd(client *gossh.Client, cmd string, fails int) (*gossh.Result, error) {
	retrys := 0
	for { //循环
		retrys++
		if retrys > fails {
			return nil, errors.New("timeout")
		}
		result, err := client.Execute(cmd)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		log.Println(result.Stdout())
		return result, nil
	}
}

func VolumeMount(ipAddress, password string) error {
	log.Println("start mount")

	client := gossh.New(ipAddress, "root")
	if client == nil {
		return errors.New("client no connect")
	}

	// my default agent authentication is used. use
	client.SetPassword(password)

	defer client.Close()

	cmd := fmt.Sprintf("echo -e \"d\n1\nw\" | fdisk /dev/%s", device_name)
	result, err := client.Execute(cmd)
	/*if err != nil {
		return err
	}*/

	cmd = fmt.Sprintf("echo -e \"n\np\n1\n\n\nw\" | fdisk /dev/%s", device_name)
	result, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("fdisk -l |grep %s", device_name)
	result, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("file -s %s |grep %s", device_name1, device_name1)
	result, err = client.Execute(cmd)
	if err != nil {
		cmd = fmt.Sprintf("mkfs -t ext4 /dev/%s", device_name1)
		result, err = client.Execute(cmd)
		if err != nil {
			return err
		}
	}

	cmd = fmt.Sprintf("mount /dev/%s %s", device_name1, mount_point)
	result, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	log.Println("success mount ", result)
	return nil
}
