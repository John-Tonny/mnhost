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

	INTER_ERROR = "8000"

	JSON_DATAERR           = "8001"
	ORDER_PROCESSING       = "8002"
	TIMEOUT_VPS            = "8003"
	TIMEOUT_VOLUME         = "8004"
	CONNECT_ERR            = "8005"
	TERMINATE_INSTANCE_ERR = "8006"
	DELETE_VOLUME_ERR      = "8007"
	RELEASE_ADDRESS_ERR    = "8008"
)

var recodeText = map[string]string{
	RECODE_OK:         "成功",
	RECODE_DBERR:      "数据库查询错误",
	RECODE_NODATA:     "无数据",
	RECODE_DATAEXIST:  "数据已存在",
	RECODE_DATAERR:    "数据错误",
	RECODE_SESSIONERR: "用户未登录",
	RECODE_LOGINERR:   "用户登录失败",
	RECODE_PARAMERR:   "参数错误",
	RECODE_USERERR:    "用户不存在或未激活",
	RECODE_USERONERR:  "用户已经注册",
	RECODE_ROLEERR:    "用户身份错误",
	RECODE_PWDERR:     "密码错误",
	RECODE_REQERR:     "非法请求或请求次数受限",
	RECODE_IPERR:      "IP受限",
	RECODE_THIRDERR:   "第三方系统错误",
	RECODE_IOERR:      "文件读写错误",
	RECODE_SERVERERR:  "内部错误",
	RECODE_UNKNOWERR:  "未知错误",
	RECODE_SMSERR:     "短信失败",
	RECODE_MOBILEERR:  "手机号错误",
	RECODE_UPDATEERR:  "更新失败",
	RECODE_INSERTERR:  "插入失败",
	RECODE_DELETEERR:  "删除失败",
	RECORD_SYSTEMERR:  "系统错误",

	INTER_ERROR:            "内部错误",
	JSON_DATAERR:           "数据错误",
	ORDER_PROCESSING:       "正在处理订单",
	TIMEOUT_VPS:            "实例超时",
	TIMEOUT_VOLUME:         "卷超时",
	CONNECT_ERR:            "连接失败",
	TERMINATE_INSTANCE_ERR: "删除实例失败",
	DELETE_VOLUME_ERR:      "删除卷失败",
	RELEASE_ADDRESS_ERR:    "释放地址失败",
}

func RecodeText(code string) string {
	str, ok := recodeText[code]
	if ok {
		return str
	}
	return recodeText[RECODE_UNKNOWERR]
}
