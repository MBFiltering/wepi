package wepi

import "net/http"

const (
	POST  = "POST"
	GET   = "GET"
	PATCH = "PATCH"
)

type Route struct {
	route        string
	method       string
	RouteHandler any
	Middlewares  []func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error)
}

type RouteHandlerWithStruct[ST any, R any] struct {
	Handler func(st ST, params ParamsManager, req *http.Request) (R, *CustomResponse, error)
}
type RouteHandlerSimple[R any] struct {
	Handler func(params ParamsManager, req *http.Request) (R, *CustomResponse, error)
}

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
