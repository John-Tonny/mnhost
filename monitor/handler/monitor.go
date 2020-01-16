package handler

import (
	"context"
	"encoding/json"
	"errors"

	"strconv"
	"strings"
	"sync"

	"fmt"
	"log"
	"time"

	//"os"

	//"github.com/dynport/gossh"
	//"github.com/go-ini/ini"
	//"github.com/pytool/ssh"

	"github.com/robfig/cron"

	"github.com/astaxie/beego/orm"
	"github.com/micro/go-micro/broker"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/conf"
	"github.com/John-Tonny/mnhost/model"
	mnhostTypes "github.com/John-Tonny/mnhost/types"
	"github.com/John-Tonny/mnhost/utils"

	uec2 "github.com/John-Tonny/micro/vps/amazon"
	"github.com/aws/aws-sdk-go/aws"

	//"github.com/aws/aws-sdk-go/service/ec2"

	logPB "github.com/John-Tonny/mnhost/interface/out/log"
	mnPB "github.com/John-Tonny/mnhost/interface/out/mnmsg"
	monitor "github.com/John-Tonny/mnhost/interface/out/monitor"
	pb "github.com/John-Tonny/mnhost/interface/out/vps"

	"github.com/John-Tonny/go-virclerpc"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	//"github.com/docker/docker/api/types/filters"
)

type Monitor struct {
	Broker broker.Broker
}

const service = "monitor"

var (
	topic       string
	serviceName string
	version     string

	newVpsFlag bool

	leaderRetrys int

	mutex sync.Mutex

	serviceRetrys mnhostTypes.RetrysMap
	appRetrys     mnhostTypes.RetrysMap
)

func init() {
	topic = config.GetBrokerTopic("log")
	serviceName = config.GetServiceName(service)

	version = config.GetVersion(service)
	if version == "" {
		version = "latest"
	}

	newVpsFlag = false
	serviceRetrys = mnhostTypes.RetrysMap{
		Retrys: make(map[string]*mnhostTypes.RetrysInfo),
	}
	appRetrys = mnhostTypes.RetrysMap{
		Retrys: make(map[string]*mnhostTypes.RetrysInfo),
	}
}

func GetHandler(bk broker.Broker) *Monitor {
	return &Monitor{
		Broker: bk,
	}
}

