package utils

const (
	RECODE_OK        = "0"
	RECODE_DBERR     = "4001"
	RECODE_NODATA    = "4002"
	RECODE_DATAEXIST = "4003"
	RECODE_DATAERR   = "4004"

	RECODE_SESSIONERR = "4101"
	RECODE_LOGINERR   = "4102"
	RECODE_PARAMERR   = "4103"
	RECODE_USERONERR  = "4104"
	RECODE_ROLEERR    = "4105"
	RECODE_PWDERR     = "4106"
	RECODE_USERERR    = "4107"
	RECODE_SMSERR     = "4108"
	RECODE_MOBILEERR  = "4109"

	RECODE_REQERR    = "4201"
	RECODE_IPERR     = "4202"
	RECODE_THIRDERR  = "4301"
	RECODE_IOERR     = "4302"
	RECODE_SERVERERR = "4500"
	RECODE_UNKNOWERR = "4501"
	RECODE_UPDATEERR = "4502"
	RECODE_INSERTERR = "4503"
	RECODE_DELETEERR = "4504"
	RECORD_SYSTEMERR = "4505"
	RECODE_QUERYERR  = "4506"

	INTER_ERROR = "8000"

	JSON_DATAERR     = "8001"
	ORDER_PROCESSING = "8002"
	TIMEOUT_VPS      = "8003"
	TIMEOUT_VOLUME   = "8004"

	CREATE_GROUP_ERR = "8102"

	CREATE_KEYPAIR_ERR = "8103"

	CREATE_INSTANCE_ERR         = "8104"
	WAIT_INSTANCE_ERR           = "8105"
	DESC_INSTANCE_ERR           = "8106"
	TERMINATE_INSTANCE_ERR      = "8107"
	WAIT_TERMINATE_INSTANCE_ERR = "8108"

	ALLOCATION_ERR = "8109"
	ASSOCIATE_ERR  = "8110"

	CREATE_VOLUME_ERR     = "8111"
	DESC_VOLUME_ERR       = "8112"
	WAIT_VOLUME_AVAIL_ERR = "8113"
	ATTRACH_VOLUME_ERR    = "8114"
	MODIFY_VOLUME_ERR     = "8115"

	MOUNT_VOLUME_ERR = "8116"
	EFS_MOUNT_ERR    = "8117"

	DELETE_VOLUME_ERR   = "8118"
	RELEASE_ADDRESS_ERR = "8119"

	PORT_ERR         = "8120"
	COIN_ENABLED_ERR = "8121"
	GROWPART_ERR     = "8122"
	RESIZE_ERR       = "8123"

	VPS_CONNECT_ERR    = "8200"
	SSH_CONNECT_ERR    = "8201"
	DOCKER_CONNECT_ERR = "8202"

	SWARM_INIT_ERR       = "8210"
	SWARM_JOIN_ERR       = "8211"
	SWARM_INSPECT_ERR    = "8212"
	SERVICE_NEW_ERR      = "8213"
	SERVICE_REMOVE_ERR   = "8214"
	MANAGER_ERR          = "8215"
	REMOVE_NODE_DATA_ERR = "8216"
)

var recodeText = map[string]string{
	RECODE_OK:                   "成功",
	RECODE_DBERR:                "数据库查询错误",
	RECODE_NODATA:               "无数据",
	RECODE_DATAEXIST:            "数据已存在",
	RECODE_DATAERR:              "数据错误",
	RECODE_SESSIONERR:           "用户未登录",
	RECODE_LOGINERR:             "用户登录失败",
	RECODE_PARAMERR:             "参数错误",
	RECODE_USERERR:              "用户不存在或未激活",
	RECODE_USERONERR:            "用户已经注册",
	RECODE_ROLEERR:              "用户身份错误",
	RECODE_PWDERR:               "密码错误",
	RECODE_REQERR:               "非法请求或请求次数受限",
	RECODE_IPERR:                "IP受限",
	RECODE_THIRDERR:             "第三方系统错误",
	RECODE_IOERR:                "文件读写错误",
	RECODE_SERVERERR:            "内部错误",
	RECODE_UNKNOWERR:            "未知错误",
	RECODE_SMSERR:               "短信失败",
	RECODE_MOBILEERR:            "手机号错误",
	RECODE_UPDATEERR:            "更新失败",
	RECODE_INSERTERR:            "插入失败",
	RECODE_DELETEERR:            "删除失败",
	RECORD_SYSTEMERR:            "系统错误",
	RECODE_QUERYERR:             "查询失败",
	INTER_ERROR:                 "内部错误",
	JSON_DATAERR:                "数据错误",
	ORDER_PROCESSING:            "正在处理订单",
	TIMEOUT_VPS:                 "实例超时",
	TIMEOUT_VOLUME:              "卷超时",
	SERVICE_NEW_ERR:             "创建服务失败",
	VPS_CONNECT_ERR:             "VPS连接失败",
	SSH_CONNECT_ERR:             "SSH连接失败",
	DOCKER_CONNECT_ERR:          "DOCKER连接失败",
	CREATE_GROUP_ERR:            "创建组失败",
	CREATE_KEYPAIR_ERR:          "创建键失败",
	CREATE_INSTANCE_ERR:         "创建实例失败",
	WAIT_INSTANCE_ERR:           "等待实例运行失败",
	DESC_INSTANCE_ERR:           "获取实例描述失败",
	ALLOCATION_ERR:              "地址分配失败",
	ASSOCIATE_ERR:               "地址关联失败",
	CREATE_VOLUME_ERR:           "创建卷失败",
	WAIT_VOLUME_AVAIL_ERR:       "等待卷有效失败",
	ATTRACH_VOLUME_ERR:          "卷关联失败",
	MOUNT_VOLUME_ERR:            "卷挂载失败",
	EFS_MOUNT_ERR:               "EFS挂载失败",
	TERMINATE_INSTANCE_ERR:      "删除实例失败",
	WAIT_TERMINATE_INSTANCE_ERR: "等待实例终止失败",
	DELETE_VOLUME_ERR:           "删除卷失败",
	RELEASE_ADDRESS_ERR:         "释放地址失败",
	PORT_ERR:                    "端口分配错误",
	COIN_ENABLED_ERR:            "币暂不支持",

	SWARM_INIT_ERR: "创建集群失败",
	SWARM_JOIN_ERR: "加入集群失败",
}

func RecodeText(code string) string {
	str, ok := recodeText[code]
	if ok {
		return str
	}
	return recodeText[RECODE_UNKNOWERR]
}
