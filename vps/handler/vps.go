package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
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

const topic_restartnode_success = "Vircle.Mnhost.Topic.RestartNode.Success"
const topic_restartnode_fail = "Vircle.Mnhost.Topic.RestartNode.Fail"
const topic_restartnode_start = "Vircle.Mnhost.Topic.RestartNode.Start"

const ssh_password = "vpub$999000"
const rpc_user = "vpub"
const rpc_password = "vpub999000"
const port_from = 19900
const port_to = 20000
const s_port = "9998"
const s_rpcport = "9999"
const s_workdir = "vircle"

//const order_tablename = "order_node"
//const order_id = "Id"
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

	var torder models.TOrder
	o := orm.NewOrm()
	qs := o.QueryTable("t_order")
	err := qs.Filter("id", orderId).One(&torder)
	if err != nil {
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	go e.processNewNode(int64(torder.Id))

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	jorder, err := json.Marshal(torder)
	rsp.Mix = jorder

	log.Println("new node request success")
	return nil
}

func (e *Vps) DelNode(ctx context.Context, req *vps.Request, rsp *vps.Response) error {
	log.Println("del node request")
	nodeId := req.Id

	var tnode models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	err := qs.Filter("id", nodeId).One(&tnode)
	if err != nil {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}

	go e.processDelNode(int64(tnode.Id))

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	jnode, err := json.Marshal(tnode)
	rsp.Mix = jnode

	log.Println("success del node request")
	return nil
}

