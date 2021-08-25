package dialect

import (
	"reflect"
	"testing"
)

func TestSqlite3_DataTypeOf(t *testing.T) {
	dialect := &sqlite3{}
	cases := []struct{
		Value interface{}
		Type string
	}{
		{"Tom", "text"},
		{123, "integer"},
		{1.2, "real"},
		{[]int{1, 2, 3}, "blob"},
	}

	for _, c := range cases {
		typ := dialect.DataTypeOf(reflect.ValueOf(c.Value))
		if typ != c.Type {
			t.Fatalf("expect %s, but got %s", c.Type, typ)
		}
		t.Logf("expect %s, and got %s", c.Type, typ)
	}
}
