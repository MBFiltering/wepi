package wepi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func setupController() *WepiController {
	return Get()
}

func TestRun_GETRoute_JSONResponse(t *testing.T) {
	w := setupController()

	type Response struct {
		Message string `json:"message"`
	}

	AddGET[Response](w, "/hello", func(params ParamsManager, req *http.Request) (Response, *CustomResponse, error) {
		name := params.GetString("name", "world")
		return Response{Message: "hello " + name}, nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/hello?name=alice", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	var resp Response
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Message != "hello alice" {
		t.Errorf("message = %q, want %q", resp.Message, "hello alice")
	}
}

func TestRun_POSTRoute_JSONBody(t *testing.T) {
	w := setupController()

	type Input struct {
		Name string `json:"name"`
	}
	type Output struct {
		Greeting string `json:"greeting"`
	}

	AddJsonPOST[Input, Output](w, "/greet", func(st Input, params ParamsManager, req *http.Request) (Output, *CustomResponse, error) {
		return Output{Greeting: "Hi " + st.Name}, nil, nil
	})

	body := strings.NewReader(`{"name":"bob"}`)
	req := httptest.NewRequest(http.MethodPost, "/greet", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	var resp Output
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Greeting != "Hi bob" {
		t.Errorf("greeting = %q, want %q", resp.Greeting, "Hi bob")
	}
}

func TestRun_NoMatchingRoute(t *testing.T) {
	w := setupController()

	req := httptest.NewRequest(http.MethodGet, "/nothing", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if handled {
		t.Error("expected handled=false for unregistered route")
	}
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestRun_HandlerReturnsError(t *testing.T) {
	w := setupController()

	AddGET[string](w, "/fail", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "", nil, errors.New("something broke")
	})

	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled {
		t.Error("expected handled=true even when handler errors")
	}
	if err == nil {
		t.Error("expected error to be returned")
	}
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestRun_CustomResponse(t *testing.T) {
	w := setupController()

	AddGET[string](w, "/custom", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ignored", Custom().
			SetStatus(http.StatusCreated).
			SetBodyString("custom body").
			SetHeader("X-Custom", "yes"), nil
	})

	req := httptest.NewRequest(http.MethodGet, "/custom", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	if rr.Body.String() != "custom body" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "custom body")
	}
}

func TestRun_WithPathParams(t *testing.T) {
	w := setupController()

	AddGET[map[string]string](w, "/users/{id}", func(params ParamsManager, req *http.Request) (map[string]string, *CustomResponse, error) {
		return map[string]string{"id": params.GetString("id", "")}, nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	var resp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["id"] != "42" {
		t.Errorf("id = %q, want %q", resp["id"], "42")
	}
}

func TestRun_Middleware_ShortCircuit(t *testing.T) {
	w := setupController()

	middleware := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		return Custom().SetStatus(http.StatusForbidden).SetBodyString("blocked"), nil
	}

	AddGET[string](w, "/blocked", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		t.Error("handler should not be called when middleware short-circuits")
		return "should not reach", nil, nil
	}, middleware)

	req := httptest.NewRequest(http.MethodGet, "/blocked", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}
	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestRun_IOReaderResponse(t *testing.T) {
	w := setupController()

	AddGET[io.Reader](w, "/download", func(params ParamsManager, req *http.Request) (io.Reader, *CustomResponse, error) {
		return strings.NewReader("file content"),
			Custom().SetHeader("Content-Type", "application/octet-stream"), nil
	})

	req := httptest.NewRequest(http.MethodGet, "/download", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}
	if rr.Body.String() != "file content" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "file content")
	}
}

func TestRun_GetWithStruct_QueryParams(t *testing.T) {
	w := setupController()

	type Filter struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	type Output struct {
		Result string `json:"result"`
	}

	AddGetWithStruct(w, "/search", func(st Filter, params ParamsManager, req *http.Request) (Output, *CustomResponse, error) {
		return Output{Result: st.Name + ":" + st.Status}, nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/search?name=alice&status=active", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	var resp Output
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Result != "alice:active" {
		t.Errorf("result = %q, want %q", resp.Result, "alice:active")
	}
}

func TestRun_GetWithStruct_Validation(t *testing.T) {
	w := setupController()

	type Filter struct {
		Email string `json:"email" validate:"required"`
	}

	AddGetWithStruct(w, "/validated-get", func(st Filter, params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/validated-get", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled {
		t.Error("expected handled=true for validation error")
	}
	if err == nil {
		t.Error("expected error for validation failure")
	}
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestRun_GetWithStruct_WithPathParams(t *testing.T) {
	w := setupController()

	type Filter struct {
		Sort string `json:"sort"`
	}
	type Output struct {
		ID   string `json:"id"`
		Sort string `json:"sort"`
	}

	AddGetWithStruct(w, "/items/{id}", func(st Filter, params ParamsManager, req *http.Request) (Output, *CustomResponse, error) {
		return Output{ID: params.GetString("id", ""), Sort: st.Sort}, nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/items/99?sort=desc", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	var resp Output
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ID != "99" {
		t.Errorf("id = %q, want %q", resp.ID, "99")
	}
	if resp.Sort != "desc" {
		t.Errorf("sort = %q, want %q", resp.Sort, "desc")
	}
}

func TestRun_ValidationError(t *testing.T) {
	w := setupController()

	type Input struct {
		Email string `json:"email" validate:"required"`
	}

	AddJsonPOST[Input, string](w, "/validate", func(st Input, params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	})

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/validate", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled {
		t.Error("expected handled=true for validation error")
	}
	if err == nil {
		t.Error("expected error for validation failure")
	}
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

// --- Middleware and handler parameter matching tests ---

func TestMiddleware_ReceivesSameStruct_POST(t *testing.T) {
	w := setupController()

	type Input struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	type Output struct {
		OK bool `json:"ok"`
	}

	var middlewareValue any
	var middlewareParams ParamsManager
	var handlerStruct Input
	var handlerParams ParamsManager

	middleware := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		middlewareValue = value
		middlewareParams = params
		return nil, nil
	}

	AddJsonPOST[Input, Output](w, "/test-mw-struct-post", func(st Input, params ParamsManager, req *http.Request) (Output, *CustomResponse, error) {
		handlerStruct = st
		handlerParams = params
		return Output{OK: true}, nil, nil
	}, middleware)

	body := strings.NewReader(`{"name":"alice","value":42}`)
	req := httptest.NewRequest(http.MethodPost, "/test-mw-struct-post?extra=hello", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	// Middleware value should be the same concrete struct type as the handler receives
	middlewareInput, ok := middlewareValue.(Input)
	if !ok {
		t.Fatalf("middleware value type = %T, want Input", middlewareValue)
	}
	if middlewareInput != handlerStruct {
		t.Errorf("middleware struct = %+v, handler struct = %+v", middlewareInput, handlerStruct)
	}

	// Both should see the same query params
	if middlewareParams.GetString("extra", "") != handlerParams.GetString("extra", "") {
		t.Errorf("params mismatch: middleware extra=%q, handler extra=%q",
			middlewareParams.GetString("extra", ""), handlerParams.GetString("extra", ""))
	}
}

func TestMiddleware_ReceivesSameStruct_GETWithStruct(t *testing.T) {
	w := setupController()

	type Filter struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	type Output struct {
		OK bool `json:"ok"`
	}

	var middlewareValue any
	var middlewareParams ParamsManager
	var handlerStruct Filter
	var handlerParams ParamsManager

	middleware := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		middlewareValue = value
		middlewareParams = params
		return nil, nil
	}

	AddGetWithStruct(w, "/test-mw-struct-get", func(st Filter, params ParamsManager, req *http.Request) (Output, *CustomResponse, error) {
		handlerStruct = st
		handlerParams = params
		return Output{OK: true}, nil, nil
	}, middleware)

	req := httptest.NewRequest(http.MethodGet, "/test-mw-struct-get?name=alice&status=active", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	// Middleware value should be the same concrete struct type
	middlewareFilter, ok := middlewareValue.(Filter)
	if !ok {
		t.Fatalf("middleware value type = %T, want Filter", middlewareValue)
	}
	if middlewareFilter != handlerStruct {
		t.Errorf("middleware struct = %+v, handler struct = %+v", middlewareFilter, handlerStruct)
	}

	// Both should see the same query params
	if middlewareParams.GetString("name", "") != handlerParams.GetString("name", "") {
		t.Errorf("params 'name' mismatch: middleware=%q, handler=%q",
			middlewareParams.GetString("name", ""), handlerParams.GetString("name", ""))
	}
	if middlewareParams.GetString("status", "") != handlerParams.GetString("status", "") {
		t.Errorf("params 'status' mismatch: middleware=%q, handler=%q",
			middlewareParams.GetString("status", ""), handlerParams.GetString("status", ""))
	}
}

func TestMiddleware_ReceivesSameParams_GETSimple(t *testing.T) {
	w := setupController()

	var middlewareValue any
	var middlewareParams ParamsManager
	var handlerParams ParamsManager

	middleware := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		middlewareValue = value
		middlewareParams = params
		return nil, nil
	}

	AddGET[string](w, "/test-mw-params/{id}", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		handlerParams = params
		return "ok", nil, nil
	}, middleware)

	req := httptest.NewRequest(http.MethodGet, "/test-mw-params/99?color=blue", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	// Middleware and handler should see the same path params
	if middlewareParams.GetString("id", "") != handlerParams.GetString("id", "") {
		t.Errorf("path param 'id' mismatch: middleware=%q, handler=%q",
			middlewareParams.GetString("id", ""), handlerParams.GetString("id", ""))
	}

	// Middleware and handler should see the same query params
	if middlewareParams.GetString("color", "") != handlerParams.GetString("color", "") {
		t.Errorf("query param 'color' mismatch: middleware=%q, handler=%q",
			middlewareParams.GetString("color", ""), handlerParams.GetString("color", ""))
	}

	// For ParamsManager routes, middleware value should be ParamsManager (not reflect.Value)
	_, ok := middlewareValue.(ParamsManager)
	if !ok {
		t.Errorf("middleware value type = %T, want ParamsManager", middlewareValue)
	}
}

func TestMiddleware_ReceivesSameParams_POSTForm(t *testing.T) {
	w := setupController()

	var middlewareParams ParamsManager
	var handlerParams ParamsManager

	middleware := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		middlewareParams = params
		return nil, nil
	}

	AddFormPost[string](w, "/test-mw-form", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		handlerParams = params
		return "ok", nil, nil
	}, middleware)

	body := strings.NewReader("field1=hello&field2=world")
	req := httptest.NewRequest(http.MethodPost, "/test-mw-form", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	if middlewareParams.GetString("field1", "") != handlerParams.GetString("field1", "") {
		t.Errorf("param 'field1' mismatch: middleware=%q, handler=%q",
			middlewareParams.GetString("field1", ""), handlerParams.GetString("field1", ""))
	}
	if middlewareParams.GetString("field2", "") != handlerParams.GetString("field2", "") {
		t.Errorf("param 'field2' mismatch: middleware=%q, handler=%q",
			middlewareParams.GetString("field2", ""), handlerParams.GetString("field2", ""))
	}
}

// --- Pointer struct parameter tests ---

func TestMiddleware_PointerStruct_POST(t *testing.T) {
	w := setupController()

	type Input struct {
		Name string `json:"name"`
	}
	type Output struct {
		OK bool `json:"ok"`
	}

	var middlewareValue any
	var handlerSt *Input

	middleware := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		middlewareValue = value
		return nil, nil
	}

	AddJsonPOST[*Input, Output](w, "/test-ptr-post", func(st *Input, params ParamsManager, req *http.Request) (Output, *CustomResponse, error) {
		handlerSt = st
		return Output{OK: true}, nil, nil
	}, middleware)

	body := strings.NewReader(`{"name":"bob"}`)
	req := httptest.NewRequest(http.MethodPost, "/test-ptr-post", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	// Middleware should receive *Input (not Input, **Input, or reflect.Value)
	middlewarePtr, ok := middlewareValue.(*Input)
	if !ok {
		t.Fatalf("middleware value type = %T, want *Input", middlewareValue)
	}

	// Both should point to the same underlying struct
	if middlewarePtr != handlerSt {
		t.Error("middleware and handler received different pointers")
	}

	// Verify the data is correct
	if handlerSt == nil {
		t.Fatal("handler received nil pointer")
	}
	if handlerSt.Name != "bob" {
		t.Errorf("handler struct name = %q, want %q", handlerSt.Name, "bob")
	}
}

func TestHandler_PointerStruct_ReceivesPointer_POST(t *testing.T) {
	w := setupController()

	type Input struct {
		Name string `json:"name"`
	}
	type Output struct {
		OK bool `json:"ok"`
	}

	var handlerArgType reflect.Type
	var handlerArgKind reflect.Kind

	AddJsonPOST[*Input, Output](w, "/test-ptr-handler-post", func(st *Input, params ParamsManager, req *http.Request) (Output, *CustomResponse, error) {
		handlerArgType = reflect.TypeOf(st)
		handlerArgKind = reflect.ValueOf(st).Kind()
		return Output{OK: true}, nil, nil
	})

	body := strings.NewReader(`{"name":"charlie"}`)
	req := httptest.NewRequest(http.MethodPost, "/test-ptr-handler-post", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	// Handler should receive *Input (pointer), not Input (value) or **Input (double pointer)
	expectedType := reflect.TypeOf((*Input)(nil))
	if handlerArgType != expectedType {
		t.Errorf("handler arg type = %v, want %v", handlerArgType, expectedType)
	}
	if handlerArgKind != reflect.Pointer {
		t.Errorf("handler arg kind = %v, want %v", handlerArgKind, reflect.Pointer)
	}
}

func TestMiddleware_PointerStruct_GETWithStruct(t *testing.T) {
	w := setupController()

	type Filter struct {
		Name string `json:"name"`
	}
	type Output struct {
		OK bool `json:"ok"`
	}

	var middlewareValue any
	var handlerSt *Filter

	middleware := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		middlewareValue = value
		return nil, nil
	}

	AddGetWithStruct(w, "/test-ptr-get", func(st *Filter, params ParamsManager, req *http.Request) (Output, *CustomResponse, error) {
		handlerSt = st
		return Output{OK: true}, nil, nil
	}, middleware)

	req := httptest.NewRequest(http.MethodGet, "/test-ptr-get?name=dan", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	// Middleware should receive *Filter (not Filter, **Filter, or reflect.Value)
	middlewarePtr, ok := middlewareValue.(*Filter)
	if !ok {
		t.Fatalf("middleware value type = %T, want *Filter", middlewareValue)
	}

	// Both should point to the same underlying struct
	if middlewarePtr != handlerSt {
		t.Error("middleware and handler received different pointers")
	}

	// Verify the data is correct
	if handlerSt == nil {
		t.Fatal("handler received nil pointer")
	}
	if handlerSt.Name != "dan" {
		t.Errorf("handler struct name = %q, want %q", handlerSt.Name, "dan")
	}
}

func TestHandler_PointerStruct_ReceivesPointer_GETWithStruct(t *testing.T) {
	w := setupController()

	type Filter struct {
		Name string `json:"name"`
	}
	type Output struct {
		OK bool `json:"ok"`
	}

	var handlerArgType reflect.Type
	var handlerArgKind reflect.Kind

	AddGetWithStruct(w, "/test-ptr-handler-get", func(st *Filter, params ParamsManager, req *http.Request) (Output, *CustomResponse, error) {
		handlerArgType = reflect.TypeOf(st)
		handlerArgKind = reflect.ValueOf(st).Kind()
		return Output{OK: true}, nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test-ptr-handler-get?name=eve", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	// Handler should receive *Filter (pointer), not Filter (value) or **Filter (double pointer)
	expectedType := reflect.TypeOf((*Filter)(nil))
	if handlerArgType != expectedType {
		t.Errorf("handler arg type = %v, want %v", handlerArgType, expectedType)
	}
	if handlerArgKind != reflect.Pointer {
		t.Errorf("handler arg kind = %v, want %v", handlerArgKind, reflect.Pointer)
	}
}

func TestMiddleware_PointerStruct_NotDoublePointer(t *testing.T) {
	w := setupController()

	type Input struct {
		Name string `json:"name"`
	}
	type Output struct {
		OK bool `json:"ok"`
	}

	var middlewareValueType reflect.Type
	var handlerArgType reflect.Type

	middleware := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		middlewareValueType = reflect.TypeOf(value)
		return nil, nil
	}

	AddJsonPOST[*Input, Output](w, "/test-no-double-ptr", func(st *Input, params ParamsManager, req *http.Request) (Output, *CustomResponse, error) {
		handlerArgType = reflect.TypeOf(st)
		return Output{OK: true}, nil, nil
	}, middleware)

	body := strings.NewReader(`{"name":"eve"}`)
	req := httptest.NewRequest(http.MethodPost, "/test-no-double-ptr", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	expectedType := reflect.TypeOf((*Input)(nil))

	// Handler should receive *Input
	if handlerArgType != expectedType {
		t.Errorf("handler arg type = %v, want %v", handlerArgType, expectedType)
	}

	// Middleware should receive *Input (not **Input, Input, or reflect.Value)
	if middlewareValueType != expectedType {
		t.Errorf("middleware value type = %v, want %v", middlewareValueType, expectedType)
	}

	// Explicitly check it's not a double pointer
	if middlewareValueType != nil && middlewareValueType.Kind() == reflect.Pointer &&
		middlewareValueType.Elem().Kind() == reflect.Pointer {
		t.Error("middleware received **Input (double pointer), want *Input")
	}
}
