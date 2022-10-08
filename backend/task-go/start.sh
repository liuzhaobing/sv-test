#/bin/bash
go mod download
go mod verify
go build task-go
nohup ./task-go -id 8089 &
