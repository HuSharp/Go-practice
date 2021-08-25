package main

import (
	"fmt"
	"geerpc"
	"log"
	"net"
	"sync"
	"time"
)

// 在 startServer 中使用了信道 addr，确保服务端端口监听成功，客户端再发起请求。
func startServer(addr chan string)  {
	// pick up a free port
	listen, err := net.Listen("tcp", ":63662")
	if err != nil {
		log.Fatalf("[startServer] network err: %v", err)
	}
	log.Println("[startServer] network err: ", err)
	addr <- listen.Addr().String()
	geerpc.Accept(listen)
}

func main() {
	// 客户端首先发送 Option 进行协议交换，
	// 接下来发送消息头 h := &codec.Header{}，和消息体 geeRpc req ${h.Seq}。

	// 删除所有标志，包括时间戳
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)
	client, _ := geerpc.Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second * 3)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("[]geeRpc req: %d", i)
			var reply string
			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)
	}
	wg.Wait()
}