package main

import (
	"log"
	//"time"
	"github.com/John-Tonny/mnhost/model"
	"github.com/astaxie/beego/orm"
)

var secrets = []string{"3334d3998d83633af194ef19902b9778bbb6daa03e78bcf158c0f19052b2dc0d",
	"521eee7a28b734062a3bcd7656f466d436a3ad1974836f471f2e13d9cabdeea8",
	"270b2d0a45c22fb904313abf16e9c5e32ebe3f4d13c557c185f6d46d09f634c9",
	"3c3d310df4e1cf2dbed8d5936e8923051da6195d7f4cdbc577c94cbea9cefb99",
	"6af69b598c3b59e171b4a518820640d0890921f4436a89ac7b18e010cc61d0c6",
	"3ad9c92e48e118d9312cd9667b3b78cdd5086308ce50f74a0c0403c2f3e288b1",
	"3dabf0b8b33fa88a148c52b943bc0d4e04a6ff0620edc3df5ac054c2d91496d6",
	"2a7d4d8fd84c423f30dc61e0c93359ebe8477d5c7751a9ee86ac0680467dd681",
	"5ad0c25659872bea266e6159f77177b20d6507233f3548d94b754cca2a1355d4",
	"49e01226bbe09ed920ff7b67dad436422d68aa517679239f800f7ac7332a6560",
	"1a2c0ce8cbe4f964c74720111c97e3ab1810c784ae045d47ad70337f0221934e",
	"478b0924f9d17e73d5a21a96d5fa6bbd1f375ced4188ed9ed7e0cbdef2439234",
	"581757f58f3a33ff136af7b9c3dcf4268ef85eb91e18a4b32cd5b37ac0a5c1f4",
	"218145227032b02c476418e4831fb1e3c2500617dd85f7a9213839af831e7479",
	"1c015b817be51f243c97559ab0e534c4aa0c4e7d8d17e8b4b6a7a9c85c12a53e",
	"55638523302d52b09a3ad7b36345d06099da063f189a25417a276408dd80ef4b",
	"33f85ccf23fd0e876849d14a22ea59fc128769bb59d48e54bf02b5eeaba7fd5b",
	"3292c29d5ac3da4fd413eeaf4e0c0c64e1dc9070f8e6a3b1687b1e0d61416295",
	"18b8f065ea95dfff6bc8f366e5a27eefa0a4419ea0f70e2b55b4059b8928abb5",
	"2996508af32f9a88948a4bccd58ed47df70986c192aa9d24b7d1c0390d950a71",
	"36412299567549a1988ee403f655da9f6788e6cc61c1d76c2b7181135b25df5e",
	"0f2280403e7342ef78e5b9caac8d1a27549e28e2e456bf56c70eec33e3d027cc",
	"40ccbc0accbb1fdecd0e936d027fa5e7283585a6cd441c82fa68c8dc41221090",
	"67ad57c76b5cf0b6691edb3d3465c0968cdf25857ad4059b35cce34c4c7d194f",
	"2a2f61f40c5bd3e1f11d66f947eabfacc3dffdc30eaea10467f7523d71db1d8f",
	"0d36fa2e634233d682ef6ed3c82ccee26bc524b7f9a0111e5660847f71b491e5",
	"4e9f795f19cdedcfb8c64f9321ff7315c842d53b1c8c5fbdb78b468a0c3174e9",
	"4fc4619e3d20ce64f0c2bdbf570cf6267d3be1ee93111cf2bc2cd579ac7c22d4",
	"30d0bab4d4b7f50219f4a4b2f552f788748f98e8d94ede3afbf4c26966d602b1",
}

func main() {
	log.Println("start orm")

	/*var tuser models.TAccount
	o := orm.NewOrm()
	qs := o.QueryTable("t_account")
	err := qs.Filter("id", 1).One(&tuser)
	if err != nil {
		log.Printf("err:%v\n", err)
	}*/

	for key, secret := range secrets {
		o := orm.NewOrm()
		torder := models.TOrder{}
		log.Printf("%d-%s\n", key, secret)
		torder.Userid = 1
		torder.Coinname = "dash"
		torder.Mnkey = secret
		torder.Timetype = 1
		torder.Price = 1
		_, err := o.Insert(&torder)
		if err != nil {
			log.Printf("error:%+v\n", err)
		} else {
			log.Printf("update:%d-%s\n", key, secret)
		}
	}

	/*var tnode models.TNode
	o := orm.NewOrm()
	qs := o.QueryTable("t_node")
	err := qs.Filter("id", 2).One(&tnode)
	if err != nil {
		log.Printf("err:%v\n", err)
	}
	log.Printf("node:%v\n", tnode)*/

	/*var tvps models.TVps
	o = orm.NewOrm()
	qs = o.QueryTable("t_vps")
	err = qs.One(&tvps)
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
	}*/

	log.Println("1")

	/*o := orm.NewOrm()
	var r RawSeter
	r = o.Raw("UPDATE tvps SET status = ? WHERE name = ?", "testing", "slene")
	*/

	/*tnvps := models.TVps{}
	log.Println("2")
	tnvps.AllocateId = "allocateid"
	log.Println("3")
	tnvps.InstanceId = "instanceId1"
	//tnvps.VolumeId = "volumeId"
	tnvps.ProviderName = "provider_name"
	tnvps.Cores = 1
	tnvps.Memory = 1
	tnvps.KeyPairName = "key_pair_name"
	tnvps.SecurityGroupName = "group_name"
	tnvps.RegionName = "regionName"
	tnvps.PublicIp = "publicIp"
	tnvps.PrivateIp = "privateIp"
	tnvps.ClusterName = "cluster1"
	tnvps.VpsRole = "manager"
	tnvps.Status = "wait-data"

	o = orm.NewOrm()
	_, err = o.Insert(&tnvps)
	if err != nil {
		log.Fatalf("err:%v\n", err)
	}*/

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
