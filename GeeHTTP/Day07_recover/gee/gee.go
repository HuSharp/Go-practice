package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type HandlerFunc func(ctx *Context)

type Engine struct {
	*RouterGroup
	router *router
	groups	[]*RouterGroup
	htmlTemplates	*template.Template	// for html render 将所有的模板加载进内存
/*
    FuncMap 类型定义了函数名字符串到函数的映射，每个函数都必须有1到2个返回值，
	如果有2个则后一个必须是error接口类型；
	如果有2个返回值的方法返回的error非nil，模板执行会中断并返回给调用者该错误。
 */
	funcMap			template.FuncMap	// for html render 是所有的自定义模板渲染函数
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

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
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

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	// StripPrefix 用来过滤掉 absolute 前缀
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(ctx *Context) {
		file := ctx.Param("filepath")
		// check if file exists or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(ctx.Writer, ctx.Req)
	}
}

// Static 这个方法是暴露给用户的。用户可以将磁盘上的某个文件夹 root 映射到路由 relativePath。
/*
例如：
	r := gee.New()
	r.Static("/assets", "/usr/geektutu/blog/static")
	// 或相对路径 r.Static("/assets", "./static")
	r.Run(":9999")
	用户访问localhost:9999/assets/js/geektutu.js，
	最终返回/usr/geektutu/blog/static/js/geektutu.js。
 */
func (group *RouterGroup) Static(relativePath, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// register GET handlers
	group.GET(urlPattern, handler)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string)  {
	// ParseGlob创建一个模板并解析匹配pattern的文件（参见glob规则）里的模板
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
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
	context.engine = engine
	engine.router.handler(context)
}

