package session

import (
	"errors"
	"geeorm/clause"
	"geeorm/log"
	"reflect"
)

// 用于实现增删改查相关代码

func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		s.CallMethod(BeforeInsert, value)	// 这是 Hook 相关
		table := s.Model(value).RefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))// 平铺数据
	}

	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}

	s.CallMethod(AfterInsert, nil)
	return result.RowsAffected()
}

func (s *Session) Find(values interface{}) error {
	s.CallMethod(BeforeQuery, nil)
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()	// 获取切片的单个元素类型
	// reflect.New(destType) 作为 destType 的实例
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}
	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var values []interface{}
		for _, name := range table.FieldNames {
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		// 调用 rows.Scan() 将该行记录每一列的值依次赋值给 values 中的每一个字段
		if err := rows.Scan(values...); err != nil {
			return err
		}
		s.CallMethod(AfterQuery, dest.Addr().Interface())
		// 将 dest 添加到切片 destSlice 中
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}

/*
Update 方法比较特别的一点在于，Update 接受 2 种入参，平铺开来的键值对和 map 类型的键值对。
因为 generator 接受的参数是 map 类型的键值对，
因此 Update 方法会动态地判断传入参数的类型，如果不是 map 类型，则会自动转换
 */
func (s *Session) Update(kv ...interface{}) (int64, error) {
	s.CallMethod(BeforeUpdate, nil)
	kvMap, ok := kv[0].(map[string]interface{})
	if !ok { // 如果不是 map 类型，则会自动转换，理解为"a","2","b","2"  : "a=2,b=2"
		kvMap = make(map[string]interface{})
		for i := 0; i < len(kv); i+=2 {
			kvMap[kv[i].(string)] = kv[i+1]
		}
	}
	log.Infof("[Update] update kvMap success! kvMap:%v", kvMap)
	s.clause.Set(clause.UPDATE, s.RefTable().Name, kvMap)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		log.Errorf("[Update] update failed! err:%v", err)
		return 0, err
	}
	s.CallMethod(AfterUpdate, nil)
	return result.RowsAffected()
}

func (s *Session) Delete() (int64, error) {
	s.CallMethod(BeforeDelete, nil)
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		log.Errorf("[Update] update failed! err:%v", err)
		return 0, err
	}
	s.CallMethod(AfterDelete, nil)
	return result.RowsAffected()
}

func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var cnt int64
	if err := row.Scan(&cnt); err != nil {
		return 0, err
	}
	return cnt, nil
}

/*
	WHERE、LIMIT、ORDER BY 等查询条件语句非常适合链式调用
 */
func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	str := append(append(vars, desc), args...)
	s.clause.Set(clause.WHERE, str...)
	log.Infof("[Where] sql statement:%v, sqlVars: %v, str:%v", s.sql.String(), s.sqlVars, str)
	return s
}

func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

func (s *Session) First(value interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("[First] NOT FOUND")
	}
	dest.Set(destSlice.Index(0))
	return nil
}