package handler

import (
	//"bytes"
	"context"
	//"errors"

	//"fmt"

	//"io/ioutil"
	//"os"
	//"os/exec"
	"time"

	"github.com/micro/go-micro/util/log"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	system "github.com/John-Tonny/mnhost/system/proto/system"
)

type System struct{}

var (
	CpuPercent float32
	MemPercent float32
)

// Call is a single request handler called via client.Call or the generated client code
func (e *System) GetSysStatus(ctx context.Context, req *system.Request, rsp *system.SysResponse) error {
	log.Log("Received System.Call request")

	v, _ := mem.VirtualMemory()
	cc, _ := cpu.Percent(time.Second, false)

	rsp.CpuPercent = float32(cc[0])
	rsp.MemPercent = float32(v.UsedPercent)

	rsp.Errno = "0"
	rsp.Errmsg = "成功"

	return nil
}

func (e *System) WriteConf(ctx context.Context, req *system.ConfRequest, rsp *system.Response) error {
	log.Log("Received System.Call request")
	rsp.Msg = "Hello " + req.ExternalIp
	return nil
}
