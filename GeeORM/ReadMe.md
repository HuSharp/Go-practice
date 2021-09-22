---
layout: post
title:  " Go 实现类 ORM 框架 GeeORM"
date:   2021-06-16 23:53:15 +0800
categories:  Go
tags: Go 编程语言
author: Hu#
typora-root-url: ..

---

* content
  {:toc}


# 实现特性

GeeORM 的设计主要参考了 xorm，一些细节上的实现参考了 gorm。GeeORM 的目的主要是了解 ORM 框架设计的原理，具体实现上鲁棒性做得不够，一些复杂的特性，例如 gorm 的关联关系，xorm 的读写分离没有实现。目前支持的特性有：

- 表的创建、删除、迁移。
- 记录的增删查改，查询条件的链式操作。
- 单一主键的设置(primary key)。
- 钩子(在创建/更新/删除/查找之前或之后)
- 事务(transaction)。
- GeeORM 的所有的开发和测试均基于 SQLite。



## 各个模块介绍

- Engine 来负责：交互前的准备工作（比如连接/测试数据库），交互后的收尾工作（关闭连接）等。是 GeeORM 与用户交互的入口。

- Session 的核心功能是与数据库进行交互。

- Dialect 实现了一些特定的 SQL 语句的转换，为适配不同的数据库，映射数据类型和特定的 SQL 语句，屏蔽数据库差异；

- Schema，利用反射(reflect)完成结构体和数据库表结构的映射，包括表名、字段名、字段类型、字段 tag 等。

- Clause  拼接各个独立的子句，通过 Build、Set 来实现

  ```go
  	var clause Clause
  	clause.Set(LIMIT, 3)
  	clause.Set(SELECT, "User", []string{"*"})
  	clause.Set(WHERE, "Name = ?", "Tom")
  	clause.Set(ORDERBY, "Age ASC")
  	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
  ```

- 链式操作

- Hook 思想



## 设计思路

> 对象关系映射（Object Relational Mapping，简称ORM）是通过使用描述对象和数据库之间映射的元数据，将面向对象语言程序中的对象自动持久化到关系数据库中。

那对象和数据库是如何映射的呢？

| 数据库              | 面向对象的编程语言  |
| :------------------ | :------------------ |
| 表(table)           | 类(class/struct)    |
| 记录(record, row)   | 对象 (object)       |
| 字段(field, column) | 对象属性(attribute) |



举一个具体的例子，来理解 ORM。

```
CREATE TABLE `User` (`Name` text, `Age` integer);
INSERT INTO `User` (`Name`, `Age`) VALUES ("Tom", 18);
SELECT * FROM `User`;
```

第一条 SQL 语句，在数据库中创建了表 `User`，并且定义了 2 个字段 `Name` 和 `Age`；第二条 SQL 语句往表中添加了一条记录；最后一条语句返回表中的所有记录。

假如我们使用了 ORM 框架，可以这么写：

```go
type User struct {
    Name string
    Age  int
}

orm.CreateTable(&User{})
orm.Save(&User{"Tom", 18})
var users []User
orm.Find(&users)
```

ORM 框架相当于对象和数据库中间的一个桥梁，借助 ORM 可以避免写繁琐的 SQL 语言，仅仅通过操作具体的对象，就能够完成对关系型数据库的操作。

那如何实现一个 ORM 框架呢？

- `CreateTable` 方法需要从参数 `&User{}` 得到对应的**结构体的名称 User 作为表名，成员变量 Name, Age 作为列名**，同时还需要知道**成员变量对应的类型**。
- `Save` 方法则需要知道**每个成员变量的值**。
- `Find` 方法仅从传入的空切片 `&[]User`，得到对应的结构体名也就是表名 User，并从数据库中取到所有的记录，将其**转换成 User 对象，添加到切片中**。

### 延伸到**所有对象**转换为数据库表和记录

如果这些方法只接受 User 类型的参数，那是很容易实现的。**但是 ORM 框架是通用的，也就是说可以将任意合法的对象转换成数据库中的表和记录**。例如：

```go
type Account struct {
    Username string
    Password string
}

orm.CreateTable(&Account{})
```

这就面临了一个很重要的问题：如何根据任意类型的指针，得到其对应的结构体的信息。这涉及到了 **Go 语言的反射机制(reflect)**，通过反射，可以获取到对象对应的结构体名称，成员变量、方法等信息，例如：

