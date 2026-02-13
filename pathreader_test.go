package wepi

import (
	"regexp"
	"testing"
)

func TestBuildRegexFromTemplate_Simple(t *testing.T) {
	re, keys := buildRegexFromTemplate("/users/{id}")
	if re == nil {
		t.Fatal("expected non-nil regex")
	}
	if len(keys) != 1 || keys[0] != "id" {
		t.Errorf("keys = %v, want [id]", keys)
	}
	if !re.MatchString("/users/123") {
		t.Error("regex should match /users/123")
	}
	if re.MatchString("/users/123/extra") {
		t.Error("regex should not match /users/123/extra")
	}
}

func TestBuildRegexFromTemplate_MultipleParams(t *testing.T) {
	re, keys := buildRegexFromTemplate("/users/{userId}/posts/{postId}")
	if re == nil {
		t.Fatal("expected non-nil regex")
	}
	if len(keys) != 2 || keys[0] != "userId" || keys[1] != "postId" {
		t.Errorf("keys = %v, want [userId postId]", keys)
	}
	if !re.MatchString("/users/abc/posts/xyz") {
		t.Error("regex should match /users/abc/posts/xyz")
	}
}

func TestBuildRegexFromTemplate_NoParams(t *testing.T) {
	re, keys := buildRegexFromTemplate("/static/path")
	if re != nil || keys != nil {
		t.Error("expected nil for template with no params")
	}
}

func TestBuildRegexFromTemplate_ConsecutiveParams(t *testing.T) {
	// Consecutive params like {a}{b} are ambiguous and should be rejected
	re, keys := buildRegexFromTemplate("/path/{a}{b}")
	if re != nil || keys != nil {
		t.Error("expected nil for consecutive params (ambiguous)")
	}
}

func TestExtractPatternValues(t *testing.T) {
	re := regexp.MustCompile(`^/users/([^/]+)/posts/([^/]+)$`)
	keys := []string{"userId", "postId"}

	values := extractPatternValues(re, keys, "/users/alice/posts/42")
	if values == nil {
		t.Fatal("expected non-nil values")
	}
	if values["userId"] != "alice" {
		t.Errorf("userId = %q, want %q", values["userId"], "alice")
	}
	if values["postId"] != "42" {
		t.Errorf("postId = %q, want %q", values["postId"], "42")
	}
}

func TestExtractPatternValues_NoMatch(t *testing.T) {
	re := regexp.MustCompile(`^/users/([^/]+)$`)
	keys := []string{"id"}

	values := extractPatternValues(re, keys, "/other/path")
	if values != nil {
		t.Error("expected nil for non-matching path")
	}
}

func TestCheckPatternsForPath(t *testing.T) {
	w := Get()
	w.addPattern("/items/{id}")
	w.addPattern("/items/{id}/details/{detailId}")

	values, pattern := w.checkPatternsForPath("/items/99")
	if pattern != "/items/{id}" {
		t.Errorf("pattern = %q, want %q", pattern, "/items/{id}")
	}
	if values["id"] != "99" {
		t.Errorf("id = %q, want %q", values["id"], "99")
	}

	values, pattern = w.checkPatternsForPath("/items/99/details/5")
	if pattern != "/items/{id}/details/{detailId}" {
		t.Errorf("pattern = %q, want %q", pattern, "/items/{id}/details/{detailId}")
	}
	if values["id"] != "99" || values["detailId"] != "5" {
		t.Errorf("values = %v, want id=99 detailId=5", values)
	}

	values, pattern = w.checkPatternsForPath("/nothing")
	if pattern != "" || values != nil {
		t.Error("expected no match for unregistered path")
	}
}

func TestLoadRouteFromRequest(t *testing.T) {
	w := Get()

	// Register a route with a path pattern
	route := &Route{
		route:  "/users/{id}",
		method: GET,
	}
	w.addRoute(&WepiComposedRoute{
		path:   "/users/{id}",
		route:  route,
		method: GET,
	})

	path, r, params := w.loadRouteFromRequest("/users/42", GET)
	if path == "" {
		t.Fatal("expected a match")
	}
	if r != route {
		t.Error("expected the registered route")
	}
	if params["id"] != "42" {
		t.Errorf("id = %q, want %q", params["id"], "42")
	}

	// Non-matching method
	path, _, _ = w.loadRouteFromRequest("/users/42", POST)
	if path != "" {
		t.Error("expected no match for wrong method")
	}

	// Non-matching path
	path, _, _ = w.loadRouteFromRequest("/other", GET)
	if path != "" {
		t.Error("expected no match for unknown path")
	}
}

func TestLoadRouteFromRequest_StaticPath(t *testing.T) {
	w := Get()

	route := &Route{
		route:  "/health",
		method: GET,
	}
	w.addRoute(&WepiComposedRoute{
		path:   "/health",
		route:  route,
		method: GET,
	})

	path, r, params := w.loadRouteFromRequest("/health", GET)
	if path != "/health" {
		t.Errorf("path = %q, want %q", path, "/health")
	}
	if r != route {
		t.Error("expected the registered route")
	}
	if params != nil {
		t.Error("expected nil params for static path")
	}
}
