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
	"github.com/hacdias/fileutils"
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
	VolumeId   string
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
	UUID_ENABLED = 0
)

var (
	g_devMode bool
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
			"磁盘参数错误",
			-1,
			-1}
	} else {
		fmt.Printf("params:%+v\n", params)
		pdir := fmt.Sprintf("%s/%s", MOUNT_PATH, params.Name)
		info, err := disk.Usage(pdir)
		if err != nil {
			data = &DiskInfo{
				"400",
				"磁盘参数错误",
				-1,
				-1}
		} else {
			data = &DiskInfo{
				"200",
				"磁盘成功",
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
			"配置参数错误"}
	} else {
		fmt.Printf("conf:%+v\n", conf)
		if conf.CoinName == "" || conf.RpcPort == 0 || conf.MnKey == "" || conf.ExternalIp == "" || conf.FileName == "" {
			data = &RespInfo{
				"400",
				"配置参数错误"}

		} else {
			data = &RespInfo{
				"200",
				"配置成功"}
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
			"查找节点参数错误",
			""}
	} else {
		fmt.Printf("params:%+v\n", params)
		if params.Name == "" {
			data = &NameRespInfo{
				"400",
				"查找节点参数错误",
				""}

		} else {
			nodeIp, err := findNodeName(params.Name)
			if err != nil {
				data = &NameRespInfo{
					"401",
					"查找节点错误",
					""}
			} else {
				data = &NameRespInfo{
					"200",
					"查找节点成功",
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
			"绑定efs错误"}
	} else {
		data = &RespInfo{
			"200",
			"绑定efs成功"}
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
			"绑定ebs参数错误"}
	} else {
		fmt.Printf("params:%+v\n", params)
		if params.DeviceName == "" || params.NodeName == "" || params.VolumeId == "" {
			data = &RespInfo{
				"400",
				"绑定ebs参数错误"}
		} else {
			err := MountEbs(params.DeviceName, params.NodeName, params.VolumeId)
			if err != nil {
				data = &RespInfo{
					"401",
					"绑定ebs错误"}
			} else {
				data = &RespInfo{
					"200",
					"绑定ebs成功"}
			}
		}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

func (this *ModifyEbsController) Post() {
	params := MountInfo{}
	body := this.Ctx.Input.RequestBody
	err := json.Unmarshal(body, &params)
	data := &RespInfo{}
	if err != nil {
		data = &RespInfo{
			"400",
			"修正ebs参数错误"}
	} else {
		fmt.Printf("params:%+v\n", params)
		if params.DeviceName == "" || params.VolumeId == "" {
			data = &RespInfo{
				"400",
				"修正ebs参数错误"}
		} else {
			err := ModifyEbs(params.DeviceName, params.VolumeId)
			if err != nil {
				data = &RespInfo{
					"401",
					"修正ebs错误"}
			} else {
				data = &RespInfo{
					"200",
					"修正ebs成功"}
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
			"解绑ebs参数错误"}
	} else {
		fmt.Printf("params:%+v\n", params)
		if params.Name == "" {
			data = &RespInfo{
				"400",
				"解绑ebs参数错误"}
		} else {
			err := UmountEbs(params.Name, true, true)
			if err != nil {
				data = &RespInfo{
					"401",
					"解绑ebs错误"}
			} else {
				data = &RespInfo{
					"200",
					"解绑ebs成功"}
			}
		}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

func writeConf(conf Conf) {
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

func restartApp(nodeName string) error {
	cmd := exec.Command("docker", "ps", "-a")
	bFind, name, err := findInfo(cmd, -1, 0, nodeName)
	if err != nil {
		fmt.Printf("fail restart %s\n", nodeName)
		return err
	}

	if bFind && (len(name) > 0) {
		cmd := exec.Command("docker", "restart", name)
		cmd.Run()
	}
	return nil
}

func findNodeName(nodeName string) (string, error) {
	cmd := exec.Command("docker", "service", "ps", nodeName)
	_, name, err := findInfo(cmd, 6, 3, nodeName, "Running")
	if err != nil {
		fmt.Printf("fail findnode %s\n", nodeName)
		return "", err
	}
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

func MountEbs(deviceName, nodeName, volumeId string) error {
	mpath := fmt.Sprintf("%s/%s", MOUNT_PATH, nodeName)
	if !g_devMode {
		deviceName = fmt.Sprintf("/dev/%s%d", deviceName, 1)
	} else {
		deviceName = fmt.Sprintf("%s%s", GetRealDeviceName(volumeId), "p1")
	}

	cmd := exec.Command("df", "-h")
	bFind, _, err := findInfo(cmd, -1, -1, deviceName, mpath)
	if err != nil {
		fmt.Printf("fail mount %s-%s\n", deviceName, nodeName)
		return err
	}

	if !bFind {
		cmd := exec.Command("df", "-h")
		bFind, result, err := findInfo(cmd, -1, -1, deviceName)
		if err != nil {
			return err
		}

		if bFind {
			err = UmountEbs(result, true, false)
			if err != nil {
				fmt.Printf("fail mount %s-%s\n", deviceName, nodeName)
				return err
			}
		}

		err = os.MkdirAll(mpath, os.ModePerm)
		if err != nil {
			fmt.Printf("fail mount %s-%s\n", deviceName, nodeName)
			return err
		}

		fmt.Printf("mount path:%s-%s\n", deviceName, mpath)
		if mpath == "/mnt/vircle" {
			panic(errors.New("error mount"))
		}
		cmd = exec.Command("mount", deviceName, mpath)
		_, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("fail mount %s-%s\n", deviceName, nodeName)
			return err
		}
	}

	err = WriteFstab(deviceName, nodeName, volumeId)
	if err != nil {
		return err
	}

	fmt.Printf("success mount %s-%s\n", deviceName, mpath)
	return nil
}

func ModifyEbs(deviceName, volumeId string) error {
	defer func() {
	}()
	if !g_devMode {
		deviceName = fmt.Sprintf("/dev/%s", deviceName)
	} else {
		deviceName = GetRealDeviceName(volumeId)
	}
	fmt.Printf("deviceName:%s\n", deviceName)
	cmd := exec.Command("growpart", deviceName, "1")
	cmd.CombinedOutput()
	/*if err != nil {
		return err
	}*/

	if !g_devMode {
		deviceName = fmt.Sprintf("%s%d", deviceName, 1)
	} else {
		deviceName = fmt.Sprintf("%s%s", deviceName, "p1")
	}
	fmt.Printf("deviceName:%s\n", deviceName)
	cmd = exec.Command("resize2fs", deviceName)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return err
	}

	fmt.Printf("modify ebs %s\n", deviceName)
	return nil
}

func UmountEbs(nodeName string, bDel, bSimple bool) error {
	mpath := nodeName
	if bSimple {
		mpath = fmt.Sprintf("%s/%s", MOUNT_PATH, nodeName)
	}
	defer func() {
		if bDel {
			os.RemoveAll(mpath)
		}
	}()

	cmd := exec.Command("df", "-h")
	_, deviceName, err := findInfo(cmd, -1, 0, mpath)
	if err != nil {
		return err
	}

	if bDel {
		if UUID_ENABLED == 1 {
			_, uuid, err := GetVolumeUUID(deviceName)
			if err != nil {
				return err
			}

			//删除匹配行
			info := fmt.Sprintf("/^UUID=%s/d", uuid)
			cmd = exec.Command("sed", "-i", info, "/etc/fstab")
			_, err = cmd.CombinedOutput()
			if err != nil {
				return err
			}
		} else {
			//删除匹配行
			deviceNames := strings.Split(deviceName, "/")
			devName := deviceNames[len(deviceNames)-1]
			info := fmt.Sprintf("/%s/d", devName)
			cmd = exec.Command("sed", "-i", info, "/etc/fstab")
			_, err = cmd.CombinedOutput()
			if err != nil {
				return err
			}
		}
	}

	cmd = exec.Command("umount", "-f", mpath)
	result, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(result), "mountpoint not found") && !strings.Contains(string(result), "not mounted") {
			fmt.Printf("fail umount %s\n", mpath)
			return err
		}
	}

	fmt.Printf("success umount %s\n", mpath)
	return nil
}

func GetDevMode() bool {
	deviceName := "/dev/nvme"
	cmd := exec.Command("nvme", "list")
	bFind, _, err := findInfo(cmd, -1, -1, deviceName)
	if err != nil {
		panic(err)
	}

	return bFind
}

func GetRealDeviceName(volumeId string) string {
	deviceName := "/dev/nvme"
	cmd := exec.Command("nvme", "list")
	_, realDevName, err := findInfo(cmd, -1, 0, deviceName, volumeId)
	if err != nil {
		panic(err)
	}
	return realDevName
}

func WriteFstab(deviceName, nodeName, volumeId string) error {
	mpath := fmt.Sprintf("%s/%s", MOUNT_PATH, nodeName)
	fmt.Printf("write fstab:%s-%s\n", deviceName, mpath)

	if UUID_ENABLED == 1 {
		bFind, uuid, err := GetVolumeUUID(deviceName)
		if err != nil {
			return err
		}

		if bFind && len(uuid) > 0 {
			err := fileutils.CopyFile("/etc/fstab", "/etc/fstab.orig")
			if err != nil {
				return err
			}

			//删除匹配行
			info := fmt.Sprintf("/^UUID=%s/d", uuid)
			cmd := exec.Command("sed", "-i", info, "/etc/fstab")
			_, err = cmd.CombinedOutput()
			if err != nil {
				return err
			}

			fmt.Printf("uuid:%s-%s\n", uuid, mpath)

			//增加新行
			info = fmt.Sprintf("$a\\UUID=%s %s ext4 defaults,nofail 0 0", uuid, mpath)
			cmd = exec.Command("sed", "-i", info, "/etc/fstab")
			_, err = cmd.CombinedOutput()
			if err != nil {
				return err
			}

			//卸载
			err = UmountEbs(nodeName, false, true)
			if err != nil {
				return err
			}

			//绑定
			cmd = exec.Command("mount", "-a")
			_, err = cmd.CombinedOutput()
			if err != nil {
				//还原
				err = fileutils.CopyFile("/etc/fstab.orig", "/etc/fstab")
				if err != nil {
					return err
				}
				return errors.New("mount error")
			}

			cmd = exec.Command("df", "-h")
			_, _, err = findInfo(cmd, -1, -1, mpath, deviceName)
			if err != nil {
				return err
			}
			return nil
		}
		return errors.New("no found")
	} else {
		err := fileutils.CopyFile("/etc/fstab", "/etc/fstab.orig")
		if err != nil {
			return err
		}

		//删除匹配行
		deviceNames := strings.Split(deviceName, "/")
		devName := deviceNames[len(deviceNames)-1]
		info := fmt.Sprintf("/%s/d", devName)
		fmt.Printf("****:%s\n", info)
		cmd := exec.Command("sed", "-i", info, "/etc/fstab")
		_, err = cmd.CombinedOutput()
		if err != nil {
			return err
		}

		//增加新行
		info = fmt.Sprintf("$a/dev/%s %s ext4 defaults,nofail 0 0", devName, mpath)
		cmd = exec.Command("sed", "-i", info, "/etc/fstab")
		_, err = cmd.CombinedOutput()
		if err != nil {
			return err
		}

		//卸载
		err = UmountEbs(nodeName, false, true)
		if err != nil {
			return err
		}

		//绑定
		cmd = exec.Command("mount", "-a")
		_, err = cmd.CombinedOutput()
		if err != nil {
			//还原
			err = fileutils.CopyFile("/etc/fstab.orig", "/etc/fstab")
			if err != nil {
				fmt.Printf("err4:%+v\n", err)
				return err
			}
			return errors.New("mount error")
		}

		cmd = exec.Command("df", "-h")
		bFind, _, err := findInfo(cmd, -1, -1, mpath, deviceName)
		if err != nil {
			return err
		}
		fmt.Printf("*****:%+v-%s-%s\n", bFind, mpath, deviceName)
		return nil
	}
}

func readByLine(filename string) (lines [][]byte, err error) {
	fp, err := os.Open(filename) // 获取文件指针
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	bufReader := bufio.NewReader(fp)
	for {
		line, _, err := bufReader.ReadLine() // 按行读
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
		} else {
			lines = append(lines, line)
		}
	}
	return
}

func findInfo(cmd *exec.Cmd, total, pos int, infos ...string) (bool, string, error) {
	result := ""
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false, "", err
	}

	cmd.Start()
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		sFind := true
		for _, info := range infos {
			if !strings.Contains(line, info) {
				sFind = false
				break
			}
		}

		if sFind {
			line = DeleteExtraSpace(line)
			tmp := strings.Split(line, " ")
			if len(tmp) >= total || total == -1 {
				if pos == -1 {
					pos = len(tmp) - 1
				}
				result = tmp[pos]
				result = strings.Replace(result, " ", "", -1)
				result = strings.Replace(result, "\n", "", -1)
				result = strings.Replace(result, "\r", "", -1)
				fmt.Printf("result:%s\n", result)
				return true, result, nil
			}
		}
	}

	return false, "", nil
}

func GetVolumeUUID(deviceName string) (bool, string, error) {
	cmd := exec.Command("lsblk", "-o", "+UUID", deviceName)
	sFinds := strings.Split(deviceName, "/")
	return findInfo(cmd, 7, -1, sFinds[len(sFinds)-1])
}

func main() {
	/*err := MountEbs("sdb", "dash11000", "333")
	if err != nil {
		panic(err)
	}*/

	/*err1 := UmountEbs("dash11000", true)
	if err1 != nil {
		panic(err1)
	}*/

	g_devMode = GetDevMode()
	fmt.Printf("devMode:%t\n", g_devMode)
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