```go
typ := reflect.Indirect(reflect.ValueOf(&Account{})).Type()
fmt.Println(typ.Name()) // Account

for i := 0; i < typ.NumField(); i++ {
    field := typ.Field(i)
    fmt.Println(field.Name) // Username Password
}
```

- `reflect.ValueOf()` 获取指针对应的反射值。
- `reflect.Indirect()` 获取指针指向的对象的反射值。
- `(reflect.Type).Name()` 返回类名(字符串)。
- `(reflect.Type).Field(i)` 获取第 i 个成员变量。







# 具体介绍

## 模块介绍

首先熟悉一下 sqlite3 的使用

```go
func main() {
	db, _ := sql.Open("sqlite3", "gee.db")
	defer func() { _ = db.Close() }()
	_, _ = db.Exec("DROP TABLE IF EXISTS User;")
	_, _ = db.Exec("CREATE TABLE User(Name text);")
	result, err := db.Exec("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam")
	if err == nil {
		affected, _ := result.RowsAffected()
		log.Println(affected)
	}
	row := db.QueryRow("SELECT Name FROM User LIMIT 1")
	var name string
	if err := row.Scan(&name); err == nil {
		log.Println(name)
	}
}
```

- 使用 `sql.Open()` 连接数据库，第一个参数是驱动名称，import 语句 `_ "github.com/mattn/go-sqlite3"` 包导入时会注册 sqlite3 的驱动，第二个参数是数据库的名称，对于 SQLite 来说，也就是文件名，不存在会新建。返回一个 `sql.DB` 实例的指针。
- `Exec()` 用于执行 SQL 语句，如果是查询语句，不会返回相关的记录。所以查询语句通常使用 `Query()` 和 `QueryRow()`，前者可以返回多条记录，后者只返回一条记录。
- `Exec()`、`Query()`、`QueryRow()` 接受1或多个入参，第一个入参是 SQL 语句，后面的入参是 SQL 语句中的占位符 `?` 对应的值，占位符一般用来防 SQL 注入。
- `QueryRow()` 的返回值类型是 `*sql.Row`，`row.Scan()` 接受1或多个指针作为参数，可以获取对应列(column)的值，在这个示例中，只有 `Name` 一列，因此传入字符串指针 `&name` 即可获取到查询的结果。



### log 实现

开发一个框架/库并不容易，详细的日志能够帮助我们快速地定位问题。因此，在写核心代码之前，我们先用几十行代码实现一个简单的 log 库。

这个简易的 log 库具备以下特性：

- 支持日志分级（Info、Error、Disabled 三级）。
- 不同层级日志显示时使用不同的颜色区分。
- 显示打印日志代码对应的文件名和行号。

#### 第一步，创建 2 个日志实例分别用于打印 Info 和 Error 日志。

```go
// 创建 2 个日志实例分别用于打印 Info 和 Error 日志。
// [info ] 颜色为蓝色，[error] 为红色。
var (
	infoLog  = log.New(os.Stdout, "\033[34m[ info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	errorLog = log.New(os.Stdout, "\033[31m[ error]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

// log methods
var (
	Error	= errorLog.Println
	Errorf	= errorLog.Printf
	Info	= infoLog.Println
	Infof	= infoLog.Printf
)
```

- `[info ]` 颜色为蓝色，`[error]` 为红色。
- 使用 `log.Lshortfile` 支持显示文件名和代码行号。
- 暴露 `Error`，`Errorf`，`Info`，`Infof` 4个方法。

#### 第二步，支持设置日志的层级(InfoLevel, ErrorLevel, Disabled)。

通过控制 Output，来控制日志是否打印。

```go
// 支持设置日志的层级(InfoLevel, ErrorLevel, Disabled)
const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

// 通过控制 Output，来控制日志是否打印
func SetLevel(level int)  {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}
	if ErrorLevel < level {
		// Discard是一个io.Writer接口，对它的所有Write调用都会无实际操作的成功返回。
		errorLog.SetOutput(ioutil.Discard)
	}
	// 如果设置为 ErrorLevel，infoLog 的输出会被定向到 ioutil.Discard，即不打印该日志。
	if InfoLevel < level {
		infoLog.SetOutput(ioutil.Discard)
	}
}
```





