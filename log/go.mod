module github.com/John-Tonny/mnhost/log

go 1.13

replace github.com/John-Tonny/mnhost/interface/out/log => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/log

replace github.com/John-Tonny/mnhost/conf => /root/mygo/src/github.com/John-Tonny/mnhost/conf

replace github.com/John-Tonny/mnhost/common => /root/mygo/src/github.com/John-Tonny/mnhost/common

replace github.com/John-Tonny/mnhost/model => /root/mygo/src/github.com/John-Tonny/mnhost/model

replace github.com/John-Tonny/mnhost/utils => /root/mygo/src/github.com/John-Tonny/mnhost/utils

replace github.com/John-Tonny/mnhost/types => /root/mygo/src/github.com/John-Tonny/mnhost/types

replace github.com/John-Tonny/mnhost/user/model => /root/mygo/src/github.com/John-Tonny/mnhost/user/model

replace github.com/John-Tonny/mnhost/user/handler => /root/mygo/src/github.com/John-Tonny/mnhost/user/handler

replace github.com/John-Tonny/mnhost/interface/out/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/out/user

replace github.com/John-Tonny/mnhost/interface/proto/user => /root/mygo/src/github.com/John-Tonny/mnhost/interface/proto/user

replace github.com/John-Tonny/micro/vps/amazon => /root/mygo/src/github.com/John-Tonny/micro/vps/amazon

require (
	github.com/John-Tonny/mnhost/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/John-Tonny/mnhost/interface/out/log v0.0.0-00010101000000-000000000000 // indirect
	github.com/docker/docker v1.13.1
)
