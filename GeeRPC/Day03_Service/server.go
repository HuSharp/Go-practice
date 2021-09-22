package geerpc

import (
	"encoding/json"
	"errors"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

/*
	Option 协商消息的编解码方式
 */
const HuSharpMagicNumber = 0x48755368617270	// HuSharp

// Option 为固定编码，拥有指定编码方式
type Option struct {
	MagicNumber int			// marks this is a geeRpc request
	CodecType   codec.Type	// client may choose different Codec to encode body
}

var DefaultOption = &Option{
	MagicNumber: HuSharpMagicNumber,
	CodecType:   codec.GobType,
}

// Register publishes in the server the set of methods of the service
func (server *Server) Register(receive interface{}) error {
	s := newService(receive)
	// 如果值存在则直接返回，若不存在则存储，返回值loaded，true表示数据被加载，false表示数据被存储
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("[Register] rpc: service already defined: " + s.name)
	}
	return nil
}

func Register(receive interface{}) error {
	return DefaultServer.Register(receive)
}

// ServiceMethod 的构成是 “Service.Method”，因此先将其分割成 2 部分，第一部分是 Service 的名称，第二部分即方法名
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


/*
Server 服务端的实现
 */
// 声明只包含方法的结构体
type Server struct {
	serviceMap	sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

// Accept ：net.Listener 作为参数，
// for 循环等待 socket 连接建立，并开启子协程处理，处理过程交给了 ServerConn 方法
func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("[Accept] rpc server accept err:", err)
			return
		}
		go server.ServerConn(conn)
	}
}

// Accept DefaultServer 是一个默认的 Server 实例，主要为了用户使用方便。
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

// ServerConn
// 首先使用 json.NewDecoder 反序列化得到 Option 实例，检查 MagicNumber 和 CodeType 的值是否正确。
// 然后根据 CodeType 得到对应的消息编解码器，接下来的处理交给 serverCodec。
func (server *Server) ServerConn(conn io.ReadWriteCloser) {
	defer conn.Close()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("[ServerConn] rpc server: Options wrong, err: ", err)
		return
	}
	if opt.MagicNumber != HuSharpMagicNumber {
		log.Println("[ServerConn] rpc server: MagicNumber wrong, now number: ", opt.MagicNumber)
		return
	}
	// 消息编解码器
	codecFunc := codec.NewCodecFuncMap[opt.CodecType]
	if codecFunc == nil {
		log.Println("[ServerConn] rpc server: CodecType wrong, now codecType: ", opt.CodecType)
		return
	}
	server.serveCodec(codecFunc(conn))
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

/* serveCodec 的过程非常简单。主要包含三个阶段
	读取请求 readRequest
	处理请求 handleRequest
	回复请求 sendResponse
 */
func (server *Server) serveCodec(cc codec.Codec) {
	// 使用锁保证报文完整性
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)		// wait until all request are handled

	// 在一次连接中，允许接收多个请求，即多个 request header 和 request body，
	// 因此这里使用了 for 无限制地等待请求的到来，直到发生错误（例如连接被关闭，接收到的报文有问题等）
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			// 只有在 header 解析失败时候 才终止循环
			if req == nil {
				break
			}
			req.header.Error = err.Error()	// 服务端出错，将 Error 信息放置 err
			server.sendResponse(cc, req.header, sending, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	cc.Close()
}

/*
request 处理的相关函数
	读取请求 readRequest
	处理请求 handleRequest
	回复请求 sendResponse
 */
type request struct {
	header			*codec.Header // header of request
	argv, replyVal 	reflect.Value // argv and replyVal of request
	mtype			*methodType
	svc				*service
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		log.Println("[readRequestHeader] rpc server: readHeader failed. err: ", err)
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	header, err := server.readRequestHeader(cc)
	if err != nil {
		log.Println("[readRequest] rpc server: readRequestHeader failed. err: ", err)
		return nil, err
	}
	req := &request{header: header}
	// 创建出两个入参实例
	req.argv = req.mtype.newArgv()
	req.replyVal = req.mtype.newReplyVal()
	req.svc, req.mtype, err = server.findService(header.ServiceMethod)

	// make sure that argvI is a pointer, ReadBody need a pointer as parameter
	argvI := req.argv.Interface()
	if req.argv.Kind() != reflect.Ptr {
		argvI = req.argv.Addr().Interface()
	} 
	if err := cc.ReadBody(argvI); err != nil {
		log.Println("[readRequest] rpc server: readBody failed. err: ", err)
	}
	return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, header *codec.Header, body interface{}, sending *sync.Mutex)  {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(header, body); err != nil {
		log.Println("[readRequest] rpc server: sendResponse failed. err: ", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("[handleRequest] ing...", req.header, req.argv.Elem())
	err := req.svc.call(req.mtype, req.argv, req.replyVal)
	if err != nil {
		req.header.Error = err.Error()
		server.sendResponse(cc, req.header, invalidRequest, sending)
		return
	}
	server.sendResponse(cc, req.header, req.replyVal.Interface(), sending)
}