// 发送vps
func (e *Monitor) pubMsg(userID, topic, msgId string) error {
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

func (e *Monitor) pubErrMsg(userID, method, errno, msg, topic string) error {
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

func (e *Monitor) pubLog(userID, method, msg string) error {
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

func (e *Monitor) GetStatus(ctx context.Context, req *monitor.Request, rsp *monitor.Response) error {
	log.Println("get status request")

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	log.Println("get status request success")
	return nil
}

func (e *Monitor) StartMonitor(ctx context.Context, req *monitor.Request, rsp *monitor.Response) error {
	log.Println("start monitor request")

	c := cron.New()
	spec := "*/30 * * * * ?"
	c.AddFunc(spec, func() {
		ret := false //GetNodeSyncStatus(req.UserId, req.MNKey, dockerid)
		if ret == true {
			c.Stop()
		}
	})
	c.Start()

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	log.Println("start monitor request success")
	return nil
}

func (e *Monitor) StopMonitor(ctx context.Context, req *monitor.Request, rsp *monitor.Response) error {
	log.Println("stop monitor request")

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	log.Println("stop monitor request success")
	return nil
}

func (e *Monitor) UpdateService(ctx context.Context, req *monitor.UpdateRequest, rsp *monitor.Response) error {
	log.Println("update service request")

	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	var tcoin models.TCoin
	o := orm.NewOrm()
	qs := o.QueryTable("t_coin")
	err := qs.Filter("Name", req.CoinName).One(&tcoin)
	if err != nil {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	tcoin.Docker = req.DockerName
	_, err = o.Update(&tcoin)
	if err != nil {
		rsp.Errno = utils.RECODE_UPDATEERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	go ProcessUpdateService(req.CoinName, req.DockerName)

	log.Println("update service request success")
	return nil
}

func MonitorVps() {
	defaultTime, err := strconv.ParseInt(config.GetMyConst("vpsTime"), 10, 64)
	if err != nil {
		panic(err)
	}
	retrys := 0
	for {
		startTime := time.Now().Unix()

		var tvpss []models.TVps
		o := orm.NewOrm()
		qs := o.QueryTable("t_vps")
		nums, err := qs.Filter("clusterName", "cluster1").All(&tvpss)
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor vps time")
			continue
		}

		if nums == 0 {
			DelayTime(startTime, defaultTime, "monitor vps time")
			continue
		}

		var managerInfo mnhostTypes.ResourceInfo
		var workerInfo mnhostTypes.ResourceInfo
		var inputQ = make(chan mnhostTypes.ResourceInfo, 50) // 生产队列

		for _, tvps := range tvpss {
			go common.GetVpsResource(tvps.PublicIp, tvps.PrivateIp, tvps.VpsRole, inputQ)
		}

		time.Sleep(time.Second * 2)

		managers := 0
		workers := 0
		for i := 0; i < len(tvpss); i++ {
			result := <-inputQ
			if result.CpuPercert < 0 || result.MemPercert < 0 {
				continue
			}
			if result.Role == mnhostTypes.ROLE_MANAGER {
				managers++
				managerInfo.CpuPercert += result.CpuPercert
				managerInfo.MemPercert += result.MemPercert
			} else {
				workers++
				workerInfo.CpuPercert += result.CpuPercert
				workerInfo.MemPercert += result.MemPercert
			}
		}

		close(inputQ)

		totals := managers + workers
		cpuPercert := (managerInfo.CpuPercert + workerInfo.CpuPercert) / float32(totals)
		memPercert := (managerInfo.MemPercert + workerInfo.MemPercert) / float32(totals)
		if totals > 0 {
			log.Printf("sysstatus:%d-%d-%f-%f", managers, workers, cpuPercert, memPercert)
		}
		if cpuPercert >= 70.0 || memPercert >= 70.0 {
			if !newVpsFlag {
				role := mnhostTypes.ROLE_MANAGER
				if workers <= managers*40 {
					role = mnhostTypes.ROLE_WORKER
				}
				if retrys > 10 {
					log.Printf("####retrys####:%s-%d\n", role, retrys)
					VpsNew(role)
					retrys = 0
				}

			}
			retrys++
		} else {
			retrys = 0
		}

		DelayTime(startTime, defaultTime, "monitor vps time")
	}
}

func MonitorNode() error {
	defaultTime, err := strconv.ParseInt(config.GetMyConst("nodeTime"), 10, 64)
	if err != nil {
		panic(err)
	}
	for {
		startTime := time.Now().Unix()

		var tvpss []models.TVps
		o := orm.NewOrm()
		qs := o.QueryTable("t_vps")
		nums, err := qs.Filter("clusterName", "cluster1").All(&tvpss)
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor node time")
			continue
		}

		if nums == 0 {
			DelayTime(startTime, defaultTime, "monitor node time")
			continue
		}

		mpublicIp, mprivateIp, err := common.GetVpsIp("cluster1")
		if err != nil {
			SwarmReInit("cluster1")
			DelayTime(startTime, defaultTime, "monitor node time")
			continue
		}

		mc, _, err := common.DockerNewClient(mpublicIp, mprivateIp)
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor node time")
			continue
		}
		if mc == nil {
			DelayTime(startTime, defaultTime, "monitor node time")
			continue
		}

		nodes, err := mc.NodeListA(types.NodeListOptions{})
		if err != nil {
			SwarmReInit("cluster1")
			DelayTime(startTime, defaultTime, "monitor node time")
			continue
		}

		leaderPrivateIp := ""
		allnodes := mnhostTypes.NodeMap{
			Node: make(map[string]*mnhostTypes.NodeInfo),
		}
		for _, node := range nodes {
			if node.Spec.Role == mnhostTypes.ROLE_MANAGER {
				if node.ManagerStatus != nil && node.ManagerStatus.Leader == true {
					leaderPrivateIp = strings.Split(node.ManagerStatus.Addr, ":")[0]
					log.Printf("leaderIp:%s\n", leaderPrivateIp)
					err = common.UpdateVpsLeader("cluster1", leaderPrivateIp)
					if err != nil {
						continue
					}
				}
			}

			role := fmt.Sprintf("%s", node.Spec.Role)
			privateIp := node.Status.Addr

			_, ok := allnodes.Node[privateIp]
			if !ok {
				allnodes.Node[privateIp] = &mnhostTypes.NodeInfo{}
			}
			allnodes.Node[privateIp].PrivateIp = privateIp
			allnodes.Node[privateIp].Role = role

			bReady := false
			if role == mnhostTypes.ROLE_MANAGER {
				if node.ManagerStatus != nil {
					if node.Status.State == "ready" || node.ManagerStatus.Reachability == "Reachable" {
						bReady = true
					}
				}
			} else {
				if node.Status.State == "ready" {
					bReady = true
				}
			}

			if bReady {
				allnodes.Node[privateIp].Status = true
			} /*else {
				var tvps models.TVps
				o := orm.NewOrm()
				qs := o.QueryTable("t_vps")
				err := qs.Filter("privateIp", privateIp).One(&tvps)
				if err != nil {
					continue
				}
				allnodes.Node[privateIp].PublicIp = tvps.PublicIp
			}*/
		}

		for _, tvps := range tvpss {
			if vpsExist(tvps.PrivateIp, nodes) {
				_, ok := allnodes.Node[tvps.PrivateIp]
				if !ok {
					allnodes.Node[tvps.PrivateIp] = &mnhostTypes.NodeInfo{}
				}
				allnodes.Node[tvps.PrivateIp].PrivateIp = tvps.PrivateIp
				allnodes.Node[tvps.PrivateIp].PublicIp = tvps.PublicIp
				allnodes.Node[tvps.PrivateIp].Role = tvps.VpsRole
			} else {
				if tvps.VpsRole == mnhostTypes.ROLE_MANAGER {
					_, ok := allnodes.Node[tvps.PrivateIp]
					if !ok {
						allnodes.Node[tvps.PrivateIp] = &mnhostTypes.NodeInfo{}
					}
					allnodes.Node[tvps.PrivateIp].PrivateIp = tvps.PrivateIp
					allnodes.Node[tvps.PrivateIp].PublicIp = tvps.PublicIp
					allnodes.Node[tvps.PrivateIp].Role = tvps.VpsRole
					allnodes.Node[tvps.PrivateIp].Status = false
				}
			}
		}

		if len(allnodes.Node) > 0 {
			var wg sync.WaitGroup
			wg.Add(len(allnodes.Node))
			for _, nodeInfo := range allnodes.Node {
				ProcessNode(nodeInfo, mpublicIp, mprivateIp, mc, &wg)
			}
			wg.Wait()
		}

		DelayTime(startTime, defaultTime, "monitor node time")
	}
}

func MonitorService() {
	defaultTime, err := strconv.ParseInt(config.GetMyConst("serviceTime"), 10, 64)
	if err != nil {
		panic(err)
	}
	for {
		startTime := time.Now().Unix()

		var tnodes []models.TNode
		o := orm.NewOrm()
		qs := o.QueryTable("t_node")
		nums, err := qs.Filter("clusterName", "cluster1").All(&tnodes)
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor service time")
			continue
		}

		if nums == 0 {
			DelayTime(startTime, defaultTime, "monitor service time")
			continue
		}

		mpublicIp, mprivateIp, err := common.GetVpsIp("cluster1")
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor service time")
			continue
		}

		mc, _, err := common.DockerNewClient(mpublicIp, mprivateIp)
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor service time")
			continue
		}
		if mc == nil {
			DelayTime(startTime, defaultTime, "monitor service time")
			continue
		}
		defer mc.Close()

		var wg sync.WaitGroup
		wg.Add(int(nums))
		for _, tnode := range tnodes {
			if tnode.Status == "wait-data" {
				if !newVpsFlag {
					common.NodeReadyData(mpublicIp, mprivateIp, tnode.CoinName, tnode.RpcPort, &wg)
				} else {
					wg.Done()
					log.Printf("newvps***service:%+v\n", tnode)
				}
			} else if tnode.Status == "wait-conf" {
				common.NodeReadyConfig(tnode.PublicIp, tnode.PrivateIp, tnode.CoinName, tnode.VolumeId, tnode.RpcPort, &wg)
			} else if tnode.Status == "finish" {
				ProcessService(tnode.ClusterName, tnode.CoinName, tnode.Status, tnode.PrivateIp, mpublicIp, mprivateIp, tnode.RpcPort, mc, &wg)
			} else {
				wg.Done()
				log.Printf("***service:%+v\n", tnode)
			}
		}
		wg.Wait()

		totalTime := time.Now().Unix() - startTime
		log.Printf("%s:%d\n", "monitor service time", totalTime)
		time.Sleep(time.Second * time.Duration(defaultTime))
		//DelayTime(startTime, defaultTime, "monitor service time")
	}
}

