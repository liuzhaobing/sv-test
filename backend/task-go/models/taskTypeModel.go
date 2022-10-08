package models

import (
	"github.com/jinzhu/gorm"
	"task-go/pkg/util"
)

const TaskTypeTableName = "task_type"

type TaskType struct {
	Id          int64         `form:"id,omitempty"    json:"id"   gorm:"column:id"  gorm:"primary_key"   `
	Name        string        `form:"name,omitempty"  json:"name" gorm:"column:name"`
	Description string        `form:"description,omitempty"  json:"description" gorm:"column:description"`
	IsDel       int64         `form:"is_del,omitempty"  json:"is_del" gorm:"column:is_del"`
	TaskType    int64         `form:"task_type,omitempty"  json:"task_type" gorm:"column:task_type"`
	CreateTime  util.JSONTime `form:"create_time,omitempty"    json:"create_time"    gorm:"column:create_time"`
	UpdateTime  util.JSONTime `form:"update_time,omitempty"    json:"update_time"    gorm:"column:update_time"`

	Session *Session `json:"-" gorm:"-"`
}

func (TaskType) TableName() string {
	return TaskTypeTableName
}

func NewTaskTypeModel() *TaskType {
	return &TaskType{Session: NewSession()}
}

func (g *TaskType) ExistTaskTypeByID(id int64) (bool, error) {
	s := &TaskType{}
	err := g.Session.db.Select("id").Where("id = ? ", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if s.Id > 0 {
		return true, nil
	}
	return false, nil
}

func (g *TaskType) GetTaskTypeTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := g.Session.db.Model(&TaskType{}).Where(query, args...).Count(&count).Error
	return count, err
}

func (g *TaskType) GetTaskTypes(pageNum int, pageSize int, maps interface{}, args ...interface{}) ([]*TaskType, error) {
	var s []*TaskType
	err := g.Session.db.Where(maps, args...).Offset(pageNum).Limit(pageSize).Find(&s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return s, nil
}

func (g *TaskType) GetTaskTypeByID(id int64) (*TaskType, error) {
	s := &TaskType{}
	err := g.Session.db.Where("id = ?", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return s, nil
}

func (g *TaskType) UpdateTaskTypeByID(id int64, s *TaskType) error {
	tx := GetSessionTx(g.Session)
	return tx.Model(&TaskType{}).Where("id = ?", id).Updates(s).Error
}

func (g *TaskType) AddTaskType(s *TaskType) (int64, error) {
	tx := GetSessionTx(g.Session)
	err := tx.Create(s).Error
	if err != nil {
		return 0, err
	}
	return s.Id, nil
}

func (g *TaskType) DeleteTaskTypeByID(id int64) error {
	tx := GetSessionTx(g.Session)
	return tx.Where("id = ?", id).Delete(TaskType{}).Error
}
