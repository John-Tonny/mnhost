module github.com/John-Tonny/mnhost/monitor

go 1.13

replace github.com/John-Tonny/mnhost/interface/out/log => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/log

replace github.com/John-Tonny/mnhost/conf => /root/mygo/src/github.com/John-Tonny/mnhost/conf

replace github.com/John-Tonny/mnhost/common => /root/mygo/src/github.com/John-Tonny/mnhost/common

replace github.com/John-Tonny/mnhost/model => /root/mygo/src/github.com/John-Tonny/mnhost/model

replace github.com/John-Tonny/mnhost/utils => /root/mygo/src/github.com/John-Tonny/mnhost/utils

replace github.com/John-Tonny/mnhost/user/model => /root/mygo/src/github.com/John-Tonny/mnhost/user/model

replace github.com/John-Tonny/mnhost/types => /root/mygo/src/github.com/John-Tonny/mnhost/types

replace github.com/John-Tonny/mnhost/user/handler => /root/mygo/src/github.com/John-Tonny/mnhost/user/handler

replace github.com/John-Tonny/mnhost/monitor/handler => /root/mygo/src/github.com/John-Tonny/mnhost/monitor/handler

replace github.com/John-Tonny/mnhost/interface/out/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/user

replace github.com/John-Tonny/mnhost/interface/proto/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/proto/user

replace github.com/John-Tonny/mnhost/interface/out/monitor => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/monitor

replace github.com/John-Tonny/mnhost/interface/out/mnmsg => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/mnmsg

replace github.com/John-Tonny/micro/vps/amazon => /root/mygo/src/github.com/John-Tonny/micro/vps/amazon

replace github.com/John-Tonny/go-virclerpc => /root/mygo/src/github.com/John-Tonny/go-virclerpc

replace github.com/John-Tonny/go-jsonrpc => /root/mygo/src/github.com/John-Tonny/go-jsonrpc

require (
	github.com/John-Tonny/go-virclerpc v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/micro/vps/amazon v0.0.0-20191214054803-9172d7f73796
	github.com/John-Tonny/mnhost/common v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/conf v0.0.0-20191225064632-834cc3555a47
	github.com/John-Tonny/mnhost/interface/out/log v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/mnmsg v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/monitor v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/vps v0.0.0-20191231073638-f4380469eb80
	github.com/John-Tonny/mnhost/model v0.0.0-20191225064632-834cc3555a47
	github.com/John-Tonny/mnhost/types v0.0.0-20191225064632-834cc3555a47
	github.com/John-Tonny/mnhost/utils v0.0.0-20191225064632-834cc3555a47
	github.com/astaxie/beego v1.12.0
	github.com/aws/aws-sdk-go v1.26.8
	github.com/docker/docker v1.13.1
	github.com/micro/go-micro v1.18.0
	github.com/robfig/cron v1.2.0
)
