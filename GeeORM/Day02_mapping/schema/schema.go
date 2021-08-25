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
