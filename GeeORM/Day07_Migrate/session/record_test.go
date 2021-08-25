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

func TestSession_Update(t *testing.T) {
	session := testRecordInit(t)
	affected, _ := session.Where("Name = ?", "Tom").Update("Age", 30)
	user := &User{}
	_ = session.OrderBy("Age DESC").First(user)

	if affected != 1 || user.Age != 30 {
		t.Fatalf("[TestSession_Limit] failed to update. u:%v", user)
	}
	t.Log("----------------------")
	t.Logf("TestSession_Limit success! user:%v", user)
}

func TestSession_Limit(t *testing.T) {
	session := testRecordInit(t)
	var users []User
	err := session.Limit(1).Find(&users)
	if err != nil {
		t.Fatalf("[TestSession_Limit] failed to limit. users:%v", users)
	}
	t.Log("----------------------")
	t.Logf("TestSession_Limit success! user:%v", users)
}

func TestSession_DeleteAndCount(t *testing.T) {
	session := testRecordInit(t)
	affected, _ := session.Where("Name = ?", "Tom").Delete()
	count, _ := session.Count()

	if affected != 1 || count != 1 {
		t.Fatalf("[TestSession_DeleteAndCount] failed to delete or count. affected:%d, count:%d", affected, count)
	}
	t.Logf("[TestSession_DeleteAndCount] success to delete or count. affected:%d, count:%d", affected, count)
}