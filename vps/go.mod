module github.com/John-Tonny/mnhost/vps

go 1.13

replace github.com/John-Tonny/mnhost/common => /root/mygo/src/github.com/John-Tonny/mnhost/common

replace github.com/John-Tonny/mnhost/interface/out/log => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/log

replace github.com/John-Tonny/mnhost/conf => /root/mygo/src/github.com/John-Tonny/mnhost/conf

replace github.com/John-Tonny/mnhost/model => /root/mygo/src/github.com/John-Tonny/mnhost/model

replace github.com/John-Tonny/mnhost/utils => /root/mygo/src/github.com/John-Tonny/mnhost/utils

replace github.com/John-Tonny/mnhost/types => /root/mygo/src/github.com/John-Tonny/mnhost/types

replace github.com/John-Tonny/mnhost/user/model => /root/mygo/src/github.com/John-Tonny/mnhost/user/model

replace github.com/John-Tonny/mnhost/user/handler => /root/mygo/src/github.com/John-Tonny/mnhost/user/handler

replace github.com/John-Tonny/mnhost/interface/out/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/user

replace github.com/John-Tonny/mnhost/interface/proto/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/proto/user

replace github.com/John-Tonny/mnhost/interface/out/vps => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/vps

replace github.com/John-Tonny/mnhost/interface/out/mnmsg => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/mnmsg

replace github.com/John-Tonny/micro/vps/amazon => /root/mygo/src/github.com/John-Tonny/micro/vps/amazon

require (
	github.com/John-Tonny/micro/vps/amazon v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/common v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/conf v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/log v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/mnmsg v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/user v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/interface/out/vps v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/model v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/types v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/utils v0.0.0-00010101000000-000000000000
	github.com/astaxie/beego v1.12.0
	github.com/aws/aws-sdk-go v1.25.49
	github.com/docker/docker v1.13.1
	github.com/dynport/gossh v0.0.0-20170809141523-122e3ee2a6b0
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/go-ini/ini v1.51.0
	github.com/micro/go-log v0.1.0 // indirect
	github.com/micro/go-micro v1.18.0
	github.com/micro/go-plugins v1.5.1 // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/pkg/sftp v1.10.1 // indirect
	github.com/pytool/ssh v0.0.0-20190312091242-5aaea5918db7
	github.com/robfig/cron v1.2.0
	github.com/uber/jaeger-client-go v2.20.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	moul.io/http2curl v1.0.0 // indirect
)
