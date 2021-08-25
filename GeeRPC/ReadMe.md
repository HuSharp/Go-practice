

## 消息的序列化与反序列化

一个典型的 RPC 调用如下：

```
err = client.Call("Arith.Multiply", args, &reply)
```

客户端发送的请求包括服务名 `Arith`，方法名 `Multiply`，参数 `args` 三个，服务端的响应包括错误 `error`，返回值 `reply` 2 个。我们将请求和响应中的参数和返回值抽象为 body，剩余的信息放在 header 中，那么就可以抽象出数据结构 Header：



### gob说明

http://c.biancheng.net/view/4597.html

为了让某个[数据结构](http://c.biancheng.net/data_structure/)能够在网络上传输或能够保存至文件，它必须被编码然后再解码。当然已经有许多可用的编码方式了，比如 [JSON](http://c.biancheng.net/view/4545.html)、[XML](http://c.biancheng.net/view/4551.html)、Google 的 protocol buffers 等等。而现在又多了一种，由Go语言 encoding/gob 包提供的方式。在编码和解码过程中用到了 Go 的反射。

Gob 是Go语言自己以二进制形式序列化和反序列化程序数据的格式，可以在 encoding 包中找到。这种格式的数据简称为 Gob（即 Go binary 的缩写）。类似于 [Python](http://c.biancheng.net/python/) 的“pickle”和 [Java](http://c.biancheng.net/java/) 的“Serialization”。





### 头格式

一般来说，涉及协议协商的这部分信息，需要设计固定的字节来传输的。但是为了实现上更简单，GeeRPC 客户端固定采用 JSON 编码 Option，后续的 header 和 body 的编码方式由 Option 中的 CodeType 指定，服务端首先使用 JSON 解码 Option，然后通过 Option 的 CodeType 解码剩余的内容。即报文将以这样的形式发送：

```
| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|
```

在一次连接中，Option 固定在报文的最开始，Header 和 Body 可以有多个，即报文可能是这样的。

```
| Option | Header1 | Body1 | Header2 | Body2 | ...
```



### 粘包拆包

json 字符串是有数据的边界的即 "{" 和 "}"所以这里并不会出现粘包的问题





## 拆编码器







### 客户端

对一个客户端端来说，接收响应、发送请求是最重要的 2 个功能。

#### 接收功能

那么首先实现接收功能，接收到的响应有三种情况：

- call 不存在，可能是请求没有发送完整，或者因为其他原因被取消，但是服务端仍旧处理了。
- call 存在，但服务端处理出错，即 h.Error 不为空。
- call 存在，服务端处理正常，那么需要从 body 中读取 Reply 的值。

#### 发送请求





### 服务端

#### 结构体映射为服务

通过反射实现结构体与服务的映射关系 Service
func (t *T) MethodName(argType T1, replyType *T2) error

```go
type methodType struct {
	method			reflect.Method	// 方法本身
	ArgType			reflect.Type	// 请求参数（第一个
	ReplyType		reflect.Type	// 回复参数（第二个
	numCalls		uint64			// 后续统计方法调用次数时会用到
}
```



#### 定义结构体 service：

```
type service struct {
	name   string
	typ    reflect.Type
	rcvr   reflect.Value
	method map[string]*methodType
}
```

service 的定义也是非常简洁的，name 即映射的结构体的名称，比如 `T`，比如 `WaitGroup`；typ 是结构体的类型；rcvr 即结构体的实例本身，保留 rcvr 是因为在调用时需要 rcvr 作为第 0 个参数；method 是 map 类型，存储映射的结构体的所有符合条件的方法。



#### 将 service 注册到服务端

```go
func (server *Server) Register(receive interface{}) error {
   s := newService(receive)
   // 如果值存在则直接返回，若不存在则存储，返回值loaded，true表示数据被加载，false表示数据被存储
   if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
      return errors.New("[Register] rpc: service already defined: " + s.name)
   }
   return nil
}
```

关于 service 的**创建实例：**

```go
// 入参是任意需要映射为服务的结构体实例
func newService(receive interface{}) *service {
   s := new(service)
   s.val = reflect.ValueOf(receive)
   s.name = reflect.Indirect(s.val).Type().Name()
   s.typ = reflect.TypeOf(receive)
   if !ast.IsExported(s.name) {
      log.Fatalf("[newService] rpc server: %s is not a valid service name", s.name)
   }
   s.registerMethods()
   log.Printf("[newService] rpc server success, s: %v", s)
   return s
}
```



进行 method 的查找

因为 ServiceMethod 的构成是 “Service.Method”，因此先将其分割成 2 部分，第一部分是 Service 的名称，第二部分即方法名。现在 serviceMap 中找到对应的 service 实例，再从 service 实例的 method 中，找到对应的 methodType。

```go
func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
   dot := strings.LastIndex(serviceMethod, ".")
   if dot < 0 {
      err = errors.New("[findService] rpc server: service/method request ill-formed: " + serviceMethod)
      return
   }
   serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
   svci, ok := server.serviceMap.Load(serviceName)
   if !ok {
      err = errors.New("[findService] rpc server: can't find service " + serviceName)
      return
   }
   svc = svci.(*service)
   mType := svc.method[methodName]
   if mType == nil {
      err = errors.New("[findService] rpc server: can't find method " + methodName)
   }
   return
}
```





## 超时机制

**使用 `time.After()` 结合 `select+chan` 完成。**

纵观整个远程调用的过程，需要客户端处理超时的地方有：

- 与服务端建立连接，导致的超时				---------------->修改 dial 即可
- 发送请求到服务端，写报文导致的超时              -------------------> Call 中采用 ctx 的 done 信道进行关闭
- 等待服务端处理时，等待处理导致的超时（比如服务端已挂死，迟迟不响应）
- 从服务端接收响应时，读报文导致的超时

需要服务端处理超时的地方有：

- 读取客户端请求报文时，读报文导致的超时
- 发送响应报文时，写报文导致的超时
- 调用映射服务的方法时，处理报文导致的超时·







## 服务端支持 HTTP 协议

那通信过程应该是这样的：

1. 客户端向 RPC 服务器发送 CONNECT 请求

```
CONNECT 10.0.0.1:9999/_geerpc_ HTTP/1.0
```

1. RPC 服务器返回 HTTP 200 状态码表示连接建立。

```
HTTP/1.0 200 Connected to Gee RPC
```

1. 客户端使用创建好的连接发送 RPC 报文，先发送 Option，再发送 N 个请求报文，服务端处理 RPC 请求并响应。