package session

import (
	"fmt"
	"geeorm/log"
	"geeorm/schema"
	logs "log"
	"reflect"
	"strings"
)

// Model 方法用于给 refTable 赋值。解析操作是比较耗时的，因此将解析的结果保存在成员变量 refTable 中，
// 即使 Model() 被调用多次，如果传入的结构体名称不发生变化，则不会更新 refTable 的值。
func (s *Session) Model(value interface{}) *Session {
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(value, s.dialect)
		log.Info("[Model] s.refTable: %v", s.refTable)
	}
	return s
}

func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		log.Error("Model is not set")
	}
	return s.refTable
}

func (s *Session) CreateTable() error {
	logs.Print("[CreateTable] start!")
	table := s.RefTable()
	var columns []string
	for _, filed := range table.Fields {
		filedStr := fmt.Sprintf("%s %s %s", filed.Name, filed.Type, filed.Tag)
		logs.Printf("[CreateTable] filedStr:%v", filedStr)
		columns = append(columns, filedStr)
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, desc)).Exec()
	return err
}

func (s *Session) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	return err
}

func (s *Session) HasTable() bool {
	sql, values := s.dialect.TableExistSQL(s.RefTable().Name)
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == s.RefTable().Name
}
