module github.com/John-Tonny/mnhost/vps

go 1.13

require (
	github.com/John-Tonny/micro/vps/amazon v0.0.0-20191104053440-f9154c36e20b
	github.com/John-Tonny/mnhost/common v0.0.0-20191122065528-036b67172205 // indirect
	github.com/John-Tonny/mnhost/conf v0.0.0-20191122065528-036b67172205
	github.com/John-Tonny/mnhost/interface/out/log v0.0.0-20191122065528-036b67172205
	github.com/John-Tonny/mnhost/interface/out/mnmsg v0.0.0-20191122065528-036b67172205
	github.com/John-Tonny/mnhost/interface/out/user v0.0.0-20191122065528-036b67172205 // indirect
	github.com/John-Tonny/mnhost/interface/out/vps v0.0.0-20191122065528-036b67172205
	github.com/John-Tonny/mnhost/model v0.0.0-20191122065528-036b67172205
	github.com/John-Tonny/mnhost/utils v0.0.0-20191122065528-036b67172205
	github.com/astaxie/beego v1.12.0
	github.com/aws/aws-sdk-go v1.25.36
	github.com/dynport/gossh v0.0.0-20170809141523-122e3ee2a6b0
	github.com/garyburd/redigo v1.6.0
	github.com/go-ini/ini v1.51.0
	github.com/micro/go-log v0.1.0 // indirect
	github.com/micro/go-micro v1.16.0
	github.com/micro/go-plugins v1.5.1 // indirect
	github.com/pkg/sftp v1.10.1 // indirect
	github.com/pytool/ssh v0.0.0-20190312091242-5aaea5918db7
	github.com/uber/jaeger-client-go v2.20.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
)

replace github.com/John-Tonny/mnhost/model => /root/mygo/src/github.com/John-Tonny/mnhost/model
