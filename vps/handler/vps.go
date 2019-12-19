package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	//"path/filepath"
	"strconv"
	"sync"
	"time"

	//"os"

	"github.com/dynport/gossh"
	"github.com/go-ini/ini"
	"github.com/pytool/ssh"

	//"github.com/robfig/cron"

	"github.com/astaxie/beego/orm"
	"github.com/micro/go-micro/broker"

	uec2 "github.com/John-Tonny/micro/vps/amazon"
	"github.com/John-Tonny/mnhost/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	logPB "github.com/John-Tonny/mnhost/interface/out/log"
	mnPB "github.com/John-Tonny/mnhost/interface/out/mnmsg"
	vps "github.com/John-Tonny/mnhost/interface/out/vps"

	"github.com/John-Tonny/mnhost/conf"
	"github.com/John-Tonny/mnhost/model"
	mnhostTypes "github.com/John-Tonny/mnhost/types"
	"github.com/John-Tonny/mnhost/utils"
)

type Vps struct {
	Broker broker.Broker
}

const service = "vps"

var (
	topic       string
	serviceName string
	version     string
	mutex       sync.Mutex
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
func (e *Vps) pubMsg(userID, topic, msgId string) error {
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

func (e *Vps) CreateVps(ctx context.Context, req *vps.CreateVpsRequest, rsp *vps.Response) error {
	log.Println("create vps request")

	go e.processCreateVps(req.ClusterName, req.Role, req.VolumeSize)

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	log.Println("create vps request success")
	return nil
}

func (e *Vps) RemoveVps(ctx context.Context, req *vps.Request, rsp *vps.Response) error {
	log.Println("remove vps request")

	var tvps models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	err := qs.Filter("instanceId", req.Id).One(&tvps)
	if err != nil {
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	go e.processRemoveVps(req.Id)

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	log.Println("remove vps request success")
	return nil
}

func (e *Vps) CreateNode(ctx context.Context, req *vps.Request, rsp *vps.Response) error {
	log.Println("create node request")

	orderId, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		rsp.Errno = utils.RECODE_DATAERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	var torder models.TOrder
	o := orm.NewOrm()
	qs := o.QueryTable("t_order")
	err = qs.Filter("id", orderId).One(&torder)
	if err != nil {
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	userId := strconv.FormatInt(torder.Userid, 10)

	var tcoin models.TCoin
	o = orm.NewOrm()
	qs = o.QueryTable("t_coin")
	err = qs.Filter("name", torder.Coinname).Filter("status", "Enabled").One(&tcoin)
	if err != nil {
		e.pubErrMsg(userId, "newnode", utils.RECODE_NODATA, err.Error(), mnhostTypes.TOPIC_NEWNODE_FAIL)
		return err
	}

	go e.processCreateNode(orderId, "cluster1")

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	jorder, err := json.Marshal(torder)
	rsp.Mix = jorder

	log.Println("create node request success")
	return nil
}

func (e *Vps) RemoveNode(ctx context.Context, req *vps.Request, rsp *vps.Response) error {
	log.Println("remove node request")
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

	go e.processRemoveNode(tnode.Id, "cluster1")

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	jnode, err := json.Marshal(tnode)
	rsp.Mix = jnode

	log.Println("success remove node request")
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

	go e.processExpandVolume("", volumeId, size, tvps.PublicIp, tvps.PrivateIp, mnhostTypes.SSH_PASSWORD)

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

func (e *Vps) GetAllNodeFromUser(ctx context.Context, req *vps.Request, rsp *vps.NodeResponse) error {
	userId := req.Id
	log.Printf("get all node from userId: %s\n", userId)

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

	log.Printf("success get all node from userId: %s\n", userId)
	return nil
}

func (e *Vps) processCreateVps(clusterName, role string, volumeSize int64) error {
	log.Printf("process create Vps from cluster:%s, role:%s, size:%d\n", clusterName, role, volumeSize)

	var vpsInfo *mnhostTypes.VpsInfo
	vpsInfo, errcode, err := NewVps(mnhostTypes.SYSTEM_IMAGE, mnhostTypes.ZONE_DEFAULT, mnhostTypes.INSTANCE_TYPE_DEFAULT, clusterName, role, volumeSize, false, false)
	if err != nil {
		e.pubErrMsg("", "newvps", errcode, err.Error(), mnhostTypes.TOPIC_NEWVPS_FAIL)
		return nil
	}
	err = EfsMount(vpsInfo.PublicIp, vpsInfo.PrivateIp, mnhostTypes.SSH_PASSWORD)
	if err != nil {
		e.pubErrMsg("", "newvps", utils.EFS_MOUNT_ERR, err.Error(), mnhostTypes.TOPIC_NEWVPS_FAIL)
		return nil
	}

	if (role == mnhostTypes.ROLE_MANAGER) || (role == mnhostTypes.ROLE_WORKER) {
		publicIp, privateIp, err := common.GetVpsIp("cluster1")
		if err != nil {
			e.pubErrMsg("", "newvps", utils.MANAGER_ERR, err.Error(), mnhostTypes.TOPIC_NEWVPS_FAIL)
			return nil
		}

		mc, _, err := common.DockerNewClient(publicIp, privateIp)
		if err != nil {
			e.pubErrMsg("", "newvps", utils.DOCKER_CONNECT_ERR, err.Error(), mnhostTypes.TOPIC_NEWVPS_FAIL)
			return nil
		}
		defer mc.Close()

		managerToken, workerToken, err := mc.SwarmInspectA()
		if err != nil {
			e.pubErrMsg("", "newvps", utils.SWARM_INSPECT_ERR, err.Error(), mnhostTypes.TOPIC_NEWVPS_FAIL)
			return nil
		} else {
			c, _, err := common.DockerNewClient(vpsInfo.PublicIp, vpsInfo.PrivateIp)
			if err != nil {
				e.pubErrMsg("", "newvps", utils.DOCKER_CONNECT_ERR, err.Error(), mnhostTypes.TOPIC_NEWVPS_FAIL)
				return nil
			}
			defer c.Close()

			token := managerToken
			status := "manager"
			if role == mnhostTypes.ROLE_WORKER {
				token = workerToken
				status = "worker"
			}
			err = c.SwarmJoinA(vpsInfo.PrivateIp, privateIp, token)
			if err != nil {
				e.pubErrMsg("", "newvps", utils.SWARM_INIT_ERR, err.Error(), mnhostTypes.TOPIC_NEWVPS_FAIL)
				return nil
			}

			var tvps models.TVps
			o := orm.NewOrm()
			qs := o.QueryTable("t_vps")
			err = qs.Filter("instanceId", vpsInfo.InstanceId).One(&tvps)
			if err != nil {
				e.pubErrMsg("", "delvps", utils.RECODE_DATAERR, err.Error(), mnhostTypes.TOPIC_DELVPS_FAIL)
				return nil
			}

			o = orm.NewOrm()
			tvps.Status = status
			_, err = o.Update(&tvps)
			if err != nil {
				return nil
			}
		}
	}

	e.pubMsg("", mnhostTypes.TOPIC_NEWVPS_SUCCESS, vpsInfo.InstanceId)
	log.Printf("success process create Vps from cluster:%s, role:%s, size:%d\n", clusterName, role, volumeSize)
	return nil
}

func (e *Vps) processRemoveVps(instanceId string) error {
	log.Printf("process remove Vps from instanceId:%s\n", instanceId)

	var tvps models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	err := qs.Filter("instanceId", instanceId).One(&tvps)
	if err != nil {
		e.pubErrMsg("", "delvps", utils.RECODE_DATAERR, err.Error(), mnhostTypes.TOPIC_DELVPS_FAIL)
		return nil
	}

	errcode, err := DelVps("", instanceId)
	if err != nil {
		e.pubErrMsg("", "delvps", errcode, err.Error(), mnhostTypes.TOPIC_DELVPS_FAIL)
		return nil
	}

	e.pubMsg("", mnhostTypes.TOPIC_DELVPS_SUCCESS, instanceId)
	log.Printf("success process remove Vps from instanceId:%s\n", instanceId)
	return nil
}

func (e *Vps) processCreateNode(orderId int64, clusterName string) error {
	log.Printf("process create node from orderid %d\n", orderId)

	var torder models.TOrder
	o := orm.NewOrm()
	qs := o.QueryTable("t_order")
	err := qs.Filter("id", orderId).One(&torder)
	if err != nil {
		e.pubErrMsg("", "newnode", utils.RECODE_NODATA, err.Error(), mnhostTypes.TOPIC_NEWNODE_FAIL)
		return err
	}
	userId := strconv.FormatInt(torder.Userid, 10)

	var tcoin models.TCoin
	o = orm.NewOrm()
	qs = o.QueryTable("t_coin")
	err = qs.Filter("name", torder.Coinname).Filter("status", "Enabled").One(&tcoin)
	if err != nil {
		e.pubErrMsg(userId, "newnode", utils.RECODE_NODATA, err.Error(), mnhostTypes.TOPIC_NEWNODE_FAIL)
		return err
	}

	errcode, err := NewNode(orderId, clusterName)
	if err != nil {
		e.pubErrMsg(userId, "newnode", errcode, err.Error(), mnhostTypes.TOPIC_NEWNODE_FAIL)
		log.Printf("err:%s-%s\n", errcode, err)
		return err
	}

	e.pubMsg(userId, mnhostTypes.TOPIC_NEWNODE_SUCCESS, strconv.FormatInt(orderId, 10))
	log.Printf("success process create node %v\n", orderId)
	return nil
}

func (e *Vps) processRemoveNode(nodeId int64, clusterName string) error {
	log.Printf("process remove node from %v\n", nodeId)

	var tnode models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	err := qs.Filter("id", nodeId).One(&tnode)
	if err != nil {
		e.pubErrMsg("", "delnode", utils.RECODE_NODATA, err.Error(), mnhostTypes.TOPIC_DELNODE_FAIL)
		return err
	}
	log.Printf("node:%v\n", tnode)

	userId := strconv.FormatInt(tnode.Userid, 10)

	errcode, err := DelNode(nodeId, clusterName)
	if err != nil {
		e.pubErrMsg(userId, "delnode", errcode, err.Error(), mnhostTypes.TOPIC_DELNODE_FAIL)
		return err
	}

	e.pubMsg(userId, mnhostTypes.TOPIC_DELNODE_SUCCESS, strconv.FormatInt(nodeId, 10))
	log.Printf("success process remove node from %v\n", nodeId)

	return nil
}

func (e *Vps) processExpandVolume(userId, volumeId string, size int64, publicIp, privateIp, password string) error {
	log.Printf("process expandvolume, volume:%s, size:%d\n", volumeId, size)

	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.VPS_CONNECT_ERR, err.Error(), mnhostTypes.TOPIC_EXPANDVOLUME_FAIL)
		return err
	}

	result, err := c.GetDescribeVolumes([]string{volumeId})
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.DESC_VOLUME_ERR, err.Error(), mnhostTypes.TOPIC_EXPANDVOLUME_FAIL)
		return err
	}
	if result == nil {
		e.pubErrMsg(userId, "expandvolume", utils.RECORD_SYSTEMERR, err.Error(), mnhostTypes.TOPIC_EXPANDVOLUME_FAIL)
		return err
	}
	log.Printf("volume:%s\n", result)

	osize := aws.Int64Value(result.Volumes[0].Size)
	if osize > size {
		e.pubErrMsg(userId, "expandvolume", utils.RECODE_DATAERR, "", mnhostTypes.TOPIC_EXPANDVOLUME_FAIL)
		return err
	}

	_, err = c.ModifyVolumes(volumeId, size)
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.MODIFY_VOLUME_ERR, err.Error(), mnhostTypes.TOPIC_EXPANDVOLUME_FAIL)
		return err
	}

	client := common.SshNewClient(publicIp, privateIp, mnhostTypes.SSH_PASSWORD)
	if client == nil {
		e.pubErrMsg(userId, "expandvolume", utils.SSH_CONNECT_ERR, err.Error(), mnhostTypes.TOPIC_EXPANDVOLUME_FAIL)
		return errors.New("client no connect")
	}
	defer client.Close()

	var rsp *gossh.Result
	cmd := fmt.Sprintf("growpart /dev/%s 1", mnhostTypes.DEVICE_NAME)
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.GROWPART_ERR, err.Error(), mnhostTypes.TOPIC_EXPANDVOLUME_FAIL)
		return err
	}

	cmd = fmt.Sprintf("resize2fs /dev/%s", mnhostTypes.DEVICE_NAME1)
	rsp, err = sshCmd(client, cmd, 5)
	if err != nil {
		e.pubErrMsg(userId, "expandvolume", utils.RESIZE_ERR, err.Error(), mnhostTypes.TOPIC_EXPANDVOLUME_FAIL)
		return err
	}
	log.Printf("rsp:%v\n", rsp)

	e.pubMsg(userId, mnhostTypes.TOPIC_EXPANDVOLUME_SUCCESS, volumeId)
	log.Printf("success process expandvolume, volume:%v, size:%d\n", volumeId, size)
	return nil
}

