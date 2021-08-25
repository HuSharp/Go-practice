package clause

import (
	"fmt"
	"geeorm/log"
	"strings"
)

// 实现各个子句的生成规则
type generator func(values ...interface{}) (string, []interface{})

var generators map[Type]generator

func init()  {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
	generators[UPDATE] = _update
	generators[DELETE] = _delete
	generators[COUNT] = _count
}

func GetBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ", ")
}

func _insert(values ...interface{}) (string, []interface{}) {
	// INSERT INTO $tableName ($fields)
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	insertStr := fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields)
	log.Infof("[_insert] sql: %v", insertStr)
	return insertStr, []interface{}{}
}

func _values(values ...interface{}) (string, []interface{}) {
	// VALUES ($v1), ($v2), ...
	var sql strings.Builder
	var vars []interface{}
	var bindStr string
	sql.WriteString("VALUES ")
	for i, value := range values {
		v := value.([]interface{})
		if bindStr == "" {
			bindStr = GetBindVars(len(v))
		}
		sql.WriteString(fmt.Sprintf("(%v)", bindStr))
		// 继续写入
		if i+1 != len(values) {
			sql.WriteString(", ")
		}
		vars = append(vars, v...)
	}
	log.Infof("[_values] sql: %v", sql.String())
	return sql.String(), vars
}

func _select(values ...interface{}) (string, []interface{}) {
	// SELECT $fields FROM $tableName
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}

func _limit(values ...interface{}) (string, []interface{}) {
	// LIMIT $num
	return "LIMIT ?", values
}

func _where(values ...interface{}) (string, []interface{}) {
	// WHERE $desc
	desc, vars := values[0], values[1:]
	return fmt.Sprintf("WHERE %s", desc), vars

}

func _orderBy(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}

// _update 设计入参是2个，第一个参数是表名(table)，第二个参数是 map 类型，表示待更新的键值对。
func _update(values ...interface{}) (string, []interface{}) {
	setMap := values[1].(map[string]interface{})
	var keys []string
	var vars []interface{}
	for k, v := range setMap {
		keys = append(keys, k+" = ?")
		vars = append(vars, v)
	}
	return fmt.Sprintf("UPDATE %s SET %s", values[0], strings.Join(keys, ", ")), vars
}

// _delete 只有一个入参，即表名。
func _delete(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

// _count 只有一个入参，即表名，并复用了 _select 生成器。
func _count(values ...interface{}) (string, []interface{}) {
	return _select(values[0], []string{"count(*)"})
}

