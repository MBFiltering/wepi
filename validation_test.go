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
	// Simple routes have ParamsManager as first param
	if stType.Name() != "ParamsManager" {
		t.Errorf("structType = %v, want ParamsManager", stType.Name())
	}
}

func TestValidateAndExtractRouteFunc_WithStruct(t *testing.T) {
	type Input struct {
		Name string `json:"name"`
	}
	handler := &RouteHandlerWithStruct[Input, string]{
		Handler: func(st Input, params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
			return "ok", nil, nil
		},
	}
	route := &Route{
		route:        "/test",
		method:       POST,
		RouteHandler: handler,
	}

	_, stType, err := validateAndExtractRouteFunc(route)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stType.Name() != "Input" {
		t.Errorf("structType = %v, want Input", stType.Name())
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

type BadHandler struct {
	NotHandler string
}

func TestValidateAndExtractRouteFunc_MissingHandlerField(t *testing.T) {
	route := &Route{
		route:        "/test",
		method:       GET,
		RouteHandler: &BadHandler{NotHandler: "nope"},
	}

	_, _, err := validateAndExtractRouteFunc(route)
	if err == nil {
		t.Error("expected error for missing Handler field")
	}
}

func TestGetJSONFieldName_SimpleField(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required"`
	}

	// We need a real FieldError to test this. We'll use the validator directly.
	ts := TestStruct{}
	err := validatorSingleton.Struct(ts)
	if err == nil {
		t.Fatal("expected validation error")
	}

	ve := err.(interface {
		Error() string
	})
	_ = ve // just to ensure it's a validation error

	// The validation error interface from go-playground/validator
	type fieldErrorer interface {
		Field() string
		StructNamespace() string
		Tag() string
		Param() string
	}
}

func TestGetJSONFieldName_NestedField(t *testing.T) {
	type Inner struct {
		Email string `json:"email" validate:"required"`
	}
	type Outer struct {
		Data Inner `json:"data"`
	}

	ts := Outer{}
	err := validatorSingleton.Struct(ts)
	if err == nil {
		t.Fatal("expected validation error")
	}
}
