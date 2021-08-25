package session

import (
	"database/sql"
	"testing"
)

var (
	user1 = &User{"Tom", 18}
	user2 = &User{"Sam", 25}
	user3 = &User{"Jack", 25}
)


func testRecordInit(t *testing.T) *Session {
	t.Helper()
	TestDB, _ = sql.Open("sqlite3", "../gee.db")
	session := NewSession().Model(&User{})
	err1 := session.DropTable()
	err2 := session.CreateTable()
	_, err3 := session.Insert(user1, user2)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("failed init! err1:%v, err2:%v, err3:%v", err1, err2, err3)
	}
	t.Log("----------------------")
	t.Log("TestRecordInit success!")
	return session
}

func TestSession_Insert(t *testing.T) {
	session := testRecordInit(t)
	affected, err := session.Insert(user3)
	if err != nil || affected != 1 {
		t.Fatalf("TestSession_Insert failed! err: %v, affected: %d", err, affected)
	}
	t.Log("----------------------")
	t.Logf("TestSession_Insert success! affected: %d", affected)
}

func TestSession_Find(t *testing.T) {
	session := testRecordInit(t)
	var users []User
	if err := session.Find(&users); err != nil || len(users) != 2 {
		t.Fatalf("TestSession_Insert failed! err: %v, affected: %d", err, len(users))
	}
	t.Log("----------------------")
	t.Logf("TestSession_Find success! users:%v, len:%d", users, len(users))
}