module github.com/John-Tonny/mnhost/mnmsg

go 1.13

replace github.com/John-Tonny/mnhost/interface/out/log => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/log

replace github.com/John-Tonny/mnhost/conf => /root/mygo/src/github.com/John-Tonny/mnhost/conf

replace github.com/John-Tonny/mnhost/common => /root/mygo/src/github.com/John-Tonny/mnhost/common

replace github.com/John-Tonny/mnhost/model => /root/mygo/src/github.com/John-Tonny/mnhost/model

replace github.com/John-Tonny/mnhost/types => /root/mygo/src/github.com/John-Tonny/mnhost/types

replace github.com/John-Tonny/mnhost/utils => /root/mygo/src/github.com/John-Tonny/mnhost/utils

replace github.com/John-Tonny/mnhost/user/model => /root/mygo/src/github.com/John-Tonny/mnhost/user/model

replace github.com/John-Tonny/mnhost/user/handler => /root/mygo/src/github.com/John-Tonny/mnhost/user/handler

replace github.com/John-Tonny/mnhost/interface/out/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/user

replace github.com/John-Tonny/mnhost/interface/proto/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/proto/user

replace github.com/John-Tonny/mnhost/interface/out/mnmsg => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/mnmsg

replace github.com/John-Tonny/mnhost/interface/out/vps => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/vps

replace github.com/John-Tonny/micro/vps/amazon => /root/mygo/src/github.com/John-Tonny/micro/vps/amazon

require (
	github.com/John-Tonny/mnhost/common v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/conf v0.0.0-20191225064632-834cc3555a47
	github.com/John-Tonny/mnhost/interface/out/log v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/mnmsg v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/interface/out/vps v0.0.0-00010101000000-000000000000
	github.com/John-Tonny/mnhost/model v0.0.0-20191225064632-834cc3555a47
	github.com/John-Tonny/mnhost/utils v0.0.0-20191225064632-834cc3555a47
	github.com/json-iterator/go v1.1.8
	github.com/micro/go-micro v1.18.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
)
