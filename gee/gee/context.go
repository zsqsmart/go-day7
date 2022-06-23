package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]any

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
	// middleware
	handlers []HandlerFunc
	index    int

	engine *Engine
}

type ResponseCode int

const (
	OK_CODE   ResponseCode = 200
	FAIL_CODE ResponseCode = 400
)

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
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Param(key string) string {
	return c.Params[key]
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

func (c *Context) String(code int, format string, values ...any) {
	c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj any) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) OK(data any) {
	c.JSON(http.StatusOK, H{
		"code":    OK_CODE,
		"message": "success",
		"data":    data,
	})
}

func (c *Context) Fail(message string) {
	c.JSON(http.StatusOK, H{
		"code":    FAIL_CODE,
		"message": message,
	})
}

func (c *Context) HTML(code int, name string, data any) {
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.Status(code)
	err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data)
	if err != nil {
		c.Fail(err.Error())
	}
}