func MonitorApp() {
	defaultTime, err := strconv.ParseInt(config.GetMyConst("appTime"), 10, 64)
	if err != nil {
		panic(err)
	}
	for {
		startTime := time.Now().Unix()

		var tnodes []models.TNode
		o := orm.NewOrm()
		qs := o.QueryTable("t_node")
		nums, err := qs.Filter("clusterName", "cluster1").All(&tnodes)
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor app time")
			continue
		}

		if nums == 0 {
			DelayTime(startTime, defaultTime, "monitor app time")
			continue
		}

		/*mpublicIp, mprivateIp, err := common.GetVpsIp("cluster1")
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor app time")
		}*/

		var wg sync.WaitGroup
		wg.Add(int(nums))
		aaa := 0
		log.Printf("node nums:%d\n", nums)
		for _, tnode := range tnodes {
			aaa++
			//if tnode.PublicIp == "" || tnode.PrivateIp == "" {
			//	common.GetNodeIp(mpublicIp, mprivateIp, tnode.CoinName, tnode.Port, &wg)
			//if tnode.Status == "wait-conf" {
			//	common.NodeReadyConfig(tnode.PublicIp, tnode.PrivateIp, tnode.CoinName, tnode.RpcPort, &wg)
			//	log.Printf("***bbb:%d---%+v\n", aaa, tnode)
			if tnode.Status == "finish" {
				GetMasterNodeStatus(tnode.PublicIp, tnode.PrivateIp, tnode.CoinName, tnode.RpcPort, &wg)
				log.Printf("***ccc:%d---%+v\n", aaa, tnode)
			} else {
				wg.Done()
				log.Printf("***ddd:%d---%+v\n", aaa, tnode)
			}
		}
		wg.Wait()

		DelayTime(startTime, defaultTime, "monitor app time")
	}
}