func NewVps(imageId, zone, instanceType, clusterName, role string, volumeSize int64, bAllocation, bVolume bool) (*mnhostTypes.VpsInfo, string, error) {
	log.Printf("start new vps")

	var vpsInfo mnhostTypes.VpsInfo
	vpsInfo.AllocationState = false
	vpsInfo.VolumeState = false

	if len(role) == 0 {
		role = mnhostTypes.ROLE_MANAGER
	}

	if len(instanceType) == 0 {
		instanceType = mnhostTypes.INSTANCE_TYPE_DEFAULT
	}

	if volumeSize == 0 {
		volumeSize = mnhostTypes.VOLUME_SIZE_DEFAULT
	}

	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		return &vpsInfo, utils.VPS_CONNECT_ERR, err
	}

	var securityGroupId string
	groupResult, err := c.GetDescribeSecurityGroupsFromName([]string{mnhostTypes.GROUP_NAME})
	if err != nil {
		ipPermissions := GetIpPermission()
		securityGroupId, err = c.CreateSecurityGroups(mnhostTypes.GROUP_DESC, mnhostTypes.GROUP_DESC, ipPermissions)
		if err != nil {
			return &vpsInfo, utils.CREATE_GROUP_ERR, err
		}
	} else {
		securityGroupId = aws.StringValue(groupResult.SecurityGroups[0].GroupId)
	}

	_, err = c.GetDescribeKeyPairs([]string{mnhostTypes.KEY_PAIR_NAME})
	if err != nil {
		_, err := c.CreateKeyPairs(mnhostTypes.KEY_PAIR_NAME)
		if err != nil {
			return &vpsInfo, utils.CREATE_KEYPAIR_ERR, err
		}
	}

	instanceId, err := c.CreateInstances(imageId, instanceType, mnhostTypes.KEY_PAIR_NAME, securityGroupId)
	if err != nil {
		return &vpsInfo, utils.CREATE_INSTANCE_ERR, err
	}
	vpsInfo.InstanceId = instanceId

	err = c.WaitUntilInstanceStatusOkA([]string{instanceId})
	//err = c.WaitUntilInstanceRun([]string{instanceId})
	if err != nil {
		return &vpsInfo, utils.WAIT_INSTANCE_ERR, err
	}

	result, err := c.GetDescribeInstance([]string{instanceId})
	if err != nil {
		return &vpsInfo, utils.DESC_INSTANCE_ERR, err
	}

	regionName := aws.StringValue(result.Reservations[0].Instances[0].Placement.AvailabilityZone)
	vpsInfo.RegionName = regionName
	vpsInfo.PrivateIp = aws.StringValue(result.Reservations[0].Instances[0].PrivateIpAddress)
	vpsInfo.PublicIp = aws.StringValue(result.Reservations[0].Instances[0].PublicIpAddress)

	if bAllocation {
		publicIp, allocationId, err := c.AllocateAddresss(instanceId)
		if err != nil {
			return &vpsInfo, utils.ALLOCATION_ERR, err
		}
		vpsInfo.AllocationId = allocationId
		vpsInfo.PublicIp = publicIp

		_, err = c.AssociateAddresss(instanceId, allocationId)
		if err != nil {
			return &vpsInfo, utils.ASSOCIATE_ERR, err
		}
		vpsInfo.AllocationState = true
	}

	if bVolume {
		volumeId, err := c.CreateVolumes(regionName, "", volumeSize)
		if err != nil {
			return &vpsInfo, utils.CREATE_VOLUME_ERR, err
		}
		vpsInfo.VolumeId = volumeId

		err = c.WaitUntilVolumeAvailables([]string{volumeId})
		if err != nil {
			return &vpsInfo, utils.WAIT_VOLUME_AVAIL_ERR, err
		}
		_, err = c.AttachVolumes(instanceId, volumeId, mnhostTypes.DEVICE_NAME)
		if err != nil {
			return &vpsInfo, utils.ATTRACH_VOLUME_ERR, err
		}
		vpsInfo.VolumeState = true

		err = VolumeMount(vpsInfo.PublicIp, vpsInfo.PrivateIp, mnhostTypes.SSH_PASSWORD)
		if err != nil {
			return &vpsInfo, utils.MOUNT_VOLUME_ERR, err
		}
	}

	tvps := models.TVps{}
	tvps.AllocationId = vpsInfo.AllocationId
	tvps.InstanceId = vpsInfo.InstanceId
	tvps.VolumeId = vpsInfo.VolumeId
	tvps.ProviderName = mnhostTypes.PROVIDER_NAME
	tvps.Cores = mnhostTypes.CORE_NUMS
	tvps.Memory = mnhostTypes.MEMORY_SIZE
	tvps.KeyPairName = mnhostTypes.KEY_PAIR_NAME
	tvps.SecurityGroupName = mnhostTypes.GROUP_NAME
	tvps.RegionName = vpsInfo.RegionName
	tvps.PublicIp = vpsInfo.PublicIp
	tvps.PrivateIp = vpsInfo.PrivateIp
	tvps.ClusterName = clusterName
	tvps.VpsRole = role
	tvps.Status = "wait-data"
	o := orm.NewOrm()
	_, err = o.Insert(&tvps)
	if err != nil {
		return &vpsInfo, utils.RECODE_INSERTERR, err
	}

	log.Printf("success new vps %s\n", vpsInfo.InstanceId)

	return &vpsInfo, utils.RECODE_OK, nil
}

