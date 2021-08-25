package gee

import (
	"log"
	"net/http"
	"strings"
)

type HandlerFunc func(ctx *Context)

type Engine struct {
	*RouterGroup
	router *router
	groups	[]*RouterGroup
}

type RouterGroup struct {
	prefix	string
	middlewares	[]HandlerFunc	// 支持中间件
	parent	*RouterGroup		// 支持分组
	engine	*Engine		// 需要有访问 Router 的能力, 将 Engine 作为最顶层的分组
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

/*
	r := gee.New()
	v1 := r.Group("/v1")
 */
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,// (*Engine).engine 是指向自己的。
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// Use is defined to add middleware to the group
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	log.Printf("Group: %v - add middlewares", group.prefix)
	group.middlewares = append(group.middlewares, middlewares...)
}

/*
可以仔细观察下addRoute函数，调用了group.engine.router.addRoute来实现了路由的映射。
由于Engine从某种意义上继承了RouterGroup的所有属性和方法，
因为 (*Engine).engine 是指向自己的。
这样实现，我们既可以像原来一样添加路由，也可以通过分组添加路由。
 */
func (group *RouterGroup) addRoute(method string, patternSuffix string, handler HandlerFunc) {
	pattern := group.prefix + patternSuffix
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc)  {
	group.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	context := newContext(w, req)
	context.handlers = middlewares
	engine.router.handler(context)
}