### **Session**

Session 的核心功能是与数据库进行交互。

session struct 是会在会话中复用的，如果使用 string 类型，string 是只读不可变的，每次修改其实都要重新申请一个内存空间，都是一个新的 string，而 string.Builder 底层使用 []byte 实现。

```go
type Session struct {
	db			*sql.DB			// 使用 sql.Open() 方法连接数据库成功之后返回的指针。
	dialect 	dialect.Dialect
	refTable	*schema.Schema
	clause		clause.Clause
	tx			*sql.Tx		// 当 tx 不为空时，则使用 tx 执行 SQL 语句，为空时，跟之前一样采用 db 执行
	// 用来拼接 SQL 语句和 SQL 语句中占位符的对应值
	sql		strings.Builder
	sqlVars	[]interface{}
}
```

- 第一个是 `db *sql.DB`，即使用 `sql.Open()` 方法连接数据库成功之后返回的指针。
- sql 和 sqlVars 成员变量用来拼接 SQL 语句和 SQL 语句中占位符的对应值。用户调用 `Raw()` 方法即可改变这两个变量的值。

Raw 函数可改变 sql 和 sqlvars 的值

```go
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}
```



#### 封装原生方法

封装 Exec()、Query() 和 QueryRow() 三个原生方法。

- 封装有 2 个目的，一是统一打印日志（包括 执行的SQL 语句和错误日志）。
- 二是执行完成后，清空 `(s *Session).sql` 和 `(s *Session).sqlVars` 两个变量。这样 Session 可以复用，开启一次会话，可以执行多次 SQL。

```go
// 封装 Exec()、Query() 和 QueryRow() 三个原生方法。
/*
封装原因：
	1. 统一打印日志（包括 执行的SQL 语句和错误日志）
	2. 执行完成后，清空 (s *Session).sql 和 (s *Session).sqlVars 两个变量。
这样 Session 可以复用，开启一次会话，可以执行多次 SQL。
*/
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	log.Infof("[Exec] sql statement:%v, sqlVars: %v", s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Errorf("[Exec] err: %v" , err)
	}
	return
}

func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Infof("[QueryRow] sql statement:%v, sqlVars: %v", s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	log.Infof("[QueryRows] sql statement:%v, sqlVars: %v", s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Errorf("[QueryRows] err: %v" , err)
	}
	return
}
```



### **Engine**

Session 负责与数据库的交互，那交互前的准备工作（比如连接/测试数据库），交互后的收尾工作（关闭连接）等就交给 Engine 来负责了。Engine 是 GeeORM 与用户交互的入口。

```go
// Engine 是 GeeORM 与用户交互的入口
type Engine struct {
	db 		*sql.DB
	dialect dialect.Dialect
}
```

`NewEngine` 主要做了两件事。

- 连接数据库，返回 `*sql.DB`。
- 调用 `db.Ping()`，检查数据库是否能够正常连接。

```
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
```

**并提供  NewSession 方法进行创建 Session，进而与数据库进行交互。**

```go
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db)
}
```





- **1）为适配不同的数据库，映射数据类型和特定的 SQL 语句，创建 Dialect 层屏蔽数据库差异。**
- **2）设计 Schema，利用反射(reflect)完成结构体和数据库表结构的映射，包括表名、字段名、字段类型、字段 tag 等。**
- **3）构造创建(create)、删除(drop)、存在性(table exists) 的 SQL 语句完成数据库表的基本操作。**





### **Dialect**

> Dialect 实现了一些**特定的** SQL 语句的转换

SQL 语句中的类型和 Go 语言中的类型是不同的，例如Go 语言中的 `int`、`int8`、`int16` 等类型均对应 SQLite 中的 `integer` 类型。因此实现 ORM 映射的第一步，需要思考如何将 Go 语言的类型映射为数据库中的类型。

同时，不同数据库支持的数据类型也是有差异的，即使功能相同，在 SQL 语句的表达上也可能有差异。ORM 框架往往需要兼容多种数据库，因此我们需要将差异的这一部分提取出来，每一种数据库分别实现，实现最大程度的复用和解耦。这部分代码称之为 `dialect`。