func DelVps(userId, instanceId string) (string, error) {
	log.Printf("start del vps %s\n", instanceId)

	var tvps models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	err := qs.Filter("instanceId", instanceId).One(&tvps)
	if err != nil {
		return utils.RECODE_DATAERR, err
	}

	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		return utils.VPS_CONNECT_ERR, err
	}

	_, err = c.TerminateInstance(tvps.InstanceId)
	if err != nil {
		log.Printf("term err:%+v\n", err)

		return utils.TERMINATE_INSTANCE_ERR, err
	}

	err = c.WaitUntilInstanceTerminate([]string{tvps.InstanceId})
	if err != nil {
		return utils.WAIT_TERMINATE_INSTANCE_ERR, err
	}

	if len(tvps.VolumeId) != 0 {
		_, err = c.DeleteVolumes(tvps.VolumeId)
		if err != nil {
			return utils.DELETE_VOLUME_ERR, err
		}
	}

	if len(tvps.AllocationId) != 0 {
		_, err = c.ReleaseAddresss(tvps.AllocationId)
		if err != nil {
			return utils.RELEASE_ADDRESS_ERR, err
		}
	}

	o = orm.NewOrm()
	_, err = o.Delete(&tvps)
	if err != nil {
		return utils.RECODE_DELETEERR, err
	}

	log.Printf("success del vps %s\n", instanceId)
	return "", nil
}

