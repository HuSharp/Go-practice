package main

import (
	"flag"
	"fmt"
	"geecache"
	"log"
	"net/http"
	"net/url"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2 << 10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[slowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// startCacheServer() 用来启动缓存服务器：创建 HTTPPool，添加节点信息，注册到 gee 中，
// 启动 HTTP 服务（共3个端口，8001/8002/8003），用户不感知。
func startCacheServer(addr string, addrs []string, gee *geecache.Group)  {
	peers := geecache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Println("[startCacheServer] geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))	// Fatal等价于{l.Print(v...); os.Exit(1)}
}

// startAPIServer() 用来启动一个 API 服务（端口 9999），与用户进行交互，用户感知。
func startAPIServer(apiAddr string, gee *geecache.Group)  {
	http.Handle("/api", http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			key := request.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "application/octet-stream")
			writer.Write(view.ByteSlice())
		}))
	log.Println("frontend Server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))	// Fatal等价于{l.Print(v...); os.Exit(1)}
}

func main() {
	// 传入 port 和 api 2 个参数，用来在指定端口启动 HTTP 服务。
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Cache Server port")
	flag.BoolVar(&api, "api", false, "Start a Api impServer")
	flag.Parse()

	// 作为多个节点
	serverAddrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	apiAddr := "http://localhost:9999"

	var serverAddrs []string
	for _, val := range serverAddrMap {
		serverAddrs = append(serverAddrs, val)
	}

	geeGroup := createGroup()
	if api {
		go startAPIServer(apiAddr, geeGroup)
	}
	startCacheServer(serverAddrMap[port], serverAddrs, geeGroup)
}

func httpEscape()  {
	webpage := "http://mywebpage.com/thumbify"
	image := "http://images.com/cat.png"
	fmt.Println(webpage +"?image" + url.QueryEscape(image))
}

