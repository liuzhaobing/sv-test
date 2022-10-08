package models

import "github.com/jinzhu/gorm"

const beginStatus = 1
const endStatus = 0

type Session struct {
	db           *gorm.DB
	tx           *gorm.DB
	commitSign   int8
	rollbackSign bool
}

func NewSession() *Session {
	session := new(Session)
	session.db = db
	return session
}

func GetSessionTx(session *Session) *gorm.DB {
	if session.tx != nil {
		return session.tx
	}

	return session.db
}

func (s *Session) Begin() {
	s.rollbackSign = true
	if s.tx == nil {
		s.tx = db.Begin()
		s.commitSign = beginStatus
	}
}

func (s *Session) Rollback() {
	if s.tx != nil && s.rollbackSign == true {
		s.tx.Rollback()
		s.tx = nil
	}
}

func (s *Session) Commit() {
	s.rollbackSign = false
	if s.tx != nil {
		if s.commitSign == beginStatus {
			s.tx.Commit()
			s.tx = nil
		} else {
			s.commitSign = endStatus
		}
	}
}
