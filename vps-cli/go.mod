module github.com/John-Tonny/mnhost/user-cli

go 1.13

replace github.com/John-Tonny/mnhost/interface/out/log => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/log

replace github.com/John-Tonny/mnhost/conf => /root/mygo/src/github.com/John-Tonny/mnhost/conf

replace github.com/John-Tonny/mnhost/common => /root/mygo/src/github.com/John-Tonny/mnhost/common

replace github.com/John-Tonny/mnhost/model => /root/mygo/src/github.com/John-Tonny/mnhost/model

replace github.com/John-Tonny/mnhost/utils => /root/mygo/src/github.com/John-Tonny/mnhost/utils

replace github.com/John-Tonny/mnhost/user/model => /root/mygo/src/github.com/John-Tonny/mnhost/user/model

replace github.com/John-Tonny/mnhost/user/handler => /root/mygo/src/github.com/John-Tonny/mnhost/user/handler

replace github.com/John-Tonny/mnhost/interface/out/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/user

replace github.com/John-Tonny/mnhost/interface/proto/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/proto/user

replace github.com/John-Tonny/mnhost/interface/out/vps => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/vps

require (
	github.com/John-Tonny/mnhost/common v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/conf v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/user v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/interface/out/vps v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/model v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/utils v0.0.0-00010101000000-000000000000 // indirect
	github.com/astaxie/beego v1.12.0
	github.com/dynport/gossh v0.0.0-20170809141523-122e3ee2a6b0 // indirect
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/json-iterator/go v1.1.8
	github.com/micro/go-log v0.1.0 // indirect
	github.com/micro/go-plugins v1.5.1 // indirect
	github.com/uber/jaeger-client-go v2.20.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/yaml.v2 v2.2.5 // indirect
)
