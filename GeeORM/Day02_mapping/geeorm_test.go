package geeorm

import (
	_ "github.com/mattn/go-sqlite3"
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
