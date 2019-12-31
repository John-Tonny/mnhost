package common

import (
	"errors"
	"fmt"
	"log"

	//"math"
	//"strconv"
	"strings"
	"sync"
	"time"

	//"github.com/robfig/cron"

	"github.com/John-Tonny/mnhost/model"
	mnhostTypes "github.com/John-Tonny/mnhost/types"

	"github.com/astaxie/beego/orm"
	"github.com/dynport/gossh"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/parnurzeal/gorequest"

	uec2 "github.com/John-Tonny/micro/vps/amazon"
	"github.com/aws/aws-sdk-go/aws"
)

/*const LOGIN_USER = "root"
const ROLE_MANAGER = "manager"
const SSH_PASSWORD = "vpub$999000"*/

func SshNewClient(publicIp, privateIp, password string) *gossh.Client {
	ipAddress := publicIp
	if mnhostTypes.PUBLIC_IP_ENABLED == 0 {
		ipAddress = privateIp
	}
	c := gossh.New(ipAddress, mnhostTypes.LOGIN_USER)
	if c == nil {
		return nil
	}
	c.SetPassword(password)
	return c
}

func GetNodeIp(publicIp, privateIp, coinName string, rpcPort int, wg *sync.WaitGroup) error {
	defer func() {
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("get node ip error:%+v\n", err)
		}
	}()
	nodeName := fmt.Sprintf("%s%d", coinName, rpcPort)
	log.Printf("start get node ip from %s\n", nodeName)

	nodeIpResponse := &mnhostTypes.NodeIpResponse{}
	request := gorequest.New().Timeout(5000 * time.Millisecond)
	ipAddress := privateIp
	if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = publicIp
	}

	url := fmt.Sprintf("http://%s:%d/FindNode", ipAddress, mnhostTypes.SYS_MONITOR_PORT)
	params := &mnhostTypes.NameRequest{
		Name: nodeName,
	}
	resp, _, err1 := request.Post(url).
		SendStruct(params).
		EndStruct(&nodeIpResponse)
	if err1 != nil {
		panic(errors.New("findnode error"))
	} else {
		if resp.StatusCode != 200 {
			panic(errors.New(resp.Status))
		}

		if nodeIpResponse.Code != "200" {
			panic(errors.New(nodeIpResponse.CodeMsg))
		}
	}
	if len(nodeIpResponse.Name) == 0 {
		panic(errors.New("no find service"))
	}

	mc, _, err := DockerNewClient(publicIp, privateIp)
	if err != nil {
		panic(err)
	}
	defer mc.Close()

	f := filters.NewArgs()
	f.Add("name", nodeIpResponse.Name)
	nodes, err := mc.NodeListA(types.NodeListOptions{
		Filters: f,
	})
	if err == nil {
		bReady := false
		for _, node := range nodes {
			role := fmt.Sprintf("%s", node.Spec.Role)
			if role == "manager" {
				if node.ManagerStatus != nil && (node.Status.State == "ready" || node.ManagerStatus.Reachability == "Reachable") {
					bReady = true
					privateIp = node.Status.Addr
					break
				}
			} else {
				if node.Status.State == "ready" {
					bReady = true
					privateIp = node.Status.Addr
					break
				}
			}
		}
		if bReady == true {
			var tvps models.TVps
			o := orm.NewOrm()
			qs := o.QueryTable("t_vps")
			err := qs.Filter("privateIp", privateIp).One(&tvps)
			if err == nil {
				var tnode models.TNode
				o = orm.NewOrm()
				qs = o.QueryTable("t_node")
				err = qs.Filter("coinName", coinName).Filter("rpcPort", rpcPort).One(&tnode)
				if err == nil {
					tnode.PublicIp = tvps.PublicIp
					tnode.PrivateIp = tvps.PrivateIp
					o.Update(&tnode)
					log.Printf("success get node ip:%s-%s", tvps.PublicIp, tvps.PrivateIp)
				}
			}
		}
	}

	return nil
}

