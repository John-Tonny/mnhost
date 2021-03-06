package common

import (
	"errors"
	"fmt"
	"log"

	//"math"
	//"strconv"
	"sort"
	"strings"
	"sync"
	"time"

	//"github.com/robfig/cron"

	"github.com/John-Tonny/mnhost/conf"
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
	ipAddress := privateIp
	if config.GetMyConst("publicIpEnabled") == "1" {
		//if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = publicIp
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
	if config.GetMyConst("publicIpEnabled") == "1" {
		//if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
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

	log.Printf("start ready data %s-%s\n", coinName, publicIp)

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

	publicIp, privateIp, instanceId, _, err := AllocateVps()
	if err != nil {
		panic(err)
	}

	rpcport, port, err := getRpcPort()
	if err != nil {
		panic(err)
	}

	deviceNo := byte(0)
	odeviceNo := tnode.DeviceNo
	volumeId := ""
	for i := 0; i < mnhostTypes.DEVICE_MAX_NUMS; i++ {
		deviceNo, err = GetDeviceNo(privateIp, odeviceNo)
		if err != nil {
			panic(err)
		}
		deviceName := GetRealDeviceName(deviceNo)
		log.Printf("####start deviceName:%s-%d,deviceNo:%d\n", deviceName, i, deviceNo)
		volumeId, err = VolumeReady(tvps.RegionName, tcoin.SnapshotId, instanceId, deviceName, tnode.VolumeId)
		if err == nil {
			log.Printf("success get deviceName:%s\n", deviceName)
			break
		}
		log.Printf("######get volume error:%+v\n", err)
		if tnode.VolumeId == "" {
			o = orm.NewOrm()
			tnode.VolumeId = volumeId
			tnode.VolumeState = ""
			_, err = o.Update(&tnode)
			if err != nil {
				log.Printf("tnode update volumeid error:%+v\n", err)
				panic(err)
			}
		}
		odeviceNo = deviceNo
	}

	if deviceNo == 0 {
		panic(errors.New("create volume error"))
	}

	o = orm.NewOrm()
	tnode.RpcPort = rpcport
	tnode.Port = port
	tnode.PublicIp = publicIp
	tnode.PrivateIp = privateIp
	tnode.InstanceId = instanceId
	tnode.DeviceNo = deviceNo
	tnode.VolumeId = volumeId
	tnode.VolumeState = "attached"
	tnode.Status = "wait-conf"
	_, err = o.Update(&tnode)
	if err != nil {
		panic(err)
	}

	log.Printf("success ready data %s%d\n", coinName, rpcport)
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
	if config.GetMyConst("publicIpEnabled") == "1" {
		//if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
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

	deviceName := GetRealDeviceName(tnode.DeviceNo)
	log.Printf("volume remove:%s-%s-%s-%s\n", tvps.RegionName, tnode.VolumeId, tnode.InstanceId, deviceName)
	err = VolumeRemove(tvps.RegionName, tnode.VolumeId, tnode.InstanceId, deviceName)
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
	if config.GetMyConst("publicIpEnabled") == "1" {
		//if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
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

func NodeReadyConfig(publicIp, privateIp, coinName, volumeId string, rpcPort int, wg *sync.WaitGroup) error {
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
	deviceName := GetRealDeviceName(tnode.DeviceNo)
	log.Printf("ready config :%s-%s\n", nodeName, deviceName)
	err = EbsMount(tnode.PublicIp, tnode.PrivateIp, deviceName, nodeName, volumeId)
	if err != nil {
		panic(err)
	}

	basicResponse := &mnhostTypes.BasicResponse{}
	request := gorequest.New().Timeout(5000 * time.Millisecond)
	ipAddress := privateIp
	if config.GetMyConst("publicIpEnabled") == "1" {
		//if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
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
	if config.GetMyConst("publicIpEnabled") == "1" {
		//if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
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

func EbsMount(publicIp, privateIp, deviceName, nodeName, volumeId string) error {
	basicResponse := &mnhostTypes.BasicResponse{}
	request := gorequest.New().Timeout(5000 * time.Millisecond)
	ipAddress := privateIp
	if config.GetMyConst("publicIpEnabled") == "1" {
		//if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = publicIp
	}
	url := fmt.Sprintf("http://%s:%d/MountEbs", ipAddress, mnhostTypes.SYS_MONITOR_PORT)
	params := &mnhostTypes.MountRequest{
		DeviceName: deviceName,
		NodeName:   nodeName,
		VolumeId:   volumeId,
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

	deviceName = fmt.Sprintf("/dev/%s", deviceName)
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

	deviceName = fmt.Sprintf("/dev/%s", deviceName)
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

func GetDeviceNo(privateIp string, odeviceNo byte) (byte, error) {
	start := byte(1)
	stop := byte(mnhostTypes.DEVICE_MAX_NUMS)
	if odeviceNo > 0 {
		start = odeviceNo + 1
	}
	var tnodes []models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	nums, err := qs.Filter("privateIp", privateIp).All(&tnodes)
	if err != nil {
		return 0, err
	}
	if nums == 0 {
		if start > stop {
			return 0, errors.New("volume1 is full")
		}
		return start, nil
	}

	var i byte
	for i = start; i <= stop; i++ {
		if deviceExist(i, &tnodes) == false {
			return i, nil
		}
	}
	return 0, errors.New("volume is full")
}

func deviceExist(deviceNo byte, tnodes *[]models.TNode) bool {
	for _, node := range *tnodes {
		if node.DeviceNo == deviceNo {
			return true
		}
	}
	return false
}

func AllocateVps() (string, string, string, string, error) {
	var tvpss []models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	nums, err := qs.All(&tvpss)
	if err != nil {
		return "", "", "", "", err
	}

	if nums == 0 {
		return "", "", "", "", errors.New("no vps")
	}

	type NodeInfo struct {
		PrivateIp string
		Nums      int
	}

	var nodes []NodeInfo
	for _, tvps := range tvpss {
		var tnodes []models.TNode
		o = orm.NewOrm()
		qs = o.QueryTable("t_node")
		nodenums, err := qs.Filter("privateIp", tvps.PrivateIp).All(&tnodes)
		if err != nil {
			return "", "", "", "", err
		}
		nodes = append(nodes, NodeInfo{tvps.PrivateIp, int(nodenums)})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Nums < nodes[j].Nums // 升序
		//return nodes[i].Nums > nodes[j].Nums // 降序
	})
	publicIp, instanceId, regionName, err := GetPublicIpFromVps(nodes[0].PrivateIp)
	if err != nil {
		return "", nodes[0].PrivateIp, "", "", err
	}

	return publicIp, nodes[0].PrivateIp, instanceId, regionName, nil
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
		if node.RpcPort == port {
			return true
		}
	}
	return false
}

func GetDiskInfo(tnode models.TNode, wg *sync.WaitGroup) (int64, int64, error) {
	o := orm.NewOrm()
	tnode.VolumeTotal = -1
	tnode.VolumeFree = -1
	diskInfo := mnhostTypes.DiskInfo{
		Total: -1.0,
		Free:  -1.0,
	}
	defer func() {
		log.Printf("***getDiskInfo exit:%s-%s%d\n", tnode.PublicIp, tnode.CoinName, tnode.RpcPort)
		_, err1 := o.Update(&tnode)
		if err1 != nil {
			log.Printf("diskinfo update error:%+v\n", err1)
		}
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("get diskinfo error:%+v\n", err)
		}
	}()

	if tnode.Status != "finish" {
		panic(errors.New("status is not finish"))
	}

	log.Printf("***get diskinfo:%s-%s%d\n", tnode.PublicIp, tnode.CoinName, tnode.RpcPort)

	request := gorequest.New().Timeout(5000 * time.Millisecond)
	ipAddress := tnode.PrivateIp
	if config.GetMyConst("publicIpEnabled") == "1" {
		//if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = tnode.PublicIp
	}

	nodeName := fmt.Sprintf("%s%d", tnode.CoinName, tnode.RpcPort)
	log.Printf("*****device:%s\n", nodeName)
	url := fmt.Sprintf("http://%s:%d/GetDiskInfo", ipAddress, mnhostTypes.SYS_MONITOR_PORT)
	params := &mnhostTypes.NameRequest{
		Name: nodeName,
	}
	for i := 0; i < 5; i++ {
		resp, _, err1 := request.Post(url).
			SendStruct(params).
			EndStruct(&diskInfo)
		log.Printf("diskinfo:%+v\n", resp)
		if err1 != nil {
			continue
		} else {
			if resp.StatusCode != 200 {
				continue
			}

			if diskInfo.Code != "200" {
				continue
			}
			tnode.VolumeTotal = diskInfo.Total
			tnode.VolumeFree = diskInfo.Free
			break
		}
	}

	log.Printf("diskinfo:%s-%d-%d-%s%d\n", tnode.PublicIp, diskInfo.Total, diskInfo.Free, tnode.CoinName, tnode.RpcPort)
	return diskInfo.Total, diskInfo.Free, nil
}

func EbsModify(tnode models.TNode, size int64, c *uec2.EC2Client, wg *sync.WaitGroup) error {
	defer func() {
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("ebs modify error:%+v\n", err)
		}
	}()
	if tnode.Status != "finish" {
		panic(errors.New("status is not finish"))
	}

	deviceName := GetRealDeviceName(tnode.DeviceNo)
	log.Printf("***ebs modify:%s-%s-%s\n", tnode.PublicIp, tnode.VolumeId, deviceName)

	result, err := c.GetDescribeVolumes([]string{tnode.VolumeId})
	if err != nil {
		panic(err)
	}
	if result == nil {
		panic(errors.New("result is nil"))
	}
	log.Printf("volume:%s\n", result)

	osize := aws.Int64Value(result.Volumes[0].Size)
	if osize > size {
		panic(errors.New("volume haved modify"))
	}

	_, err = c.ModifyVolumes(tnode.VolumeId, size)
	if err != nil {
		//panic(err)
	}

	time.Sleep(time.Second * 10)

	request := gorequest.New().Timeout(5000 * time.Millisecond)
	ipAddress := tnode.PrivateIp
	if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = tnode.PublicIp
	}

	url := fmt.Sprintf("http://%s:%d/ModifyEbs", ipAddress, mnhostTypes.SYS_MONITOR_PORT)
	basicResponse := &mnhostTypes.BasicResponse{}
	params := &mnhostTypes.MountRequest{
		DeviceName: deviceName,
		NodeName:   fmt.Sprintf("%s%d", tnode.CoinName, tnode.RpcPort),
		VolumeId:   tnode.VolumeId,
	}
	for i := 0; i < 5; i++ {
		resp, _, err1 := request.Post(url).
			SendStruct(params).
			EndStruct(&basicResponse)
		log.Printf("ebs modify:%+v\n", resp)
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

	time.Sleep(time.Second * 5)

	result, err = c.GetDescribeVolumes([]string{tnode.VolumeId})
	if err != nil {
		panic(err)
	}
	if result == nil {
		panic(errors.New("result is nil"))
	}
	log.Printf("volume:%s\n", result)

	if aws.Int64Value(result.Volumes[0].Size) != size {
		panic(errors.New("volume size dismatch"))
	}

	log.Printf("success ebs modify:%s-%s-%s\n", tnode.PublicIp, tnode.VolumeId, deviceName)
	return nil
}

func GetRealDeviceName(deviceNo byte) string {
	start := mnhostTypes.DEVICE_NAME_FROM[0]
	if config.GetMyConst("deviceMode") == "1" {
		return fmt.Sprintf("%s%d%s", "nvme", deviceNo, "n1p1")
	} else {
		if deviceNo <= 26 {
			return fmt.Sprintf("%s%s", "xvda", string(start+deviceNo-1))
		}
		return fmt.Sprintf("%s%s", "xvdb", string(byte(start+deviceNo-27)))
	}
}
