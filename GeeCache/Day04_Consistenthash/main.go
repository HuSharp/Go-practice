package main

import (
	"fmt"
	"strconv"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	for i := 0; i < 3; i++ {
		p := strconv.Itoa(i) + "key"
		bytes := []byte(p)
		fmt.Println(p, ", ", bytes)
	}

	p := []int{1, 2, 3, 4}
	test(p[1:]...)
}

func test(args... int)  {
	for _, v := range args {
		fmt.Println(v)
	}
}