func (e *Vps) ExpandVolume(ctx context.Context, req *vps.VolumeRequest, rsp *vps.Response) error {
	log.Println("expand volume request")
	volumeId := req.VolumeId
	size := req.Size

	var tvps models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	err := qs.Filter("volumeId", volumeId).One(&tvps)
	if err != nil {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}

	if size == 0 {
		rsp.Errno = utils.RECODE_DATAERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	go e.processExpandVolume("", volumeId, size, tvps.IpAddress, ssh_password)

	jvps, err := json.Marshal(tvps)
	rsp.Mix = jvps

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	log.Println("success expand volume request")
	return nil
}

func (e *Vps) GetAllVps(ctx context.Context, req *vps.Request, rsp *vps.VpsResponse) error {
	log.Printf("get all vps")

	var tvpss []models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	nums, err := qs.All(&tvpss)
	if err != nil {
		rsp.Errno = utils.RECODE_QUERYERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}
	if nums == 0 {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	pbVpss := make([]*vps.MVps, len(tvpss))
	for i, vps := range tvpss {
		pbVps := Vps2PBVps(&vps)
		pbVpss[i] = &pbVps
	}

	rsp.Vpss = pbVpss

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	log.Printf("success get all vps")
	return nil
}

func (e *Vps) GetAllNodeFromVps(ctx context.Context, req *vps.Request, rsp *vps.NodeResponse) error {
	vpsId := req.Id
	log.Printf("get all node from vpsId: %d\n", vpsId)
	var tnodes []models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	nums, err := qs.Filter("vps_id", vpsId).All(&tnodes)
	if err != nil {
		rsp.Errno = utils.RECODE_QUERYERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}
	if nums == 0 {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	pbNodes := make([]*vps.Node, len(tnodes))
	for i, node := range tnodes {
		pbNode := Node2PBNode(&node)
		pbNodes[i] = &pbNode
	}

	rsp.Nodes = pbNodes

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)
	log.Printf("success get all node from vpsId: %d\n", vpsId)
	return nil
}

func (e *Vps) GetAllNodeFromUser(ctx context.Context, req *vps.Request, rsp *vps.NodeResponse) error {
	userId := req.Id
	log.Printf("get all node from userId: %d\n", userId)

	var tnodes []models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	nums, err := qs.Filter("userid", userId).All(&tnodes)
	if err != nil {
		rsp.Errno = utils.RECODE_QUERYERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}
	if nums == 0 {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	pbNodes := make([]*vps.Node, len(tnodes))
	for i, node := range tnodes {
		pbNode := Node2PBNode(&node)
		pbNodes[i] = &pbNode
	}

	rsp.Nodes = pbNodes

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	log.Printf("success get all node from userId: %d\n", userId)
	return nil
}

func (e *Vps) RestartNode(ctx context.Context, req *vps.Request, rsp *vps.Response) error {
	nodeId := req.Id
	log.Printf("restart node request from %v\n", nodeId)

	var tnode models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	err := qs.Filter("id", nodeId).One(&tnode)
	if err != nil {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}

	userId := strconv.FormatInt(tnode.Userid, 10)

	var tvps models.TVps
	o = orm.NewOrm()
	qs = o.QueryTable("t_vps")
	err = qs.Filter("id", tnode.Vps.Id).One(&tvps)
	if err != nil {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}

	go e.processRestartNode(userId, tvps.IpAddress, ssh_password, tnode.Port, nodeId)

	type Rmsg struct {
		NodeId    string
		IpAddress string
		RpcPort   string
	}
	msg := Rmsg{
		NodeId:    strconv.FormatInt(nodeId, 10),
		IpAddress: tvps.IpAddress,
		RpcPort:   strconv.Itoa(tnode.Port),
	}
	jmsg, err := json.Marshal(msg)
	rsp.Mix = jmsg

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)
	log.Printf("success restart node request from %d\n", nodeId)
	return nil
}

func (e *Vps) processNewNode(orderId int64) error {
	log.Printf("process new node from orderid %d\n", orderId)

	var torder models.TOrder
	o := orm.NewOrm()
	qs := o.QueryTable("t_order")
	err := qs.Filter("id", orderId).One(&torder)
	if err != nil {
		e.pubErrMsg("", "newnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
		return err
	}

	userId := strconv.FormatInt(torder.Userid, 10)

	var nvpsInfo VpsInfo
	var vpsInfo *VpsInfo
	var tvps models.TVps
	tnvps := models.TVps{}
	o = orm.NewOrm()
	qs = o.QueryTable("t_vps")
	err = qs.Filter("usable_nodes__gt", 0).One(&tvps)
	retrys := 0
	if err != nil {
		log.Printf("err1:%v\n", err)
		for { //循环
			retrys++
			if retrys > vps_retrys {
				e.pubErrMsg(userId, "newnode", utils.TIMEOUT_VPS, err.Error(), topic_newnode_fail)
				return nil
			}
			vpsInfo, err = newVps("ami-0b0426f6bc13cbfe4", "us-east-2", "", test_volume_size, nvpsInfo)
			log.Printf("new vps:%v\n", vpsInfo)
			nvpsInfo.allocationId = vpsInfo.allocationId
			nvpsInfo.allocationState = vpsInfo.allocationState
			nvpsInfo.instanceId = vpsInfo.instanceId
			nvpsInfo.publicIp = vpsInfo.publicIp
			nvpsInfo.regionName = vpsInfo.regionName
			nvpsInfo.volumeId = vpsInfo.volumeId
			nvpsInfo.volumeState = vpsInfo.volumeState
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			break
		}
		tnvps.AllocateId = vpsInfo.allocationId
		tnvps.InstanceId = vpsInfo.instanceId
		tnvps.VolumeId = vpsInfo.volumeId
		tnvps.ProviderName = provider_name
		tnvps.Cores = core_nums
		tnvps.Memory = memory_size
		tnvps.KeyPairName = key_pair_name
		tnvps.MaxNodes = masternode_max_nums
		tnvps.UsableNodes = masternode_max_nums
		tnvps.SecurityGroupName = group_name
		tnvps.RegionName = vpsInfo.regionName
		tnvps.IpAddress = vpsInfo.publicIp
		log.Printf("tnvps:%v\n", tnvps)
		o = orm.NewOrm()
		_, err = o.Insert(&tnvps)
		if err != nil {
			e.pubErrMsg(userId, "newnode", utils.RECODE_INSERTERR, err.Error(), topic_newnode_fail)
			return nil
		}
		o = orm.NewOrm()
		qs = o.QueryTable("t_vps")
		err = qs.Filter("usable_nodes__gt", 0).One(&tvps)
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
		err = newNode(tvps.Id, torder.Coinname, tvps.IpAddress, ssh_password, torder.Mnkey)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	rpcPort, _, err := getRpcPort(tvps.Id)
	if err != nil {
		e.pubErrMsg(userId, "newnode", utils.RECODE_INSERTERR, err.Error(), topic_newnode_fail)
		return nil
	}

	tnode := models.TNode{}
	tnode.CoinName = torder.Coinname
	tnode.Userid = torder.Userid
	tnode.Vps = &tvps
	tnode.Order = &torder
	tnode.Port = rpcPort
	o = orm.NewOrm()
	_, err = o.Insert(&tnode)
	if err != nil {
		e.pubErrMsg(userId, "newnode", utils.RECODE_INSERTERR, err.Error(), topic_newnode_fail)
		return nil
	}

	o = orm.NewOrm()
	tvps.UsableNodes = tvps.UsableNodes - 1
	_, err = o.Update(&tvps)
	if err != nil {
		e.pubErrMsg(userId, "newnode", utils.RECODE_UPDATEERR, err.Error(), topic_newnode_fail)
		return nil
	}

	/*o = orm.NewOrm()
	order.Status = models.ORDER_STATUS_COMPLETE
	_, err = o.Update(&order)
	if err != nil {
		e.pubErrMsg(userId, "newnode", utils.RECODE_UPDATEERR, err.Error(), topic_newnode_fail)
		return nil
	}*/

	e.pubMsg(userId, topic_newnode_success, orderId)
	log.Printf("success process new node %v\n", orderId)
	return nil
}

func (e *Vps) processDelNode(nodeId int64) error {
	log.Printf("process del node from %v\n", nodeId)

	var tnode models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	err := qs.Filter("id", nodeId).One(&tnode)
	if err != nil {
		e.pubErrMsg("", "delnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
		return err
	}
	log.Printf("node:%v\n", tnode)

	userId := strconv.FormatInt(tnode.Userid, 10)

	var tvps models.TVps
	o = orm.NewOrm()
	qs = o.QueryTable("t_vps")
	err = qs.Filter("id", tnode.Vps.Id).One(&tvps)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
		return err
	}
	log.Printf("vps:%v\n", tvps)

	var torder models.TOrder
	o = orm.NewOrm()
	qs = o.QueryTable("t_order")
	err = qs.Filter("id", tnode.Order.Id).One(&torder)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
		return err
	}
	log.Printf("tnode:%v\n", torder)

	client := gossh.New(tvps.IpAddress, "root")
	if client == nil {
		return errors.New("client no connect")
	}
	client.SetPassword(ssh_password)

	defer client.Close()

	port := tnode.Port
	volumeName := "mn" + strconv.Itoa(port)

	cmd := "docker stop  `docker ps -aq --filter name=" + volumeName + "`"
	sshCmd(client, cmd, 5)
	cmd = "docker rm  `docker ps -aq --filter name=" + volumeName + "`"
	sshCmd(client, cmd, 5)
	cmd = "docker volume rm " + volumeName
	sshCmd(client, cmd, 5)

	o = orm.NewOrm()
	_, err = o.Delete(&tnode)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_DELETEERR, "", topic_delnode_fail)
		return nil
	}

	o = orm.NewOrm()
	tvps.UsableNodes = tvps.UsableNodes + 1
	_, err = o.Update(&tvps)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_UPDATEERR, "", topic_delnode_fail)
		return nil
	}

	/*o = orm.NewOrm()
	order.Status = models.ORDER_STATUS_EXPIRED
	_, err = o.Update(&order)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_UPDATEERR, "", topic_delnode_fail)
		return nil
	}*/

	isDel, err := e.delVps(userId, tvps.Id)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_DELETEERR, "", topic_delnode_fail)
	}

	if isDel == true {
		o = orm.NewOrm()
		_, err = o.Delete(&tvps)
		if err != nil {
			e.pubErrMsg(userId, "delnode", utils.RECODE_DELETEERR, "", topic_newnode_fail)
			return nil
		}
	}

	e.pubMsg(userId, topic_delnode_success, nodeId)
	log.Printf("success process del node from %v\n", nodeId)

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
	if result == nil {
		e.pubErrMsg(userId, "expandvolume", utils.RECORD_SYSTEMERR, err.Error(), topic_expandvolume_fail)
		return err
	}
	log.Printf("volume:%s\n", result)

	osize := aws.Int64Value(result.Volumes[0].Size)
	if osize > size {
		e.pubErrMsg(userId, "expandvolume", utils.RECODE_DATAERR, "", topic_expandvolume_fail)
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
	cmd := fmt.Sprintf("growpart /dev/%s 1", device_name)
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.CONNECT_ERR, err.Error(), topic_expandvolume_fail)
		return err
	}

	cmd = fmt.Sprintf("resize2fs /dev/%s", device_name1)
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.CONNECT_ERR, err.Error(), topic_expandvolume_fail)
		return err
	}
	log.Printf("rsp:%v\n", rsp)

	e.pubMsg(userId, topic_expandvolume_success, 1)
	log.Printf("success process expandvolume, volume:%v, size:%d\n", volumeId, size)
	return nil
}

