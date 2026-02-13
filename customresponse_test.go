package wepi

import (
	"testing"
)

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

func TestCustomResponse_AddHeader(t *testing.T) {
	c := Custom().AddHeader("X-Test", "val1").AddHeader("X-Test", "val2")
	vals := c.headers.Values("X-Test")
	if len(vals) != 2 || vals[0] != "val1" || vals[1] != "val2" {
		t.Errorf("X-Test values = %v, want [val1 val2]", vals)
	}
}
