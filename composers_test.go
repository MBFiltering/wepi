package wepi

import (
	"net/http"
	"testing"
)

func TestAddJsonPOST(t *testing.T) {
	w := Get()

	type Input struct {
		Name string `json:"name"`
	}

	AddJsonPOST[Input, string](w, "/post", func(st Input, params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	})

	// Verify the route was registered
	r, ok := w.routes.Load("/post" + POST)
	if !ok {
		t.Fatal("expected route to be registered")
	}
	route := r.(*Route)
	if route.route != "/post" {
		t.Errorf("route = %q, want %q", route.route, "/post")
	}
	if route.method != POST {
		t.Errorf("method = %q, want %q", route.method, POST)
	}
}

func TestAddFormPost(t *testing.T) {
	w := Get()

	AddFormPost[string](w, "/form", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	})

	r, ok := w.routes.Load("/form" + POST)
	if !ok {
		t.Fatal("expected route to be registered")
	}
	route := r.(*Route)
	if route.method != POST {
		t.Errorf("method = %q, want %q", route.method, POST)
	}
}

func TestAddGET(t *testing.T) {
	w := Get()

	AddGET[string](w, "/get", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	})

	r, ok := w.routes.Load("/get" + GET)
	if !ok {
		t.Fatal("expected route to be registered")
	}
	route := r.(*Route)
	if route.method != GET {
		t.Errorf("method = %q, want %q", route.method, GET)
	}
}

func TestAddJsonPOST_WithMiddleware(t *testing.T) {
	w := Get()

	type Input struct {
		Name string `json:"name"`
	}

	mw := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		return nil, nil
	}

	AddJsonPOST[Input, string](w, "/mw", func(st Input, params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	}, mw)

	r, ok := w.routes.Load("/mw" + POST)
	if !ok {
		t.Fatal("expected route to be registered")
	}
	route := r.(*Route)
	if len(route.Middlewares) != 1 {
		t.Errorf("middleware count = %d, want 1", len(route.Middlewares))
	}
}

func TestAddGET_WithPathParams(t *testing.T) {
	w := Get()

	AddGET[string](w, "/users/{id}", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return params.GetString("id", ""), nil, nil
	})

	// Verify the route is stored under the template path
	r, ok := w.routes.Load("/users/{id}" + GET)
	if !ok {
		t.Fatal("expected route to be registered with template path")
	}
	route := r.(*Route)
	if route.route != "/users/{id}" {
		t.Errorf("route = %q, want %q", route.route, "/users/{id}")
	}
}