func NodeReadyData(publicIp, privateIp, coinName string, rpcPort int, wg *sync.WaitGroup) error {
	defer func() { //匿名函数捕获错误
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("ready data error:%+v\n", err)
		}
	}()

	log.Printf("start ready data %s%d-%s\n", coinName, rpcPort, publicIp)

	var tcoin models.TCoin
	o := orm.NewOrm()
	qs := o.QueryTable("t_coin")
	err := qs.Filter("name", coinName).Filter("status", "Enabled").One(&tcoin)
	if err != nil {
		panic(err)
	}

	var tnode models.TNode
	o = orm.NewOrm()
	qs = o.QueryTable("t_node")
	err = qs.Filter("coinName", coinName).Filter("rpcPort", rpcPort).One(&tnode)
	if err != nil {
		panic(err)
	}

	var tvps models.TVps
	o = orm.NewOrm()
	qs = o.QueryTable("t_vps")
	err = qs.Filter("privateIp", privateIp).One(&tvps)
	if err != nil {
		panic(err)
	}

	deviceName := ""
	odeviceName := tnode.DeviceName
	volumeId := ""
	for i := 0; i < 21; i++ {
		deviceName, err = GetDeviceName(privateIp, odeviceName)
		if err != nil {
			return err
		}
		log.Printf("####start deviceName:%s=%d\n", deviceName, i)
		volumeId, err = VolumeReady(tvps.RegionName, tcoin.SnapshotId, tnode.InstanceId, deviceName, tnode.VolumeId)
		if err == nil {
			break
		}
		if tnode.VolumeId == "" {
			o = orm.NewOrm()
			tnode.VolumeId = volumeId
			tnode.VolumeState = ""
			_, err = o.Update(&tnode)
			if err != nil {
				log.Printf("######update error:%+v\n", err)
				panic(err)
			}
		}
		log.Printf("######get error:%+v\n", err)
		odeviceName = deviceName
	}
	log.Printf("deviceName1:%s\n", deviceName)

	o = orm.NewOrm()
	tnode.DeviceName = deviceName
	tnode.VolumeId = volumeId
	tnode.VolumeState = "attached"
	tnode.Status = "wait-conf"
	_, err = o.Update(&tnode)
	if err != nil {
		panic(err)
	}

	log.Printf("success ready data %s%d\n", coinName, rpcPort)
	return nil
}

func NodeRemoveData(publicIp, privateIp, coinName string, rpcPort int) error {
	defer func() { //匿名函数捕获错误
		err := recover()
		if err != nil {
			log.Printf("remove data error:%+v\n", err)
		}
	}()

	log.Printf("start remove data %s%d-%s\n", coinName, rpcPort, publicIp)

	var tcoin models.TCoin
	o := orm.NewOrm()
	qs := o.QueryTable("t_coin")
	err := qs.Filter("name", coinName).Filter("status", "Enabled").One(&tcoin)
	if err != nil {
		panic(err)
	}

	var tnode models.TNode
	o = orm.NewOrm()
	qs = o.QueryTable("t_node")
	err = qs.Filter("coinName", coinName).Filter("rpcPort", rpcPort).One(&tnode)
	if err != nil {
		panic(err)
	}

	var tvps models.TVps
	o = orm.NewOrm()
	qs = o.QueryTable("t_vps")
	err = qs.Filter("privateIp", privateIp).One(&tvps)
	if err != nil {
		panic(err)
	}

	basicResponse := &mnhostTypes.BasicResponse{}
	request := gorequest.New().Timeout(5000 * time.Millisecond)
	ipAddress := privateIp
	if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = publicIp
	}

	nodeName := fmt.Sprintf("%s%d", coinName, rpcPort)
	url := fmt.Sprintf("http://%s:%d/UmountEbs", ipAddress, mnhostTypes.SYS_MONITOR_PORT)
	params := &mnhostTypes.NameRequest{
		Name: nodeName,
	}
	for i := 0; i < 5; i++ {
		resp, _, err1 := request.Post(url).
			SendStruct(params).
			EndStruct(&basicResponse)
		log.Printf("umount:%+v\n", resp)
		if err1 != nil {
			continue
		} else {
			if resp.StatusCode != 200 {
				continue
			}

			if basicResponse.Code != "200" {
				continue
			}
			break
		}
	}

	deviceName := fmt.Sprintf("%s%s", mnhostTypes.DEVICE_NAME_PREFIX, tnode.DeviceName)
	log.Printf("volume remove:%s-%s-%s-%s\n", tvps.RegionName, tnode.VolumeId, tnode.InstanceId, deviceName)
	err = VolumeRemove(tvps.RegionName, tnode.VolumeId, tnode.InstanceId, tnode.DeviceName)
	if err != nil {
		log.Printf("volume remove:%+v\n", err)
		panic(err)
	}

	o = orm.NewOrm()
	_, err = o.Delete(&tnode)
	if err != nil {
		panic(err)
	}

	log.Printf("success remove data %s%d\n", coinName, rpcPort)
	return nil
}

