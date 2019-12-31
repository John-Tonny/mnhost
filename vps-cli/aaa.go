package main

import (
	"fmt"
	"log"

	//"strconv"

	"os"
	"path"

	"bufio"
	"errors"
	"io"
	"os/exec"
	"strings"

	uec2 "github.com/John-Tonny/micro/vps/amazon"
	"github.com/John-Tonny/mnhost/common"
	"github.com/aws/aws-sdk-go/aws"

	//"github.com/aws/aws-sdk-go/service/ec2"

	mnhostTypes "github.com/John-Tonny/mnhost/types"
	"github.com/docker/docker/api/types"
)

func main() {
	log.Println("start ssh")

	mc, _, err := common.DockerNewClient("3.136.202.223", "172.31.38.42")
	if err != nil {
		log.Printf("%+v\n", err)
	}
	if mc == nil {
		log.Printf("%+v\n", mc)
	}
	defer mc.Close()

	nodes, err := mc.NodeListA(types.NodeListOptions{})
	if err != nil {
		log.Printf("%+v\n", err)
	}
	log.Printf("%+v\n", nodes)

	fileName := fmt.Sprintf("%s/%s%d/%s", "/mnt/vircle", "dash", 10000, ".dashcore/dash.conf")
	path1 := path.Dir(fileName)
	log.Printf("remove lock :%s\n", path1)
	os.RemoveAll(fmt.Sprintf("%s/%s", path1, ".lock"))
	os.RemoveAll(fmt.Sprintf("%s/%s", path1, "testnet3/.lock"))

	//path := "/mnt/vircle/dash10000"
	//os.RemoveAll(path)
	theme := "f"
	for i := 0; i < len(theme); i++ {
		fmt.Printf("ascii: %d\n", theme[i])
	}

	//device_name_from := "b"
	//device_name_to := "z"

	//bbb, _ := strconv.Atoi(device_name_from)
	aaa := mnhostTypes.DEVICE_NAME_FROM[0]
	var i byte
	for i = 0; i < 5; i++ {
		tmp := fmt.Sprintf("xvd%s", string(aaa+i))
		//bbb, err := fmt.Printf("ascii:%d\n", device_name_from)
		VolumeReady("us-east-2c", "snap-0ad4e0a1bde0756b0", "i-083b1853579c7b922", tmp, "")
		//VolumeReady("us-east-2c", "snap-0ad4e0a1bde0756b0", "i-0bd52b6e186750423", tmp, "")
		//VolumeReady("us-east-2c", "snap-0ad4e0a1bde0756b0", "i-037766032a069aac6", tmp, "")
		fmt.Println(tmp)
	}

	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
	}

	info, err := c.GetDescribeVolumes([]string{"vol-0a333929a1195f72d"})
	fmt.Printf("%+v\n", info)

	ppp := "dkfdsfasdf "
	log.Printf("voluem:%t\n", strings.Contains(ppp, "dddd"))

	log.Println(strings.Contains("widuu", "/widd"))

	deviceName := "/dev/shm2"
	cmd := exec.Command("df", "-h")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(errors.New("exec cmd error"))
	}

	cmd.Start()
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		if strings.Contains(line, deviceName) {
			log.Printf("success mount %s", deviceName)
			break
		}
	}

	for i := 0; i < 20; i++ {
		//VolumeReady("us-east-2c", "snap-0bfa04e44bd28a30b", "i-083b1853579c7b922", "/dev/xvdm")
	}

	//err = VolumeRemove("us-east-2c", "vol-08733508d3278507f", "i-06e42874160198fda", "/dev/xvdp")
	log.Println(err)
}

func VolumeReady(zone, snapshotId, instanceId, deviceName, nvolumeId string) error {

	c, err := uec2.NewEc2Client(mnhostTypes.ZONE_DEFAULT, mnhostTypes.AWS_ACCOUNT)
	if err != nil {
		return err
	}

	if nvolumeId == "" {
		result, err := c.SnapshotsDescribe(snapshotId)
		if err != nil {
			return err
		}
		log.Println(result)

		volumeId, err := c.CreateVolumes(zone, snapshotId, aws.Int64Value(result.Snapshots[0].VolumeSize))
		if err != nil {
			return err
		}
		log.Printf("volumeId:%s\n", volumeId)

		err = c.WaitUntilVolumeAvailables([]string{volumeId})
		if err != nil {
			return err
		}
		log.Printf("create volume wait finish:%s\n", volumeId)

		nvolumeId = volumeId
	}

	deviceName = fmt.Sprintf("/dev/%s", deviceName)
	resp, err := c.AttachVolumes(instanceId, nvolumeId, deviceName)
	if err != nil {
		return err
	}

	err = c.WaitUntilVolumeAttach(5, nvolumeId)
	if err != nil {
		return err
	}

	log.Printf("attach volume :%+v\n", resp)
	return nil
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
	log.Printf("volume wait finish:%s\n", volumeId)

	resp, err := c.DeleteVolumes(volumeId)
	if err != nil {
		return err
	}
	log.Printf("del volume:%+v\n", resp)

	return nil
}
