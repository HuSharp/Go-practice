## GeeGin



使用 Go 语言实现一个简单的 Web 框架，起名叫做`Gee`，[`geektutu.com`](https://geektutu.com/)的前三个字母。，参考了`Gin`。

实验环境为 GoLand， 采用 curl 调用观察。

 

## net/http 的介绍

可以查看[此文章](https://juejin.cn/post/6844903998869209095 )

### 需要实现 ServerHTTP 方法

定义了一个结构体`Engine`，实现了方法`ServeHTTP`。这个方法有2个参数，第二个参数是 *Request* ，该对象包含了该HTTP请求的所有的信息，比如请求地址、Header和Body等信息；第一个参数是 *ResponseWriter* ，利用 *ResponseWriter* 可以构造针对该请求的响应。



#### Engine 结构体

在`Engine`中，添加了一张路由映射表`router`，key 由请求方法和静态路由地址构成，例如`GET-/`、`GET-/hello`、`POST-/hello`，这样针对相同的路由，如果请求方法不同,可以映射不同的处理方法(Handler)，value 是用户映射的处理方法。

```go
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
```

- `Engine`实现的 *ServeHTTP* 方法的作用就是，解析请求的路径，查找路由映射表，如果查到，就执行注册的处理方法。如果查不到，就返回 *404 NOT FOUND* 。



实现了路由映射表，提供了用户注册静态路由的方法，包装了启动服务的函数。



## 设计Context

### 必要性

1. 对Web服务来说，无非是根据请求`*http.Request`，构造响应`http.ResponseWriter`。但是这两个对象提供的接口粒度太细，比如我们要构造一个完整的响应，需要考虑消息头(Header)和消息体(Body)，而 Header 包含了状态码(StatusCode)，消息类型(ContentType)等几乎每次请求都需要设置的信息。因此，如果不进行有效的封装，那么框架的用户将需要写大量重复，繁杂的代码，而且容易出错。针对常用场景，能够高效地构造出 HTTP 响应是一个好的框架必须考虑的点。

用返回 JSON 数据作比较，感受下封装前后的差距。

封装前

```go
obj = map[string]interface{}{
    "name": "geektutu",
    "password": "1234",
}
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
encoder := json.NewEncoder(w)
if err := encoder.Encode(obj); err != nil {
    http.Error(w, err.Error(), 500)
}
```

VS 封装后：

```
c.JSON(http.StatusOK, gee.H{
    "username": c.PostForm("username"),
    "password": c.PostForm("password"),
})
```

- 将`路由(router)`独立出来，方便之后增强。



具体设计

```go
package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

/*
	中间件  非业务的技术类组件
接收到请求后，应查找所有应作用于该路由的中间件，保存在 Context 中，依次进行调用。
为什么依次调用后，还需要在Context中保存呢？
因为在设计中，中间件不仅作用在处理流程前，也可以作用在处理流程后，
即在用户定义的 Handler 处理完毕后，还可以执行剩下的操作。
 */
type Context struct {
	// origin obj
	Writer	http.ResponseWriter
	Req		*http.Request
	// request info
	Path	string
	Method	string
	params 	map[string]string	// 将解析后的参数存储到Params中，通过c.Param("lang")的方式获取到对应的值。
	// response info
	StatusCode int
	// middleware
	handlers	[]HandlerFunc
	index	int	// 指示 handlerFunc 目前到了哪一个中间件
	engine *Engine	// 使可以通过 context 访问 Engine 中的 HTML 模版
 }

func newContext(w http.ResponseWriter, req *http.Request) *Context {		
	return &Context{
		Writer: w,
		Req: req,
		Path: req.URL.Path,
		Method: req.Method,
		index: -1,
	}
}

/*
	c.Next()表示等待执行其他的中间件或用户的Handler：
	index是记录当前执行到第几个中间件，当在中间件中调用Next方法时，控制权交给了下一个中间件，
	直到调用到最后一个中间件，然后再从后往前，调用每个中间件在Next方法之后定义的部分。
 */
func (c *Context) Next(){
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message" : err})
}

func (c *Context) Param(key string) string {
	val, _ := c.params[key]
	return val
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string	{
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, val string) {
	c.Writer.Header().Set(key, val)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	// 	Encoder 主要负责将结构对象编码成 JSON 数据
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	// ExecuteTemplate方法类似Execute，但是使用名为name的t关联的模板产生输出。
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
}

```





- 设计`上下文(Context)`，封装 Request 和 Response ，提供对 JSON、HTML 等返回类型的支持。