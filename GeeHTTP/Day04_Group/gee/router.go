package gee

import (
	"log"
	"net/http"
	"strings"
)

// roots key eg, roots['GET'] roots['POST']
// handlers key eg, handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']
type router struct {
	roots map[string]*node	// 用 roots 来存储每种请求方式的 Trie 树根节点
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots: make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// 只允许有一个 *
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)

	log.Printf("[addRoute] Route %4s - %s", method, pattern)
	// 放入 handler
	key := method + "-" + pattern
	r.handlers[key] = handler
	// 放入 roots
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
}

/*
getRoute 函数中，还解析了:和*两种匹配符的参数，返回一个 map 。
例如/p/go/doc匹配到/p/:lang/doc，解析结果为：{lang: "go"}，
/static/css/geektutu.css匹配到/static/*filepath，
解析结果为{filepath: "css/geektutu.css"}。
 */
func (r *router) getRoute(method string, path string) (*node, map[string]string)  {
	searchParts := parsePattern(path)
	params := make(map[string]string)

	root, ok := r.roots[method]		// 得到每个方法下的 根节点
	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)
	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			// 将 * 后面的所有直接拼接
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router)getRoutes(method string) []*node {
	root, ok := r.roots[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

func (r *router) handler(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.params = params
		key := c.Method + "-" + n.pattern
		log.Printf("[handler] Route %4s - %s", c.Method, n.pattern)
		if handler, ok := r.handlers[key]; ok {
			handler(c)
		} else {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		}
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
