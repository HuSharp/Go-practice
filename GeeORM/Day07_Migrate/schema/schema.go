package schema

import (
	"geeorm/dialect"
	"go/ast"
	"reflect"
)

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

func Parse(dest interface{}, d dialect.Dialect) *Schema {
	// reflect.Indirect() 获取指针指向的实例
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema {
		Model:    dest,
		Name:     modelType.Name(),	// 结构体的名称作为表名
		filedMap: make(map[string]*Field),
	}

	// NumField() 获取实例的字段的个数，然后通过下标获取到特定字段 p := modelType.Field(i)
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