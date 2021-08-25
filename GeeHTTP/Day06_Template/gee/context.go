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
