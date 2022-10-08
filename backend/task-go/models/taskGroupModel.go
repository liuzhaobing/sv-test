package models

import (
	"github.com/jinzhu/gorm"
	"task-go/pkg/util"
)

const TaskGroupTableName = "task_group"

type TaskGroup struct {
	Id          int64         `form:"id,omitempty"    json:"id"   gorm:"column:id"  gorm:"primary_key"   `
	Name        string        `form:"name,omitempty"  json:"name" gorm:"column:name"`
	Description string        `form:"description,omitempty"  json:"description" gorm:"column:description"`
	IsDel       int64         `form:"is_del,omitempty"  json:"is_del" gorm:"column:is_del"`
	CreateTime  util.JSONTime `form:"create_time,omitempty"    json:"create_time"    gorm:"column:create_time"`
	UpdateTime  util.JSONTime `form:"update_time,omitempty"    json:"update_time"    gorm:"column:update_time"`

	Session *Session `json:"-" gorm:"-"`
}

func (TaskGroup) TableName() string {
	return TaskGroupTableName
}

func NewTaskGroupModel() *TaskGroup {
	return &TaskGroup{Session: NewSession()}
}

func (g *TaskGroup) ExistTaskGroupByID(id int64) (bool, error) {
	s := &TaskGroup{}
	err := g.Session.db.Select("id").Where("id = ? ", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if s.Id > 0 {
		return true, nil
	}
	return false, nil
}

func (g *TaskGroup) GetTaskGroupTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := g.Session.db.Model(&TaskGroup{}).Where(query, args...).Count(&count).Error
	return count, err
}

func (g *TaskGroup) GetTaskGroups(pageNum int, pageSize int, maps interface{}, args ...interface{}) ([]*TaskGroup, error) {
	var s []*TaskGroup
	err := g.Session.db.Where(maps, args...).Offset(pageNum).Limit(pageSize).Find(&s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return s, nil
}

func (g *TaskGroup) GetTaskGroupByID(id int64) (*TaskGroup, error) {
	s := &TaskGroup{}
	err := g.Session.db.Where("id = ?", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return s, nil
}

func (g *TaskGroup) UpdateTaskGroupByID(id int64, s *TaskGroup) error {
	tx := GetSessionTx(g.Session)
	return tx.Model(&TaskGroup{}).Where("id = ?", id).Updates(s).Error
}

func (g *TaskGroup) AddTaskGroup(s *TaskGroup) (int64, error) {
	tx := GetSessionTx(g.Session)
	err := tx.Create(s).Error
	if err != nil {
		return 0, err
	}
	return s.Id, nil
}

func (g *TaskGroup) DeleteTaskGroupByID(id int64) error {
	tx := GetSessionTx(g.Session)
	return tx.Where("id = ?", id).Delete(TaskGroup{}).Error
}
