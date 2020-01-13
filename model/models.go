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

	"strconv"
)

/* 用户 table_name = user */
type TAccount struct {
	Id            int64     `orm:"size(20)" json:"account_id"`                    //用户编号
	Account       string    `orm:"size(255);unique" json:"account"`               //用户昵称
	Passwd        string    `orm:"size(255)" json:"passwd"`                       //用户密码加密的
	Walletaddress string    `orm:"size(128)" json:"walletaddress`                 //手机号
	Createtime    time.Time `orm:"auto_now_add;type(datetime)" json:"createtime"` //创建时间
	Updatetime    time.Time `orm:"auto_now;type(datetime)" json:"updatetime"`     //更新时间
	//Orders        []*TOrder `orm:"reverse(many)" json:"orders"`                   //用户下的订单  一个人多次订单
}

/* 云主机 table_name = Vps */
type TVps struct {
	Id                int64     `orm:"size(20)" json:"vps_id"`                        //主机编号
	ClusterName       string    `orm:"size(32)" json:"cluster_name"`                  //集群名称
	ProviderName      string    `orm:"size(32)" json:"provider_name"`                 //主机服务商名称
	Cores             int       `orm:"default(2)" json:"cpus"`                        //核数量
	Memory            int       `orm:"default(4)" json:"memory"`                      //内存
	RegionName        string    `orm:"size(64)" json:"region_name"`                   //区域
	InstanceId        string    `orm:"size(64);unique" json:"instance_id"`            //实例ID
	VolumeId          string    `orm:"size(64)" json:"volume_id"`                     //磁盘ID
	SecurityGroupName string    `orm:"size(64)" json:"security_group_name"`           //安全组名称
	KeyPairName       string    `orm:"size(64)" json:"key_pair_name"`                 //密钥名称
	AllocationId      string    `orm:"size(64)" json:"allocation_id"`                 //分配地址id
	PublicIp          string    `orm:"size(64)" json:"public_ip"`                     //主机公共IP
	PrivateIp         string    `orm:"size(64)" json:"private_ip"`                    //主机私有IP
	VpsRole           string    `orm:"size(64)" json:"vps_role"`                      //主机角色  manager,worker,nfs
	Status            string    `orm:"size(64)" json:"status"`                        //主机角色状态 complete,wait,process
	Createtime        time.Time `orm:"auto_now_add;type(datetime)" json:"createtime"` //创建时间
	Updatetime        time.Time `orm:"auto_now;type(datetime)" json:"updatetime"`     //更新时间
}

/* 信息 table_name = Node */
type TNode struct {
	Id     int64 `orm:"size(20)" json:"node_id"` //节点编号
	Userid int64 `json:"userid"`
	//User     		*TAccount 	`orm:"rel(fk)" json:"user_id"`   						//用户编号  	与用户进行关联
	ClusterName string    `orm:"size(32)" json:"cluster_name"`                  //集群名称
	Order       *TOrder   `orm:"rel(fk)" json:"order_id"`                       //订单编号		与订单表进行关联
	CoinName    string    `orm:"size(32)" json:"coin_name"`                     //币名称
	RpcPort     int       `json:"rpc_port"`                                     //rpc端口号
	Port        int       `json:"port"`                                         //rpc端口号
	PublicIp    string    `orm:"size(64)" json:"public_ip"`                     //主机公共IP
	PrivateIp   string    `orm:"size(64)" json:"private_ip"`                    //主机私有IP
	InstanceId  string    `orm:"size(64)" json:"instance_id"`                   //实例ID
	VolumeId    string    `orm:"size(64)" json:"volume_id"`                     //EBS卷标Id
	VolumeState string    `orm:"size(64)" json:"volume_state"`                  //EBS卷标状态
	VolumeTotal int64     `json:"volume_total"`                                 //EBS卷总空间
	VolumeFree  int64     `json:"volume_free"`                                  //EBS卷剩余空间
	DeviceNo    byte      `json:"device_no"`                                    //设备名称编号
	State       string    `orm:"size(64)" json:"state"`                         //主节点状态
	Status      string    `orm:"size(64)" json:"staus"`                         //系统状态 (配置和数据准备)
	Createtime  time.Time `orm:"auto_now_add;type(datetime)" json:"createtime"` //创建时间
	Updatetime  time.Time `orm:"auto_now;type(datetime)" json:"updatetime"`     //更新时间
}

