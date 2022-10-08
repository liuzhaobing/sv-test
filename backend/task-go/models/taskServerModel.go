package models

import (
	"github.com/jinzhu/gorm"
	"task-go/pkg/util"
)

const TaskServerTableName = "task_server"

type TaskServer struct {
	Id          int64         `form:"id,omitempty"    json:"id"   gorm:"column:id"  gorm:"primary_key"   `
	Name        string        `form:"name,omitempty"  json:"name" gorm:"column:name"`
	Description string        `form:"description,omitempty"  json:"description" gorm:"column:description"`
	IsDel       int64         `form:"is_del,omitempty"  json:"is_del" gorm:"column:is_del"`
	CreateTime  util.JSONTime `form:"create_time,omitempty"    json:"create_time"    gorm:"column:create_time"`
	UpdateTime  util.JSONTime `form:"update_time,omitempty"    json:"update_time"    gorm:"column:update_time"`

	Session *Session `json:"-" gorm:"-"`
}

func (TaskServer) TableName() string {
	return TaskServerTableName
}

func NewTaskServerModel() *TaskServer {
	return &TaskServer{Session: NewSession()}
}

func (g *TaskServer) ExistTaskServerByID(id int64) (bool, error) {
	s := &TaskServer{}
	err := g.Session.db.Select("id").Where("id = ? ", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if s.Id > 0 {
		return true, nil
	}
	return false, nil
}

func (g *TaskServer) GetTaskServerTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := g.Session.db.Model(&TaskServer{}).Where(query, args...).Count(&count).Error
	return count, err
}

func (g *TaskServer) GetTaskServers(pageNum int, pageSize int, maps interface{}, args ...interface{}) ([]*TaskServer, error) {
	var s []*TaskServer
	err := g.Session.db.Where(maps, args...).Offset(pageNum).Limit(pageSize).Find(&s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return s, nil
}

func (g *TaskServer) GetTaskServerByID(id int64) (*TaskServer, error) {
	s := &TaskServer{}
	err := g.Session.db.Where("id = ?", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return s, nil
}

func (g *TaskServer) UpdateTaskServerByID(id int64, s *TaskServer) error {
	tx := GetSessionTx(g.Session)
	return tx.Model(&TaskServer{}).Where("id = ?", id).Updates(s).Error
}

func (g *TaskServer) AddTaskServer(s *TaskServer) (int64, error) {
	tx := GetSessionTx(g.Session)
	err := tx.Create(s).Error
	if err != nil {
		return 0, err
	}
	return s.Id, nil
}

func (g *TaskServer) DeleteTaskServerByID(id int64) error {
	tx := GetSessionTx(g.Session)
	return tx.Where("id = ?", id).Delete(TaskServer{}).Error
}
