package task

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"task-go/models"
	"task-go/pkg/logf"
	"time"
)

type Task interface {
	setUp()
	run()
	tearDown()
	setBaseConfig(*models.BaseConfig)
	setTaskConfig(interface{})
	getFeiShuContent() string
}

type backendTask struct {
	baseTask
}

func (b *backendTask) run() {}

func (b *backendTask) setTaskConfig(interface{}) {}

func (b *backendTask) setBaseConfig(baseConfig *models.BaseConfig) {
	b.baseConfig = baseConfig
}

func (b *backendTask) getFeiShuContent() string {
	return "backendTask"
}

var _ Task = &backendTask{}

type baseTask struct {
	t          *Task
	baseConfig *models.BaseConfig
	startTime  time.Time
	endTime    time.Time
}

func (b *baseTask) setUp() {
	b.startTime = time.Now()
}

func (b *baseTask) Run() {
	(*b.t).setUp()
	(*b.t).run()
	(*b.t).tearDown()

	text := (*b.t).getFeiShuContent()
	for _, addr := range b.baseConfig.WebhookConfig {
		if addr.IsWebhook == models.Available {
			HttpToFeiShu(text, addr.WebhookString)
		}
	}
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()
}

func HttpToFeiShu(text, url string) {

	method := "POST"
	newString := strings.ReplaceAll(text, "\n", "\\n")
	newString = strings.ReplaceAll(newString, "\t", "\\t")
	payload := strings.NewReader(fmt.Sprintf("{\n    \"msg_type\": \"text\",\n    \"content\": {\n        \"text\": \"%s\"\n    }\n}", newString))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		logf.Error("err is ", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		logf.Error("http err ,err is ", err)
	}

	logf.Info(string(body))
}

func (b *baseTask) tearDown() {
	b.endTime = time.Now()
}

type CornManage struct {
	CronServer *cron.Cron
	maxTask    int
	taskMap    map[cron.EntryID]*Task
}

func (c *CornManage) GetCronNum() int {
	return len(c.CronServer.Entries())
}

func (c *CornManage) AddTask(baseConfig *models.BaseConfig, taskConfig interface{}, job Task) (cron.EntryID, error) {
	// 检测到新增/修改测试计划时 将任务加入到crontab
	if c.GetCronNum()+1 > c.maxTask {
		return 0, errors.New("max task error")
	}
	if job == nil {
		return 0, errors.New("task must not nil")
	}
	job.setTaskConfig(taskConfig)
	job.setBaseConfig(baseConfig)

	var b = &baseTask{
		t:          &job,
		baseConfig: baseConfig,
	}
	res, err := c.CronServer.AddJob(baseConfig.CronConfig.CronString, b)
	if err != nil {
		return res, err
	}
	c.taskMap[res] = &job
	return res, err
}

func (c *CornManage) RemoveTask(id cron.EntryID) cron.EntryID {
	// 检测到删除/修改测试计划时 先删除对应的cron
	c.CronServer.Remove(id)
	delete(c.taskMap, id)
	return id
}

type taskInfo struct {
	CronId     int         `json:"cron_id"`
	NextTime   time.Time   `json:"next_time"`
	TaskConfig interface{} `json:"task_config"`
}

func (c *CornManage) GetCronList() (resList []*taskInfo) {
	keys := make([]int, 0)
	for i, _ := range c.taskMap {
		keys = append(keys, int(i))
	}
	sort.Ints(keys)
	for _, key := range keys {
		resList = append(resList, &taskInfo{
			CronId:     key,
			NextTime:   time.Time{},
			TaskConfig: nil,
		})
	}
	return
}

var CM = &CornManage{}

func init() {
	CM.CronServer = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)))
	CM.taskMap = make(map[cron.EntryID]*Task)
	CM.maxTask = 1000

	CM.CronServer.Start()
}
