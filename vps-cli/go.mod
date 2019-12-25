module github.com/John-Tonny/mnhost/user-cli

go 1.13

replace github.com/John-Tonny/mnhost/system/proto/system => /root/mygo/src/github.com/John-Tonny/mnhost/system/proto/system

replace github.com/John-Tonny/micro/vps/amazon => /root/mygo/src/github.com/John-Tonny/micro/vps/amazon

replace github.com/John-Tonny/mnhost/interface/out/log => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/log

replace github.com/John-Tonny/mnhost/conf => /root/mygo/src/github.com/John-Tonny/mnhost/conf

replace github.com/John-Tonny/mnhost/common => /root/mygo/src/github.com/John-Tonny/mnhost/common

replace github.com/John-Tonny/mnhost/model => /root/mygo/src/github.com/John-Tonny/mnhost/model

replace github.com/John-Tonny/mnhost/utils => /root/mygo/src/github.com/John-Tonny/mnhost/utils

replace github.com/John-Tonny/mnhost/user/model => /root/mygo/src/github.com/John-Tonny/mnhost/user/model

replace github.com/John-Tonny/mnhost/types => /root/mygo/src/github.com/John-Tonny/mnhost/types

replace github.com/John-Tonny/mnhost/user/handler => /root/mygo/src/github.com/John-Tonny/mnhost/user/handler

replace github.com/John-Tonny/mnhost/interface/out/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/user

replace github.com/John-Tonny/mnhost/interface/proto/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/proto/user

replace github.com/John-Tonny/mnhost/interface/out/vps => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/vps

require (
	github.com/John-Tonny/micro/vps/amazon v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/conf v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/interface/out/user v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/interface/out/vps v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/model v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/types v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/utils v0.0.0-00010101000000-000000000000 // indirect
	github.com/astaxie/beego v1.12.0 // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/dongjialong2006/log v0.0.0-20181213075813-1c1f8dfa5a28 // indirect
	github.com/dongjialong2006/monitor v0.0.0-20180808054500-5bd25b886bae // indirect
	github.com/dynport/gossh v0.0.0-20170809141523-122e3ee2a6b0 // indirect
	github.com/evalphobia/logrus_fluent v0.5.4 // indirect
	github.com/fluent/fluent-logger-golang v1.4.0 // indirect
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/micro/go-log v0.1.0 // indirect
	github.com/micro/go-micro v1.18.0 // indirect
	github.com/micro/go-plugins v1.5.1 // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/robfig/cron/v3 v3.0.0 // indirect
	github.com/shirou/gopsutil v2.19.11+incompatible // indirect
	github.com/simulatedsimian/cpuusage v0.0.0-20190419175213-c393e527f3fa // indirect
	github.com/uber/jaeger-client-go v2.20.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
	moul.io/http2curl v1.0.0 // indirect
)