func GetVpsResource(publicIp, privateIp, role string, inputQ chan mnhostTypes.ResourceInfo) (float32, float32, error) {
	resourceInfo := mnhostTypes.ResourceInfo{
		CpuPercert: -1.0,
		MemPercert: -1.0,
		Role:       role,
	}

	defer func() {
		inputQ <- resourceInfo
	}()

	request := gorequest.New().Timeout(5000 * time.Millisecond)
	ipAddress := privateIp
	if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = publicIp
	}
	url := fmt.Sprintf("http://%s:%d/GetSysStatus", ipAddress, mnhostTypes.SYS_MONITOR_PORT)
	resp, _, err := request.Get(url).EndStruct(&resourceInfo)
	if err != nil {
		return float32(-1.0), float32(-1.0), err[0]
	} else {
		if resp.StatusCode != 200 {
			return float32(-1.0), float32(-1.0), errors.New(resp.Status)
		}

		if resourceInfo.Code != "200" {
			return float32(-1.0), float32(-1.0), errors.New(resourceInfo.CodeMsg)
		}
	}

	//log.Printf("get vps cpu memory from %s, mem:%f, cpu:%f\n", publicIp, resourceInfo.MemPercert, resourceInfo.CpuPercert)
	return resourceInfo.MemPercert, resourceInfo.CpuPercert, nil
}

func NodeReadyConfig(publicIp, privateIp, coinName string, rpcPort int, wg *sync.WaitGroup) error {
	defer func() { //匿名函数捕获错误
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("ready config error:%+v\n", err)
		}
	}()

	log.Printf("ready config:%s%d", coinName, rpcPort)
	var tcoin models.TCoin
	o := orm.NewOrm()
	qs := o.QueryTable("t_coin")
	err := qs.Filter("name", coinName).Filter("status", "Enabled").One(&tcoin)
	if err != nil {
		panic(err)
	}

	var tnode models.TNode
	o = orm.NewOrm()
	qs = o.QueryTable("t_node")
	err = qs.Filter("coinName", coinName).Filter("rpcPort", rpcPort).One(&tnode)
	if err != nil {
		panic(err)
	}

	var torder models.TOrder
	o = orm.NewOrm()
	qs = o.QueryTable("t_order")
	err = qs.Filter("id", tnode.Order.Id).One(&torder)
	if err != nil {
		panic(err)
	}

	nodeName := fmt.Sprintf("%s%d", coinName, rpcPort)
	deviceName := fmt.Sprintf("%s%s", mnhostTypes.DEVICE_NAME_PREFIX, tnode.DeviceName)
	err = EbsMount(tnode.PublicIp, tnode.PrivateIp, deviceName, nodeName)
	if err != nil {
		panic(err)
	}

	basicResponse := &mnhostTypes.BasicResponse{}
	request := gorequest.New().Timeout(5000 * time.Millisecond)
	ipAddress := privateIp
	if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = publicIp
	}

	url := fmt.Sprintf("http://%s:%d/WriteConf", ipAddress, mnhostTypes.SYS_MONITOR_PORT)
	fileName := fmt.Sprintf("%s/%s", tcoin.Path, tcoin.Conf)
	conf := &mnhostTypes.CoinConf{
		CoinName:   coinName,
		RpcPort:    rpcPort,
		MnKey:      torder.Mnkey,
		ExternalIp: publicIp,
		FileName:   fileName,
	}
	log.Printf("conf:%+v\n", conf)
	resp, _, err1 := request.Post(url).
		SendStruct(conf).
		EndStruct(&basicResponse)
	log.Printf("resp***:%+v\n", resp)
	if err1 != nil {
		panic(err)
	} else {
		if resp.StatusCode != 200 {
			panic(errors.New(resp.Status))
		}

		if basicResponse.Code != "200" {
			panic(errors.New(basicResponse.CodeMsg))
		}
	}

	/*
		//重启应用
		nodeName := fmt.Sprintf("%s%d", coinName, rpcPort)
		url = fmt.Sprintf("http://%s:%d/Restart", ipAddress, mnhostTypes.SYS_MONITOR_PORT)
		params := &mnhostTypes.NameRequest{
			Name: nodeName,
		}
		resp, _, err1 = request.Post(url).
			SendStruct(params).
			EndStruct(&basicResponse)
		log.Printf("resp***:%+v\n", resp)
		if err1 != nil {
			return errors.New("http post error")
		} else {
			if resp.StatusCode != 200 {
				return errors.New(resp.Status)
			}

			if basicResponse.Code != "200" {
				return errors.New(basicResponse.CodeMsg)
			}
		}*/

	mpublicIp, mprivateIp, err := GetVpsIp("cluster1")
	if err != nil {
		panic(err)
	}

	mc, _, err := DockerNewClient(mpublicIp, mprivateIp)
	if err != nil {
		panic(err)
	}
	defer mc.Close()

	log.Printf("***remove service %s%d\n", coinName, rpcPort)
	err = mc.ServiceRemoveA(coinName, rpcPort)
	if err != nil {
		log.Printf("remove service error:%+v\n", err)
		errInfo := fmt.Sprintf("service %s%d not found", coinName, rpcPort)
		if !strings.Contains(err.Error(), errInfo) {
			panic(err)
		}
	} else {
		log.Printf("***wait restart %s%d\n", coinName, rpcPort)
		time.Sleep(time.Second * 70)
	}

	mc.ServiceCreateA(coinName, rpcPort, tcoin.Docker, tnode.PrivateIp)
	/*if err != nil {
		log.Printf("ppp:%+v\n", err)
		return err
	}*/

	o = orm.NewOrm()
	tnode.Status = "finish"
	o.Update(&tnode)
	if err != nil {
		panic(err)
	}

	log.Printf("success ready config %s%d", coinName, rpcPort)
	return nil
}

