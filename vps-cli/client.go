package main

import (
	"fmt"
	"log"

	"strings"

	"github.com/dynport/gossh"
	"github.com/robfig/cron"

	mnhostTypes "github.com/John-Tonny/mnhost/types"
)

type NodeInfo struct {
	publicIp  string
	privateIp string
	role      string
	status    string
}

type NodeMap struct {
	nodeInfo map[string]*NodeInfo
}

type NameType struct { //定义一个结构体，类似Python语言中定义一个类的概念。
	Blog string
}

func main() {
	log.Println("start ssh")

	BlogAddress := make(map[string]*NameType)
	BlogAddress["yinzhengjie"] = &NameType{} //这里赋值的是指针。
	BlogAddress["yinzhengjie"].Blog = "http://www.cnblogs.com/yinzhengjie"
	fmt.Println(BlogAddress)
	fmt.Println(BlogAddress["yinzhengjie"].Blog)

	bb := NodeMap{
		nodeInfo: make(map[string]*NodeInfo),
	}

	bb.nodeInfo["ip1"] = &NodeInfo{}
	bb.nodeInfo["ip1"].privateIp = "33"
	bb.nodeInfo["ip1"].publicIp = "44"
	bb.nodeInfo["ip1"].role = "55"
	bb.nodeInfo["ip1"].status = "ready"

	bb.nodeInfo["ip2"] = &NodeInfo{}
	bb.nodeInfo["ip2"].privateIp = "33a"
	bb.nodeInfo["ip2"].publicIp = "44a"
	bb.nodeInfo["ip2"].role = "55a"
	bb.nodeInfo["ip2"].status = "down"

	bb.nodeInfo["ip1"].status = "aaaaabbbb"
	log.Printf("bb:%+v\n", bb.nodeInfo["ip1"])

	_, ok := bb.nodeInfo["ip1"]
	if ok == false {
		log.Println("error")
	} else {
		log.Println("correct")
	}

	a := make(map[string]map[string]string, 100)
	a["key1"] = make(map[string]string)
	a["key1"]["key2"] = "abc2"
	a["key1"]["key3"] = "abc3"
	a["key1"]["key4"] = "abc4"
	a["key1"]["key5"] = "abc5"
	fmt.Println(a)
	fmt.Printf("aa:%s\n", a["key1"]["key4"])

	aa := make(map[string]map[string]string, 100)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%s%d", "aa", i)
		value := fmt.Sprintf("%s%d", "aa", i+10)
		log.Printf("key:%s,value:%s", key, value)
		aa[key] = make(map[string]string)
		aa[key][key] = value
	}
	log.Printf("result:%s\n", aa["aa1"]["aa1"])
	log.Printf("aa:%+v\n", aa)

	/*if aa["aa10"]["aa10"] {
		log.Println("error")
	}*/
	log.Println("correct")

	client := gossh.New("3.134.78.231", "root")
	if client == nil {
		log.Printf("err:%v\n", client)
	}
	//client.SetPassword("htjonny")
	client.SetPassword("vpub$999000")
	// client.SetPrivateKey("/root/myaws.pem")
	defer client.Close()

	cmd := fmt.Sprintf("rm -rf %s/%s/%s%d", mnhostTypes.NFS_PATH, "dash", mnhostTypes.NODE_PREFIX, 10000)
	fmt.Printf("cmd:%s\n", cmd)
	_, err := client.Execute(cmd)

	if err == nil {
		log.Printf("erra:%+v\n", err)
	} else {
		log.Println("finish remove")
	}

	cmd = fmt.Sprintf("find %s/%s/ -name %s%d", "/mnt/efs", "dash", "node", 10000)
	fmt.Printf("cmd:%s\n", cmd)
	result, err := client.Execute(cmd)
	if err != nil {
		log.Printf("err1:%+v", err)
	}
	log.Printf("res:%s-%d\n", result.Stdout(), len(result.Stdout()))
	resp := fmt.Sprintf("%s", result.Stdout())
	log.Printf("aaa:%d\n", strings.Index(resp, "node10000"))

	for {
		log.Printf("%s-%d", resp, len(resp))
		ll := strings.Index(resp, "\n")
		log.Printf("pos:%d", ll)
		if ll > 0 {
			cc := strings.Replace(resp[:ll], "\r", "", -1)
			//cc = strings.Replace(resp[:ll], "\n", "", -1)
			log.Printf("%s-%d", cc, len(cc))
			resp = resp[ll+1:]
		} else {
			break
		}

	}

	/*resp = strings.Split(resp, "\n")
	for _, aa = range resp {
		log.Printf("%s\n", aa)
	}*/

	/*cmd := fmt.Sprintf("docker %s", "node ls")
	result, err := client.Execute(cmd)
	if err != nil {
		log.Printf("err1:%v\n", err)
	}
	log.Printf("docker :%s\n", result.Stdout())
	*/
	/*bb := result.Stdout()
	aa := bb[:8]
	cmd = fmt.Sprintf("docker restart %s", aa)
	result, err = client.Execute(cmd)
	if err != nil {
		log.Printf("err2:%v\n", err.Error())
	}
	log.Printf("restart:%s\n", result.Stdout())*/

	// test1(1, 8)
	///test2(2, 10)

	c := cron.New()
	spec := "*/2 * * * * ?"
	err = c.AddFunc(spec, func() {
		log.Printf("a55")
	})
	if err != nil {
		log.Println(err)
	}
	//log.Printf("resp:%+v\n", resp)
	spec1 := "*/6 * * * * ?"
	err = c.AddFunc(spec1, func() {
		log.Printf("a66")

	})
	if err != nil {
		log.Println(err)
	}
	//log.Printf("resp1:%+v\n", resp)
	c.Start()

	select {}

}

func test1(nums, bb int) {
	c := cron.New()
	spec := "*/2 * * * * ?"
	c.AddFunc(spec, func() {
		log.Printf("nums:%d-%d", nums, 55)
		nums++

		if nums == bb {
			c.Stop()
		}

	})
	c.Start()
}

func test2(nums, bb int) {
	c := cron.New()
	spec := "*/3 * * * * ?"
	c.AddFunc(spec, func() {
		log.Printf("nums:%d-%d", nums, 66)
		nums++

		if nums == bb {
			c.Stop()
		}

	})
	c.Start()
}

func aaa() {
}

/*
//定义一个类型 包含一个int类型参数和函数体
type funcIntJob struct {
	num      int
	function func(int)
}

//实现这个类型的Run()方法 使得可以传入Job接口
func (this *funcIntJob) Run() {
	if nil != this.function {
		this.function(this.num)
	}
}

//非必须  返回一个urlServeJob指针
func newfuncIntJob(num int, function funcInt) *urlServeJob {
	instance := &funcIntJob{
		num:      num,
		function: function,
	}
	return instance
}

//示例任务
func shownum(num int) {
	fmt.Println(num)
}

func main() {
	var c = cron.New()
	job := newfuncIntJob(3, shownum)
	spec := "*5 * * * * ?"
	c.AddJob(spec, job)
	c.Start()
	defer c.Stop()
	select {}
}*/