func MonitorVolume() {
	defaultTime, err := strconv.ParseInt(config.GetMyConst("volumeTime"), 10, 64)
	if err != nil {
		panic(err)
	}
	minSize, err := strconv.ParseInt(config.GetMyConst("volumeModifySize"), 10, 64)
	if err != nil {
		panic(err)
	}
	for {
		startTime := time.Now().Unix()

		var tnodes []models.TNode
		o := orm.NewOrm()
		qs := o.QueryTable("t_node")
		nums, err := qs.All(&tnodes)
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor volume time")
			continue
		}

		if nums == 0 {
			DelayTime(startTime, defaultTime, "monitor volume time")
			continue
		}

		var wg sync.WaitGroup
		wg.Add(int(nums))
		for _, tnode := range tnodes {
			common.GetDiskInfo(tnode, &wg)
		}
		wg.Wait()

		o = orm.NewOrm()
		qs = o.QueryTable("t_node")
		num1s, err := qs.Filter("volumeFree__lt", minSize).All(&tnodes)
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor volume time")
			continue
		}

		if num1s == 0 {
			DelayTime(startTime, defaultTime, "monitor volume time")
			continue
		}

		c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor volume time")
			continue
		}

		if c == nil {
			DelayTime(startTime, defaultTime, "monitor volume time")
			continue
		}

		wg.Add(int(num1s))
		for _, tnode := range tnodes {
			log.Printf("disk node:%+v\n", tnode)
			size := minSize/1024/1024/1024 + int64((tnode.VolumeTotal+512*1024*1024)/1024/1024/1024)
			common.EbsModify(tnode, size, c, &wg)
		}
		wg.Wait()

		DelayTime(startTime, defaultTime, "monitor volume time")
	}
}