```go
var dialectMap = map[string]Dialect{}

type Dialect interface {
	DataTypeOf(typ reflect.Value) string // 转换 go 类型到该数据库的类型
	TableExistSQL(tableName string) (string, []interface{})	// 返回某个表是否存在的语句
}

func RegisterDialect(name string, dialect Dialect)  {
	dialectMap[name] = dialect
}

func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectMap[name]
	return
}
```

`Dialect` 接口包含 2 个方法：

- `DataTypeOf` 用于将 Go 语言的类型转换为该数据库的数据类型。
- `TableExistSQL` 返回某个表是否存在的 SQL 语句，参数是表名(table)。

```go
type sqlite3 struct {}

func init() {
	RegisterDialect("sqlite3", &sqlite3{})
}

func (s *sqlite3) DataTypeOf(typ reflect.Value) string  {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		return "integer"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.String:
		return "text"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind().String()))
}

func (s *sqlite3) TableExistSQL(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT name FROM sqlite_master WHERE type='table' and name = ?", args
}
```

`init()` 函数，包在第一次加载时，会将 sqlite3 的 dialect 自动注册到全局。



### **Schema**

Schema 利用反射(reflect)完成**结构体和数据库表结构的映射**，包括表名、字段名、字段类型、字段 tag 等。

```go
type Field struct {
	Name string
	Type string
	Tag  string
}

// Schema 主要包含被映射的对象 Model、表名 Name 和字段 Fields。
type Schema struct {
	Model      interface{}
	Name       string
	Fields     []*Field
	FieldNames []string		// FieldNames 包含所有的字段名(列名)
	filedMap   map[string]*Field	// fieldMap 记录字段名和 Field 的映射关系
}

func (schema *Schema) GetField(name string) *Field {
	return schema.filedMap[name]
}
```

- 表名(table name) —— 结构体名(struct name)
- 字段名和字段类型 —— 成员变量和类型。
- 额外的约束条件(例如非空、主键等) —— 成员变量的Tag（Go 语言通过 Tag 实现，Java、Python 等语言通过注解实现）

#### Parse 函数

```go
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),	// 结构体的名称作为表名
		filedMap: make(map[string]*Field),
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		//  IsExported 报告名称是否为导出的 Go 符号（即，它是否以大写字母开头）。
		//  Anonymous 是否是匿名字段
		if !field.Anonymous && ast.IsExported(field.Name) {
			goField := &Field{
				Name: field.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(field.Type))),
			}
			if v, ok := field.Tag.Lookup("geeorm"); ok {
				goField.Tag = v
			}
			schema.Fields = append(schema.Fields, goField)
			schema.FieldNames = append(schema.FieldNames, field.Name)
			schema.filedMap[field.Name] = goField
		}
	}
	return schema
}
```

值得注意的是，反射是很耗时的，因此对于 Parse 解析出的 table 将其放到 Session 的成员变量中，只要传入的结构体名称不发生变化，则不会更新 refTable 的值

```go
Model 一般这么被调用：
	NewSession().Model(&Account{})
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
```



#### 表相关函数

除了建立映射之外，在表层次提供 create、drop、isExist 函数

```go
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

```



### Clause

GeeORM 需要涉及一些较为复杂的操作，例如查询操作。查询语句一般由很多个子句(clause) 构成。SELECT 语句的构成通常是这样的：

```sql
SELECT col1, col2, ...
    FROM table_name
    WHERE [ conditions ]
    GROUP BY col1
    HAVING [ conditions ]
```

也就是说，如果想一次构造出完整的 SQL 语句是比较困难的，因此我们将构造 SQL 语句这一部分独立出来，放在子package clause 中实现。

需要实现的结果如下，使用到 Set 和 Build 函数

```go
	var clause Clause
	clause.Set(LIMIT, 3)
	clause.Set(SELECT, "User", []string{"*"})
	clause.Set(WHERE, "Name = ?", "Tom")
	clause.Set(ORDERBY, "Age ASC")
	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
```

结构体 `Clause` 拼接各个独立的子句。

> 当然 Clause 变量也是放在 Session 中。

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
    generators[SELECT] = _select
    ...
    generators[DELETE] = _delete
	generators[COUNT] = _count
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

> SELECT * FROM User





### 增删改查语句的实现

#### Insert

INSERT 对应的 SQL 语句一般是这样的：

```
INSERT INTO table_name(col1, col2, col3, ...) VALUES
    (A1, A2, A3, ...),
    (B1, B2, B3, ...),
    ...
```

