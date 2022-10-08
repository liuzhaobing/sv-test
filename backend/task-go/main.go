package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"task-go/models"
	"task-go/pkg/logf"
	"task-go/pkg/setting"
	"task-go/routers"
)

func init() {
	setting.Setup()
	logf.Setup()
	models.MysqlSetup()
	models.MongoSetup()
}

func main() {
	gin.SetMode(setting.ServerSetting.RunMode)
	r := routers.InitRouter()
	endPoint := fmt.Sprintf(":%d", setting.ServerSetting.HttpPort)
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        r,
		MaxHeaderBytes: maxHeaderBytes,
	}

	log.Printf("[info] start httpweb server listening %s", endPoint)

	err := server.ListenAndServe()
	if err != nil {
		return
	}
}
