package wepi

import (
	"testing"
)

func TestCustom_Defaults(t *testing.T) {
	c := Custom()
	if c.status != 0 {
		t.Errorf("status = %d, want 0", c.status)
	}
	if c.body != nil {
		t.Error("body should be nil")
	}
	if c.headers != nil {
		t.Error("headers should be nil")
	}
}

func TestCustomResponse_SetStatus(t *testing.T) {
	c := Custom().SetStatus(201)
	if c.status != 201 {
		t.Errorf("status = %d, want 201", c.status)
	}
}

func TestCustomResponse_SetBody(t *testing.T) {
	c := Custom().SetBody([]byte("hello"))
	if string(c.body) != "hello" {
		t.Errorf("body = %q, want %q", string(c.body), "hello")
	}
}

func TestCustomResponse_SetBodyString(t *testing.T) {
	c := Custom().SetBodyString("world")
	if string(c.body) != "world" {
		t.Errorf("body = %q, want %q", string(c.body), "world")
	}
}

func TestCustomResponse_AddHeader(t *testing.T) {
	c := Custom().AddHeader("X-Test", "val1").AddHeader("X-Test", "val2")
	if c.headers == nil {
		t.Fatal("headers should not be nil")
	}
	vals := c.headers.Values("X-Test")
	if len(vals) != 2 || vals[0] != "val1" || vals[1] != "val2" {
		t.Errorf("X-Test values = %v, want [val1 val2]", vals)
	}
}

func TestCustomResponse_SetHeader(t *testing.T) {
	c := Custom().SetHeader("X-Test", "val1").SetHeader("X-Test", "val2")
	if c.headers == nil {
		t.Fatal("headers should not be nil")
	}
	vals := c.headers.Values("X-Test")
	// SetHeader replaces, so only the last value should remain
	if len(vals) != 1 || vals[0] != "val2" {
		t.Errorf("X-Test values = %v, want [val2]", vals)
	}
}

func TestCustomResponse_Chaining(t *testing.T) {
	c := Custom().
		SetStatus(404).
		SetBodyString("not found").
		SetHeader("Content-Type", "text/plain")

	if c.status != 404 {
		t.Errorf("status = %d, want 404", c.status)
	}
	if string(c.body) != "not found" {
		t.Errorf("body = %q, want %q", string(c.body), "not found")
	}
	if c.headers.Get("Content-Type") != "text/plain" {
		t.Errorf("Content-Type = %q, want text/plain", c.headers.Get("Content-Type"))
	}
}