在 ORM 框架中期望 Insert 的调用方式如下：

```
s := geeorm.NewEngine("sqlite3", "gee.db").NewSession()
u1 := &User{Name: "Tom", Age: 18}
u2 := &User{Name: "Sam", Age: 25}
s.Insert(u1, u2, ...)
```

也就是说，我们还需要一个步骤，根据数据库中列的顺序，从对象中找到对应的值，按顺序平铺。即 `u1`、`u2` 转换为 `("Tom", 18), ("Same", 25)` 这样的格式。

因此在实现 Insert 功能之前，还需要给 `Schema` 新增一个函数 `RecordValues` 完成上述的转换。

```go
// RecordValues 根据数据库中列的顺序，从对象中找到对应的值，按顺序平铺。
// 即 u1 := &User{Name: "Tom", Age: 18} 转换为 ("Tom", 18) 这样的格式。
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
```

需要实现的 Insert 被调用时语句如下

> ```go
> s := geeorm.NewEngine("sqlite3", "gee.db").NewSession()
> u1 := &User{Name: "Tom", Age: 18}
> u2 := &User{Name: "Sam", Age: 25}
> s.Insert(u1, u2, ...)
> ```

后续所有构造 SQL 语句的方式都将与 `Insert` 中构造 SQL 语句的方式一致。分两步：

- 1）多次调用 `clause.Set()` 构造好每一个子句。
- 2）调用一次 `clause.Build()` 按照传入的顺序构造出最终的 SQL 语句。

```go
func (s *Session) Insert(values ...interface{}) (int64, error) {
   recordValues := make([]interface{}, 0)
   for _, value := range values {
		s.CallMethod(BeforeInsert, value)	// 这是 Hook 相关
		table := s.Model(value).RefTable()
       // table.Name 只需要设置一次，这是为了书写方便
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



#### Select

期望的调用方式是这样的：传入一个切片指针，查询的结果保存在切片中。

```go
s := geeorm.NewEngine("sqlite3", "gee.db").NewSession()
var users []User
s.Find(&users);
```

Find 功能的难点和 Insert 恰好反了过来。Insert 需要将已经存在的对象的每一个字段的值平铺开来，而 Find 则是需要根据平铺开的字段的值构造出对象。同样，也需要用到反射(reflect)。

```go
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
```

Find 的代码实现比较复杂，主要分为以下几步：

- 1）`destSlice.Type().Elem()` 获取切片的单个元素的类型 `destType`，使用 `reflect.New()` 方法创建一个 `destType` 的实例，作为 `Model()` 的入参，映射出表结构 `RefTable()`。
- 2）根据表结构，使用 clause 构造出 SELECT 语句，查询到所有符合条件的记录 `rows`。
- 3）遍历每一行记录，利用反射创建 `destType` 的实例 `dest`，将 `dest` 的所有字段平铺开，构造切片 `values`。
- 4）调用 `rows.Scan()` 将该行记录每一列的值依次赋值给 values 中的每一个字段。
- 5）将 `dest` 添加到切片 `destSlice` 中。循环直到所有的记录都添加到切片 `destSlice` 中。



#### Update

Update 方法比较特别的一点在于，Update 接受 2 种入参，平铺开来的键值对和 map 类型的键值对。

因为 generator 接受的参数是 map 类型的键值对，因此 Update 方法会动态地判断传入参数的类型，如果不是 map 类型，**则会自动转换**。

> 自动转换：可以理解为"a","2","b","2"  : "a=2,b=2"
>
> `_update` 设计入参是2个，第一个参数是表名(table)，第二个参数是 map 类型，表示待更新的键值对。

```go
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
```



#### Delete

> `_delete` 只有一个入参，即表名。
>
> ```go
> func _delete(values ...interface{}) (string, []interface{}) {
> 	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
> }
> ```

```go
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
```



#### Count

> `_count` 只有一个入参，即表名，并复用了 `_select` 生成器。
>
> ```go
> func _count(values ...interface{}) (string, []interface{}) {
> 	return _select(values[0], []string{"count(*)"})
> }
> ```

```go
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
```



## 链式操作

链式调用是一种简化代码的编程方式，能够使代码更简洁、易读。链式调用的原理也非常简单，某个对象调用某个方法后，将该对象的引用/指针返回，即可以继续调用该对象的其他方法。通常来说，当某个对象需要一次调用多个方法来设置其属性时，就非常适合改造为链式调用了。

SQL 语句的构造过程就非常符合这个条件。SQL 语句由多个子句构成，典型的例如 SELECT 语句，往往需要设置查询条件（WHERE）、限制返回行数（LIMIT）等。理想的调用方式应该是这样的：

```go
s := geeorm.NewEngine("sqlite3", "gee.db").NewSession()
var users []User
s.Where("Age > 18").Limit(3).Find(&users)
```

可以看出，`WHERE`、`LIMIT`、`ORDER BY` 等查询条件语句非常适合链式调用。

```go
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
```

#### First

很多时候，我们期望 SQL 语句只返回一条记录，比如根据某个童鞋的学号查询他的信息，返回结果有且只有一条。结合链式调用，我们可以非常容易地实现 First 方法。

```go
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
```





## Hook

Hook，翻译为钩子，其主要思想是提前在可能增加功能的地方埋好(预设)一个钩子，当我们需要重新修改或者增加这个地方的逻辑的时候，把扩展的类或者方法挂载到这个点即可。钩子的应用非常广泛，例如 Github 支持的 travis 持续集成服务，当有 `git push` 事件发生时，会触发 travis 拉取新的代码进行构建。IDE 中钩子也非常常见，比如，当按下 `Ctrl + s` 后，自动格式化代码。再比如前端常用的 `hot reload` 机制，前端代码发生变更时，自动编译打包，通知浏览器自动刷新页面，实现所写即所得。

钩子机制设计的好坏，取决于扩展点选择的是否合适。例如对于持续集成来说，代码如果不发生变更，反复构建是没有意义的，因此钩子应设计在代码可能发生变更的地方，比如 MR、PR 合并前后。

那对于 ORM 框架来说，合适的扩展点在哪里呢？很显然，记录的增删查改前后都是非常合适的。

比如，我们设计一个 `Account` 类，`Account` 包含有一个隐私字段 `Password`，那么每次查询后都需要做脱敏处理，才能继续使用。如果提供了 `AfterQuery` 的钩子，查询后，自动地将 `Password` 字段的值脱敏，是不是能省去很多冗余的代码呢？

### 实现钩子

#### 采用反射实现

GeeORM 的钩子与结构体绑定，即每个结构体需要实现各自的钩子。

```go
// Hooks constants
const (
	BeforeQuery  = "BeforeQuery"
	AfterQuery   = "AfterQuery"
	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"
	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"
)

