module github.com/John-Tonny/mnhost/api/vps

go 1.13

replace github.com/John-Tonny/mnhost/interface/out/log => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/log

replace github.com/John-Tonny/mnhost/conf => /root/mygo/src/github.com/John-Tonny/mnhost/conf

replace github.com/John-Tonny/mnhost/common => /root/mygo/src/github.com/John-Tonny/mnhost/common

replace github.com/John-Tonny/mnhost/model => /root/mygo/src/github.com/John-Tonny/mnhost/model

replace github.com/John-Tonny/mnhost/utils => /root/mygo/src/github.com/John-Tonny/mnhost/utils

replace github.com/John-Tonny/mnhost/types => /root/mygo/src/github.com/John-Tonny/mnhost/types

replace github.com/John-Tonny/mnhost/vps/handler => /root/mygo/src/github.com/John-Tonny/mnhost/vps/handler

replace github.com/John-Tonny/mnhost/interface/out/vps => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/vps

replace github.com/John-Tonny/mnhost/api/vps/handler => /root/mygo/src/github.com/John-Tonny/mnhost/api/vps/handler

require (
	github.com/John-Tonny/mnhost/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/conf v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/interface/out/user v0.0.0-20191128093822-f81a1f9dd5df // indirect
	github.com/John-Tonny/mnhost/interface/out/vps v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/model v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/types v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/utils v0.0.0-00010101000000-000000000000 // indirect
	github.com/astaxie/beego v1.12.0 // indirect
	github.com/docker/docker v1.13.1
	github.com/dynport/gossh v0.0.0-20170809141523-122e3ee2a6b0 // indirect
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/gin-gonic/gin v1.5.0
	github.com/micro/go-log v0.1.0 // indirect
	github.com/micro/go-micro v1.18.0 // indirect
	github.com/micro/go-plugins v1.5.1 // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/uber/jaeger-client-go v2.20.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	moul.io/http2curl v1.0.0 // indirect
)
