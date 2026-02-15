package wepi

import "net/http"

const (
	POST  = "POST"
	GET   = "GET"
	PATCH = "PATCH"
)

// Route represents a registered route with its handler and middleware chain.
type Route struct {
	route        string
	method       string
	RouteHandler any
	Middlewares  []func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error)
}

// RouteHandlerWithStruct handles routes that expect a typed request body.
type RouteHandlerWithStruct[ST any, R any] struct {
	Handler func(st ST, params ParamsManager, req *http.Request) (R, *CustomResponse, error)
}

// RouteHandlerSimple handles routes that use only query/form params.
type RouteHandlerSimple[R any] struct {
	Handler func(params ParamsManager, req *http.Request) (R, *CustomResponse, error)
}

// AddJsonPOST registers a POST route that expects a JSON request body deserialized into type T.
func AddJsonPOST[T any, R any](wepiController *WepiController, path string, function func(st T, params ParamsManager, req *http.Request) (R, *CustomResponse, error), middlewares ...func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error)) {
	r := &RouteHandlerWithStruct[T, R]{
		Handler: function,
	}
	method := POST
	ro := &Route{
		route:        path,
		method:       method,
		RouteHandler: r,
		Middlewares:  middlewares,
	}
	wepiController.addRoute(&WepiComposedRoute{
		path:   path,
		route:  ro,
		method: method,
	})
}

// AddFormPost registers a POST route that reads form-encoded data via ParamsManager.
func AddFormPost[R any](wepiController *WepiController, path string, function func(params ParamsManager, req *http.Request) (R, *CustomResponse, error), middlewares ...func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error)) {
	r := &RouteHandlerSimple[R]{
		Handler: function,
	}
	method := POST
	ro := &Route{
		route:        path,
		method:       method,
		RouteHandler: r,
		Middlewares:  middlewares,
	}
	wepiController.addRoute(&WepiComposedRoute{
		path:   path,
		route:  ro,
		method: method,
	})
}

// AddGetWithStruct registers a GET route that deserializes query parameters into type T.
func AddGetWithStruct[T any, R any](wepiController *WepiController, path string, function func(st T, params ParamsManager, req *http.Request) (R, *CustomResponse, error), middlewares ...func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error)) {
	r := &RouteHandlerWithStruct[T, R]{
		Handler: function,
	}
	method := GET
	ro := &Route{
		route:        path,
		method:       method,
		RouteHandler: r,
		Middlewares:  middlewares,
	}
	wepiController.addRoute(&WepiComposedRoute{
		path:   path,
		route:  ro,
		method: method,
	})
}

// AddGET registers a GET route that reads query parameters via ParamsManager.
func AddGET[R any](wepiController *WepiController, path string, function func(params ParamsManager, req *http.Request) (R, *CustomResponse, error), middlewares ...func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error)) {
	r := &RouteHandlerSimple[R]{
		Handler: function,
	}
	method := GET
	ro := &Route{
		route:        path,
		method:       method,
		RouteHandler: r,
		Middlewares:  middlewares,
	}
	wepiController.addRoute(&WepiComposedRoute{
		path:   path,
		route:  ro,
		method: method,
	})
}
