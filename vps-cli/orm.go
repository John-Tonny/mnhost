package main

import (
	"log"
	//"time"

	"github.com/astaxie/beego/orm"

	"github.com/John-Tonny/mnhost/model"
)

func main() {
	log.Println("start orm")

	var tnode models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	err := qs.Filter("id", 2).One(&tnode)
	if err != nil {
		log.Printf("err:%v\n", err)
	}
	log.Printf("node:%v\n", tnode)

	var tvps models.TVps
	o = orm.NewOrm()
	qs = o.QueryTable("t_vps")
	err = qs.Filter("usable_nodes__gt", 1).One(&tvps)
	if err != nil {
		log.Printf("err:%v\n", err)
	}
	log.Printf("vps:%v\n", tvps)

	var torder models.TOrder
	o = orm.NewOrm()
	qs = o.QueryTable("t_order")
	err = qs.Filter("id", 1).One(&torder)
	if err != nil {
		log.Printf("err:%v\n", err)
	}
	log.Printf("order:%v\n", torder)

	var user models.TAccount
	o = orm.NewOrm()
	qs = o.QueryTable("t_account")
	err = qs.Filter("id", torder.Userid).One(&user)
	if err != nil {
		log.Printf("err:%v\n", err)
	}
	log.Printf("account:%v\n", user)

	coin := models.TCoin{}
	coin.Name = "vircle"
	coin.Status = "Enabled"
	coin.Path = ".vircle"
	coin.Conf = "vircle.conf"
	coin.Docker = "vpub/vircle:0.1"
	coin.FilePath = "/root/data/vpub-vircle-0.1.tar"
	//插入数据到数据库中
	o = orm.NewOrm()
	_, err = o.Insert(&coin)
	if err != nil {
		log.Fatalf("err:%v\n", err)
	}

	coin = models.TCoin{}
	coin.Name = "dash"
	coin.Status = "Enabled"
	coin.Path = ".dashcore"
	coin.Conf = "dash.conf"
	coin.Docker = "mnhosted/dashcore:v1.0"
	coin.FilePath = "/root/data/mnhosted-dashcore-v1.0.tar"
	//插入数据到数据库中
	o = orm.NewOrm()
	_, err = o.Insert(&coin)
	if err != nil {
		log.Fatalf("err:%v\n", err)
	}
	/*
		tnvps := models.TVps{}
		tnvps.AllocateId = "allocateid"
		tnvps.InstanceId = "instanceId"
		tnvps.VolumeId = "volumeId"
		tnvps.ProviderName = "provider_name"
		tnvps.Cores = 1
		tnvps.Memory = 1
		tnvps.KeyPairName = "key_pair_name"
		tnvps.MaxNodes = 3
		tnvps.UsableNodes = 3
		tnvps.SecurityGroupName = "group_name"
		tnvps.RegionName = "regionName"
		tnvps.IpAddress = "publicIp"
		o = orm.NewOrm()
		_, err = o.Insert(&tnvps)
		if err != nil {
			log.Fatalf("err:%v\n", err)
		}
	*/

	/*
		var order models.OrderNode
		o = orm.NewOrm()
		order.User = &user

		order.CoinName = "vircle"
		order.Alias = "vircle"
		order.Txid = "123456-7890123"
		order.OutputIndex = 1
		order.RewardAddress = "aaaa-bbbb-cccc-dddd"

		order.Begin_date, _ = time.ParseInLocation("2006-01-02 15:04:05", "2019-11-05 16:00:00", time.Local)
		if err != nil {
			log.Fatalf("err:%v\n", err)
		}

		order.End_date, _ = time.ParseInLocation("2006-01-02 15:04:05", "2019-12-05 15:59:59", time.Local)
		order.Period = "1month"
		order.Amount = 30
		order.Status = models.ORDER_STATUS_WAIT_PAYMENT

		_, err = o.Insert(&order)
		if err != nil {
			log.Fatalf("err:%v\n", err)
		}

		log.Println("add order success")

		coin := models.Coin{}
		coin.Name = "vircle"
		coin.Status = "Enabled"
		coin.Path = ".vircle"
		coin.Conf = "vircle.conf"
		coin.Docker = "vpub/vircle:0.1"
		//插入数据到数据库中
		o = orm.NewOrm()
		_, err = o.Insert(&coin)
		if err != nil {
			log.Fatalf("err:%v\n", err)
		}
	*/
}
