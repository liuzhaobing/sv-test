package project

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

func ListProject(context *gin.Context) {
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
	model := models.NewTaskProjectModel()

	result, _ := model.GetTaskProjects((req.PageNum-1)*req.PageSize, req.PageSize, fmt.Sprintf("is_del=%d", models.Available))
	app.SuccessResponseData(context, len(result), result)
}

func GetProject(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewTaskProjectModel()
	exist, err := model.ExistTaskProjectByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	result, _ := model.GetTaskProjectByID(id)
	app.SuccessResponseData(context, 1, result)
}

func DeleteProject(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewTaskProjectModel()
	exist, err := model.ExistTaskProjectByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	err = model.UpdateTaskProjectByID(id, &models.TaskProject{IsDel: models.Deleted})
	result, _ := model.GetTaskProjectByID(id)
	if err != nil {
		app.ErrorResp(context, e.ERROR, "update is_del failed reason: "+err.Error())
		return
	}
	app.SuccessResponseData(context, 1, result)
}

func AddProject(context *gin.Context) {
	//BindAndValid
	copyReq := &models.TaskProject{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error())
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(util.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(util.REQUEST_KEY).(*models.TaskProject)
	model := models.NewTaskProjectModel()
	id, err := model.AddTaskProject(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error())
		return
	}

	result, err := model.GetTaskProjectByID(id)
	app.SuccessResponseData(context, 1, result)
}

func UpdateProject(context *gin.Context) {
	//BindAndValid
	copyReq := &models.TaskProject{}
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

	req := context.MustGet(util.REQUEST_KEY).(*models.TaskProject)
	model := models.NewTaskProjectModel()
	exist, err := model.ExistTaskProjectByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	err = model.UpdateTaskProjectByID(id, req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error())
		return
	}

	result, err := model.GetTaskProjectByID(id)
	app.SuccessResponseData(context, 1, result)
}
