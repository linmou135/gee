package gee

import (
	"fmt"
	"net/http"
	"strings"
)

type router struct {
	root     map[string]*node //为了可以动态匹配
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		root:     make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func parsePatten(patten string) []string {
	v := strings.Split(patten, "/")
	var result []string
	result = make([]string, 0)
	for _, ve := range v {
		if ve != "" {
			result = append(result, ve)
			if ve[0] == '*' { //这样就可以随意匹配了
				break
			}
		}
	}
	return result
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePatten(pattern)
	key := method + "-" + pattern
	_, ok := r.root[method]
	if !ok {
		r.root[method] = &node{}
	}
	r.root[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchpart := parsePatten(path)
	params := make(map[string]string)
	root, ok := r.root[method]
	if !ok {
		return nil, nil
	}
	n := root.search(searchpart, 0)
	if n != nil {
		parts := parsePatten(n.pattern)
		for idx, part := range parts {
			if part[0] == ':' { //当前这个可以任意匹配
				params[part[1:]] = searchpart[idx]
				fmt.Println(part[1:])
				fmt.Println(searchpart[idx])
				fmt.Println(n.pattern)
			}
			if part[0] == '*' && len(part) > 1 { //可以随意匹配,用最后那些做一个点
				params[part[1:]] = strings.Join(searchpart[idx:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router) getRoutes(method string) []*node {
	root, ok := r.root[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil { //这个路径存在
		c.Params = params //将这个点
		key := c.Method + "-" + n.pattern
		c.handler = append(c.handler, r.handlers[key]) //放进c那边待会next就会自动执行
	} else {
		//增加一个失败的func让c处理
		c.handler = append(c.handler, func(context *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
