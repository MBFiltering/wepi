package wepi

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestGetURLQuery(t *testing.T) {
	vals := url.Values{
		"name": {"alice"},
		"age":  {"30"},
		"tags": {"a", "b"},
	}
	result := GetURLQuery(vals)

	if result["name"] != "alice" {
		t.Errorf("name = %v, want alice", result["name"])
	}
	if result["tags"] != "a" {
		t.Errorf("tags = %v, want a (first value)", result["tags"])
	}
}

func TestGetPostFormValues(t *testing.T) {
	body := strings.NewReader("name=bob&age=25")
	req, _ := http.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	values, err := GetPostFormValues(req)
	if err != nil {
		t.Fatalf("GetPostFormValues error: %v", err)
	}
	if values["name"] != "bob" {
		t.Errorf("name = %v, want bob", values["name"])
	}
}

func TestGetPostFormValuesMultipart(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("name", "dave")
	writer.WriteField("age", "40")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/test", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	values, err := GetPostFormValues(req)
	if err != nil {
		t.Fatalf("GetPostFormValues multipart error: %v", err)
	}
	if values["name"] != "dave" {
		t.Errorf("name = %v, want dave", values["name"])
	}
	if values["age"] != "40" {
		t.Errorf("age = %v, want 40", values["age"])
	}
}

func TestGetReqJson(t *testing.T) {
	body := strings.NewReader(`{"name":"charlie","age":35}`)
	req, _ := http.NewRequest(http.MethodPost, "/test", body)

	result, err := GetReqJson(req)
	if err != nil {
		t.Fatalf("GetReqJson error: %v", err)
	}
	if result["name"] != "charlie" {
		t.Errorf("name = %v, want charlie", result["name"])
	}
}

func TestJsonify(t *testing.T) {
	mp := map[string]any{"key": "value"}
	result, err := Jsonify(mp)
	if err != nil {
		t.Fatalf("Jsonify error: %v", err)
	}
	if result != `{"key":"value"}` {
		t.Errorf("Jsonify = %q, want %q", result, `{"key":"value"}`)
	}
}
