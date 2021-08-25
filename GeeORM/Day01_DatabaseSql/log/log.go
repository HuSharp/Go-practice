package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// 创建 2 个日志实例分别用于打印 Info 和 Error 日志。
// [info ] 颜色为蓝色，[error] 为红色。
var (
	errorLog = log.New(os.Stdout, "\033[31m[ error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[ info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

// log methods
var (
	Error	= errorLog.Println
	Errorf	= errorLog.Printf
	Info	= infoLog.Println
	Infof	= infoLog.Printf
)

// 支持设置日志的层级(InfoLevel, ErrorLevel, Disabled)
const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

// 通过控制 Output，来控制日志是否打印
func SetLevel(level int)  {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}
	if ErrorLevel < level {
		// Discard是一个io.Writer接口，对它的所有Write调用都会无实际操作的成功返回。
		errorLog.SetOutput(ioutil.Discard)
	}
	// 如果设置为 ErrorLevel，infoLog 的输出会被定向到 ioutil.Discard，即不打印该日志。
	if InfoLevel < level {
		infoLog.SetOutput(ioutil.Discard)
	}
}