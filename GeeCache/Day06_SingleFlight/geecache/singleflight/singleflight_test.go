package singleflight

import "testing"

func TestDo(t *testing.T)  {
	var g Group
	do, err := g.Do("key", func() (interface{}, error) {
		return "bar", nil
	})
	if do != "bar" || err != nil {
		t.Errorf("Do v = %v, error = %v", do, err)
	}
}
