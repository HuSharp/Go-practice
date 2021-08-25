package session

import (
	"database/sql"
	"geeorm/dialect"
	"geeorm/log"
	"geeorm/schema"
	"strings"
)

/*
	session struct 是会在会话中复用的，如果使用 string 类型，
	string 是只读不可变的，每次修改其实都要重新申请一个内存空间，都是一个新的 string，
	而 string.Builder 底层使用 []byte 实现。
 */
type Session struct {
	db			*sql.DB			// 使用 sql.Open() 方法连接数据库成功之后返回的指针。
	dialect 	dialect.Dialect
	refTable	*schema.Schema
	// 用来拼接 SQL 语句和 SQL 语句中占位符的对应值
	sql		strings.Builder
	sqlVars	[]interface{}

}

func New(db *sql.DB, dialect dialect.Dialect) *Session {

	return &Session{
		db: db,
		dialect: dialect,
	}
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
}

func (s *Session) DB() *sql.DB {
	return s.db
}

func (s *Session) Raw(sql string, vals ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, vals...)
	return s
}

// 封装 Exec()、Query() 和 QueryRow() 三个原生方法。
/*
封装原因：
	1. 统一打印日志（包括 执行的SQL 语句和错误日志）
	2. 执行完成后，清空 (s *Session).sql 和 (s *Session).sqlVars 两个变量。
这样 Session 可以复用，开启一次会话，可以执行多次 SQL。
*/


func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	log.Infof("[Exec] sql statement:%v, sqlVars: %v", s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Errorf("[Exec] err: %v" , err)
	}
	return
}

func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Infof("[QueryRow] sql statement:%v, sqlVars: %v", s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	log.Infof("[QueryRows] sql statement:%v, sqlVars: %v", s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Errorf("[QueryRows] err: %v" , err)
	}
	return
}


