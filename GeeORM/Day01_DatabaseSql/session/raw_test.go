package session

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

var TestDB *sql.DB

func NewSession() *Session {
	return New(TestDB)
}

func TestSession_Exec(t *testing.T) {
	TestDB, _ = sql.Open("sqlite3", "../gee.db")
	session := NewSession()
	_, _ = session.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = session.Raw("CREATE TABLE User(Name text);").Exec()
	result, _ := session.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	// RowsAffected: 得到本次操作影响的行数
	if count, err := result.RowsAffected(); err != nil || count != 2 {
		t.Fatal("[TestSession_Exec] expect 2 row, but got", count)
	}
}

func TestSession_QueryRow(t *testing.T) {
	TestDB, _ = sql.Open("sqlite3", "../gee.db")
	session := NewSession()
	_, _ = session.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = session.Raw("CREATE TABLE User(Name text);").Exec()
	result := session.Raw("SELECT COUNT(*) FROM User;").QueryRow()
	var count int
	if err := result.Scan(&count); err != nil || count != 0 {
		t.Fatal("[TestSession_Exec] expect 0 row, but got", count)
	}
}

func TestSession_QueryRows(t *testing.T) {
	TestDB, _ = sql.Open("sqlite3", "../gee.db")
	session := NewSession()
	_, _ = session.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = session.Raw("CREATE TABLE User(Name text);").Exec()
	result, _ := session.Raw("SELECT COUNT(*) FROM User;").QueryRows()
	t.Log("[TestSession_QueryRows] query val: ", result)
}