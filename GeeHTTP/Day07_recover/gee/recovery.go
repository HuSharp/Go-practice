package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// 用来获取触发 panic 的堆栈信息
func trace(message string) string {
	/*
	Callers 用来返回调用栈的程序计数器, 第 0 个 Caller 是 Callers 本身，
	第 1 个是上一层 trace，第 2 个是再上一层的 defer func。
	因此，为了日志简洁一点，我们跳过了前 3 个 Caller。
	 */
	var pcs	[32]uintptr
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback: ")

	// 通过 runtime.FuncForPC(pc) 获取对应的函数，
	// 再通过 fn.FileLine(pc) 获取到调用该函数的文件名和行号，打印在日志中。
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s: %d", file, line))
	}
	return str.String()
}

/*
	因为defer recover机制只能针对于当前函数以及直接调用的函数可能参数的panic，
所以在Recovery里面的c.Next()会执行下面这个handler

func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
}
从而捕获到panic并恢复
如果没有c.Next()，则handler不是Recovery直接调用的函数，无法recover，
panic被net/http自带的recover机制捕获
 */
func Recovery() HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				ctx.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		ctx.Next()
	}
}