func MonitorSnapshot() {
	defaultTime, err := strconv.ParseInt(config.GetMyConst("snapshotTime"), 10, 64)
	if err != nil {
		panic(err)
	}
	for {
		startTime := time.Now().Unix()

		var tcoins []models.TCoin
		o := orm.NewOrm()
		qs := o.QueryTable("t_coin")
		nums, err := qs.Filter("status", "Enabled").All(&tcoins)
		if err != nil {
			DelayTime(startTime, defaultTime, "monitor snapshot time")
			continue
		}

		if nums == 0 {
			DelayTime(startTime, defaultTime, "monitor snapshot time")
			continue
		}

		var wg sync.WaitGroup
		wg.Add(int(nums))
		for _, tcoin := range tcoins {
			ProcessSnapshot(tcoin, &wg)
		}
		wg.Wait()

		DelayTime(startTime, defaultTime, "monitor snapshot time")
	}
}

func ProcessNode(nodeInfo *mnhostTypes.NodeInfo, managerPublicIp, managerPrivateIp string, mc *common.DockerClient, wg *sync.WaitGroup) error {
	defer func() {
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("process node error:%+v\n", err)
		}
	}()

	if !nodeInfo.Status && len(nodeInfo.PublicIp) > 0 {
		log.Printf("start node repaire %s-%s-%s-%s-%s", nodeInfo.PublicIp, nodeInfo.PrivateIp, managerPublicIp, managerPrivateIp, nodeInfo.Role)
		if mc == nil {
			panic(errors.New("no client mc"))
		}
		managerToken, workerToken, err := mc.SwarmInspectA()
		if err != nil {
			panic(err)
		} else {
			err = SwarmReJoin(managerToken, workerToken, nodeInfo.PublicIp, nodeInfo.PrivateIp, managerPrivateIp, nodeInfo.Role, false)
			if err != nil {
				log.Printf("get node ip error2:%+v\n", err)
			} else {
				log.Printf("success node repaire %s-%s-%s-%s-%s", nodeInfo.PublicIp, nodeInfo.PrivateIp, managerPublicIp, managerPrivateIp, nodeInfo.Role)
			}
		}
	}
	return nil
}

