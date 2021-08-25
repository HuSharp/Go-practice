package lru

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T)  {
	if testing.Short() {
		t.Skip("skipping this test")
	}
	lru := New(int64(0), nil)
	lru.Add("key1", String("1234"))
	if get, ok := lru.Get("key1"); !ok || string(get.(String)) != "1234" {
		t.Fatal("cache hit key1 = 1234 failed!")
	}
	if _, ok := lru.Get("key2"); !ok {
		t.Fatal("cache missed key2 failed!")
	}
}

func TestRemoveOldest(t *testing.T)  {
	if testing.Short() {
		t.Skip("skipping this test")
	}
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)

	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if val, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed, val: %v", val)
	}

	if val, ok := lru.Get("key2"); ok && lru.Len() == 2 {
		t.Fatalf("Removeoldest key2 success, val: %v", val)
	}
}

func TestOnEvicted(t *testing.T) {

	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	lru := New(int64(18), callback)
	key := []string{"k1", "k2", "k3", "k1", "k4"}
	val := []string{"123456", "val2", "val3", "val1", "val4"}
	for i := 0; i < 5; i++ {
		lru.Add(key[i], String(val[i]))
		t.Logf("cur maxBytes: %v. usedBytes: %v, ll: %v", lru.maxBytes, lru.usedBytes, lru.ll)
	}

	expect := []string{"k1", "k2"}
	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys:%s equals to %s", keys, expect)
	}
}