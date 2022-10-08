package models

import (
	"github.com/jinzhu/gorm"
	"task-go/pkg/util"
)

const TaskPlanTableName = "task_plan"

type TaskPlan struct {
	Id          int64  `form:"id,omitempty"    json:"id"   gorm:"column:id"  gorm:"primary_key"`       // 任务id
	ProjectName string `form:"project_name,omitempty"  json:"project_name" gorm:"column:project_name"` // 项目名
	GroupName   string `form:"group_name,omitempty"  json:"group_name" gorm:"column:group_name"`       // 组名
	Name        string `form:"name,omitempty"  json:"name" gorm:"column:name"`                         // 任务名
	Description string `form:"description,omitempty"  json:"description" gorm:"column:description"`    // 任务描述
	TaskType    int64  `form:"task_type,omitempty"  json:"task_type" gorm:"column:task_type"`          // 任务类型
	IsDel       int64  `form:"is_del,omitempty"  json:"is_del" gorm:"column:is_del"`                   // 是否已删除
	IsRun       int64  `form:"is_run,omitempty"  json:"is_run" gorm:"column:is_run"`                   // 是否正在运行 1空闲 4运行中
	BaseConfig  string `form:"base_config,omitempty"  json:"base_config" gorm:"column:base_config"`    // JSON string
	TaskConfig  string `form:"task_config,omitempty"  json:"task_config" gorm:"column:task_config"`    // JSON string

	CreateTime util.JSONTime `form:"create_time,omitempty"    json:"create_time"    gorm:"column:create_time"`
	UpdateTime util.JSONTime `form:"update_time,omitempty"    json:"update_time"    gorm:"column:update_time"`

	Session *Session `json:"-" gorm:"-"`
}

type CronSwitch struct {
	IsCron      int64  `form:"is_cron,omitempty"  json:"is_cron" gorm:"column:is_cron"`
	CronString  string `form:"cron_string,omitempty"  json:"cron_string" gorm:"column:cron_string"`
	Description string `form:"description,omitempty"  json:"description"  gorm:"column:description"`
}

type WebhookSwitch struct {
	IsWebhook     int64  `form:"is_webhook,omitempty"  json:"is_webhook" gorm:"column:is_webhook"`
	WebhookString string `form:"webhook_string,omitempty"  json:"webhook_string" gorm:"column:webhook_string"`
	Description   string `form:"description,omitempty"  json:"description"  gorm:"column:description"`
}

type BaseConfig struct {
	CronConfig    CronSwitch      `form:"cron_config,omitempty"  json:"cron_config" gorm:"column:cron_config"`
	WebhookConfig []WebhookSwitch `form:"webhook_config,omitempty"  json:"webhook_config" gorm:"column:webhook_config"`
}

type PlanList struct {
	PageNum  int `form:"pagenum,default=1" json:"pagenum"`
	PageSize int `form:"pagesize,default=15" json:"pagesize"`
}

type RunPlanByID struct {
	Id     int64 `form:"id"    json:"id"   gorm:"column:id"`
	Status int64 `form:"status"    json:"status"   gorm:"column:status"`
}

func (TaskPlan) TableName() string {
	return TaskPlanTableName
}

func NewTaskPlanModel() *TaskPlan {
	return &TaskPlan{Session: NewSession()}
}

func (g *TaskPlan) ExistTaskPlanByID(id int64) (bool, error) {
	s := &TaskPlan{}
	err := g.Session.db.Select("id").Where("id = ? ", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if s.Id > 0 {
		return true, nil
	}
	return false, nil
}

func (g *TaskPlan) GetTaskPlanTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := g.Session.db.Model(&TaskPlan{}).Where(query, args...).Count(&count).Error
	return count, err
}

func (g *TaskPlan) GetTaskPlans(pageNum int, pageSize int, maps interface{}, args ...interface{}) ([]*TaskPlan, error) {
	var s []*TaskPlan
	err := g.Session.db.Where(maps, args...).Offset(pageNum).Limit(pageSize).Find(&s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return s, nil
}

func (g *TaskPlan) GetTaskPlanByID(id int64) (*TaskPlan, error) {
	s := &TaskPlan{}
	err := g.Session.db.Where("id = ?", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return s, nil
}

func (g *TaskPlan) UpdateTaskPlanByID(id int64, s *TaskPlan) error {
	tx := GetSessionTx(g.Session)
	return tx.Model(&TaskPlan{}).Where("id = ?", id).Updates(s).Error
}

func (g *TaskPlan) AddTaskPlan(s *TaskPlan) (int64, error) {
	tx := GetSessionTx(g.Session)
	err := tx.Create(s).Error
	if err != nil {
		return 0, err
	}
	return s.Id, nil
}

func (g *TaskPlan) DeleteTaskPlanByID(id int64) error {
	tx := GetSessionTx(g.Session)
	return tx.Where("id = ?", id).Delete(TaskPlan{}).Error
}
