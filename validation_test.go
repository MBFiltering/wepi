package wepi

import (
	"net/http"
	"testing"
)

func TestValidateAndExtractRouteFunc_Simple(t *testing.T) {
	handler := &RouteHandlerSimple[string]{
		Handler: func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
			return "ok", nil, nil
		},
	}
	route := &Route{
		route:        "/test",
		method:       GET,
		RouteHandler: handler,
	}

	handlerFunc, stType, err := validateAndExtractRouteFunc(route)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerFunc.IsValid() {
		t.Error("expected valid handler func")
	}
	if stType.Name() != "ParamsManager" {
		t.Errorf("structType = %v, want ParamsManager", stType.Name())
	}
}

func TestValidateAndExtractRouteFunc_NilHandler(t *testing.T) {
	route := &Route{
		route:        "/test",
		method:       GET,
		RouteHandler: nil,
	}

	_, _, err := validateAndExtractRouteFunc(route)
	if err == nil {
		t.Error("expected error for nil handler")
	}
}

func TestValidateAndExtractRouteFunc_NotAStruct(t *testing.T) {
	route := &Route{
		route:        "/test",
		method:       GET,
		RouteHandler: "not a struct",
	}

	_, _, err := validateAndExtractRouteFunc(route)
	if err == nil {
		t.Error("expected error for non-struct handler")
	}
}
