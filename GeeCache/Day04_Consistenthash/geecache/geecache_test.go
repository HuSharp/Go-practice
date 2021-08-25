package geecache

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"
)

var m sync.Mutex
var set = make(map[int]bool, 0)

func printOnce(num int) {
	m.Lock()
	defer m.Unlock()
	if _, exist := set[num]; !exist {
		fmt.Println(num)
	}
	set[num] = true
}

func main() {
	for i := 0; i < 10;  i++ {
		go printOnce(100)
	}
	time.Sleep(time.Second)
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed!")
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T)  {
	loadDB := make(map[string]int, len(db))
	group := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[TestDB] --SlowDB-- search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadDB[key]; !ok {
					loadDB[key] = 0
				}
				loadDB[key] += 1	// 如果次数大于1，则表示调用了多次回调函数
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not found", key)
		}))

	for k, v := range db {
		if view, err := group.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get val of %s", k)
		} // load from callback func
		if _, err := group.Get(k); err != nil || loadDB[k] > 1 {
			t.Fatalf("cache %s missed", k)
		} // cache hit
		/*
		在缓存已经存在的情况下，是否直接从缓存中获取，
		为了实现这一点，使用 loadCounts 统计某个键调用回调函数的次数，
			如果次数大于1，则表示调用了多次回调函数，没有缓存。
		 */
	}

	if view, err := group.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}