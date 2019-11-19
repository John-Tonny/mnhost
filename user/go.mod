module github.com/John-Tonny/mnhost/user

go 1.13

require (
	github.com/John-Tonny/mnhost/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/conf v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/log v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/user v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/model v0.0.0-00010101000000-000000000000
	github.com/astaxie/beego v1.12.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/golang/protobuf v1.3.2
	github.com/json-iterator/go v1.1.8
	github.com/micro/go-log v0.1.0 // indirect
	github.com/micro/go-micro v1.16.0
	github.com/micro/go-plugins v1.5.1 // indirect
	github.com/uber/jaeger-client-go v2.20.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	golang.org/x/crypto v0.0.0-20191112222119-e1110fd1c708
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
)

replace github.com/John-Tonny/mnhost/interface/out/log => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/log

replace github.com/John-Tonny/mnhost/conf => /root/mygo/src/github.com/John-Tonny/mnhost/conf

replace github.com/John-Tonny/mnhost/common => /root/mygo/src/github.com/John-Tonny/mnhost/common

replace github.com/John-Tonny/mnhost/model => /root/mygo/src/github.com/John-Tonny/mnhost/model

replace github.com/John-Tonny/mnhost/utils => /root/mygo/src/github.com/John-Tonny/mnhost/utils

replace github.com/John-Tonny/mnhost/user/model => /root/mygo/src/github.com/John-Tonny/mnhost/user/model

replace github.com/John-Tonny/mnhost/user/handler => /root/mygo/src/github.com/John-Tonny/mnhost/user/handler

replace github.com/John-Tonny/mnhost/interface/out/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/user

replace github.com/John-Tonny/mnhost/interface/proto/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/proto/user
