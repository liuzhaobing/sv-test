package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"task-go/models"
	"task-go/pkg/app"
	"task-go/pkg/e"
	"task-go/pkg/logf"
	util "task-go/pkg/util/const"
)

func ListServer(context *gin.Context) {
	//BindAndValid
	copyReq := &models.List{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error())
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(util.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(util.REQUEST_KEY).(*models.List)
	model := models.NewTaskServerModel()

	result, _ := model.GetTaskServers((req.PageNum-1)*req.PageSize, req.PageSize, fmt.Sprintf("is_del=%d", models.Available))
	app.SuccessResponseData(context, len(result), result)
}

func GetServer(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewTaskServerModel()
	exist, err := model.ExistTaskServerByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	result, _ := model.GetTaskServerByID(id)
	app.SuccessResponseData(context, 1, result)
}

func DeleteServer(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewTaskServerModel()
	exist, err := model.ExistTaskServerByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	err = model.UpdateTaskServerByID(id, &models.TaskServer{IsDel: models.Deleted})
	result, _ := model.GetTaskServerByID(id)
	if err != nil {
		app.ErrorResp(context, e.ERROR, "update is_del failed reason: "+err.Error())
		return
	}
	app.SuccessResponseData(context, 1, result)
}

func AddServer(context *gin.Context) {
	//BindAndValid
	copyReq := &models.TaskServer{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error())
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(util.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(util.REQUEST_KEY).(*models.TaskServer)
	model := models.NewTaskServerModel()
	id, err := model.AddTaskServer(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error())
		return
	}

	result, err := model.GetTaskServerByID(id)
	app.SuccessResponseData(context, 1, result)
}

func UpdateServer(context *gin.Context) {
	//BindAndValid
	copyReq := &models.TaskServer{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error())
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(util.REQUEST_KEY, copyReq)

	//VerifyReq
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)

	req := context.MustGet(util.REQUEST_KEY).(*models.TaskServer)
	model := models.NewTaskServerModel()
	exist, err := model.ExistTaskServerByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	err = model.UpdateTaskServerByID(id, req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error())
		return
	}

	result, err := model.GetTaskServerByID(id)
	app.SuccessResponseData(context, 1, result)
}
