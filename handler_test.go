package wepi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestRun_GETRoute_StringResponse(t *testing.T) {
	w := setupController()

	AddGET[string](w, "/html", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "<h1>Hello</h1>", nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/html", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	if rr.Body.String() != "<h1>Hello</h1>" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "<h1>Hello</h1>")
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html" {
		t.Errorf("Content-Type = %q, want text/html", ct)
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

func TestRun_POSTRoute_FormBody(t *testing.T) {
	w := setupController()

	AddFormPost[map[string]string](w, "/form", func(params ParamsManager, req *http.Request) (map[string]string, *CustomResponse, error) {
		return map[string]string{"got": params.GetString("field", "")}, nil, nil
	})

	body := strings.NewReader("field=value123")
	req := httptest.NewRequest(http.MethodPost, "/form", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["got"] != "value123" {
		t.Errorf("got = %q, want %q", resp["got"], "value123")
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
	if rr.Header().Get("X-Custom") != "yes" {
		t.Errorf("X-Custom = %q, want %q", rr.Header().Get("X-Custom"), "yes")
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

func TestRun_WithPathHead(t *testing.T) {
	w := setupController()

	AddGET[string](w, "/endpoint", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/endpoint", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("/api/v1", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}
	if rr.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "ok")
	}
}

func TestRun_PUTTreatedAsPOST(t *testing.T) {
	w := setupController()

	type Input struct {
		Val string `json:"val"`
	}

	AddJsonPOST[Input, string](w, "/update", func(st Input, params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "updated: " + st.Val, nil, nil
	})

	body := strings.NewReader(`{"val":"hello"}`)
	req := httptest.NewRequest(http.MethodPut, "/update", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}
}

func TestRun_Middleware_PassThrough(t *testing.T) {
	w := setupController()

	middleware := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		// Pass through — don't block the request
		return nil, nil
	}

	AddGET[string](w, "/mw", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "reached handler", nil, nil
	}, middleware)

	req := httptest.NewRequest(http.MethodGet, "/mw", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}
	if rr.Body.String() != "reached handler" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "reached handler")
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
	if rr.Body.String() != "blocked" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "blocked")
	}
}

func TestRun_Middleware_Error(t *testing.T) {
	w := setupController()

	middleware := func(value any, params ParamsManager, req *http.Request) (*CustomResponse, error) {
		return nil, errors.New("middleware failure")
	}

	AddGET[string](w, "/mwerr", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "should not reach", nil, nil
	}, middleware)

	req := httptest.NewRequest(http.MethodGet, "/mwerr", nil)
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled {
		t.Error("expected handled=true")
	}
	if err == nil {
		t.Error("expected error from middleware")
	}
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestRun_IOReaderResponse_WithCustomHeaders(t *testing.T) {
	w := setupController()

	// Note: io.Reader responses require a CustomResponse with headers set,
	// otherwise the nil custom.headers check causes a panic (existing behavior).
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

func TestRun_CORSHeaders(t *testing.T) {
	w := setupController()
	w.AddAllowedCORS("https://example.com")

	AddGET[string](w, "/cors", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/cors", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	handled, err := w.Run("", req, rr)
	if !handled || err != nil {
		t.Fatalf("Run returned handled=%v, err=%v", handled, err)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("ACAO = %q, want %q", got, "https://example.com")
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

	// Send empty JSON object — "email" is required
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

func TestRun_ShowErrors(t *testing.T) {
	w := setupController()
	w.SetShowErrors()

	AddGET[string](w, "/err", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "", nil, errors.New("visible error")
	})

	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	rr := httptest.NewRecorder()

	w.Run("", req, rr)
	if !strings.Contains(rr.Body.String(), "visible error") {
		t.Error("expected error message in response body when ShowErrors is enabled")
	}
}
