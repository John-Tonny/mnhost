package main

import (
	"encoding/json"
	"fmt"

	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"

	"github.com/go-ini/ini"
)

type SysInfo struct {
	//必须的大写开头
	Code       string
	CodeMsg    string
	CpuPercert float32
	MemPercert float32
}

type DiskInfo struct {
	Code    string
	CodeMsg string
	Total   int64
	Free    int64
}

type NameInfo struct {
	//必须的大写开头
	Name string
}

type NodeNameInfo struct {
	CoinName string
	RpcPort  int
}

type MountInfo struct {
	//必须的大写开头
	DeviceName string
	NodeName   string
}

type Conf struct {
	CoinName   string
	RpcPort    int
	MnKey      string
	ExternalIp string
	FileName   string
}

type RespInfo struct {
	//必须的大写开头
	Code    string
	CodeMsg string
}

type NameRespInfo struct {
	//必须的大写开头
	Code    string
	CodeMsg string
	Name    string
}

type SysController struct {
	beego.Controller
}

type DiskController struct {
	beego.Controller
}

type ConfController struct {
	beego.Controller
}

type RestartController struct {
	beego.Controller
}

type FindNodeController struct {
	beego.Controller
}

type MountController struct {
	beego.Controller
}

type MountEbsController struct {
	beego.Controller
}

type ModifyEbsController struct {
	beego.Controller
}

type UmountEbsController struct {
	beego.Controller
}

const (
	MOUNT_PATH   = "/mnt/vircle"
	NFS_PATH     = "/mnt/efs"
	NODE_PREFIX  = "node"
	RPC_USER     = "vpub"
	RPC_PASSWORD = "vpub999000"
	NFS_HOST     = "172.31.43.253:/"
)

func (this *SysController) Get() {
	v, _ := mem.VirtualMemory()
	cc, _ := cpu.Percent(time.Second, false)
	data := &SysInfo{
		"200",
		"成功",
		float32(cc[0]),
		float32(v.UsedPercent)}
	this.Data["json"] = data
	this.ServeJSON()
}

