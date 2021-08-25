





### Dialect

Dialect 实现了一些特定的 SQL 语句的转换

SQL 语句中的类型和 Go 语言中的类型是不同的，例如Go 语言中的 `int`、`int8`、`int16` 等类型均对应 SQLite 中的 `integer` 类型。因此实现 ORM 映射的第一步，需要思考如何将 Go 语言的类型映射为数据库中的类型。

同时，不同数据库支持的数据类型也是有差异的，即使功能相同，在 SQL 语句的表达上也可能有差异。ORM 框架往往需要兼容多种数据库，因此我们需要将差异的这一部分提取出来，每一种数据库分别实现，实现最大程度的复用和解耦。这部分代码称之为 `dialect`。

`Dialect` 接口包含 2 个方法：

- `DataTypeOf` 用于将 Go 语言的类型转换为该数据库的数据类型。
- `TableExistSQL` 返回某个表是否存在的 SQL 语句，参数是表名(table)。



### Schema

Schema 利用反射(reflect)完成结构体和数据库表结构的映射，包括表名、字段名、字段类型、字段 tag 等。

```go
type Field struct {
	Name string
	Type string
	Tag  string
}
```

- 表名(table name) —— 结构体名(struct name)
- 字段名和字段类型 —— 成员变量和类型。
- 额外的约束条件(例如非空、主键等) —— 成员变量的Tag（Go 语言通过 Tag 实现，Java、Python 等语言通过注解实现）





### Session

Session 的核心功能是与数据库进行交互.

 session struct 是会在会话中复用的，如果使用 string 类型，.string 是只读不可变的，每次修改其实都要重新申请一个内存空间，都是一个新的 string，而 string.Builder 底层使用 []byte 实现。

```
type Session struct {
   db       *sql.DB          // 使用 sql.Open() 方法连接数据库成功之后返回的指针。
   dialect    dialect.Dialect
   refTable   *schema.Schema
   clause    clause.Clause
   // 用来拼接 SQL 语句和 SQL 语句中占位符的对应值
   sql       strings.Builder
   sqlVars    []interface{}
}
```

- 第一个是 `db *sql.DB`，即使用 `sql.Open()` 方法连接数据库成功之后返回的指针。
-  sql 和 sqlVars 成员变量用来拼接 SQL 语句和 SQL 语句中占位符的对应值。用户调用 `Raw()` 方法即可改变这两个变量的值。

Raw 函数可改变 sql 和 sqlvars 的值

```go
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}
```





- 1）为适配不同的数据库，映射数据类型和特定的 SQL 语句，创建 Dialect 层屏蔽数据库差异。
- 2）设计 Schema，利用反射(reflect)完成结构体和数据库表结构的映射，包括表名、字段名、字段类型、字段 tag 等。
- 3）构造创建(create)、删除(drop)、存在性(table exists) 的 SQL 语句完成数据库表的基本操作。





### Engine

Session 负责与数据库的交互，那交互前的准备工作（比如连接/测试数据库），交互后的收尾工作（关闭连接）等就交给 Engine 来负责了。Engine 是 GeeORM 与用户交互的入口。

```go
type Engine struct {
	db 		*sql.DB
	dialect dialect.Dialect
}
```

`NewEngine` 主要做了两件事。

- 连接数据库，返回 `*sql.DB`。
- 调用 `db.Ping()`，检查数据库是否能够正常连接。
- 确保 dialect 是存在的





### Clause

GeeORM 需要涉及一些较为复杂的操作，例如查询操作。查询语句一般由很多个子句(clause) 构成。SELECT 语句的构成通常是这样的：

```
SELECT col1, col2, ...
    FROM table_name
    WHERE [ conditions ]
    GROUP BY col1
    HAVING [ conditions ]
```

也就是说，如果想一次构造出完整的 SQL 语句是比较困难的，因此我们将构造 SQL 语句这一部分独立出来，放在子package clause 中实现。



结构体 `Clause` 拼接各个独立的子句。

```go
type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}
```

- `Set` 方法根据 `Type` 调用对应的 generator，生成该子句对应的 SQL 语句。
- `Build` 方法根据传入的 `Type` 的顺序，构造出最终的 SQL 语句。

```go
func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = vars
}
```

其中的 `generators[name](vars...)`方法表示

```go
// 实现各个子句的生成规则
type generator func(values ...interface{}) (string, []interface{})

var generators map[Type]generator

func init()  {
	generators = make(map[Type]generator)
	generators[INSERT] = _select
}
```

比如

SELECT 语句的调用：

```go
clause.Set(SELECT, "User", []string{"*"})
```

实现 generate 映射如下：

```go
func _select(values ...interface{}) (string, []interface{}) {
	// SELECT $fields FROM $tableName
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}
```

因此输出应为：

SELECT * FROM User





### 增删改查语句的实现

后续所有构造 SQL 语句的方式都将与 `Insert` 中构造 SQL 语句的方式一致。分两步：

- 1）多次调用 `clause.Set()` 构造好每一个子句。
- 2）调用一次 `clause.Build()` 按照传入的顺序构造出最终的 SQL 语句。

```go
func (s *Session) Insert(values ...interface{}) (int64, error) {
   recordValues := make([]interface{}, 0)
   for _, value := range values {
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

   return result.RowsAffected()
}
```

> Insert 需要将已经存在的对象的每一个字段的值平铺开来。