// CallMethod 
// Hook 通过反射来实现
// s.RefTable().Model 或 value 即当前会话正在操作的对象，使用 MethodByName 方法反射得到该对象的方法。
func (s *Session) CallMethod(method string, value interface{}) {
	fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)
	if value != nil {
		fm = reflect.ValueOf(value).MethodByName(method)
	}
	param := []reflect.Value{reflect.ValueOf(s)}
	if fm.IsValid() {
		if v := fm.Call(param); len(v) > 0 {
			if err, ok := v[0].Interface().(error); ok {
				log.Errorf("[CallMethod] failed! err:%v", err)
			}
		}
	}
	return
}
```

- 钩子机制同样是通过反射来实现的，`s.RefTable().Model` 或 `value` 即当前会话正在操作的对象，使用 `MethodByName` 方法反射得到该对象的方法。
- 将 `s *Session` 作为入参调用。每一个钩子的入参类型均是 `*Session`。

接下来，将 `CallMethod()` 方法在 Find、Insert、Update、Delete 方法内部调用即可。例如，`Find` 方法修改为：

```go
// Find gets all eligible records
func (s *Session) Find(values interface{}) error {
	s.CallMethod(BeforeQuery, nil)
    // ...
    for rows.Next() {
        dest := reflect.New(destType).Elem()
        // ...
        s.CallMethod(AfterQuery, dest.Addr().Interface())
        // ...
	}
	return rows.Close()
}
```

- `AfterQuery` 钩子可以操作每一行记录。

#### 举个栗子

```go
type Account struct {
	ID			int `geeorm:"PRIMARY KEY"`
	Password	string
}

func (account *Account) BeforeInsert(s *Session) error {
	log.Info("before insert", account)
	account.ID += 1000
	return nil
}