func UpdateVpsLeader(clusterName, privateIp string) error {
	var tvps models.TVps
	var tvps1 models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	err := qs.Filter("clusterName", clusterName).Filter("privateIp", privateIp).One(&tvps)
	if err != nil {
		return err
	}
	o = orm.NewOrm()
	qs = o.QueryTable("t_vps")
	err = qs.Filter("clusterName", clusterName).Filter("status", "leader").One(&tvps1)
	if err != nil {
		o = orm.NewOrm()
		tvps.Status = "leader"
		_, err = o.Update(&tvps)
		if err != nil {
			return err
		}
	} else {
		if tvps1.PrivateIp != tvps.PrivateIp {
			o = orm.NewOrm()
			tvps.Status = "leader"
			_, err = o.Update(&tvps)
			if err != nil {
				return err
			}
			// 取消所有字段的leader
			if tvps.PrivateIp != tvps1.PrivateIp {
				o = orm.NewOrm()
				tvps1.Status = "manager"
				_, err = o.Update(&tvps1)
				if err != nil {
					return err
				}
			}

		}
	}
	return nil
}

func EfsMount(publicIp, privateIp, password string) error {
	basicResponse := &mnhostTypes.BasicResponse{}
	request := gorequest.New().Timeout(5000 * time.Millisecond)
	ipAddress := privateIp
	if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = publicIp
	}
	url := fmt.Sprintf("http://%s:%d/Mount", ipAddress, mnhostTypes.SYS_MONITOR_PORT)
	resp, _, err := request.Get(url).EndStruct(&basicResponse)
	if err != nil {
		return err[0]
	} else {
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
		}

		if basicResponse.Code != "200" {
			return errors.New(basicResponse.CodeMsg)
		}
	}

	return nil
}

func EbsMount(publicIp, privateIp, deviceName, nodeName string) error {
	basicResponse := &mnhostTypes.BasicResponse{}
	request := gorequest.New().Timeout(5000 * time.Millisecond)
	ipAddress := privateIp
	if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = publicIp
	}
	url := fmt.Sprintf("http://%s:%d/MountEbs", ipAddress, mnhostTypes.SYS_MONITOR_PORT)
	params := &mnhostTypes.MountRequest{
		DeviceName: deviceName,
		NodeName:   nodeName,
	}
	resp, _, err1 := request.Post(url).
		SendStruct(params).
		EndStruct(&basicResponse)
	if err1 != nil {
		return errors.New("findnode error")
	} else {
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
		}

		if basicResponse.Code != "200" {
			return errors.New(basicResponse.CodeMsg)
		}
	}
	return nil
}