func ProcessService(clusterName, coinName, status, privateIp, managerPublicIp, managerPrivateIp string, rpcPort int, mc *common.DockerClient, wg *sync.WaitGroup) error {
	defer func() {
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("process service error:%+v\n", err)
		}
	}()
	var err error

	nodeName := fmt.Sprintf("%s%d", coinName, rpcPort)
	_, _, err = mc.ServiceInspectWithRaw(context.Background(), nodeName)
	if err != nil {
		_, ok := serviceRetrys.Retrys[nodeName]
		if !ok {
			serviceRetrys.Retrys[nodeName] = &mnhostTypes.RetrysInfo{}
		}
		serviceRetrys.Retrys[nodeName].Nums++
		log.Printf("***serviceRetrys:%s-%d\n", nodeName, serviceRetrys.Retrys[nodeName].Nums)
		if serviceRetrys.Retrys[nodeName].Nums <= 3 {
			return nil
		}
		delete(serviceRetrys.Retrys, nodeName)

		log.Printf("###service inspect %s", nodeName)
		dockerId, err := GetDocker(coinName)
		if err != nil {
			panic(err)
		}
		log.Printf("start service repaire, ip:%s-%s\n", managerPublicIp, managerPrivateIp)
		err = mc.ServiceCreateA(coinName, rpcPort, dockerId, privateIp)
		if err != nil {
			panic(err)
		}

		var tnode models.TNode
		o := orm.NewOrm()
		qs := o.QueryTable("t_node")
		err = qs.Filter("clusterName", clusterName).Filter("coinName", coinName).Filter("rpcPort", rpcPort).One(&tnode)
		if err != nil {
			panic(err)
		}

		mutex.Lock()
		defer mutex.Unlock()
		o = orm.NewOrm()
		tnode.State = ""
		tnode.Status = "wait-conf"
		_, err = o.Update(&tnode)
		if err != nil {
			panic(err)
		}

		log.Printf("success service repaire, ip:%s-%s\n", managerPublicIp, managerPrivateIp)
	} else {
		_, ok := serviceRetrys.Retrys[nodeName]
		if ok {
			log.Printf("service retrys:%d\n", serviceRetrys.Retrys[nodeName].Nums)
			delete(serviceRetrys.Retrys, nodeName)
		}
		log.Printf("***service inspect %s", nodeName)
	}
	return nil
}

func ProcessApp() error {
	return nil
}

func ProcessSnapshot(tcoin models.TCoin, wg *sync.WaitGroup) error {
	defaultTime, err := strconv.ParseInt(config.GetMyConst("snapshotTime"), 10, 64)
	if err != nil {
		panic(err)
	}
	defer func() {
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("process snapshot error:%+v\n", err)
		}
	}()
	log.Printf("process snapshot start:%s\n", tcoin.Name)

	var tnodes []models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	nums, err := qs.Filter("coinName", tcoin.Name).All(&tnodes)
	if err != nil {
		panic(err)
	}

	if nums == 0 {
		panic(errors.New("node no found"))
	}

	last := tcoin.Updatetime.UTC().Unix()
	now := time.Now().UTC().Unix()

	if (now - last) <= defaultTime {
		return nil
	}

	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		panic(err)
	}

	if c == nil {
		panic(err)
	}

	snapshotId := ""
	for _, tnode := range tnodes {
		result, err := c.SnapshotCreate(tnode.VolumeId)
		if err == nil {
			snapshotId = aws.StringValue(result.SnapshotId)

			c.SnapshotDelete(tcoin.SnapshotId)

			o = orm.NewOrm()
			tcoin.SnapshotId = snapshotId
			o.Update(&tcoin)

			log.Printf("process snapshot success:%s-%s\n", tcoin.Name, snapshotId)
			return nil
		}
	}

	return errors.New("volumeid novalid")
}

func ProcessUpdateService(coinName, dockerId string) error {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("process update service error:%s-%+v\n", coinName, err)
		} else {
			log.Printf("success process update service:%s-%s\n", coinName, dockerId)
		}
	}()

	log.Println("start process update service")
	var tnodes []models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	nums, err := qs.Filter("clusterName", "cluster1").Filter("coinName", coinName).All(&tnodes)
	if err != nil {
		panic(err)
	}

	if nums == 0 {
		panic(errors.New("no node"))
	}

	for _, tnode := range tnodes {
		go ServiceUpdate(tnode.PublicIp, tnode.PrivateIp, tnode.CoinName, dockerId, tnode.RpcPort)
	}

	return nil
}

func GetDocker(coinName string) (string, error) {
	var tcoin models.TCoin
	o := orm.NewOrm()
	qs := o.QueryTable("t_coin")
	err := qs.Filter("name", coinName).One(&tcoin)
	if err != nil {
		return "", err
	}
	return tcoin.Docker, nil
}

