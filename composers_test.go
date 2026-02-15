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

	r, ok := w.routes.Load("/post" + POST)
	if !ok {
		t.Fatal("expected route to be registered")
	}
	route := r.(*Route)
	if route.method != POST {
		t.Errorf("method = %q, want %q", route.method, POST)
	}
}

func TestAddGetWithStruct(t *testing.T) {
	w := Get()

	type Filter struct {
		Name string `json:"name"`
	}

	AddGetWithStruct(w, "/search", func(st Filter, params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	})

	r, ok := w.routes.Load("/search" + GET)
	if !ok {
		t.Fatal("expected route to be registered")
	}
	route := r.(*Route)
	if route.method != GET {
		t.Errorf("method = %q, want %q", route.method, GET)
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
