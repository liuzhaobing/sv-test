package models

import (
	"github.com/jinzhu/gorm"
	"task-go/pkg/util"
)

const TaskProjectTableName = "task_project"

type TaskProject struct {
	Id          int64         `form:"id,omitempty"    json:"id"   gorm:"column:id"  gorm:"primary_key"   `
	Name        string        `form:"name,omitempty"  json:"name" gorm:"column:name"`
	Description string        `form:"description,omitempty"  json:"description" gorm:"column:description"`
	IsDel       int64         `form:"is_del,omitempty"  json:"is_del" gorm:"column:is_del"`
	CreateTime  util.JSONTime `form:"create_time,omitempty"    json:"create_time"    gorm:"column:create_time"`
	UpdateTime  util.JSONTime `form:"update_time,omitempty"    json:"update_time"    gorm:"column:update_time"`

	Session *Session `json:"-" gorm:"-"`
}

func (TaskProject) TableName() string {
	return TaskProjectTableName
}

func NewTaskProjectModel() *TaskProject {
	return &TaskProject{Session: NewSession()}
}

func (g *TaskProject) ExistTaskProjectByID(id int64) (bool, error) {
	s := &TaskProject{}
	err := g.Session.db.Select("id").Where("id = ? ", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if s.Id > 0 {
		return true, nil
	}
	return false, nil
}

func (g *TaskProject) GetTaskProjectTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := g.Session.db.Model(&TaskProject{}).Where(query, args...).Count(&count).Error
	return count, err
}

func (g *TaskProject) GetTaskProjects(pageNum int, pageSize int, maps interface{}, args ...interface{}) ([]*TaskProject, error) {
	var s []*TaskProject
	err := g.Session.db.Where(maps, args...).Offset(pageNum).Limit(pageSize).Find(&s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return s, nil
}

func (g *TaskProject) GetTaskProjectByID(id int64) (*TaskProject, error) {
	s := &TaskProject{}
	err := g.Session.db.Where("id = ?", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return s, nil
}

func (g *TaskProject) UpdateTaskProjectByID(id int64, s *TaskProject) error {
	tx := GetSessionTx(g.Session)
	return tx.Model(&TaskProject{}).Where("id = ?", id).Updates(s).Error
}

func (g *TaskProject) AddTaskProject(s *TaskProject) (int64, error) {
	tx := GetSessionTx(g.Session)
	err := tx.Create(s).Error
	if err != nil {
		return 0, err
	}
	return s.Id, nil
}

func (g *TaskProject) DeleteTaskProjectByID(id int64) error {
	tx := GetSessionTx(g.Session)
	return tx.Where("id = ?", id).Delete(TaskProject{}).Error
}
