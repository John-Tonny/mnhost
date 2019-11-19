package main

import (
	"log"
	"time"

	"github.com/astaxie/beego/orm"

	"github.com/John-Tonny/mnhost/model"
)

func main() {
	var user models.User
	o := orm.NewOrm()
	qs := o.QueryTable("user")
	err := qs.Filter("id", 3).One(&user)
	if err != nil {
		log.Fatalf("err:%v\n", err)
	}

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
}
