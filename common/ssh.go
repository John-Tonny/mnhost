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

	/*client := SshNewClient(publicIp, privateIp, mnhostTypes.SSH_PASSWORD)
	if client == nil {
		panic(errors.New("client no connect"))
	}
	defer client.Close()

	cmd := fmt.Sprintf("find %s/%s/ -name %s", mnhostTypes.NFS_PATH, coinName, tcoin.Path)
	fmt.Printf("cmd:%s\n", cmd)
	result, err := client.Execute(cmd)
	if err != nil {
		panic(err)
	}

	sourcePath := ""
	destPath := ""
	resp := fmt.Sprintf("%s", result.Stdout())
	for {
		pos := strings.Index(resp, "\n")
		if pos > 0 {
			sourcePath = strings.Replace(resp[:pos], "\r", "", -1)
			sourcePath = sourcePath[:len(sourcePath)-len(tcoin.Path)-1]
		}
		break
	}

	if len(sourcePath) > 0 {
		pos := strings.Index(sourcePath, tcoin.Path)
		if pos > 0 {
			sourcePath = sourcePath[:pos]
		}
		destPath = fmt.Sprintf("%s/%s/%s%d", mnhostTypes.NFS_PATH, coinName, mnhostTypes.NODE_PREFIX, rpcPort)
		if sourcePath != destPath {
			cmd = fmt.Sprintf("find %s/%s/ -name %s%d", mnhostTypes.NFS_PATH, coinName, mnhostTypes.NODE_PREFIX, rpcPort)
			fmt.Printf("cmd:%s\n", cmd)
			result, err = client.Execute(cmd)
			if err != nil {
				panic(err)
			}
			tmp := fmt.Sprintf("%s", result.Stdout())
			tmp1 := fmt.Sprintf("%s%d", mnhostTypes.NODE_PREFIX, rpcPort)
			if strings.Index(tmp, tmp1) < 0 {
				cmd = fmt.Sprintf("cp -r %s %s", sourcePath, destPath)
				fmt.Printf("cmd:%s\n", cmd)
				_, err = client.Execute(cmd)
				if err != nil {
					panic(err)
				}
			}
		}

		cmd = fmt.Sprintf("rm %s/%s/%s", destPath, tcoin.Path, ".lock")
		fmt.Printf("cmd:%s\n", cmd)
		client.Execute(cmd)

		cmd = fmt.Sprintf("rm %s/%s/%s", destPath, tcoin.Path, "testnet3/.lock")
		fmt.Printf("cmd:%s\n", cmd)
		client.Execute(cmd)
	}*/

	volumeId := ""
	if tnode.VolumeId == "" {
		volumeId, err = VolumeReady(tvps.RegionName, tcoin.SnapshotId, tnode.InstanceId, tnode.DeviceName)
		if err != nil {
			panic(err)
		}
	}

	o = orm.NewOrm()
	tnode.VolumeId = volumeId
	tnode.Status = "wait-conf"
	o.Update(&tnode)
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
	resp, _, _ := request.Post(url).
		SendStruct(params).
		EndStruct(&basicResponse)
	log.Printf("umount:%+v\n", resp)
	/*if err1 != nil {
		panic(errors.New("umount ebs error"))
	} else {
		if resp.StatusCode != 200 {
			panic(errors.New(resp.Status))
		}

		if basicResponse.Code != "200" {
			panic(errors.New(basicResponse.CodeMsg))
		}
	}*/

	log.Printf("volume remove:%s-%s-%s-%s\n", tvps.RegionName, tnode.VolumeId, tnode.InstanceId, tnode.DeviceName)
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
	err = EbsMount(tnode.PublicIp, tnode.PrivateIp, tnode.DeviceName, nodeName)
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

func VolumeReady(zone, snapshotId, instanceId, deviceName string) (string, error) {
	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		return "", err
	}

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
		return "", err
	}
	log.Printf("create volume wait finish:%s\n", volumeId)

	resp, err := c.AttachVolumes(instanceId, volumeId, deviceName)
	if err != nil {
		return "", err
	}

	err = c.WaitUntilVolumeAttach(20, volumeId)
	if err != nil {
		return "", err
	}

	log.Printf("attach volume :%s\n", aws.StringValue(resp.VolumeId))
	return aws.StringValue(resp.VolumeId), nil
}

func VolumeRemove(zone, volumeId, instanceId, deviceName string) error {
	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		return err
	}

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
