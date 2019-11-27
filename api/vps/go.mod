module github.com/John-Tonny/mnhost/api/vps

go 1.13

require (
	github.com/John-Tonny/mnhost/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/conf v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/interface/out/user v0.0.0-20191119033601-3f8fb810845b // indirect
	github.com/John-Tonny/mnhost/interface/out/vps v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.4.0
	github.com/micro/go-log v0.1.0 // indirect
	github.com/micro/go-micro v1.16.0
	github.com/micro/go-plugins v1.5.1 // indirect
	github.com/uber/jaeger-client-go v2.20.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
)

replace github.com/John-Tonny/mnhost/interface/out/log => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/log

replace github.com/John-Tonny/mnhost/conf => /root/mygo/src/github.com/John-Tonny/mnhost/conf

replace github.com/John-Tonny/mnhost/common => /root/mygo/src/github.com/John-Tonny/mnhost/common

replace github.com/John-Tonny/mnhost/model => /root/mygo/src/github.com/John-Tonny/mnhost/model

replace github.com/John-Tonny/mnhost/utils => /root/mygo/src/github.com/John-Tonny/mnhost/utils

replace github.com/John-Tonny/mnhost/vps/handler => /root/mygo/src/github.com/John-Tonny/mnhost/vps/handler

replace github.com/John-Tonny/mnhost/interface/out/vps => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/vps

replace github.com/John-Tonny/mnhost/api/vps/handler => /root/mygo/src/github.com/John-Tonny/mnhost/api/vps/handler
