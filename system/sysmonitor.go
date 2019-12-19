package main

import (
	"encoding/json"
	"fmt"

	"strconv"
	"time"

	"github.com/astaxie/beego"

	"github.com/shirou/gopsutil/cpu"
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

type SysController struct {
	beego.Controller
}

type ConfController struct {
	beego.Controller
}

const (
	NFS_PATH     = "/mnt/efs"
	NODE_PREFIX  = "node"
	RPC_USER     = "vpub"
	RPC_PASSWORD = "vpub999000"
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

func (this *ConfController) Post() {
	conf := Conf{}
	body := this.Ctx.Input.RequestBody
	err := json.Unmarshal(body, &conf)
	data := &RespInfo{}
	if err != nil {
		data = &RespInfo{
			"414",
			"参数错误"}
	} else {
		fmt.Printf("conf:%+v\n", conf)
		if conf.CoinName == "" || conf.RpcPort == 0 || conf.MnKey == "" || conf.ExternalIp == "" || conf.FileName == "" {
			data = &RespInfo{
				"414",
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

func writeConf(conf Conf) {
	fileName := fmt.Sprintf("%s/%s/%s%d/%s", NFS_PATH, conf.CoinName, NODE_PREFIX, conf.RpcPort, conf.FileName)
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

}

func main() {
	beego.BConfig.CopyRequestBody = true
	beego.Router("/GetSysStatus", &SysController{})
	beego.Router("/WriteConf", &ConfController{})
	beego.Run(":8844")
}
