package Group

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

func ListGroup(context *gin.Context) {
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
	model := models.NewTaskGroupModel()
	result, _ := model.GetTaskGroups((req.PageNum-1)*req.PageSize, req.PageSize, fmt.Sprintf("is_del=%d", models.Available))
	app.SuccessResponseData(context, len(result), result)
}

func GetGroup(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewTaskGroupModel()
	exist, err := model.ExistTaskGroupByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	result, _ := model.GetTaskGroupByID(id)
	app.SuccessResponseData(context, 1, result)
}

func DeleteGroup(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewTaskGroupModel()
	exist, err := model.ExistTaskGroupByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	err = model.UpdateTaskGroupByID(id, &models.TaskGroup{IsDel: models.Deleted})
	result, _ := model.GetTaskGroupByID(id)
	if err != nil {
		app.ErrorResp(context, e.ERROR, "update is_del failed reason: "+err.Error())
		return
	}
	app.SuccessResponseData(context, 1, result)
}

func AddGroup(context *gin.Context) {
	//BindAndValid
	copyReq := &models.TaskGroup{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error())
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(util.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(util.REQUEST_KEY).(*models.TaskGroup)
	model := models.NewTaskGroupModel()
	id, err := model.AddTaskGroup(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error())
		return
	}

	result, err := model.GetTaskGroupByID(id)
	app.SuccessResponseData(context, 1, result)
}

func UpdateGroup(context *gin.Context) {
	//BindAndValid
	copyReq := &models.TaskGroup{}
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

	req := context.MustGet(util.REQUEST_KEY).(*models.TaskGroup)
	model := models.NewTaskGroupModel()
	exist, err := model.ExistTaskGroupByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	err = model.UpdateTaskGroupByID(id, req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error())
		return
	}

	result, err := model.GetTaskGroupByID(id)
	app.SuccessResponseData(context, 1, result)
}
