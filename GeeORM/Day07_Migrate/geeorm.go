package geeorm

import (
	"database/sql"
	"fmt"
	"geeorm/dialect"
	"geeorm/log"
	"geeorm/session"
	"strings"
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

// difference returns a - b
func difference(a, b []string) (diff []string) {
	mapB := make(map[string]bool)
	for _, v := range b {
		mapB[v] = true
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}

/*
	第一步：从 old_table 中挑选需要保留的字段到 new_table 中。
	第二步：删除 old_table。
	第三步：重命名 new_table 为 old_table。
BEGIN TRANSACTION;
CREATE TABLE t1_backup(a,b);
INSERT INTO t1_backup SELECT a,b FROM t1;
DROP TABLE t1;
ALTER TABLE t1_backup NAME TO t1;
COMMIT;
 */
func (engine *Engine) Migrate(value interface{}) (err error) {
	_, err = engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		if !s.Model(value).HasTable() {
			log.Infof("[Migrate] table %s doesn't exist!", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		table := s.RefTable()
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		columns, _ := rows.Columns()
		addCols := difference(table.FieldNames, columns)
		delCols := difference(columns, table.FieldNames)
		log.Infof("[Migrate] added cols %v, deleted cols %v", addCols, delCols)

		for _, col := range addCols {
			field := table.GetField(col)
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table.Name, field.Name, field.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}

		// 如果没有要删除的，那么就直接加上就行
		if len(delCols) == 0 {
			return
		}

		tmp := "tmp_" + table.Name
		filedStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s FROM %s;", tmp, filedStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return
}