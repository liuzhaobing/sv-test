package _type

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

func ListType(context *gin.Context) {
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
	model := models.NewTaskTypeModel()
	result, _ := model.GetTaskTypes((req.PageNum-1)*req.PageSize, req.PageSize, fmt.Sprintf("is_del=%d", models.Available))
	app.SuccessResponseData(context, len(result), result)
}

func GetType(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewTaskTypeModel()
	exist, err := model.ExistTaskTypeByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	result, _ := model.GetTaskTypeByID(id)
	app.SuccessResponseData(context, 1, result)
}

func DeleteType(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewTaskTypeModel()
	exist, err := model.ExistTaskTypeByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	err = model.UpdateTaskTypeByID(id, &models.TaskType{IsDel: models.Deleted})
	result, _ := model.GetTaskTypeByID(id)
	if err != nil {
		app.ErrorResp(context, e.ERROR, "update is_del failed reason: "+err.Error())
		return
	}
	app.SuccessResponseData(context, 1, result)
}

func AddType(context *gin.Context) {
	//BindAndValid
	copyReq := &models.TaskType{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error())
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(util.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(util.REQUEST_KEY).(*models.TaskType)
	model := models.NewTaskTypeModel()
	id, err := model.AddTaskType(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error())
		return
	}

	result, err := model.GetTaskTypeByID(id)
	app.SuccessResponseData(context, 1, result)
}

func UpdateType(context *gin.Context) {
	//BindAndValid
	copyReq := &models.TaskType{}
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

	req := context.MustGet(util.REQUEST_KEY).(*models.TaskType)
	model := models.NewTaskTypeModel()
	exist, err := model.ExistTaskTypeByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	err = model.UpdateTaskTypeByID(id, req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error())
		return
	}

	result, err := model.GetTaskTypeByID(id)
	app.SuccessResponseData(context, 1, result)
}
