package wepi

import "net/http"

// CustomResponse allows handlers to return custom HTTP status codes, headers, and body content.
type CustomResponse struct {
	status  int
	headers http.Header
	body    []byte
}

// Custom creates a new empty CustomResponse builder.
func Custom() *CustomResponse {
	return &CustomResponse{
		body:    nil,
		status:  0,
		headers: nil,
	}
}

func (c *CustomResponse) SetStatus(status int) *CustomResponse {
	c.status = status
	return c
}

func (c *CustomResponse) SetBody(body []byte) *CustomResponse {
	c.body = body
	return c
}

func (c *CustomResponse) SetBodyString(s string) *CustomResponse {
	c.body = []byte(s)
	return c
}

func (c *CustomResponse) AddHeader(key string, value string) *CustomResponse {
	if c.headers == nil {
		c.headers = make(http.Header)
	}
	c.headers.Add(key, value)
	return c
}

func (c *CustomResponse) SetHeader(key string, value string) *CustomResponse {
	if c.headers == nil {
		c.headers = make(http.Header)
	}
	c.headers.Set(key, value)
	return c
}
