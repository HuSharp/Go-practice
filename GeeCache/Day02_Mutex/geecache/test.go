package geecache

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

func download(url string) {
	fmt.Println("start do", url)
	time.Sleep(time.Second)

}

func test() {
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func(i int) {
			fmt.Println(i)
			wg.Done()
		}(i)
	}
	fmt.Println("aaaaa")
	wg.Wait()	// Wait()用来等待所有需要等待的goroutine完成。
	fmt.Println("bbbb")
}
