package geeorm

import (
	"errors"
	"geeorm/session"
	_ "github.com/mattn/go-sqlite3"
	"reflect"
	"testing"
)

/*
t.Helper() 的作用是标记一个函数为测试辅助函数，
这样的话，该函数将不会在测试日志输出文件名和行号信息时出现。
当 go testing 系统在查找调用栈帧的时候，
通过 Helper 标记过的函数将被略过，
因此这有助于找到更确切的调用者及其相关信息。
这个函数的用途在于削减日志输出中（尤其是在打印调用栈帧信息时）的杂音。
 */
func OpenDB(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("sqlite3", "gee.db")
	if err != nil {
		t.Fatal("failed to connect!", err)
	}
	return engine
}

func TestNewEngine(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
}

type User struct {
	Name 	string	`geeorm:"PRIMARY KEY"`
	Age 	int
}

func TestEngine_Transaction(t *testing.T) {
	t.Run("rollback", func(t *testing.T) {
		transactionRollback(t)
	})
	t.Run("commit", func(t *testing.T) {
		transactionCommit(t)
	})
}

func transactionRollback(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{Name: "Tom", Age: 18})
		return nil, errors.New("Error")// 故意返回 导致回滚
	})
	if err == nil || s.HasTable() {
		t.Fatal("[transactionRollback] failed to rollback")
	}
}

func transactionCommit(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{Name: "Tom", Age: 18})
		return
	})
	user := &User{}
	err = s.First(user)
	if err != nil || user.Name != "Tom" {
		t.Fatal("[transactionCommit] fail to commit")
	}
}

func TestEngine_Migrate(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text PRIMARY KEY, XXX integer);").Exec()
	_, _ = s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	engine.Migrate(&User{})

	rows, _ := s.Raw("SELECT * FROM User").QueryRows()
	columns, _ := rows.Columns()
	if !reflect.DeepEqual(columns, []string{"Name", "Age"}) {
		t.Fatal("Failed to migrate table User, got columns", columns)
	}
}