func (e *Vps) processRestartNode(userId, ipAddress, password string, port int, nodeId int64) error {
	log.Printf("process restart node from %s:%d\n", ipAddress, port)

	client := gossh.New(ipAddress, "root")
	if client == nil {
		return errors.New("client no connect")
	}
	// my default agent authentication is used. use
	client.SetPassword(password)
	defer client.Close()

	var rsp *gossh.Result
	cmd := fmt.Sprintf("docker ps -a | grep mn%d | awk  '{print $1}'", port)
	rsp, err := sshCmd(client, cmd, 5)
	if err != nil {
		e.pubErrMsg(userId, "restartnode", utils.CONNECT_ERR, err.Error(), topic_restartnode_fail)
		return err
	}
	containerId := rsp.Stdout()
	containerId = containerId[:8]
	cmd = fmt.Sprintf("docker restart %s", containerId)
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		log.Printf("err:%v\n", err)
		e.pubErrMsg(userId, "restartnode", utils.CONNECT_ERR, err.Error(), topic_restartnode_fail)
		return err
	}
	log.Printf("container:%v\n", rsp.Stdout())

	e.pubMsg(userId, topic_restartnode_success, nodeId)
	log.Printf("success process restart node from %s:%d\n", ipAddress, port)
	return nil
}