func GetPublicIpFromVps(privateIp string) (string, string, string, error) {
	var tvps models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	err := qs.Filter("privateIp", privateIp).One(&tvps)
	if err != nil {
		return "", "", "", err
	}

	return tvps.PublicIp, tvps.InstanceId, tvps.RegionName, nil
}

func VolumeReady(zone, snapshotId, instanceId, deviceName, ovolumeId string) (string, error) {
	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		return "", err
	}

	if ovolumeId == "" {
		result, err := c.SnapshotsDescribe(snapshotId)
		if err != nil {
			return "", err
		}
		log.Println(result)

		volumeId, err := c.CreateVolumes(zone, snapshotId, aws.Int64Value(result.Snapshots[0].VolumeSize))
		if err != nil {
			return "", err
		}
		log.Printf("volumeId:%s\n", volumeId)

		err = c.WaitUntilVolumeAvailables([]string{volumeId})
		if err != nil {
			return volumeId, err
		}
		log.Printf("create volume wait finish:%s\n", volumeId)
		ovolumeId = volumeId
	}

	deviceName = fmt.Sprintf("/dev/%s%s", mnhostTypes.DEVICE_NAME_PREFIX, deviceName)
	log.Printf("device***:%s\n", deviceName)
	resp, err := c.AttachVolumes(instanceId, ovolumeId, deviceName)
	if err != nil {
		return ovolumeId, err
	}

	err = c.WaitUntilVolumeAttach(3, ovolumeId)
	if err != nil {
		log.Printf("***device err:%+v\n", err)
		resp1, err1 := c.DetachVolumes(instanceId, ovolumeId, deviceName)
		if err1 != nil {
			log.Printf("***device err1:%+v\n", err1)
			return ovolumeId, err1
		}
		log.Printf("***detach volume :%+v\n", resp1)

		err1 = c.WaitUntilVolumeAvailables([]string{ovolumeId})
		if err1 != nil {
			log.Printf("***device err2:%+v\n", err1)
			return ovolumeId, err1
		}
		log.Printf("volume wait finish:%s\n", ovolumeId)

		return ovolumeId, err
	}

	log.Printf("attach volume :%s\n", aws.StringValue(resp.VolumeId))
	return aws.StringValue(resp.VolumeId), nil
}

func VolumeRemove(zone, volumeId, instanceId, deviceName string) error {
	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		return err
	}

	deviceName = fmt.Sprintf("/dev/%s%s", mnhostTypes.DEVICE_NAME_PREFIX, deviceName)
	resp1, err := c.DetachVolumes(instanceId, volumeId, deviceName)
	if err != nil {
		return err
	}
	log.Printf("detach volume :%+v\n", resp1)

	err = c.WaitUntilVolumeAvailables([]string{volumeId})
	if err != nil {
		return err
	}
	log.Printf("create volume wait finish:%s\n", volumeId)

	resp, err := c.DeleteVolumes(volumeId)
	if err != nil {
		return err
	}
	log.Printf("del volume:%+v\n", resp)

	return nil
}

func GetDeviceName(privateIp, odeviceName string) (string, error) {
	start := mnhostTypes.DEVICE_NAME_FROM[0]
	stop := mnhostTypes.DEVICE_NAME_TO[0]
	log.Printf("start:%d--%s,stop:%d\n", start, string(start), stop)
	if len(odeviceName) > 0 {
		start = odeviceName[0] + 1
	}
	log.Printf("start1:%d--%s,stop:%d\n", start, string(start), stop)
	var tnodes []models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	nums, err := qs.Filter("privateIp", privateIp).All(&tnodes)
	if err != nil {
		return "", err
	}
	if nums == 0 {
		if start > stop {
			return "", errors.New("volume1 is full")
		}
		return fmt.Sprintf("%s", string(start)), nil
	}

	var i byte
	for i = start; i <= stop; i++ {
		tmp := fmt.Sprintf("%s", string(i))
		//log.Printf("devicename:%s\n", tmp)
		if deviceExist(tmp, &tnodes) == false {
			return tmp, nil
		}
	}
	return "", errors.New("volume is full")
}

func deviceExist(deviceName string, tnodes *[]models.TNode) bool {
	for _, node := range *tnodes {
		if node.DeviceName == deviceName {
			return true
		}
	}
	return false
}
