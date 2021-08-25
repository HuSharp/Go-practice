package session

import (
	"database/sql"
	"geeorm/log"
	"testing"
)

type Account struct {
	ID			int `geeorm:"PRIMARY KEY"`
	Password	string
}

func (account *Account) BeforeInsert(s *Session) error {
	log.Info("before insert", account)
	account.ID += 1000
	return nil
}

func (account *Account) AfterQuery(s *Session) error {
	log.Info("after find", account)
	account.Password = "******"
	return nil
}

func TestSession_CallMethod(t *testing.T) {
	TestDB, _ = sql.Open("sqlite3", "../gee.db")
	session := NewSession().Model(&Account{})
	_ = session.DropTable()
	_ = session.CreateTable()
	session.Insert(&Account{
		ID:       1,
		Password: "123456",},
		&Account{
			ID:       2,
			Password: "324354",
		})
	account := &Account{}
	err := session.First(account)
	if err != nil || account.ID != 1001 || account.Password != "******" {
		t.Fatal("Failed to call hooks after query, got", account)
	}
	t.Logf("Success to call hooks after query, got %v", account)
}