func (account *Account) AfterQuery(s *Session) error {
	log.Info("after find", account)
	account.Password = "******"
	return nil
}

func TestSession_CallMethod(t *testing.T) {
	TestDB, _ = sql.Open("sqlite3", "../gee.db")
	session := NewSession().Model(&Account{})
	_ = session.DropTable()
	_ = session.CreateTable()
	session.Insert(&Account{
		ID:       1,
		Password: "123456",},
		&Account{
			ID:       2,
			Password: "324354",
		})
	account := &Account{}
	err := session.First(account)
	if err != nil || account.ID != 1001 || account.Password != "******" {
		t.Fatal("Failed to call hooks after query, got", account)
	}
	t.Logf("Success to call hooks after query, got %v", account)
}
```

在这个测试用例中，测试了 `BeforeInsert` 和 `AfterQuery` 2 个钩子。

- `BeforeInsert` 将 account.ID 的值增加 1000
- `AfterQuery` 将密码脱敏，显示为 6 个 `*`。





#### 还可以采用 interface 实现

`ITableName` 自定义表名的，如果实现了该接口，就使用 `ITableName` 返回的字符串作为表名。

```go
type IBeforeQuery interface {
      BeforeQuery(s *Session) error
}

type IAfterQuery interface {
      AfterQuery(s *Session) error
}
.....
等等

//然后修改CallMethod
func (s *Session) CallMethod(method string, value interface{}) {
	 ...
     if i, ok := dest.(IBeforQuery); ok == true {
        i. BeforeQuery(s) 
     }
     ...
	return
}
```





## 支持事务

### 背景说明

事务解释可以看[博主这篇 Blog](http://husharp.today/2021/04/10/MySQL-Transaction-isolation/)

SQLite 中创建一个事务的原生 SQL 长什么样子呢？

```sqlite
sqlite> BEGIN;
sqlite> DELETE FROM User WHERE Age > 25;
sqlite> INSERT INTO User VALUES ("Tom", 25), ("Jack", 18);
sqlite> COMMIT;
```

`BEGIN` 开启事务，`COMMIT` 提交事务，`ROLLBACK` 回滚事务。任何一个事务，均以 `BEGIN` 开始，`COMMIT` 或 `ROLLBACK` 结束。

Go 语言标准库 database/sql 提供了支持事务的接口。用一个简单的例子，看一看 Go 语言标准是如何支持事务的。

```go
package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func main() {
	db, _ := sql.Open("sqlite3", "gee.db")
	defer func() { _ = db.Close() }()
	_, _ = db.Exec("CREATE TABLE IF NOT EXISTS User(`Name` text);")

	tx, _ := db.Begin()
	_, err1 := tx.Exec("INSERT INTO User(`Name`) VALUES (?)", "Tom")
	_, err2 := tx.Exec("INSERT INTO User(`Name`) VALUES (?)", "Jack")
	if err1 != nil || err2 != nil {
		_ = tx.Rollback()
		log.Println("Rollback", err1, err2)
	} else {
		_ = tx.Commit()
		log.Println("Commit")
	}
}
```

Go 语言中实现事务和 SQL 原生语句其实是非常接近的。调用 `db.Begin()` 得到 `*sql.Tx` 对象，使用 `tx.Exec()` 执行一系列操作，如果发生错误，通过 `tx.Rollback()` 回滚，如果没有发生错误，则通过 `tx.Commit()` 提交。



###  GeeORM 支持事务

Transaction 的实现参考了 [stackoverflow](https://stackoverflow.com/questions/16184238/database-sql-tx-detecting-commit-or-rollback)

GeeORM 之前的操作均是执行完即自动提交的，每个操作是相互独立的。之前直接使用 `sql.DB` 对象执行 SQL 语句，如果要支持事务，需要更改为 `sql.Tx` 执行。在 Session 结构体中新增成员变量 `tx *sql.Tx`，当 `tx` 不为空时，则使用 `tx` 执行 SQL 语句，否则使用 `db` 执行 SQL 语句。这样既兼容了原有的执行方式，又提供了对事务的支持。

```go
type Session struct {
	db			*sql.DB			// 使用 sql.Open() 方法连接数据库成功之后返回的指针。
	dialect 	dialect.Dialect
	refTable	*schema.Schema
	clause		clause.Clause
	tx			*sql.Tx		// 当 tx 不为空时，则使用 tx 执行 SQL 语句，为空时，跟之前一样采用 db 执行
	// 用来拼接 SQL 语句和 SQL 语句中占位符的对应值
	sql		strings.Builder
	sqlVars	[]interface{}
}

