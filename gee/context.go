package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	Params map[string]string
	// response info
	StatusCode int
	handler    []HandlerFunc //多个处理函数
	index      int           //handler的下表，目前到第几个函数
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

func (c *Context) Next() {
	c.index++
	len := len(c.handler)
	for ; c.index < len; c.index++ { //为了保证即使中间件没有调用next也可以继续调用所有的中间件
		c.handler[c.index](c)
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handler)
	c.JSON(code, H{"message": err})
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

func (c *Context) Auto(code int, value ...interface{}) {
	if len(value) >= 2 { //无法解析出html和format的不同，如果string也想用可以多加个空容器
		switch value[0].(type) {
		case string:
			str := value[0].(string)
			c.String(code, str, (value[1:])...)
		default:
			fmt.Println("unsupport type！")
		}
	} else {
		switch value[0].(type) {
		case string:
			str := value[0].(string)
			c.HTML(code, str)
			//		fmt.Println("执行的是网页")
		case []byte:
			str := value[0].([]byte)
			c.Data(code, str)
			//		fmt.Println("执行的是数据")
		default:
			c.JSON(code, value) //除了这两种类型都是json本身
			//		fmt.Println("执行的是json")
		}
	}
}
