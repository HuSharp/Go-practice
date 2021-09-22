package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	//http.HandleFunc("/", indexHandler)
	http.HandleFunc("/", say)
	err := http.ListenAndServe(":9091", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main1() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	// nil 代表使用标准库中的实例处理
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
}

func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
}

func say(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // 解析参数
	fmt.Println(r.Form)
	fmt.Println("URL:", r.URL.Path)
	fmt.Println("Scheme", r.URL.Scheme)
	for k, v := range r.Form {
		fmt.Println(k, ":", strings.Join(v, " "))
	}
	fmt.Fprintf(w, "Hello！")
}
