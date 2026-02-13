package wepi

import (
	"net/http"
	"net/url"
	"reflect"
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
	if result["age"] != "30" {
		t.Errorf("age = %v, want 30", result["age"])
	}
	// Only the first value should be taken for multi-valued keys
	if result["tags"] != "a" {
		t.Errorf("tags = %v, want a", result["tags"])
	}
}

func TestGetURLQuery_Empty(t *testing.T) {
	result := GetURLQuery(url.Values{})
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
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
	if values["age"] != "25" {
		t.Errorf("age = %v, want 25", values["age"])
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
	if result["age"] != float64(35) {
		t.Errorf("age = %v, want 35", result["age"])
	}
}

func TestGetReqJson_Invalid(t *testing.T) {
	body := strings.NewReader(`not json`)
	req, _ := http.NewRequest(http.MethodPost, "/test", body)

	_, err := GetReqJson(req)
	if err == nil {
		t.Error("expected error for invalid JSON")
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

func TestJsonifyPretty(t *testing.T) {
	mp := map[string]any{"key": "value"}
	result, err := JsonifyPretty(mp, "", "  ")
	if err != nil {
		t.Fatalf("JsonifyPretty error: %v", err)
	}
	expected := "{\n  \"key\": \"value\"\n}"
	if result != expected {
		t.Errorf("JsonifyPretty = %q, want %q", result, expected)
	}
}

func TestReadRequestValues_GET(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/test?name=alice&age=30", nil)

	values, structVal, err := readRequestValues(req, reflect.TypeOf((*ParamsManager)(nil)).Elem())
	if err != nil {
		t.Fatalf("readRequestValues error: %v", err)
	}
	if values["name"] != "alice" {
		t.Errorf("name = %v, want alice", values["name"])
	}
	if structVal.IsValid() {
		t.Error("expected invalid structVal for GET request")
	}
}

func TestReadRequestValues_JSONPost(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	body := strings.NewReader(`{"name":"bob","age":25}`)
	req, _ := http.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")

	values, structVal, err := readRequestValues(req, reflect.TypeOf(TestStruct{}))
	if err != nil {
		t.Fatalf("readRequestValues error: %v", err)
	}
	if values != nil {
		t.Error("expected nil values for JSON POST")
	}
	if !structVal.IsValid() {
		t.Fatal("expected valid structVal for JSON POST")
	}

	ts := structVal.Elem().Interface().(TestStruct)
	if ts.Name != "bob" || ts.Age != 25 {
		t.Errorf("got %+v, want {Name:bob Age:25}", ts)
	}
}

func TestReadRequestValues_FormPost_SimpleRoute(t *testing.T) {
	body := strings.NewReader("name=charlie&age=35")
	req, _ := http.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	values, structVal, err := readRequestValues(req, reflect.TypeOf((*ParamsManager)(nil)).Elem())
	if err != nil {
		t.Fatalf("readRequestValues error: %v", err)
	}
	if values["name"] != "charlie" {
		t.Errorf("name = %v, want charlie", values["name"])
	}
	if structVal.IsValid() {
		t.Error("expected invalid structVal for ParamsManager route")
	}
}

func TestReadRequestValues_FormPost_StructRoute(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  string `json:"age"`
	}

	body := strings.NewReader("name=dave&age=40")
	req, _ := http.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	values, structVal, err := readRequestValues(req, reflect.TypeOf(TestStruct{}))
	if err != nil {
		t.Fatalf("readRequestValues error: %v", err)
	}
	if values["name"] != "dave" {
		t.Errorf("name = %v, want dave", values["name"])
	}
	if !structVal.IsValid() {
		t.Fatal("expected valid structVal for struct route with form data")
	}

	ts := structVal.Elem().Interface().(TestStruct)
	if ts.Name != "dave" || ts.Age != "40" {
		t.Errorf("got %+v, want {Name:dave Age:40}", ts)
	}
}

func TestReadRequestValues_InvalidJSON(t *testing.T) {
	body := strings.NewReader(`{invalid}`)
	req, _ := http.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")

	type Dummy struct{}
	_, _, err := readRequestValues(req, reflect.TypeOf(Dummy{}))
	if err == nil {
		t.Error("expected error for invalid JSON body")
	}
}
