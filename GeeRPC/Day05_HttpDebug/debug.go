package geerpc

import (
	"fmt"
	"html/template"
	"net/http"
)

const debudText =  `<html>
	<body>
	<title>GeeRPC Services</title>
	{{range .}}
	<hr>
	Service {{.Name}}
	<hr>
		<table>
		<th align=center>Method</th><th align=center>Calls</th>
		{{range $name, $mtype := .Method}}
			<tr>
			<td align=left font=fixed>{{$name}}({{$mtype.ArgType}}, {{$mtype.ReplyType}}) error</td>
			<td align=center>{{$mtype.NumCalls}}</td>
			</tr>
		{{end}}
		</table>
	{{end}}
	</body>
	</html>`

var debug = template.Must(template.New("RPC debug").Parse(debudText))

type debugHTTP struct {
	*Server
}

type debugService struct {
	Name 	string
	Method 	map[string]*methodType
}

func (server debugHTTP) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var services []debugService
	// 遍历所有sync.Map中的键值对
	server.serviceMap.Range(func(namei, svci interface{}) bool {
		svc := svci.(*service)
		services = append(services, debugService{
			Name:   namei.(string),
			Method: svc.method,
		})
		return true
	})

	err := debug.Execute(w, services)
	if err != nil {
		_, _ = fmt.Fprintln(w, "[debugHTTP.ServeHTTP] rpc: error execute template: ", err.Error())
	}
}
