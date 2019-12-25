run:
	go run log/main.go &
	go run mnmsg/main.go &
	go run vps/main.go &
	go run api/vps/main.go &
