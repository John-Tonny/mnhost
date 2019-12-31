package handler

import (
	"context"
	"encoding/json"
	"errors"

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

	//uec2 "github.com/John-Tonny/micro/vps/amazon"
	//"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/service/ec2"

	logPB "github.com/John-Tonny/mnhost/interface/out/log"
	mnPB "github.com/John-Tonny/mnhost/interface/out/mnmsg"
	monitor "github.com/John-Tonny/mnhost/interface/out/monitor"

	"github.com/John-Tonny/go-virclerpc"

	"github.com/docker/docker/api/types"
	//"github.com/docker/docker/api/types/swarm"
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

	go ProcessUpdateService(req.CoinName, req.DockerName)

	log.Println("update service request success")
	return nil
}

func MonitorVps() {
	for {
		startTime := time.Now().Unix()

		var tvpss []models.TVps
		o := orm.NewOrm()
		qs := o.QueryTable("t_vps")
		nums, err := qs.Filter("clusterName", "cluster1").All(&tvpss)
		if err != nil {
			DelayTime(startTime, 20, "monitor vps time")
		}

		if nums == 0 {
			DelayTime(startTime, 20, "monitor vps time")
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

		if managers > 0 {
			log.Printf("manager:%d-%f-%f", managers, managerInfo.CpuPercert/float32(managers), managerInfo.MemPercert/float32(managers))
		} else {
			log.Printf("manager:%d-%f-%f", managers, managerInfo.CpuPercert, managerInfo.MemPercert)
		}
		log.Printf("worker:%d-%f-%f", workers, workerInfo.CpuPercert, workerInfo.MemPercert)

		close(inputQ)

		if managerInfo.CpuPercert <= 20 || managerInfo.MemPercert <= 20 {

		}

		DelayTime(startTime, 20, "monitor vps time")
	}
}

func MonitorNode() error {
	for {
		startTime := time.Now().Unix()

		var tvpss []models.TVps
		o := orm.NewOrm()
		qs := o.QueryTable("t_vps")
		nums, err := qs.Filter("clusterName", "cluster1").All(&tvpss)
		if err != nil {
			DelayTime(startTime, 20, "monitor node time")
		}

		if nums == 0 {
			DelayTime(startTime, 20, "monitor node time")
		}

		mpublicIp, mprivateIp, err := common.GetVpsIp("cluster1")
		if err != nil {
			DelayTime(startTime, 20, "monitor node time")
		}

		mc, _, err := common.DockerNewClient(mpublicIp, mprivateIp)
		if err != nil {
			DelayTime(startTime, 20, "monitor node time")
		}
		if mc == nil {
			DelayTime(startTime, 20, "monitor node time")
		}
		defer mc.Close()

		nodes, err := mc.NodeListA(types.NodeListOptions{})
		if err != nil {
			log.Printf("%+v\n", err)
			DelayTime(startTime, 20, "monitor node time")
		}

		leaderPrivateIp := ""
		allnodes := mnhostTypes.NodeMap{
			Node: make(map[string]*mnhostTypes.NodeInfo),
		}
		for _, node := range nodes {
			if node.Spec.Role == "manager" {
				if node.ManagerStatus != nil && node.ManagerStatus.Leader == true {
					leaderPrivateIp = strings.Split(node.ManagerStatus.Addr, ":")[0]
					//log.Printf("leaderIp:%s\n", leaderPrivateIp)
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
			if role == "manager" {
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
			} else {
				var tvps models.TVps
				o := orm.NewOrm()
				qs := o.QueryTable("t_vps")
				err := qs.Filter("privateIp", privateIp).One(&tvps)
				if err != nil {
					return err
				}
				allnodes.Node[privateIp].PublicIp = tvps.PublicIp
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

		DelayTime(startTime, 20, "monitor node time")
	}
}

func MonitorService() {
	for {
		startTime := time.Now().Unix()

		var tnodes []models.TNode
		o := orm.NewOrm()
		qs := o.QueryTable("t_node")
		nums, err := qs.Filter("clusterName", "cluster1").All(&tnodes)
		if err != nil {
			DelayTime(startTime, 20, "monitor service time")
		}

		if nums == 0 {
			DelayTime(startTime, 20, "monitor service time")
		}

		mpublicIp, mprivateIp, err := common.GetVpsIp("cluster1")
		if err != nil {
			DelayTime(startTime, 20, "monitor service time")
		}

		mc, _, err := common.DockerNewClient(mpublicIp, mprivateIp)
		if err != nil {
			DelayTime(startTime, 20, "monitor service time")
		}
		if mc == nil {
			DelayTime(startTime, 20, "monitor service time")
		}
		defer mc.Close()

		var wg sync.WaitGroup
		wg.Add(int(nums))
		for _, tnode := range tnodes {
			if tnode.Status == "wait-data" {
				common.NodeReadyData(mpublicIp, mprivateIp, tnode.CoinName, tnode.RpcPort, &wg)
			} else {
				ProcessService(tnode.ClusterName, tnode.CoinName, tnode.Status, tnode.PrivateIp, mpublicIp, mprivateIp, tnode.RpcPort, mc, &wg)
			}
		}
		wg.Wait()

		DelayTime(startTime, 20, "monitor service time")
	}
}

func MonitorApp() {
	for {
		startTime := time.Now().Unix()

		var tnodes []models.TNode
		o := orm.NewOrm()
		qs := o.QueryTable("t_node")
		nums, err := qs.Filter("clusterName", "cluster1").All(&tnodes)
		if err != nil {
			DelayTime(startTime, 20, "monitor app time")
		}

		if nums == 0 {
			DelayTime(startTime, 20, "monitor app time")
		}

		/*mpublicIp, mprivateIp, err := common.GetVpsIp("cluster1")
		if err != nil {
			DelayTime(startTime, 20, "monitor app time")
		}*/

		var wg sync.WaitGroup
		wg.Add(int(nums))
		aaa := 0
		log.Printf("node nums:%d\n", nums)
		for _, tnode := range tnodes {
			aaa++
			//if tnode.PublicIp == "" || tnode.PrivateIp == "" {
			//	common.GetNodeIp(mpublicIp, mprivateIp, tnode.CoinName, tnode.Port, &wg)
			if tnode.Status == "wait-conf" {
				common.NodeReadyConfig(tnode.PublicIp, tnode.PrivateIp, tnode.CoinName, tnode.RpcPort, &wg)
				log.Printf("***bbb:%d---%+v\n", aaa, tnode)
			} else if tnode.Status == "finish" {
				GetMasterNodeStatus(tnode.PublicIp, tnode.PrivateIp, tnode.CoinName, tnode.RpcPort, &wg)
				log.Printf("***ccc:%d---%+v\n", aaa, tnode)
			} else {
				wg.Done()
				log.Printf("***ddd:%d---%+v\n", aaa, tnode)
			}
		}
		wg.Wait()

		DelayTime(startTime, 20, "monitor app time")
	}
}

func ProcessNode(nodeInfo *mnhostTypes.NodeInfo, managerPublicIp, managerPrivateIp string, mc *common.DockerClient, wg *sync.WaitGroup) error {
	defer func() {
		wg.Done()
		err := recover()
		if err != nil {
			log.Printf("get node ip error:%+v\n", err)
		}
	}()

	if !nodeInfo.Status {
		log.Printf("start node repaire %s-%s-%s-%s-%s", nodeInfo.PublicIp, nodeInfo.PrivateIp, managerPublicIp, managerPrivateIp, nodeInfo.Role)
		if mc == nil {
			panic(errors.New("no client mc"))
		}
		managerToken, workerToken, err := mc.SwarmInspectA()
		if err != nil {
			panic(err)
		} else {
			err = SwarmReJoin(managerToken, workerToken, nodeInfo.PublicIp, nodeInfo.PrivateIp, managerPrivateIp, nodeInfo.Role, true)
			if err != nil {
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
		//tnode.PublicIp = ""
		//tnode.PrivateIp = ""
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

func ProcessUpdateService(coinName, dockerId string) error {
	log.Println("start update service")
	var tnodes []models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	nums, err := qs.Filter("clusterName", "cluster1").Filter("coinName", coinName).All(&tnodes)
	if err != nil {
		return err
	}

	if nums == 0 {
		return errors.New("no node")
	}

	for _, tnode := range tnodes {
		go ServiceUpdate(tnode.PublicIp, tnode.PrivateIp, tnode.CoinName, dockerId, tnode.RpcPort)
	}
	log.Println("finish monitor app")
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
	if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
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
	//c := cron.New()
	//spec := "*/20 * * * * ?"
	/*c.AddFunc(spec, func() {
		err := MonitorNode()
		if err != nil {
			log.Printf("node err:%+v\n", err)
		}
	})*/
	//spec1 := "*/20 * * * * ?"
	/*c.AddFunc(spec1, func() {
		err := MonitorService()
		if err != nil {
			log.Printf("service err:%+v\n", err)
		}
	})*/

	//spec2 := "*/20 * * * * ?"
	/*c.AddFunc(spec2, func() {
		err := MonitorApp()
		if err != nil {
			log.Printf("app err:%+v\n", err)
		}
	})*/
	///spec3 := "*/20 * * * * ?"
	/*c.AddFunc(spec3, func() {
		err := MonitorVps()
		if err != nil {
			log.Printf("node err:%+v\n", err)
		}
	})*/
	//c.Start()

	go MonitorVps()
	go MonitorNode()
	go MonitorService()
	go MonitorApp()

}

func SwarmReInit(clusterName string) error {
	log.Println("start repair swarm")
	var tvps models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	err := qs.Filter("clusterName", clusterName).Filter("status", "leader").One(&tvps)
	if err != nil {
		return err
	}

	mc, _, err := common.DockerNewClient(tvps.PublicIp, tvps.PrivateIp)
	if err != nil {
		return err
	}
	if mc == nil {
		return errors.New("ip error")
	}
	defer mc.Close()

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
	//tnode.PublicIp = ""
	//tnode.PrivateIp = ""
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
