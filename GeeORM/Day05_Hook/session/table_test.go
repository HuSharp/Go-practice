package session

import (
	"database/sql"
	logs "log"
	"testing"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age int
}

func TestSession_CreateTable(t *testing.T) {
	TestDB, _ = sql.Open("sqlite3", "../gee.db")
	session := NewSession().Model(&User{})
	logs.Printf("[TestSession_CreateTable] session.refTable:%v", session.refTable)
	_ = session.DropTable()
	_ = session.CreateTable()
	if !session.HasTable() {
		t.Fatal("Failed to create table User!")
	}
}

func TestSession_Model(t *testing.T) {
	s := NewSession().Model(&User{})
	table := s.RefTable()
	s.Model(&Session{})
	if table.Name != "User" || s.RefTable().Name != "Session" {
		t.Fatal("Failed to change model")
	}
}