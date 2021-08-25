package codec

import "io"

/*
一个典型的 RPC 调用如下：
err = client.Call("Arith.Multiply", args, &reply)
 */


type Header struct {
	// ServiceMethod 是服务名和方法名，通常与 Go 语言中的结构体和方法相映射。
	ServiceMethod	string	// format "Service.Method"
	Seq				uint64	// Seq
	Error 			string	// 错误信息。服务端若发生错误，则将错误信息放在 Error 中
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCode
}