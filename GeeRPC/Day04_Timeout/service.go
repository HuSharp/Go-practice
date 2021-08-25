package geerpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)


/*
通过反射实现结构体与服务的映射关系 Service
func (t *T) MethodName(argType T1, replyType *T2) error
*/
type methodType struct {
	method			reflect.Method	// 方法本身
	ArgType			reflect.Type	// 请求参数（第一个
	ReplyType		reflect.Type	// 回复参数（第二个
	numCalls		uint64			// 后续统计方法调用次数时会用到
}

func (m *methodType) NumCalls() uint64 {
	// LoadUint64原子性的获取*addr的值。
	return atomic.LoadUint64(&m.numCalls)
}

// newArgv 指针类型和值类型创建实例的方式有细微区别
func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}


func (m *methodType) newReplyVal() reflect.Value {
	var argv reflect.Value
	replyVal := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyVal.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyVal.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return argv
}

/*
Service 实现
*/
type service struct {
	name 	string
	typ		reflect.Type
	val		reflect.Value
	method 	map[string]*methodType
}

func (s *service) call(m *methodType, argv, replyVal reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnVals := f.Call([]reflect.Value{s.val, argv, replyVal})
	if errInter := returnVals[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}

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

/*
registerMethods 过滤出了符合条件的方法：
	两个导出或内置类型的入参（反射时为 3 个，第 0 个是自身，类似于 python 的 self，java 中的 this）
	返回值有且只有 1 个，类型为 error
*/
func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 ||  mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType)	{
			continue
		}
		s.method[method.Name] = &methodType{
			method: method,
			ArgType: argType,
			ReplyType: replyType,
		}
		log.Printf("[registerMethods] rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

