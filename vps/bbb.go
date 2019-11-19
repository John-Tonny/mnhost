package main

import (
	"log"
	// log "github.com/sirupsen/logrus"

	"github.com/dynport/gossh"
)

func main() {
	log.Printf("aaa:%v-%d", "1234", 5678)
	log.Println("test:", "abce")

	client := gossh.New("52.14.4.149", "root")

	// my default agent authentication is used. use
	client.SetPassword("vpub$999000")

	defer client.Close()

	cmd := "echo -e \"d\n1\nw\" | fdisk /dev/xvdk"
	result, err := client.Execute(cmd)

	cmd = "echo -e \"n\np\n1\n\n\nw\" | fdisk /dev/xvdk"
	result, err = client.Execute(cmd)
	if err != nil {
		log.Fatalf("err1a:%v\n", err)
	}

	log.Println("1")
	cmd = "fdisk -l |grep xvdk"
	result, err = client.Execute(cmd)
	if err != nil {
		log.Fatalf("err1:%v\n", err)
	}
	log.Println("2")
	log.Println(result.Stdout())

	cmd = "file -s /dev/xvdk1 |grep ext4"
	result, err = client.Execute(cmd)
	if err != nil {
		log.Printf("err5:%v\n", err)
		cmd = "mkfs -t ext4 /dev/xvdk"
		result, err = client.Execute(cmd)
		if err != nil {
			log.Fatalf("err2:%v\n", err)
		}
		log.Println("3a")
		log.Println(result.Stdout())
	}
	log.Println("3")
	log.Println(result.Stdout())

	cmd = "mount /dev/xvdk1 /var/lib/docker/volumes/"
	result, err = client.Execute(cmd)
	if err != nil {
		log.Fatalf("err3:%v\n", err)
	}
	log.Println("4")
	log.Println(result.Stdout())

	/*cmd = "umount /dev/xvdk1"
	result, err = client.Execute(cmd)
	if err != nil {
		log.Fatalf("err6:%v\n", err)
	}
	log.Println("6")
	log.Println(result.Stdout())*/

}
