package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node       // "GET"
	handlers map[string]HandlerFunc // "/a/:id/c"
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// Only one * is allowed
// /a/b/c	-->		["a","b","c"]
// /a/:b/c	-->		["a",":b","c"]
// /a/**	-->		["a","**"]
// /a/*b/c	-->		["a","*b"]
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
	parts := parsePattern(pattern) // "/a/:id/c"  --> ["a",":id","c"]

	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{} // 创建顶层node
	}
	r.roots[method].insert(pattern, parts, 0) // 插入到对应method的路由树
	r.handlers[key] = handler                 // 插入到handler map
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path) // "/a/123/c"  --> ["a","123","c"]
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok { // 请求方法不在路由中
		return nil, nil
	}

	n := root.search(searchParts, 0) // 查找第一个匹配到的路由树节点

	if n != nil {
		parts := parsePattern(n.pattern) // "/a/:b/c" --> ["a",":id","c"]
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

func (r *router) getRoutes(method string) []*node {
	root, ok := r.roots[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params // 路由参数挂载到context
		key := c.Method + "-" + n.pattern
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