/* 云主机 table_name = Coin */
type TCoin struct {
	Id         int64     `orm:"size(20)" json:"coin_id"`                       //币编号
	Name       string    `orm:"size(32);unique" json:"name"`                   //币名称
	Path       string    `orm:"size(32);unique" json:"path"`                   //节点缺省安装路径
	Conf       string    `orm:"size(32);unique" json:"conf"`                   //节点缺省配置文件名称
	SnapshotId string    `orm:"size(128);unique" json:"snapshot_id"`           //上传节点快照id
	Docker     string    `orm:"size(32);unique" json:"docker"`                 //主节点docker名称
	Status     string    `orm:"default(Enabled)" json:"staus"`                 //状态
	Port       int       `json:"port"`                                         //rpc端口(主)
	Createtime time.Time `orm:"auto_now_add;type(datetime)" json:"createtime"` //创建时间
	Updatetime time.Time `orm:"auto_now;type(datetime)" json:"updatetime"`     //更新时间
}

/* 产品 table_name = Product */
type TProduct struct {
	Id         int64     `orm:"size(20)" json:"product_id"`                    //产品编号
	Name       string    `orm:"size(32)" json:"title"`                         //产品名称
	Period     string    `orm:"size(32);unique" json:"period"`                 //服务的周期（天、月、半年、一年、三年）
	Amount     int       `json:amount`                                         //总金额
	Createtime time.Time `orm:"auto_now_add;type(datetime)" json:"createtime"` //创建时间
	Updatetime time.Time `orm:"auto_now;type(datetime)" json:"updatetime"`     //更新时间
}

/* 订单 table_name = order_node */
type TOrder struct {
	Id     int64 `orm:"size(20)" json:"order_id"` //订单编号
	Userid int64 `json:"userid"`                  //用户编号
	//User       	*TAccount 	`orm:"rel(fk)" json:"userid"`    			//下单的用户编号   	//与用户表进行关联
	Coinname   string    `orm:"size(32)" json:"coinname"`                      //
	Mnkey      string    `orm:"size(128)" json:"mnkey"`                        //别名
	Timetype   int8      `json:"timetype"`                                     //
	Price      int       `json:"price"`                                        //交易ID
	Txid       string    `json:"txid"`                                         //收益地址
	Isrenew    int       `json:"isnew"`                                        //交易ID
	Status     int       `json:"status"`                                       //交易ID
	Createtime time.Time `orm:"auto_now_add;type(datetime)" json:"createtime"` //创建时间
	Updatetime time.Time `orm:"auto_now;type(datetime)" json:"updatetime"`     //更新时间
}

const (
	ORDER_STATUS_WAIT_PAYMENT = "WAIT_PAYMENT" //待支付
	ORDER_STATUS_PAID         = "PAID"         //已支付
	ORDER_STATUS_COMPLETE     = "COMPLETE"     //已完成
	ORDER_STATUS_CANCELED     = "CANCELED"     //已取消
	ORDER_STATUS_EXPIRED      = "EXPIRED"      //已过期
)

//数据库的初始化
func init() {
	//调用什么驱动
	orm.RegisterDriver("mysql", orm.DRMySQL)

	// set default database
	//连接数据   ( 默认参数 ，mysql数据库 ，"数据库的用户名 ：数据库密码@tcp("+数据库地址+":"+数据库端口+")/库名？格式",默认参数）
	host := config.GetDB("user").Host
	dbname := config.GetDB("user").DBName
	port := strconv.Itoa(int(config.GetDB("user").Port))
	user := config.GetDB("user").User
	pw := config.GetDB("user").PW
	dburl := user + ":" + pw + "@tcp(" + host + ":" + port + ")/" + dbname + "?charset=utf8&loc=Local"
	orm.RegisterDataBase("default", "mysql", dburl, 30)

	//注册model 建表
	orm.RegisterModel(new(TAccount), new(TVps), new(TNode), new(TCoin), new(TProduct), new(TOrder))

	// create table
	//第一个是别名
	// 第二个是是否强制替换模块   如果表变更就将false 换成true 之后再换回来表就便更好来了
	//第三个参数是如果没有则同步或创建
	orm.RunSyncdb("default", false, true)
}
