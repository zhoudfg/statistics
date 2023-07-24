package dbsession

import (
	"database/sql"
)

const beginStatus = 1

// SessionFactory 会话工厂
type DBSessionFactory struct {
	*sql.DB
}

// Session 会话
type DBSession struct {
	DB           *sql.DB // 原生db
	tx           *sql.Tx // 原生事务
	commitSign   int8    // 提交标记，控制是否提交事务
	rollbackSign bool    // 回滚标记，控制是否回滚事务
}

// NewSessionFactory 创建一个会话工厂
func NewDBSessionFactory(driverName, dataSourseName string) (*DBSessionFactory, error) {
	db, err := sql.Open(driverName, dataSourseName)
	if err != nil {
		panic(err)
	}
	factory := new(DBSessionFactory)
	factory.DB = db
	return factory, nil
}

// GetSession 获取一个Session
func (sf *DBSessionFactory) GetSession() *DBSession {
	session := new(DBSession)
	session.DB = sf.DB
	return session
}

// Begin 开启事务
func (s *DBSession) Begin() error {
	s.rollbackSign = true
	if s.tx == nil {
		tx, err := s.DB.Begin()
		if err != nil {
			return err
		}
		s.tx = tx
		s.commitSign = beginStatus
		return nil
	}
	s.commitSign++
	return nil
}

// Rollback 回滚事务
func (s *DBSession) Rollback() error {
	if s.tx != nil && s.rollbackSign == true {
		err := s.tx.Rollback()
		if err != nil {
			return err
		}
		s.tx = nil
		return nil
	}
	return nil
}

// Commit 提交事务
func (s *DBSession) Commit() error {
	s.rollbackSign = false
	if s.tx != nil {
		if s.commitSign == beginStatus {
			err := s.tx.Commit()
			if err != nil {
				return err
			}
			s.tx = nil
			return nil
		} else {
			s.commitSign--
		}
		return nil
	}
	return nil
}

// Exec 执行sql语句，如果已经开启事务，就以事务方式执行，如果没有开启事务，就以非事务方式执行
func (s *DBSession) Exec(query string, args ...interface{}) (sql.Result, error) {
	if s.tx != nil {
		return s.tx.Exec(query, args...)
	}
	return s.DB.Exec(query, args...)
}

// QueryRow 如果已经开启事务，就以事务方式执行，如果没有开启事务，就以非事务方式执行
func (s *DBSession) QueryRow(query string, args ...interface{}) *sql.Row {
	if s.tx != nil {
		return s.tx.QueryRow(query, args...)
	}
	return s.DB.QueryRow(query, args...)
}

// Query 查询数据，如果已经开启事务，就以事务方式执行，如果没有开启事务，就以非事务方式执行
func (s *DBSession) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.Query(query, args...)
	}
	return s.DB.Query(query, args...)
}

// Prepare 预执行，如果已经开启事务，就以事务方式执行，如果没有开启事务，就以非事务方式执行
func (s *DBSession) Prepare(query string) (*sql.Stmt, error) {
	if s.tx != nil {
		return s.tx.Prepare(query)
	}
	return s.DB.Prepare(query)
}