/*
CommonDB is a minimal function set of db
*/
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// DB returns tx if a tx begins. otherwise return *sql.DB
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}
```



#### 封装事务接口

封装事务相关的 Begin、Commit 和 Rollback 三个接口，统一打印日志。

```go
/*
	封装事务相关，统一打印日志
 */
func (s *Session) Begin() (err error) {
	log.Info("[Begin] transaction begin")
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error("[Begin] transaction begin err:", err)
		return
	}
	return
}

func (s *Session) Commit() (err error) {
	log.Info("[Commit] transaction commit")
	if err = s.tx.Commit(); err != nil {
		log.Error("[Commit] transaction commit err:", err)
		return
	}
	return
}

func (s *Session) Rollback() (err error) {
	log.Info("[Rollback] transaction rollback")
	if err := s.tx.Rollback(); err != nil {
		log.Error("[Rollback] transaction rollback err:", err)
	}
	return
}
```

并为用户提供傻瓜式/一键式使用的接口

调用类似下面

```go
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{Name: "Tom", Age: 18})
		return nil, errors.New("Error")// 故意返回 导致回滚
	})
```

用户只需要将所有的操作放到一个回调函数中，作为入参传递给 `engine.Transaction()`，发生任何错误，自动回滚，如果没有错误发生，则提交。

```go
// TxFunc 回调
type TxFunc func(*session.Session) (interface{}, error)

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
```



## 进行数据库迁移

### 1、使用 SQL 语句 Migrate

数据库 Migrate 一直是数据库运维人员最为头痛的问题，如果仅仅是一张表增删字段还比较容易，那如果涉及到外键等复杂的关联关系，数据库的迁移就会变得非常困难。

GeeORM 的 Migrate 操作仅针对最为简单的场景，即支持字段的新增与删除，不支持字段类型变更。

在实现 Migrate 之前，我们先看看如何使用原生的 SQL 语句增删字段。

#### 1.1 新增字段

```
ALTER TABLE table_name ADD COLUMN col_name, col_type;
```

大部分数据支持使用 `ALTER` 关键字新增字段，或者重命名字段。

#### 1.2 删除字段

> 参考 [sqlite delete or add column - stackoverflow](https://stackoverflow.com/questions/8442147/how-to-delete-or-add-column-in-sqlite)

对于 SQLite 来说，删除字段并不像新增字段那么容易，一个比较可行的方法需要执行下列几个步骤：

```
CREATE TABLE new_table AS SELECT col1, col2, ... from old_table
DROP TABLE old_table
ALTER TABLE new_table RENAME TO old_table;
```

- 第一步：从 `old_table` 中挑选需要保留的字段到 `new_table` 中。
- 第二步：删除 `old_table`。
- 第三步：重命名 `new_table` 为 `old_table`。





### 2、GeeORM 实现 Migrate

按照原生的 SQL 命令，利用之前实现的事务，在 `geeorm.go` 中实现 Migrate 方法。

- 第一步：从 old_table 中挑选需要保留的字段到 new_table 中。
- 第二步：删除 old_table。
- 第三步：重命名 new_table 为 old_table。

> 大致实现思路如下：
>
> BEGIN TRANSACTION;
> CREATE TABLE t1_backup(a,b);
> INSERT INTO t1_backup SELECT a,b FROM t1;
> DROP TABLE t1;
> ALTER TABLE t1_backup NAME TO t1;
> COMMIT;

```
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
```







## 收获：

GeeORM 的目的并不是实现一个可以在生产使用的 ORM 框架，而是为了尽可能多地了解 ORM 框架大致的实现原理，例如：

- 在框架中如何屏蔽不同数据库之间的差异；
- 数据库中表结构和编程语言中的对象是如何映射的；
- 如何优雅地模拟查询条件，链式调用是个不错的选择；
- 为什么 ORM 框架通常会提供 hooks 扩展的能力；
- 事务的原理和 ORM 框架如何集成对事务的支持；
- 一些难点问题，例如数据库迁移。
- …