func GetMasterNodeStatus(publicIp, privateIp, coinName string, rpcPort int, wg *sync.WaitGroup) (string, error) {
	defer func() {
		log.Printf("***getStatus exit:%s-%s%d\n", publicIp, coinName, rpcPort)
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("get masternode status error:%+v\n", err)
		}
	}()
	log.Printf("***getStatus:%s-%s%d\n", publicIp, coinName, rpcPort)
	ipAddress := privateIp
	if config.GetMyConst("publicIpEnabled") == "1" {
		//if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = publicIp
	}

	basicAuth := &virclerpc.BasicAuth{
		Username: mnhostTypes.RPC_USER,
		Password: mnhostTypes.RPC_PASSWORD,
	}
	url := fmt.Sprintf("http://%s:%d", ipAddress, rpcPort)
	client := virclerpc.NewRPCClient(url, basicAuth)
	if client == nil {
		MasterNodeRefused(coinName, rpcPort)
		panic(errors.New("no connect"))
	}

	msg, err := client.GetMasterNode("status")
	if err != nil {
		MasterNodeRefused(coinName, rpcPort)
		panic(err)
	}

	nodeName := fmt.Sprintf("%s%d", coinName, rpcPort)
	_, ok := appRetrys.Retrys[nodeName]
	if ok {
		delete(appRetrys.Retrys, nodeName)
	}

	var tnode models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	err = qs.Filter("coinName", coinName).Filter("rpcPort", rpcPort).One(&tnode)
	if err != nil {
		panic(err)
	}

	tnode.State = msg.State
	_, err = o.Update(&tnode)
	if err != nil {
		panic(err)
	}

	log.Printf("state:%s-%s-%s%d\n", publicIp, msg.State, coinName, rpcPort)
	return msg.State, nil
}

func Init() {
	go MonitorVps()
	go MonitorNode()
	go MonitorService()
	go MonitorApp()
	go MonitorVolume()
	go MonitorSnapshot()
}

func SwarmReInit(clusterName string) error {
	log.Println("start repair swarm")
	var tvps models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	err := qs.Filter("clusterName", clusterName).Filter("status", "leader").One(&tvps)
	if err != nil {
		err = qs.Filter("clusterName", clusterName).One(&tvps)
		if err != nil {
			return err
		}
		tvps.Status = "leader"
		_, err = o.Update(&tvps)
		if err != nil {
			return err
		}
	}

	mc, _, err := common.DockerNewClient(tvps.PublicIp, tvps.PrivateIp)
	if err != nil {
		return err
	}
	if mc == nil {
		return errors.New("ip error")
	}
	defer mc.Close()

	mc.SwarmLeave(context.Background(), true)

	_, err = mc.SwarmInitA(tvps.PublicIp, tvps.PrivateIp, true)
	if err != nil {
		return err
	}

	managerToken, workerToken, err := mc.SwarmInspectA()
	if err != nil {
		return err
	}

	var tvpss []models.TVps
	o = orm.NewOrm()
	qs = o.QueryTable("t_vps")
	_, err = qs.Filter("clusterName", clusterName).All(&tvpss)
	if err != nil {
		return err
	}

	for _, vps := range tvpss {
		if vps.Status != "leader" {
			SwarmReJoin(managerToken, workerToken, vps.PublicIp, vps.PrivateIp, tvps.PrivateIp, vps.VpsRole, false)
		}
	}

	log.Println("success repair swarm")
	return nil
}

func SwarmReJoin(managerToken, workerToken, publicIp, privateIp, leaderPrivateIp, role string, bNew bool) error {
	c, _, err := common.DockerNewClient(publicIp, privateIp)
	if err != nil {
		return err
	}
	defer c.Close()

	token := managerToken
	status := "manager"
	if role == mnhostTypes.ROLE_WORKER {
		token = workerToken
		status = "worker"
	}

	if !bNew {
		log.Printf("swarm leave %s-%s", status, privateIp)
		c.SwarmLeave(context.Background(), true)
	}

	log.Printf("start join %s-%s-%s", status, privateIp, leaderPrivateIp)
	err = c.SwarmJoinA(privateIp, leaderPrivateIp, token)
	if err != nil {
		if strings.Contains(err.Error(), "node is already part of a swarm") == true {
			return nil
		}
		return err
	}
	return nil
}

