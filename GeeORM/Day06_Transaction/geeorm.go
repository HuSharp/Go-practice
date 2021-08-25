package geeorm

import (
	"database/sql"
	"geeorm/dialect"
	"geeorm/log"
	"geeorm/session"
)

// Engine 是 GeeORM 与用户交互的入口
type Engine struct {
	db 		*sql.DB
	dialect dialect.Dialect
}

// TxFunc 回调
type TxFunc func(*session.Session) (interface{}, error)

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
	// 确保 dialect 是存在的
	dialect, ok := dialect.GetDialect(driver)
	if !ok {
		log.Errorf("[NewEngine] dialect %s Not Found", driver)
		return
	}
	e = &Engine{
		db:      db,
		dialect: dialect,
	}
	log.Infof("[NewEngine] Connect DB success!e: &v", e)
	return
}

func (engine *Engine) Close()  {
	if err := engine.db.Close(); err != nil {
		log.Errorf("[Engine.Close] close DB failed! err: %v", err)
	}
	log.Info("[Engine.Close] close DB success!")
}

func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}

func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	sessionNew := engine.NewSession()
	if err := sessionNew.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = sessionNew.Rollback()
			panic(p)
		} else if err != nil {
			_ = sessionNew.Rollback()
		} else {
			err = sessionNew.Commit()
		}
	}()
	return f(sessionNew)
}

