package geeorm

import (
	"database/sql"
	"geeorm/log"
	"geeorm/session"
)

// Engine 是 GeeORM 与用户交互的入口
type Engine struct {
	db *sql.DB
}

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Errorf("[Engine.NewEngine] open DB failed! err: %v", err)
		return
	}

	// ping 一下 看是否连接成功
	if err = db.Ping(); err != nil {
		log.Errorf("[NewEngine] connect DB failed! err: %v", err)
		return
	}
	e = &Engine{db: db}
	log.Info("[NewEngine] Connect DB success!")
	return
}

func (engine *Engine) Close()  {
	if err := engine.db.Close(); err != nil {
		log.Errorf("[Engine.Close] close DB failed! err: %v", err)
	}
	log.Info("[Engine.Close] close DB success!")
}

func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db)
}