func (e *Vps) delVps(userId string, vpsId int64) (bool, error) {
	log.Println("start del vps ", vpsId)
	var tnodes []models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	nums, err := qs.Filter("vps_id", vpsId).All(&tnodes)
	if err != nil {
		e.pubErrMsg(userId, "delnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
		return false, err
	}
	if nums == 0 {
		var tvps models.TVps
		o = orm.NewOrm()
		qs = o.QueryTable("t_vps")
		err = qs.Filter("id", vpsId).One(&tvps)
		if err != nil {
			e.pubErrMsg(userId, "delnode", utils.RECODE_NODATA, err.Error(), topic_newnode_fail)
			return false, err
		}

		c, err := uec2.NewEc2Client("us-east-2", "test-account")
		if err != nil {
			e.pubErrMsg(userId, "delnode", utils.CONNECT_ERR, err.Error(), topic_newnode_fail)
			return false, err
		}

		_, err = c.TerminateInstance(tvps.InstanceId)
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
			result, err := c.GetDescribeInstance([]string{tvps.InstanceId})
			if err == nil {
				state := aws.StringValue(result.Reservations[0].Instances[0].State.Name)
				if state == "terminated" {
					break
				}
			}
			time.Sleep(time.Second)
		}

		_, err = c.DeleteVolumes(tvps.VolumeId)
		if err != nil {
			e.pubErrMsg(userId, "delnode", utils.DELETE_VOLUME_ERR, err.Error(), topic_newnode_fail)
			return false, err
		}

		_, err = c.ReleaseAddresss(tvps.AllocateId)
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
	vpsInfo.volumeState = true

	err = VolumeMount(vpsInfo.publicIp, ssh_password)
	if err != nil {
		return &vpsInfo, err
	}

	log.Println("new vps success")

	return &vpsInfo, nil
}

func newNode(vpsId int64, coinName string, ipAddress string, password string, mnKey string) error {
	log.Printf("new node from vpsid %d\n", vpsId)
	var tcoin models.TCoin
	o := orm.NewOrm()
	qs := o.QueryTable("t_coin")
	err := qs.Filter("name", coinName).One(&tcoin)
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

	coinPath := mountPoint + "/" + tcoin.Path
	cmd = "cd " + coinPath
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		cmd = "mkdir " + coinPath
		_, err = sshCmd(client, cmd, 5)
		if err != nil {
			return err
		}
	}

	localFile := "/tmp/" + tcoin.Conf
	err = WriteConf(localFile, ipAddress, mnKey)
	if err != nil {
		return err
	}

	err = UploadFile(ipAddress, "root", password, localFile, "/tmp/")
	if err != nil {
		return err
	}

	cmd = "mv /tmp/" + tcoin.Conf + " " + coinPath
	_, err = sshCmd(client, cmd, 5)
	if err != nil {
		return err
	}

	cmd = "docker images | grep " + coinName
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		localPath := tcoin.FilePath
		baseFile := filepath.Base(localPath)
		err = UploadFile(ipAddress, "root", password, localPath, "/tmp/")
		if err != nil {
			return err
		}

		cmd = "docker load  --input /tmp/" + baseFile
		rsp, err = sshCmd(client, cmd, 5)
		if err != nil {
			return err
		}

		cmd = "rm -rf /tmp/" + baseFile
		rsp, err = sshCmd(client, cmd, 5)
		if err != nil {
			return err
		}
	}

	srpcPort := strconv.Itoa(rpcPort)
	sport := strconv.Itoa(port)
	cmd = "docker run -v " + volumeName + ":/" + s_workdir + " --name=" + volumeName + " -d -p " + sport + ":" + s_port + " -p " + srpcPort + ":" + s_rpcport + " " + tcoin.Docker
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