func (this *DiskController) Post() {
	params := NameInfo{}
	body := this.Ctx.Input.RequestBody
	err := json.Unmarshal(body, &params)
	data := &DiskInfo{}
	if err != nil {
		data = &DiskInfo{
			"400",
			"参数错误",
			-1,
			-1}
	} else {
		fmt.Printf("params:%+v\n", params)
		pdir := fmt.Sprintf("%s/%s", MOUNT_PATH, params.Name)
		info, err := disk.Usage(pdir)
		if err != nil {
			data = &DiskInfo{
				"400",
				"参数错误",
				-1,
				-1}
		} else {
			data = &DiskInfo{
				"200",
				"成功",
				int64(info.Total),
				int64(info.Free)}
		}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

func (this *ConfController) Post() {
	conf := Conf{}
	body := this.Ctx.Input.RequestBody
	err := json.Unmarshal(body, &conf)
	data := &RespInfo{}
	if err != nil {
		data = &RespInfo{
			"400",
			"参数错误"}
	} else {
		fmt.Printf("conf:%+v\n", conf)
		if conf.CoinName == "" || conf.RpcPort == 0 || conf.MnKey == "" || conf.ExternalIp == "" || conf.FileName == "" {
			data = &RespInfo{
				"400",
				"参数错误"}

		} else {
			data = &RespInfo{
				"200",
				"成功"}
			writeConf(conf)

		}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

func (this *RestartController) Post() {
	params := NameInfo{}
	body := this.Ctx.Input.RequestBody
	err := json.Unmarshal(body, &params)
	data := &RespInfo{}
	if err != nil {
		data = &RespInfo{
			"400",
			"参数错误"}
	} else {
		fmt.Printf("params:%+v\n", params)
		if params.Name == "" {
			data = &RespInfo{
				"400",
				"参数错误"}

		} else {
			data = &RespInfo{
				"200",
				"成功"}
			restartApp(params.Name)
		}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

func (this *FindNodeController) Post() {
	params := NameInfo{}
	body := this.Ctx.Input.RequestBody
	err := json.Unmarshal(body, &params)
	data := &NameRespInfo{}
	if err != nil {
		data = &NameRespInfo{
			"400",
			"参数错误",
			""}
	} else {
		fmt.Printf("params:%+v\n", params)
		if params.Name == "" {
			data = &NameRespInfo{
				"400",
				"参数错误",
				""}

		} else {
			nodeIp, err := findNodeName(params.Name)
			if err != nil {
				data = &NameRespInfo{
					"401",
					"命令执行错误",
					""}
			} else {
				data = &NameRespInfo{
					"200",
					"成功",
					nodeIp}
			}
		}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

func (this *MountController) Get() {
	fmt.Println("mount efs")
	err := MountEfs()
	data := &RespInfo{}
	if err != nil {
		data = &RespInfo{
			"401",
			"命令执行错误"}
	} else {
		data = &RespInfo{
			"200",
			"成功"}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

func (this *MountEbsController) Post() {
	params := MountInfo{}
	body := this.Ctx.Input.RequestBody
	err := json.Unmarshal(body, &params)
	data := &RespInfo{}
	if err != nil {
		data = &RespInfo{
			"400",
			"参数错误"}
	} else {
		fmt.Printf("params:%+v\n", params)
		if params.DeviceName == "" || params.NodeName == "" {
			data = &RespInfo{
				"400",
				"参数错误"}
		} else {
			err := MountEbs(params.DeviceName, params.NodeName)
			if err != nil {
				data = &RespInfo{
					"401",
					"命令执行错误"}
			} else {
				data = &RespInfo{
					"200",
					"成功"}
			}
		}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

func (this *ModifyEbsController) Post() {
	params := NameInfo{}
	body := this.Ctx.Input.RequestBody
	err := json.Unmarshal(body, &params)
	data := &RespInfo{}
	if err != nil {
		data = &RespInfo{
			"400",
			"参数错误"}
	} else {
		fmt.Printf("params:%+v\n", params)
		if params.Name == "" {
			data = &RespInfo{
				"400",
				"参数错误"}
		} else {
			err := ModifyEbs(params.Name)
			if err != nil {
				data = &RespInfo{
					"401",
					"命令执行错误"}
			} else {
				data = &RespInfo{
					"200",
					"成功"}
			}
		}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

func (this *UmountEbsController) Post() {
	params := NameInfo{}
	body := this.Ctx.Input.RequestBody
	err := json.Unmarshal(body, &params)
	data := &RespInfo{}
	if err != nil {
		data = &RespInfo{
			"400",
			"参数错误"}
	} else {
		fmt.Printf("params:%+v\n", params)
		if params.Name == "" {
			data = &RespInfo{
				"400",
				"参数错误"}
		} else {
			err := UmountEbs(params.Name)
			if err != nil {
				data = &RespInfo{
					"401",
					"命令执行错误"}
			} else {
				data = &RespInfo{
					"200",
					"成功"}
			}
		}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

func writeConf(conf Conf) {
	//fileName := fmt.Sprintf("%s/%s/%s%d/%s", NFS_PATH, conf.CoinName, NODE_PREFIX, conf.RpcPort, conf.FileName)
	fileName := fmt.Sprintf("%s/%s%d/%s", MOUNT_PATH, conf.CoinName, conf.RpcPort, conf.FileName)
	cfg, err := ini.Load(fileName)

	srpcPort := strconv.Itoa(conf.RpcPort)
	sport := strconv.Itoa(conf.RpcPort + 1)

	if err != nil {
		cfg = ini.Empty()

		cfg.Section("").Key("listen").SetValue("1")
		cfg.Section("").Key("server").SetValue("1")
		cfg.Section("").Key("rpcuser").SetValue(RPC_USER)
		cfg.Section("").Key("rpcpassword").SetValue(RPC_PASSWORD)
		cfg.Section("").Key("rpcallowip").SetValue("1.2.3.4/0.0.0.0")
		cfg.Section("").Key("rpcbind").SetValue("0.0.0.0")
		cfg.Section("").Key("rpcport").SetValue(srpcPort)
		cfg.Section("").Key("port").SetValue(sport)
		cfg.Section("").Key("masternode").SetValue("1")
		cfg.Section("").Key("masternodeblsprivkey").SetValue(conf.MnKey)
		cfg.Section("").Key("externalip").SetValue(conf.ExternalIp)

		cfg.SaveTo(fileName)
	} else {
		cfg.Section("").Key("rpcport").SetValue(srpcPort)
		cfg.Section("").Key("port").SetValue(sport)
		cfg.Section("").Key("masternode").SetValue("1")
		cfg.Section("").Key("masternodeblsprivkey").SetValue(conf.MnKey)
		cfg.Section("").Key("externalip").SetValue(conf.ExternalIp)
		cfg.SaveTo(fileName)
	}

	//删除lock
	path1 := path.Dir(fileName)
	fmt.Printf("remove lock :%s\n", path1)
	os.RemoveAll(fmt.Sprintf("%s/%s", path1, ".lock"))
	os.RemoveAll(fmt.Sprintf("%s/%s", path1, "testnet3/.lock"))
}

func restartApp(nodeName string) {
	name := ""
	cmd := exec.Command("docker", "ps", "-a")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	cmd.Start()
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		if strings.Contains(line, nodeName) == true {
			tmp := strings.Split(line, " ")
			name = tmp[0]
			fmt.Printf("name:%s-%d\n", name, len(name))
			break
		}
	}
	cmd.Wait()

	if len(name) > 0 {
		cmd := exec.Command("docker", "restart", name)
		cmd.Run()
	}
	return
}

func findNodeName(nodeName string) (string, error) {
	name := ""
	cmd := exec.Command("docker", "service", "ps", nodeName)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", errors.New("exec cmd error")
	}

	cmd.Start()
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		if strings.Contains(line, nodeName) && strings.Contains(line, "Running") {
			line = DeleteExtraSpace(line)
			tmp := strings.Split(line, " ")
			if len(tmp) >= 6 {
				name = tmp[3]
				fmt.Printf("name:%s-%d\n", name, len(name))
				break
			}
		}
	}
	//cmd.Wait()
	return name, nil
}

func DeleteExtraSpace(s string) string {
	//删除字符串中的多余空格，有多个空格时，仅保留一个空格
	regstr := "\\s{2,}"                          //两个及两个以上空格的正则表达式
	reg, _ := regexp.Compile(regstr)             //编译正则表达式
	s2 := make([]byte, len(s))                   //定义字符数组切片
	copy(s2, s)                                  //将字符串复制到切片
	spc_index := reg.FindStringIndex(string(s2)) //在字符串中搜索
	for len(spc_index) > 0 {                     //找到适配项
		s2 = append(s2[:spc_index[0]+1], s2[spc_index[1]:]...) //删除多余空格
		spc_index = reg.FindStringIndex(string(s2))            //继续在字符串中搜索
	}
	return string(s2)
}

func MountEfs() error {
	cmd := exec.Command("mkdir", NFS_PATH)
	cmd.CombinedOutput()
	/*if err != nil {
		return err
	}*/

	cmd = exec.Command("mount", "-t", "nfs4", "-o", "nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2", NFS_HOST, NFS_PATH)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	/*cmd = exec.Command("chmod", "777", "-R", NFS_PATH)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}*/

	return nil
}

func MountEbs(deviceName, nodeName string) error {
	path := fmt.Sprintf("%s/%s", MOUNT_PATH, nodeName)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	deviceName = fmt.Sprintf("/dev/%s1", deviceName)
	cmd := exec.Command("mount", deviceName, path)
	cmd.CombinedOutput()

	//already mounted on

	cmd = exec.Command("df", "-h")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.New("exec cmd error")
	}

	cmd.Start()
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	fmt.Printf("%s-%s\n", deviceName, path)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		fmt.Printf("%s\n", line)
		if strings.Contains(line, deviceName) && strings.Contains(line, path) {
			fmt.Printf("success mount %s-%s\n", deviceName, nodeName)
			return nil
		}
	}

	return errors.New("no found")
}

func ModifyEbs(deviceName string) error {
	deviceName = fmt.Sprintf("/dev/%s", deviceName)
	cmd := exec.Command("growpart", deviceName, "1")
	cmd.CombinedOutput()
	/*if err != nil {
		return err
	}*/

	deviceName = fmt.Sprintf("%s1", deviceName)
	cmd = exec.Command("resize2fs", deviceName)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return err
	}

	fmt.Printf("modify ebs %s\n", deviceName)
	return nil
}

func UmountEbs(nodeName string) error {
	path := fmt.Sprintf("%s/%s", MOUNT_PATH, nodeName)
	cmd := exec.Command("umount", path)
	cmd.CombinedOutput()

	os.RemoveAll(path)

	cmd = exec.Command("df", "-h")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.New("exec cmd error")
	}

	cmd.Start()
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	fmt.Printf("%s\n", path)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		fmt.Printf("%s\n", line)
		if strings.Contains(line, path) {
			return errors.New("fail umount error")
		}
	}

	fmt.Printf("umount %s\n", path)
	return nil
}

func main() {
	beego.BConfig.CopyRequestBody = true
	beego.Router("/GetSysStatus", &SysController{})
	beego.Router("/GetDiskInfo", &DiskController{})
	beego.Router("/WriteConf", &ConfController{})
	beego.Router("/Restart", &RestartController{})
	beego.Router("/FindNode", &FindNodeController{})
	beego.Router("/Mount", &MountController{})
	beego.Router("/MountEbs", &MountEbsController{})
	beego.Router("/ModifyEbs", &ModifyEbsController{})
	beego.Router("/UmountEbs", &UmountEbsController{})
	beego.Run(":8844")
}
