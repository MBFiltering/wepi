package wepi

import (
	"math"
	"testing"
)

func TestGetParamsManager(t *testing.T) {
	data := map[string]any{"key": "value"}
	pm := GetParamsManager(data)
	if pm.data == nil {
		t.Fatal("data should not be nil")
	}
	if pm.additional == nil {
		t.Fatal("additional should not be nil")
	}
}

func TestHasKey(t *testing.T) {
	pm := GetParamsManager(map[string]any{"exists": "yes"})
	if !pm.HasKey("exists") {
		t.Error("expected HasKey to return true for existing key")
	}
	if pm.HasKey("missing") {
		t.Error("expected HasKey to return false for missing key")
	}
}

func TestGetString(t *testing.T) {
	pm := GetParamsManager(map[string]any{
		"name": "alice",
		"num":  42,
	})

	if got := pm.GetString("name", "default"); got != "alice" {
		t.Errorf("GetString(name) = %q, want %q", got, "alice")
	}
	if got := pm.GetString("missing", "default"); got != "default" {
		t.Errorf("GetString(missing) = %q, want %q", got, "default")
	}
	// Non-string value should return default
	if got := pm.GetString("num", "default"); got != "default" {
		t.Errorf("GetString(num) = %q, want %q", got, "default")
	}
}

func TestGetBoolMethod(t *testing.T) {
	pm := GetParamsManager(map[string]any{
		"flag_true":  true,
		"flag_false": false,
		"str_true":   "true",
		"str_false":  "false",
		"other":      "yes",
	})

	tests := []struct {
		key  string
		want bool
	}{
		{"flag_true", true},
		{"flag_false", false},
		{"str_true", true},
		{"str_false", false},
		{"other", false},
		{"missing", false},
	}
	for _, tt := range tests {
		if got := pm.GetBool(tt.key); got != tt.want {
			t.Errorf("GetBool(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestGetBoolStandalone(t *testing.T) {
	m := map[string]any{
		"b": true,
		"s": "true",
		"n": 123,
	}
	if !GetBool(m, "b") {
		t.Error("expected true for bool true")
	}
	if !GetBool(m, "s") {
		t.Error("expected true for string \"true\"")
	}
	if GetBool(m, "n") {
		t.Error("expected false for non-bool/non-string")
	}
	if GetBool(m, "missing") {
		t.Error("expected false for missing key")
	}
}

func TestGetFloat64(t *testing.T) {
	pm := GetParamsManager(map[string]any{
		"int":   42,
		"float": 3.14,
		"str":   "2.5",
	})

	f, err := pm.GetFloat64("int")
	if err != nil || f != 42.0 {
		t.Errorf("GetFloat64(int) = %v, %v; want 42.0, nil", f, err)
	}

	f, err = pm.GetFloat64("float")
	if err != nil || f != 3.14 {
		t.Errorf("GetFloat64(float) = %v, %v; want 3.14, nil", f, err)
	}

	f, err = pm.GetFloat64("str")
	if err != nil || f != 2.5 {
		t.Errorf("GetFloat64(str) = %v, %v; want 2.5, nil", f, err)
	}

	_, err = pm.GetFloat64("missing")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestGetFloat64OrNAN(t *testing.T) {
	pm := GetParamsManager(map[string]any{
		"valid":   1.5,
		"invalid": "not_a_number",
	})

	if got := pm.GetFloat64OrNAN("valid"); got != 1.5 {
		t.Errorf("GetFloat64OrNAN(valid) = %v, want 1.5", got)
	}
	if got := pm.GetFloat64OrNAN("missing"); !math.IsNaN(got) {
		t.Errorf("GetFloat64OrNAN(missing) = %v, want NaN", got)
	}
	if got := pm.GetFloat64OrNAN("invalid"); !math.IsNaN(got) {
		t.Errorf("GetFloat64OrNAN(invalid) = %v, want NaN", got)
	}
}

func TestGetInt64(t *testing.T) {
	pm := GetParamsManager(map[string]any{
		"int":   int64(99),
		"float": float64(42),
		"str":   "7",
	})

	i, err := pm.GetInt64("int")
	if err != nil || i != 99 {
		t.Errorf("GetInt64(int) = %v, %v; want 99, nil", i, err)
	}

	i, err = pm.GetInt64("float")
	if err != nil || i != 42 {
		t.Errorf("GetInt64(float) = %v, %v; want 42, nil", i, err)
	}

	i, err = pm.GetInt64("str")
	if err != nil || i != 7 {
		t.Errorf("GetInt64(str) = %v, %v; want 7, nil", i, err)
	}

	_, err = pm.GetInt64("missing")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestGetDataMap(t *testing.T) {
	data := map[string]any{"a": 1}
	pm := GetParamsManager(data)
	m := pm.GetDataMap()
	if m["a"] != 1 {
		t.Error("GetDataMap should return the underlying data map")
	}
}

func TestAdditionalData(t *testing.T) {
	pm := GetParamsManager(map[string]any{})
	pm.SetAdditionalData("extra", "value")
	if got := pm.GetAdditionalData("extra"); got != "value" {
		t.Errorf("GetAdditionalData(extra) = %v, want %q", got, "value")
	}
	if got := pm.GetAdditionalData("missing"); got != nil {
		t.Errorf("GetAdditionalData(missing) = %v, want nil", got)
	}
}

func TestGetFloat_AllNumericTypes(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want float64
	}{
		{"float64", float64(1.1), 1.1},
		{"float32", float32(2.2), float64(float32(2.2))},
		{"int64", int64(3), 3.0},
		{"int32", int32(4), 4.0},
		{"int16", int16(5), 5.0},
		{"int8", int8(6), 6.0},
		{"int", int(7), 7.0},
		{"uint64", uint64(8), 8.0},
		{"uint32", uint32(9), 9.0},
		{"uint16", uint16(10), 10.0},
		{"uint8", uint8(11), 11.0},
		{"uint", uint(12), 12.0},
		{"string_int", "13", 13.0},
		{"string_float", "14.5", 14.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFloat(tt.val)
			if err != nil {
				t.Fatalf("getFloat(%v) error: %v", tt.val, err)
			}
			if got != tt.want {
				t.Errorf("getFloat(%v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestGetFloat_InvalidString(t *testing.T) {
	_, err := getFloat("not_a_number")
	if err == nil {
		t.Error("expected error for invalid string")
	}
}

func TestGetInt_AllNumericTypes(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want int64
	}{
		{"float64", float64(1), 1},
		{"float32", float32(2), 2},
		{"int64", int64(3), 3},
		{"int32", int32(4), 4},
		{"int16", int16(5), 5},
		{"int8", int8(6), 6},
		{"int", int(7), 7},
		{"uint64", uint64(8), 8},
		{"uint32", uint32(9), 9},
		{"uint16", uint16(10), 10},
		{"uint8", uint8(11), 11},
		{"uint", uint(12), 12},
		{"string", "13", 13},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getInt(tt.val)
			if err != nil {
				t.Fatalf("getInt(%v) error: %v", tt.val, err)
			}
			if got != tt.want {
				t.Errorf("getInt(%v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestGetInt_InvalidString(t *testing.T) {
	_, err := getInt("not_a_number")
	if err == nil {
		t.Error("expected error for invalid string")
	}
}