func WriteConf(confName, externIp, mnKey string) error {
	cfg := ini.Empty()

	cfg.Section("").Key("listen").SetValue("1")
	cfg.Section("").Key("server").SetValue("1")
	cfg.Section("").Key("rpcuser").SetValue(rpc_user)
	cfg.Section("").Key("rpcpassword").SetValue(rpc_password)
	cfg.Section("").Key("rpcallowip").SetValue("1.2.3.4/0.0.0.0")
	cfg.Section("").Key("rpcbind").SetValue("0.0.0.0")
	cfg.Section("").Key("rpcport").SetValue(s_rpcport)
	cfg.Section("").Key("port").SetValue(s_port)
	cfg.Section("").Key("masternode").SetValue("1")
	cfg.Section("").Key("masternodeblsprivkey").SetValue(mnKey)
	cfg.Section("").Key("externalip").SetValue(externIp)

	cfg.SaveTo(confName)

	return nil
}

func getRpcPort(vpsId int64) (int, int, error) {
	rpcport := port_from
	port := rpcport + 1

	var tnodes []models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	nums, err := qs.Filter("vps_id", vpsId).All(&tnodes)
	if err != nil {
		return 0, 0, err
	}
	if nums == 0 {
		return rpcport, port, nil
	}

	for i := port_from; i < port_to; i = i + 2 {
		if portExist(i, &tnodes) == false {
			return i, i + 1, nil
		}
	}
	return 0, 0, errors.New("port is full")
}

func portExist(port int, tnodes *[]models.TNode) bool {
	for _, node := range *tnodes {
		if node.Port == port {
			return true
		}
	}
	return false
}

func sshCmd(client *gossh.Client, cmd string, fails int) (*gossh.Result, error) {
	log.Printf("cmd:%s\n", cmd)
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
	log.Printf("cmd:%s\n", cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("fdisk -l |grep %s", device_name1)
	log.Printf("cmd:%s\n", cmd)
	result, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("file -s /dev/%s |grep ext4", device_name1)
	log.Printf("cmd:%s\n", cmd)
	result, err = client.Execute(cmd)
	if err != nil {
		cmd = fmt.Sprintf("mkfs -t ext4 /dev/%s", device_name1)
		result, err = client.Execute(cmd)
		if err != nil {
			return err
		}
	}

	cmd = fmt.Sprintf("mkdir -p %s", mount_point)
	log.Printf("cmd:%s\n", cmd)
	result, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("mount /dev/%s %s", device_name1, mount_point)
	log.Printf("cmd:%s\n", cmd)
	result, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("df -h |grep %s", device_name1)
	log.Printf("cmd:%s\n", cmd)
	result, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	log.Println("success mount ", result)
	return nil
}

func Vps2PBVps(u *models.TVps) vps.MVps {
	return vps.MVps{
		Id:                strconv.FormatInt(u.Id, 10),
		ProviderName:      u.ProviderName,
		Cores:             strconv.Itoa(u.Cores),
		Memory:            strconv.Itoa(u.Memory),
		MaxNodes:          strconv.Itoa(u.MaxNodes),
		UsableNodes:       strconv.Itoa(u.UsableNodes),
		RegionName:        u.RegionName,
		InstanceId:        u.InstanceId,
		VolumeId:          u.VolumeId,
		SecurityGroupName: u.SecurityGroupName,
		KeyPairName:       u.KeyPairName,
		AllocateId:        u.AllocateId,
		IpAddress:         u.IpAddress,
	}
}

func Node2PBNode(u *models.TNode) vps.Node {
	return vps.Node{
		Id:       strconv.FormatInt(u.Id, 10),
		UserId:   strconv.FormatInt(u.Userid, 10),
		VpsId:    strconv.FormatInt(u.Vps.Id, 10),
		OrderId:  strconv.FormatInt(u.Order.Id, 10),
		CoinName: u.CoinName,
		Port:     strconv.Itoa(u.Port),
	}
}
