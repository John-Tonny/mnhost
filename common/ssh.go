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

/*
func GetVpsIp(clusterName string) (string, string, error) {
	var tvpss []models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	nums, err := qs.Filter("clusterName", clusterName).Filter("vps_role", ROLE_MANAGER).All(&tvpss)
	if err != nil {
		return "", "", err
	}
	if nums == 0 {
		return "", "", errors.New("no manager")
	}

	for _, tvps := range tvpss {
		client := SshNewClient(tvps.PublicIp, tvps.PrivateIp, SSH_PASSWORD)
		if client == nil {
			return "", "", errors.New("client no connect")
		}
		defer client.Close()

		cmd := "docker node ls"
		_, err = client.Execute(cmd)
		if err == nil {

			return tvps.PublicIp, tvps.PrivateIp, nil
		}
	}
	return "", "", errors.New("no manager")
}
*/

func GetNodeIp(publicIp, privateIp, coinName string, rpcPort int, wg *sync.WaitGroup) error {
	defer func() {
		wg.Done()
	}()
	nodeName := fmt.Sprintf("%s%d", coinName, rpcPort)
	log.Printf("start get node ip from %s\n", nodeName)

	nodeIpResponse := &mnhostTypes.NodeIpResponse{}
	request := gorequest.New()
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
		return errors.New("http post error")
	} else {
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
		}

		if nodeIpResponse.Code != "200" {
			return errors.New(nodeIpResponse.CodeMsg)
		}
	}
	if len(nodeIpResponse.Name) == 0 {
		return errors.New("no find service")
	}

	mc, _, err := DockerNewClient(publicIp, privateIp)
	if err != nil {
		return err
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
				err = qs.Filter("coinName", coinName).Filter("port", rpcPort).One(&tnode)
				if err == nil {
					tnode.PublicIp = tvps.PublicIp
					tnode.PrivateIp = tvps.PrivateIp
					o.Update(&tnode)
					log.Printf("success get node ip:%s-%s", tvps.PublicIp, tvps.PrivateIp)
				}
			}
		}
	}

	/*client := SshNewClient(publicIp, privateIp, mnhostTypes.SSH_PASSWORD)
	if client == nil {
		return errors.New("client no connect")
	}
	defer client.Close()

	cmd := fmt.Sprintf("docker service ps %s | awk '{if($5==\"Running\"){print $4;}}'", nodeName)
	log.Println(cmd)
	result, err := client.Execute(cmd)
	if err == nil {
		hostName := strings.Replace(result.Stdout(), "\r", "", -1)
		hostName = strings.Replace(hostName, "\n", "", -1)
		privateIp := ""
		if len(hostName) > 0 {
			f := filters.NewArgs()
			f.Add("name", hostName)
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
						err = qs.Filter("coinName", coinName).Filter("port", rpcPort).One(&tnode)
						if err == nil {
							tnode.PublicIp = tvps.PublicIp
							tnode.PrivateIp = tvps.PrivateIp
							o.Update(&tnode)
							log.Printf("success get node ip:%s-%s", tvps.PublicIp, tvps.PrivateIp)
						}
					}
				}
			}
		}
	}*/

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
		return err
	}

	var tnode models.TNode
	o = orm.NewOrm()
	qs = o.QueryTable("t_node")
	err = qs.Filter("coinName", coinName).Filter("port", rpcPort).One(&tnode)
	if err != nil {
		return err
	}
	/*tnode.Status = "processing"
	o.Update(&tnode)
	if err != nil {
		panic(err)
	}

	var torder models.TOrder
	o = orm.NewOrm()
	qs = o.QueryTable("t_order")
	err = qs.Filter("id", tnode.Order.Id).One(&torder)
	if err != nil {
		panic(err)
	}*/

	client := SshNewClient(publicIp, privateIp, mnhostTypes.SSH_PASSWORD)
	if client == nil {
		log.Printf("client****%s\n", client)
		panic(errors.New("client no connect"))
	}
	defer client.Close()

	cmd := fmt.Sprintf("find %s/%s/ -name %s", mnhostTypes.NFS_PATH, coinName, tcoin.Path)
	fmt.Printf("cmd:%s\n", cmd)
	result, err := client.Execute(cmd)
	if err != nil {
		log.Printf("err1****#########%s\n", err)
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
				/*cmd = fmt.Sprintf("chmod 777 -R %s", destPath)
				fmt.Printf("cmd:%s\n", cmd)
				_, err = client.Execute(cmd)
				if err != nil {
					panic(err)
				}*/
			}
		}

		cmd = fmt.Sprintf("rm %s/%s/%s", destPath, tcoin.Path, ".lock")
		fmt.Printf("cmd:%s\n", cmd)
		client.Execute(cmd)

		cmd = fmt.Sprintf("rm %s/%s/%s", destPath, tcoin.Path, "testnet3/.lock")
		fmt.Printf("cmd:%s\n", cmd)
		client.Execute(cmd)
	}

	o = orm.NewOrm()
	tnode.Status = "wait-conf"
	o.Update(&tnode)
	if err != nil {
		panic(err)
	}

	log.Printf("success ready data %s%d\n", coinName, rpcPort)
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
	//log.Printf("get vps cpu memory from %s\n", publicIp)
	/*client := SshNewClient(publicIp, privateIp, mnhostTypes.SSH_PASSWORD)
	if client == nil {
		return -1, -1, errors.New("client no connect")
	}
	defer client.Close()

	cmd := "free -m  | awk -F '[ :]+' 'NR==2{print $2,$7}'"
	fmt.Printf("cmd:%s\n", cmd)
	result, err := client.Execute(cmd)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return -1, -1, err
	}

	resp := fmt.Sprintf("%s", result.Stdout())

	resp1 := strings.Split(resp, " ")
	if len(resp1) < 2 {
		return -1, -1, errors.New("params error")
	}

	totalMemory := strings.Replace(resp1[0], "\r", "", -1)
	totalMemory = strings.Replace(totalMemory, "\n", "", -1)

	availMemory := strings.Replace(resp1[1], "\r", "", -1)
	availMemory = strings.Replace(availMemory, "\n", "", -1)

	if len(totalMemory) > 0 && len(availMemory) > 0 {
		total, err := strconv.ParseInt(totalMemory, 10, 64)
		if err != nil {
			return -1, -1, err
		}
		avail, err := strconv.ParseInt(availMemory, 10, 64)
		if err != nil {
			return -1, -1, err
		}

		tmp := fmt.Sprintf("%.2f", float64(avail)/float64(total))
		tmp1, err := strconv.ParseFloat(tmp, 64)
		if err != nil {
			return -1, -1, err
		}
		mem = int(tmp1 * 100)
	}

	log.Printf("%s-%s-%d", totalMemory, availMemory, mem)

	cmd = "top -n 1 | awk -F '[ %]+' 'NR==3 {print $9}'"
	fmt.Printf("cmd:%s\n", cmd)
	result, err = client.Execute(cmd)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return mem, -1, err
	}

	resp = fmt.Sprintf("%s", result.Stdout())
	if len(resp) < 2 {
		return mem, -1, errors.New("params error")
	}
	resp = resp[:2]
	cpu1, err := strconv.ParseInt(resp, 10, 64)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return mem, -1, err
	}
	//cpu := int(math.Floor(tmp1 + 0.5))
	cpu = int(cpu1)*/

	request := gorequest.New()
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
	}()

	var tcoin models.TCoin
	o := orm.NewOrm()
	qs := o.QueryTable("t_coin")
	err := qs.Filter("name", coinName).Filter("status", "Enabled").One(&tcoin)
	if err != nil {
		return err
	}

	var tnode models.TNode
	o = orm.NewOrm()
	qs = o.QueryTable("t_node")
	err = qs.Filter("coinName", coinName).Filter("port", rpcPort).One(&tnode)
	if err != nil {
		return err
	}

	var torder models.TOrder
	o = orm.NewOrm()
	qs = o.QueryTable("t_order")
	err = qs.Filter("id", tnode.Order.Id).One(&torder)
	if err != nil {
		return err
	}

	/*client := SshNewClient(publicIp, privateIp, mnhostTypes.SSH_PASSWORD)
	if client == nil {
		return errors.New("client no connect")
	}
	defer client.Close()

	destPath := fmt.Sprintf("%s/%s/%s%d", mnhostTypes.NFS_PATH, coinName, mnhostTypes.NODE_PREFIX, rpcPort)
	cmd := fmt.Sprintf("sed -i '/^rpcport/c rpcport='%d'' %s/%s/%s", rpcPort, destPath, tcoin.Path, tcoin.Conf)
	log.Printf("cmd:%s\n", cmd)
	_, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("sed -i '/^port/c port='%d'' %s/%s/%s", rpcPort+1, destPath, tcoin.Path, tcoin.Conf)
	log.Printf("cmd:%s\n", cmd)
	_, err = client.Execute(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("sed -i '/masternodeblsprivkey/c masternodeblsprivkey='%s'' %s/%s/%s", torder.Mnkey, destPath, tcoin.Path, tcoin.Conf)
	log.Printf("cmd:%s\n", cmd)
	_, err = client.Execute(cmd)
	if err != nil {
		log.Printf("aaa3:%+v\n", err)
		return err
	}

	cmd = fmt.Sprintf("sed -i '/externalip/c externalip='%s'' %s/%s/%s", tnode.PublicIp, destPath, tcoin.Path, tcoin.Conf)
	log.Printf("cmd:%s\n", cmd)
	_, err = client.Execute(cmd)
	if err != nil {
		log.Printf("aaa2:%+v\n", err)
		return err
	}*/

	basicResponse := &mnhostTypes.BasicResponse{}
	request := gorequest.New()
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
	//log.Printf("resp***:%+v\n", resp)
	if err1 != nil {
		return err
	} else {
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
		}

		if basicResponse.Code != "200" {
			return errors.New(basicResponse.CodeMsg)
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
		return err
	}

	mc, _, err := DockerNewClient(mpublicIp, mprivateIp)
	if err != nil {
		return err
	}
	defer mc.Close()

	log.Printf("***remove service %s%d\n", coinName, rpcPort)
	err = mc.ServiceRemoveA(coinName, rpcPort)
	if err != nil {
		log.Printf("ppp:%+v\n", err)
		errInfo := fmt.Sprintf("service %s%d not found", coinName, rpcPort)
		if !strings.ContainsAny(err.Error(), errInfo) == true {
			return err
		}
	} else {
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
		return err
	}

	log.Println("success ready config")

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
	request := gorequest.New()
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