func NewNode(orderId int64, clusterName string) (string, error) {
	defer func() {
		mutex.Unlock()
	}()

	log.Printf("start new node from orderid %d\n", orderId)
	var torder models.TOrder
	o := orm.NewOrm()
	qs := o.QueryTable("t_order")
	err := qs.Filter("id", orderId).One(&torder)
	if err != nil {
		return utils.RECODE_NODATA, err
	}

	var tcoin models.TCoin
	o = orm.NewOrm()
	qs = o.QueryTable("t_coin")
	err = qs.Filter("name", torder.Coinname).Filter("status", "Enabled").One(&tcoin)
	if err != nil {
		return utils.RECODE_NODATA, err
	}

	mutex.Lock()
	rpcport, _, err := getRpcPort()
	if err != nil {
		return utils.PORT_ERR, err
	}

	/*publicIp, privateIp, err := common.GetVpsIp(clusterName)
	if err != nil {
		return utils.MANAGER_ERR, err
	}

	errcode, err := ReadyNodeData(torder.Coinname, rpcport, torder.Mnkey, publicIp, privateIp)
	if err != nil {
		return errcode, err
	}

	mc, _, err := common.DockerNewClient(publicIp, privateIp)
	if err != nil {
		return utils.DOCKER_CONNECT_ERR, err
	}
	defer mc.Close()

	err = mc.ServiceCreateA(torder.Coinname, rpcport, tcoin.Docker)
	if err != nil {
		return utils.SERVICE_NEW_ERR, err
	}*/

	var tnode models.TNode
	tnode.ClusterName = clusterName
	tnode.CoinName = torder.Coinname
	tnode.Port = rpcport
	tnode.Userid = torder.Userid
	tnode.Order = &torder
	tnode.Status = "wait-data"
	o = orm.NewOrm()
	_, err = o.Insert(&tnode)
	if err != nil {
		return "", err
	}

	log.Printf("success new node from orderid %d\n", orderId)
	return "", nil
}

