package main

import (
	"encoding/json"
	"fmt"
	"geerpc"
	"geerpc/codec"
	"log"
	"net"
	"time"
)

// 在 startServer 中使用了信道 addr，确保服务端端口监听成功，客户端再发起请求。
func startServer(addr chan string)  {
	// pick up a free port
	listen, err := net.Listen("tcp", "localhost:43561")
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
	addr := make(chan string)
	go startServer(addr)

	// a simple rpc client
	// Dial在指定的网络上连接指定的地址
	// 确保服务端端口监听成功，客户端再发起请求
	conn, _ := net.Dial("tcp", <-addr)
	defer conn.Close()

	time.Sleep(time.Second * 3)
	// send options
	// 首先发送 Option 进行协议交换，
	json.NewEncoder(conn).Encode(geerpc.DefaultOption)
	cc := codec.NewGobCode(conn)
	// 其次发送 request 和 body
	for i := 0; i < 5; i++ {
		header := &codec.Header {
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		cc.Write(header, fmt.Sprintf("geeRpc req %d", i))

		//var replyHeader codec.Header
		cc.ReadHeader(header)
		var replyBody string
		cc.ReadBody(&replyBody)
		log.Println("client data:", header, replyBody)
	}
}