package models

import (
	//使用了beego的orm模块
	"github.com/astaxie/beego/orm"
	//go语言的sql的驱动
	_ "github.com/go-sql-driver/mysql"
	//已经创建好的工具包
	"github.com/John-Tonny/mnhost/conf"
	//time包关于时间信息
	"time"

	//"log"
	//"reflect"

	"strconv"
)

/* 用户 table_name = user */
type User struct {
	Id            int          `json:"user_id"`                                 //用户编号
	Name          string       `orm:"null;size(32)"  json:"name"`               //用户昵称
	Password      string       `orm:"size(128)" json:"password"`                //用户密码加密的
	Mobile        string       `orm:"size(11);unique" json:"mobile"`            //手机号
	RealName      string       `orm:"null;size(32)" json:"realName"`            //真实姓名  实名认证
	IdCard        string       `orm:"null;size(20)" json:"idCard"`              //身份证号  实名认证
	RewardAddress string       `orm:"null;size(128)" json:"rewardAddress"`      //收益地址
	Ctime         time.Time    `orm:"auto_now_add;type(datetime)" json:"ctime"` //创建时间
	Utime         time.Time    `orm:"auto_now;type(datetime)" json:"utime"`     //更新时间
	Orders        []*OrderNode `orm:"reverse(many)" json:"orders"`              //用户下的订单  一个人多次订单
}

/* 云主机 table_name = Vps */
type Vps struct {
	Id                int       `json:"vps_id"`                                  //主机编号
	ProviderName      string    `orm:"size(32)" json:"title"`                    //主机服务商名称
	Cores             int       `orm:"default(2)" json:"cpus"`                   //核数量
	Memory            int       `orm:"default(4)" json:"memorys"`                //内存
	MaxNodes          int       `orm:"default(15)" json:maxnodes`                //
	UsableNodes       int       `orm:"default(15)" json:usablenodes`             //
	RegionName        string    `orm:"size(64)" json:"regionName"`               //区域
	InstanceId        string    `orm:"size(64);unique" json:"instanceid"`        //实例ID
	VolumeId          string    `orm:"size(64)" json:"VolumeId"`                 //磁盘ID
	SecurityGroupName string    `orm:"size(64)" json:"securitygroupId"`          //安全组名称
	KeyPairName       string    `orm:"size(64)" json:"KeyPairName"`              //密钥名称
	AllocateId        string    `orm:"size(64)" json:"AllocateId"`               //主机IP
	IpAddress         string    `orm:"size(64)" json:"AllocateId"`               //主机IP
	Ctime             time.Time `orm:"auto_now_add;type(datetime)" json:"ctime"` //创建时间
	Utime             time.Time `orm:"auto_now;type(datetime)" json:"utime"`     //更新时间
	Nodes             []*Node   `orm:"reverse(many)" json:"nodes"`               //用户下的订单       一个人多次订单
}

/* 房屋信息 table_name = Node */
type Node struct {
	Id        int        `json:"node_id"`                //节点编号
	User      *User      `orm:"rel(fk)" json:"user_id"`  //用户编号  	与用户进行关联
	Vps       *Vps       `orm:"rel(fk)" json:"vps_id"`   //主机编号		与主机表进行关联
	OrderNode *OrderNode `orm:"rel(fk)" json:"order_id"` //主机编号		与主机表进行关联
	CoinName  string     `orm:"size(32)" json:"coin"`    //币名称
	Port      int        //端口号
	Ctime     time.Time  `orm:"auto_now_add;type(datetime)" json:"ctime"` //创建时间
	Utime     time.Time  `orm:"auto_now;type(datetime)" json:"utime"`     //更新时间
}

/* 云主机 table_name = Coin */
type Coin struct {
	Id     int       `json:"coin_id"`                                 //币编号
	Name   string    `orm:"size(32);unique" json:"name"`              //币名称
	Path   string    `orm:"size(32);unique" json:"path"`              //币名称
	Conf   string    `orm:"size(32);unique" json:"conf"`              //币名称
	Docker string    `orm:"size(32);unique" json:"version"`           //币名称
	Status string    `orm:"default(Enabled)"`                         //状态
	Ctime  time.Time `orm:"auto_now_add;type(datetime)" json:"ctime"` //创建时间
	Utime  time.Time `orm:"auto_now;type(datetime)" json:"utime"`     //更新时间
}

/* 产品 table_name = Product */
type Product struct {
	Id     int       `json:"product_id"`                   //产品编号
	Name   string    `orm:"size(32)" json:"title"`         //产品名称
	Period string    `orm:"size(32);unique" json:"period"` //服务的周期（天、月、半年、一年、三年）
	Amount int       //总金额
	Ctime  time.Time `orm:"auto_now_add;type(datetime)" json:"ctime"` //创建时间
	Utime  time.Time `orm:"auto_now;type(datetime)" json:"utime"`     //更新时间
}

const (
	ORDER_STATUS_WAIT_PAYMENT = "WAIT_PAYMENT" //待支付
	ORDER_STATUS_PAID         = "PAID"         //已支付
	ORDER_STATUS_COMPLETE     = "COMPLETE"     //已完成
	ORDER_STATUS_CANCELED     = "CANCELED"     //已取消
	ORDER_STATUS_EXPIRED      = "EXPIRED"      //已过期
)

/* 订单 table_name = order_node */
type OrderNode struct {
	Id            int       `json:"order_id"`              //订单编号
	User          *User     `orm:"rel(fk)" json:"user_id"` //下单的用户编号   	//与用户表进行关联
	CoinName      string    `orm:"size(32)" json:"coin"`   //
	Alias         string    `orm:"size(32)" json:"alias"`  //别名
	Txid          string    `orm:"size(64)" json:"txid"`   //交易ID
	OutputIndex   int       //交易index
	RewardAddress string    `orm:"size(64)" json:"rewardaddress"` //收益地址
	Begin_date    time.Time `orm:"type(datetime)"`                //服务的起始时间
	End_date      time.Time `orm:"type(datetime)"`                //服务的结束时间
	Period        string    //服务的周期（天、月、半年、一年、三年）
	Amount        int       //订单总金额
	Status        string    `orm:"default(WAIT_PAYMENT)"`                    //订单状态
	Ctime         time.Time `orm:"auto_now_add;type(datetime)" json:"ctime"` //创建时间
	Utime         time.Time `orm:"auto_now;type(datetime)" json:"utime"`     //更新时间
}

//数据库的初始化
func init() {
	//调用什么驱动
	orm.RegisterDriver("mysql", orm.DRMySQL)

	// set default database
	//连接数据   ( 默认参数 ，mysql数据库 ，"数据库的用户名 ：数据库密码@tcp("+数据库地址+":"+数据库端口+")/库名？格式",默认参数）
	host := config.GetDB("user").Host
	dbname := config.GetDB("user").DBName
	port := strconv.Itoa(int(config.GetDB("user").Port))
	dburl := "root:vpub999000@tcp(" + host + ":" + port + ")/" + dbname + "?charset=utf8"
	orm.RegisterDataBase("default", "mysql", dburl, 30)

	//注册model 建表
	orm.RegisterModel(new(User), new(Vps), new(Node), new(Coin), new(Product), new(OrderNode))

	// create table
	//第一个是别名
	// 第二个是是否强制替换模块   如果表变更就将false 换成true 之后再换回来表就便更好来了
	//第三个参数是如果没有则同步或创建
	orm.RunSyncdb("default", false, true)
}