func DelNode(nodeId int64, clusterName string) (string, error) {
	defer func() {
	}()
	log.Printf("start del node from nodeId:%d\n", nodeId)

	var tnode models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	err := qs.Filter("id", nodeId).One(&tnode)
	if err != nil {
		return utils.RECODE_NODATA, err
	}

	publicIp, privateIp, err := common.GetVpsIp(clusterName)
	if err != nil {
		return utils.MANAGER_ERR, err
	}

	mc, _, err := common.DockerNewClient(publicIp, privateIp)
	if err != nil {
		return utils.DOCKER_CONNECT_ERR, err
	}
	defer mc.Close()

	nodeName := fmt.Sprintf("%s%d", tnode.CoinName, tnode.Port)
	_, err = mc.ServiceInspectA(nodeName)
	if err != nil {
		bError := true
		if strings.ContainsAny(err.Error(), "No such service") == true {
			bError = false
		}
		if bError {
			return utils.SERVICE_REMOVE_ERR, err
		}
	} else {
		err = mc.ServiceRemove(context.Background(), nodeName)
		if err != nil {
			return utils.SERVICE_REMOVE_ERR, err
		}
	}

	var tcoin models.TCoin
	o = orm.NewOrm()
	qs = o.QueryTable("t_coin")
	err = qs.Filter("name", tnode.CoinName).Filter("port", tnode.Port).One(&tcoin)
	if err != nil {
		log.Printf("******")
		errcode, err := NodeRemoveData(tnode.CoinName, tnode.Port, publicIp, privateIp)
		if err != nil {
			log.Printf("remove data error:%s-%+v\n", errcode, err)
			return errcode, err
		}
	}

	o = orm.NewOrm()
	_, err = o.Delete(&tnode)
	if err != nil {
		return utils.RECODE_DELETEERR, err
	}

	log.Printf("success del node from orderid:%d\n", nodeId)

	return utils.RECODE_OK, nil
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
			SetFromPort(2375).
			SetToPort(2375).
			SetIpRanges([]*ec2.IpRange{
				(&ec2.IpRange{}).
					SetCidrIp("0.0.0.0/0"),
			}),
		(&ec2.IpPermission{}).
			SetIpProtocol("tcp").
			SetFromPort(2049).
			SetToPort(2049).
			SetIpRanges([]*ec2.IpRange{
				(&ec2.IpRange{}).
					SetCidrIp("0.0.0.0/0"),
			}),
		(&ec2.IpPermission{}).
			SetIpProtocol("tcp").
			SetFromPort(mnhostTypes.PORT_FROM).
			SetToPort(mnhostTypes.PORT_TO).
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

func UploadFile(publicIp string, username string, password string, localFile string, remoteDir string) error {
	client, err := ssh.NewClient(publicIp, "22", username, password)
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
	cfg.Section("").Key("rpcuser").SetValue(mnhostTypes.RPC_USER)
	cfg.Section("").Key("rpcpassword").SetValue(mnhostTypes.RPC_PASSWORD)
	cfg.Section("").Key("rpcallowip").SetValue("1.2.3.4/0.0.0.0")
	cfg.Section("").Key("rpcbind").SetValue("0.0.0.0")
	cfg.Section("").Key("rpcport").SetValue(mnhostTypes.S_RPCPROT)
	cfg.Section("").Key("port").SetValue(mnhostTypes.S_PORT)
	cfg.Section("").Key("masternode").SetValue("1")
	cfg.Section("").Key("masternodeblsprivkey").SetValue(mnKey)
	cfg.Section("").Key("externalip").SetValue(externIp)

	cfg.SaveTo(confName)

	return nil
}

func getRpcPort() (int, int, error) {
	rpcport := mnhostTypes.PORT_FROM
	port := rpcport + 1

	var tnodes []models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	nums, err := qs.All(&tnodes)
	if err != nil {
		return 0, 0, err
	}
	if nums == 0 {
		return rpcport, port, nil
	}

	for i := mnhostTypes.PORT_FROM; i < mnhostTypes.PORT_TO; i = i + 2 {
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

func VolumeMount(publicIp, privateIp, password string) error {
	log.Println("start mount")

	client := common.SshNewClient(publicIp, privateIp, password)
	if client == nil {
		return errors.New("client no connect")
	}

	defer client.Close()

	cmd := fmt.Sprintf("echo -e \"d\n1\nw\" | fdisk /dev/%s", mnhostTypes.DEVICE_NAME)
	result, err := client.Execute(cmd)
	/*if err != nil {
		return err
	}*/

	cmd = fmt.Sprintf("echo -e \"n\np\n1\n\n\nw\" | fdisk /dev/%s", mnhostTypes.DEVICE_NAME)
	result, err = client.Execute(cmd)
	log.Printf("cmd:%s\n", cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("fdisk -l |grep %s", mnhostTypes.DEVICE_NAME1)
	log.Printf("cmd:%s\n", cmd)
	result, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("file -s /dev/%s |grep ext4", mnhostTypes.DEVICE_NAME1)
	log.Printf("cmd:%s\n", cmd)
	result, err = client.Execute(cmd)
	if err != nil {
		cmd = fmt.Sprintf("mkfs -t ext4 /dev/%s", mnhostTypes.DEVICE_NAME1)
		result, err = client.Execute(cmd)
		if err != nil {
			return err
		}
	}

	cmd = fmt.Sprintf("mkdir -p %s", mnhostTypes.MOUNT_POINT)
	log.Printf("cmd:%s\n", cmd)
	result, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("mount /dev/%s %s", mnhostTypes.DEVICE_NAME1, mnhostTypes.MOUNT_POINT)
	log.Printf("cmd:%s\n", cmd)
	result, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("df -h |grep %s", mnhostTypes.DEVICE_NAME1)
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
		ClusterName:       u.ClusterName,
		ProviderName:      u.ProviderName,
		Cores:             strconv.Itoa(u.Cores),
		Memory:            strconv.Itoa(u.Memory),
		RegionName:        u.RegionName,
		InstanceId:        u.InstanceId,
		VolumeId:          u.VolumeId,
		SecurityGroupName: u.SecurityGroupName,
		KeyPairName:       u.KeyPairName,
		AllocationId:      u.AllocationId,
		PublicIp:          u.PublicIp,
		PrivateIp:         u.PrivateIp,
	}
}

func Node2PBNode(u *models.TNode) vps.Node {
	return vps.Node{
		Id:          strconv.FormatInt(u.Id, 10),
		UserId:      strconv.FormatInt(u.Userid, 10),
		OrderId:     strconv.FormatInt(u.Order.Id, 10),
		ClusterName: u.ClusterName,
		CoinName:    u.CoinName,
		Port:        strconv.Itoa(u.Port),
	}
}

func Init(clusterName string) error {
	log.Printf("start init %s\n", clusterName)

	var tvpss []models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	nums, err := qs.Filter("clusterName", clusterName).Filter("vps_role", mnhostTypes.ROLE_MANAGER).All(&tvpss)
	if err != nil {
		log.Fatalf("init: query db tvps error:%+v!\n", err)
		return err
	}

	for _, tvps := range tvpss {
		c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
		if err != nil {
			log.Fatalf("init: client error:%+v!\n", err)
			return err
		}
		result, err := c.GetDescribeInstance([]string{tvps.InstanceId})
		if err != nil {
			log.Fatalf("init: desc instance error:%+v!\n", err)
			return err
		}
		tvps.PrivateIp = aws.StringValue(result.Reservations[0].Instances[0].PrivateIpAddress)
		tvps.PublicIp = aws.StringValue(result.Reservations[0].Instances[0].PublicIpAddress)
		//o = orm.NewOrm()
		_, err = o.Update(&tvps)
		if err != nil {
			log.Fatalf("init: update db tvps :%+v!\n", err)
			return err
		}

		/*err = EfsMount(tvps.PublicIp, tvps.PrivateIp, mnhostTypes.SSH_PASSWORD)
		if err != nil {
			log.Fatalf("init: mount efs :%+v!\n", err)
			return err
		}*/

	}

	if nums < mnhostTypes.INIT_MANAGER_NUMS {
		for i := nums; i < mnhostTypes.INIT_MANAGER_NUMS; i++ {
			err = InitNewVps(clusterName)
			if err != nil {
				log.Fatalf("init: new vps error:%+v!\n", err)
				return err
			}
		}
	}

	/*err = EfsMount(publicIp, privateIp, SSH_PASSWORD)
	if err != nil {
		return err
	}

	c, _, err := common.DockerNewClient(publicIp, privateIp)
	if err != nil {
		return err
	}
	c.Close()

	_, _, err = c.SwarmInspectA()
	if err != nil {
		_, err = c.SwarmInitA(publicIp, privateIp)
		if err != nil {
			return err
		}
	}

	o = orm.NewOrm()
	tvps.Status = "leader"
	_, err = o.Update(&tvps)
	if err != nil {
		return err
	}*/

	log.Printf("success start init %s\n", clusterName)
	return nil
}

func InitNewVps(clusterName string) error {
	role := mnhostTypes.ROLE_MANAGER
	volumeSize := int64(0)
	bAllocation := false
	bVolume := false

	mpublicIp := ""
	mprivateIp := ""

	var tvpss []models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	nums, err := qs.Filter("clusterName", clusterName).Filter("vps_role", mnhostTypes.ROLE_MANAGER).All(&tvpss)
	if err != nil {
		return err
	}

	if nums > 0 {
		mpublicIp, mprivateIp, err = common.GetVpsIp("cluster1")
		if err != nil {
			return err
		}
	}

	vpsInfo, errcode, err := NewVps(mnhostTypes.SYSTEM_IMAGE, mnhostTypes.ZONE_DEFAULT, mnhostTypes.INSTANCE_TYPE_DEFAULT, clusterName, role, volumeSize, bAllocation, bVolume)
	if err != nil {
		log.Panicf("init: new vps error:%s!\n", errcode)
		return err
	}

	if nums == 0 {
		mpublicIp = vpsInfo.PublicIp
		mprivateIp = vpsInfo.PrivateIp
	}

	err = EfsMount(vpsInfo.PublicIp, vpsInfo.PrivateIp, mnhostTypes.SSH_PASSWORD)
	if err != nil {
		return err
	}

	mc, _, err := common.DockerNewClient(mpublicIp, mprivateIp)
	if err != nil {
		return err
	}
	defer mc.Close()

	managerToken, _, err := mc.SwarmInspectA()
	if err != nil {
		_, err = mc.SwarmInitA(mpublicIp, mprivateIp, false)
		if err != nil {
			return err
		}
	} else {
		c, _, err := common.DockerNewClient(vpsInfo.PublicIp, vpsInfo.PrivateIp)
		if err != nil {
			return err
		}
		defer c.Close()

		err = c.SwarmJoinA(vpsInfo.PrivateIp, mprivateIp, managerToken)
		if err != nil {
			return err
		}
	}
	return nil
}

func EfsMount(publicIp, privateIp, password string) error {
	ipAddress := publicIp
	if mnhostTypes.PUBLIC_IP_ENABLED == 0 {
		ipAddress = privateIp
	}
	log.Printf("start efs mount %s\n", ipAddress)

	client := common.SshNewClient(publicIp, privateIp, password)
	if client == nil {
		return errors.New("client no connect")
	}

	defer client.Close()

	/*cmd := "apt-get -y install nfs-common"
	fmt.Printf("cmd:%s\n", cmd)
	_, err := client.Execute(cmd)
	if err != nil {
		return err
	}*/

	cmd := fmt.Sprintf("mkdir -p %s", mnhostTypes.NFS_PATH)
	fmt.Printf("cmd:%s\n", cmd)
	_, err := client.Execute(cmd)
	if err != nil {
		fmt.Println(err)
		return err
	}

	cmd = fmt.Sprintf("mount -t nfs4 -o nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2 %s:/ %s", mnhostTypes.NFS_HOST, mnhostTypes.NFS_PATH)
	fmt.Printf("cmd:%s\n", cmd)
	_, err = client.Execute(cmd)
	if err != nil {
		fmt.Println(err)
		return err
	}

	cmd = fmt.Sprintf("chmod 777 -R %s", mnhostTypes.NFS_PATH)
	fmt.Printf("cmd:%s\n", cmd)
	_, err = client.Execute(cmd)
	if err != nil {
		fmt.Println(err)
		return err
	}

	log.Println("success efs mount ")
	return nil
}

/*func ReadyNodeData(coinName string, rpcport int, mnKey, publicIp, privateIp string) (string, error) {
	log.Printf("start ready data %s%d\n", coinName, rpcport)

	var tcoin models.TCoin
	o := orm.NewOrm()
	qs := o.QueryTable("t_coin")
	err := qs.Filter("name", coinName).Filter("status", "Enabled").One(&tcoin)
	if err != nil {
		return utils.COIN_ENABLED_ERR, err
	}

	client := common.SshNewClient(publicIp, privateIp, mnhostTypes.SSH_PASSWORD)
	if client == nil {
		return "", errors.New("client no connect")
	}
	defer client.Close()

	destPath := fmt.Sprintf("%s/%s/%s%d", mnhostTypes.NFS_PATH, coinName, mnhostTypes.NODE_PREFIX, rpcport)
	cmd := fmt.Sprintf("mkdir -p %s", destPath)
	fmt.Printf("cmd:%s\n", cmd)
	_, err = client.Execute(cmd)
	if err == nil {
		return "", err
	}

	cmd = fmt.Sprintf("chmod 777 -R %s", destPath)
	fmt.Printf("cmd:%s\n", cmd)
	_, err = client.Execute(cmd)
	if err != nil {
		return "", err
	}

	log.Printf("success ready data %s%d\n", coinName, rpcport)
	return "", nil
}*/

func NodeRemoveData(coinName string, rpcport int, publicIp, privateIp string) (string, error) {
	log.Printf("start remove data %s%d\n", coinName, rpcport)

	client := common.SshNewClient(publicIp, privateIp, mnhostTypes.SSH_PASSWORD)
	if client == nil {
		return "", errors.New("ssh client no connect")
	}
	defer client.Close()

	retrys := 0
	for {
		cmd := fmt.Sprintf("rm -rf %s/%s/%s%d", mnhostTypes.NFS_PATH, coinName, mnhostTypes.NODE_PREFIX, rpcport)
		fmt.Printf("cmd:%s\n", cmd)
		_, err := client.Execute(cmd)
		if err == nil {
			log.Printf("success remove data %s%d\n", coinName, rpcport)
			return "", nil
		}

		retrys++
		if retrys >= 5 {
			break
		}

		time.Sleep(time.Second * 5)
	}
	return "", errors.New("remove data timeout")
}
