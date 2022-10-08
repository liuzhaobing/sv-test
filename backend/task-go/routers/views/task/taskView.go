package task

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"task-go/models"
	"task-go/pkg/app"
	"task-go/pkg/e"
	"task-go/pkg/logf"
	util "task-go/pkg/util/const"
	"time"
)

func ListPlan(context *gin.Context) {
	//BindAndValid
	copyReq := &models.PlanList{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error())
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(util.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(util.REQUEST_KEY).(*models.PlanList)
	model := models.NewTaskPlanModel()

	result, _ := model.GetTaskPlans((req.PageNum-1)*req.PageSize, req.PageSize, fmt.Sprintf("is_del=%d", models.Available))
	app.SuccessResponseData(context, len(result), result)
}

func GetPlan(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewTaskPlanModel()
	exist, err := model.ExistTaskPlanByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	result, _ := model.GetTaskPlanByID(id)
	app.SuccessResponseData(context, 1, result)
}

func DeletePlan(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewTaskPlanModel()
	exist, err := model.ExistTaskPlanByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	err = model.UpdateTaskPlanByID(id, &models.TaskPlan{IsDel: models.Deleted})
	result, _ := model.GetTaskPlanByID(id)
	if err != nil {
		app.ErrorResp(context, e.ERROR, "update is_del failed reason: "+err.Error())
		return
	}
	app.SuccessResponseData(context, 1, result)
}

func AddPlan(context *gin.Context) {
	//BindAndValid
	copyReq := &models.TaskPlan{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error())
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(util.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(util.REQUEST_KEY).(*models.TaskPlan)
	model := models.NewTaskPlanModel()
	id, err := model.AddTaskPlan(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error())
		return
	}

	result, err := model.GetTaskPlanByID(id)
	app.SuccessResponseData(context, 1, result)
}

func UpdatePlan(context *gin.Context) {
	//BindAndValid
	copyReq := &models.TaskPlan{}
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

	req := context.MustGet(util.REQUEST_KEY).(*models.TaskPlan)
	model := models.NewTaskPlanModel()
	exist, err := model.ExistTaskPlanByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!")
		return
	}
	err = model.UpdateTaskPlanByID(id, req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error())
		return
	}

	result, err := model.GetTaskPlanByID(id)
	app.SuccessResponseData(context, 1, result)
}

func ChangeRunningStatus(id, status int64) (*models.TaskPlan, bool, string) {
	model := models.NewTaskPlanModel()
	exist, err := model.ExistTaskPlanByID(id)
	if err != nil || !exist {
		return nil, false, err.Error()
	}
	result, _ := model.GetTaskPlanByID(id)
	if result.IsRun == status {
		return result, false, "can not change running status!"
	}
	err = model.UpdateTaskPlanByID(id, &models.TaskPlan{IsRun: status})
	if err != nil {
		return nil, false, "change running status failed!"
	}
	result.IsRun = status
	return result, true, "change running status succeed!"
}

func RunPlanOrEndPlan(context *gin.Context) {
	//BindAndValid
	copyReq := &models.RunPlanByID{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error())
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(util.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(util.REQUEST_KEY).(*models.RunPlanByID)

	// start run or stop run by id
	result, status, message := ChangeRunningStatus(req.Id, req.Status)
	// backend run task for long time
	if status {
		if req.Status == models.Running {
			go runSomeThing(req.Id)
		}
		if req.Status == models.Available {
			go endSomeThing(req.Id)
		}
	}

	if status {
		// fronted received response message
		app.SuccessResponseData(context, 1, result)
		return
	} else {
		app.ErrorResp(context, e.ERROR, message)
		return
	}
}

// RunningChannel 存储每个taskID的cancel
var RunningChannel = make(map[int64]context.CancelFunc)

// RunningProgress 存储每个taskID的progress
var RunningProgress = make(map[int64]float32)

func ListInstance(context *gin.Context) {
	app.SuccessResponseData(context, len(RunningProgress), RunningProgress)
}

func runSomeThing(taskID int64) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	RunningChannel[taskID] = cancelFunc

	total := 999
	nowCaseId := make(chan int)
	go worker(ctx, fmt.Sprintf("task id = %d is stopped", taskID), nowCaseId, total)

	getProgress(ctx, nowCaseId, taskID, total)
	cancelFunc()
}

// 统计进度
func getProgress(ctx context.Context, nowCaseId chan int, taskID int64, total int) {
	for {
		select {
		// 收到channel取消信号 则不再继续统计进度
		case <-ctx.Done():
			return

		// 持续从out这个channel中提取进度信息
		case t := <-nowCaseId:
			RunningProgress[taskID] = float32(t) / float32(total)
			if t == total {
				return
			}
		}
	}
}

func worker(ctx context.Context, name string, out chan<- int, total int) {
	go func() {
		err := Stream(ctx, out, total)
		if err != nil {
		}
	}()
	select {
	case <-ctx.Done():
		fmt.Println(name, "got the stop channel")
		return
	}
}

func Stream(ctx context.Context, out chan<- int, total int) error {
	for {
		go DoSomething(ctx, out, total)
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func DoSomething(ctx context.Context, out chan<- int, total int) {
	for i := 0; i < total; i++ {
		time.Sleep(100 * time.Millisecond)
		out <- i
	}
}

func stopSomething(cancelFunc context.CancelFunc) {
	defer cancelFunc()
}

func endSomeThing(taskID int64) {
	if _, ok := RunningChannel[taskID]; ok {
		stopSomething(RunningChannel[taskID])
		delete(RunningChannel, taskID)
		delete(RunningProgress, taskID)
	}
}