func ServiceUpdate(publicIp, privateIp, coinName, dockerId string, rpcPort int) error {
	c, _, err := common.DockerNewClient(publicIp, privateIp)
	if err != nil {
		return err
	}
	defer c.Close()

	nodeName := fmt.Sprintf("%s%d", coinName, rpcPort)
	result, _, err := c.ServiceInspectWithRaw(context.Background(), nodeName)
	if err != nil {
		return err
	}

	err = c.ServiceUpdateA(coinName, rpcPort, dockerId, result.Version)
	if err != nil {
		log.Printf("%+v\n", err)
		return err
	}
	return nil
}

func MasterNodeRefused(coinName string, rpcPort int) error {
	nodeName := fmt.Sprintf("%s%d", coinName, rpcPort)
	_, ok := appRetrys.Retrys[nodeName]
	if !ok {
		appRetrys.Retrys[nodeName] = &mnhostTypes.RetrysInfo{}
	}
	appRetrys.Retrys[nodeName].Nums++
	log.Printf("***appRetrys:%s-%d\n", nodeName, appRetrys.Retrys[nodeName].Nums)
	if appRetrys.Retrys[nodeName].Nums < 5 {
		return nil
	}
	delete(appRetrys.Retrys, nodeName)

	var tnode models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	err := qs.Filter("coinName", coinName).Filter("rpcPort", rpcPort).One(&tnode)
	if err != nil {
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	tnode.Status = "wait-conf"
	_, err = o.Update(&tnode)
	if err != nil {
		return err
	}
	return nil
}

func DelayTime(startTime, defaultTime int64, name string) {
	totalTime := time.Now().Unix() - startTime
	log.Printf("%s:%d\n", name, totalTime)
	if totalTime < defaultTime {
		time.Sleep(time.Second * time.Duration(defaultTime))
	} else {
		time.Sleep(time.Second * time.Duration(totalTime))
	}
}

//Error: Cannot obtain a lock on data directory
//Please restart with -reindex

func VpsNewSuccess(pub broker.Event) error {
	newVpsFlag = false
	log.Println("#############new nodesuccess start")
	var msg *mnPB.MnMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}
	userId := pub.Message().Header["user_id"]
	InstanceId := (*msg).MsgId
	log.Printf("new vpssuccess finish, userId:%v,instanceId:%s\n", userId, InstanceId)
	return nil
}

func VpsNewFail(pub broker.Event) error {
	newVpsFlag = false
	log.Printf("##############new vpsfail start")
	var msg *mnPB.MnErrMsg
	if err := json.Unmarshal(pub.Message().Body, &msg); err != nil {
		return err
	}
	userId := pub.Message().Header["user_id"]
	log.Printf("new vpsfail finish, userId:%v,failmsg:%v\n", userId, msg)
	return nil
}

func VpsNew(role string) error {
	service := "vps"
	serviceName = config.GetServiceName(service)
	srv := common.GetMicroClient(service)

	client := pb.NewVpsService(serviceName, srv.Client())
	_, err := client.CreateVps(context.Background(), &pb.CreateVpsRequest{
		ClusterName: "cluster1",
		Role:        role,
		VolumeSize:  0,
	})
	if err != nil {
		return err
	}
	newVpsFlag = true

	return nil
}

func vpsExist(privateIp string, nodes []swarm.Node) bool {
	for _, node := range nodes {
		if privateIp == node.Status.Addr {
			return true
		}
	}
	return false
}

func RemoveNovalidVolume() error {
	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		return err
	}

	infos, err := c.GetDescribeVolumes([]string{})
	for _, info := range infos.Volumes {
		instanceId := aws.StringValue(info.Attachments[0].InstanceId)
		log.Printf("InstanceId:%s\n", instanceId)
	}

	return nil
}
