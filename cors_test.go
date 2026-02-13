package wepi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsOriginAllowed_ExactMatch(t *testing.T) {
	w := Get()
	w.AddAllowedCORS("https://example.com")

	if !w.isOriginAllowed("https://example.com") {
		t.Error("expected exact origin to be allowed")
	}
	if w.isOriginAllowed("https://other.com") {
		t.Error("expected non-listed origin to be rejected")
	}
}

func TestIsOriginAllowed_Wildcard(t *testing.T) {
	w := Get()
	w.AddAllowedCORS("*")

	if !w.isOriginAllowed("https://anything.com") {
		t.Error("expected any origin to be allowed with wildcard")
	}
}

func TestIsOriginAllowed_NoCORS(t *testing.T) {
	w := Get()
	if w.isOriginAllowed("https://example.com") {
		t.Error("expected no origins allowed when CORS not configured")
	}
}

func TestOptionsInterceptor_NoCORS(t *testing.T) {
	w := Get()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	rr := httptest.NewRecorder()

	handled := w.optionsInterceptor("/test", rr, req)
	if handled {
		t.Error("expected OPTIONS to not be handled when CORS not configured")
	}
}

func TestOptionsInterceptor_NotOptionsMethod(t *testing.T) {
	w := Get()
	w.AddAllowedCORS("*")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handled := w.optionsInterceptor("/test", rr, req)
	if handled {
		t.Error("expected non-OPTIONS request to not be handled")
	}
}

func TestOptionsInterceptor_ValidPreflight(t *testing.T) {
	w := Get()
	w.AddAllowedCORS("https://example.com")

	// Register a GET route so the interceptor finds a matching route
	AddGET[string](w, "/api/data", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	handled := w.optionsInterceptor("/api/data", rr, req)
	if !handled {
		t.Fatal("expected OPTIONS to be handled")
	}
	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("Allow-Origin = %q, want %q", got, "https://example.com")
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Allow-Methods header to be set")
	}
}

func TestOptionsInterceptor_DisallowedOrigin(t *testing.T) {
	w := Get()
	w.AddAllowedCORS("https://allowed.com")

	AddGET[string](w, "/api/data", func(params ParamsManager, req *http.Request) (string, *CustomResponse, error) {
		return "ok", nil, nil
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	handled := w.optionsInterceptor("/api/data", rr, req)
	if handled {
		t.Error("expected OPTIONS to not be handled for disallowed origin")
	}
}

func TestOptionsInterceptor_NoMatchingRoute(t *testing.T) {
	w := Get()
	w.AddAllowedCORS("*")

	req := httptest.NewRequest(http.MethodOptions, "/nonexistent", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	handled := w.optionsInterceptor("/nonexistent", rr, req)
	if handled {
		t.Error("expected OPTIONS to not be handled when no route matches")
	}
}
