package gee

import (
"fmt"
"net/http"
)

// 用来定义路由映射的处理方法
type HandleFunc func(http.ResponseWriter, *http.Request)

// 添加了一张路由映射表router
type Engine struct {
	router map[string]HandleFunc
}

func New() *Engine {
	return &Engine{router: make(map[string]HandleFunc)}
}

// key 由请求方法和静态路由地址构成，例如GET-/、GET-/hello、POST-/hello
func (engine *Engine) addRoute(method string, pattern string, handler HandleFunc) {
	key := method + "-" + pattern
	engine.router[key] = handler
}

func (engine *Engine) GET(pattern string, handler HandleFunc) {
	engine.addRoute("GET", pattern, handler)
}

func (engine *Engine) POST(pattern string, handler HandleFunc) {
	engine.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.Method + "-" + req.URL.Path
	if handler, ok := engine.router[key]; ok {
		handler(w, req)
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 NOT FOUND: %s \n", req.URL)
	}
}


