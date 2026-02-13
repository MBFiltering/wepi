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
	pm := GetParamsManager(map[string]any{"name": "alice", "num": 42})

	if got := pm.GetString("name", "default"); got != "alice" {
		t.Errorf("GetString(name) = %q, want %q", got, "alice")
	}
	if got := pm.GetString("missing", "default"); got != "default" {
		t.Errorf("GetString(missing) = %q, want %q", got, "default")
	}
}

func TestGetBool(t *testing.T) {
	pm := GetParamsManager(map[string]any{
		"flag_true": true,
		"str_true":  "true",
		"str_false": "false",
	})

	if !pm.GetBool("flag_true") {
		t.Error("expected true for bool true")
	}
	if !pm.GetBool("str_true") {
		t.Error("expected true for string \"true\"")
	}
	if pm.GetBool("str_false") {
		t.Error("expected false for string \"false\"")
	}
	if pm.GetBool("missing") {
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

	_, err = pm.GetFloat64("missing")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestGetFloat64OrNAN(t *testing.T) {
	pm := GetParamsManager(map[string]any{"valid": 1.5})

	if got := pm.GetFloat64OrNAN("valid"); got != 1.5 {
		t.Errorf("GetFloat64OrNAN(valid) = %v, want 1.5", got)
	}
	if got := pm.GetFloat64OrNAN("missing"); !math.IsNaN(got) {
		t.Errorf("GetFloat64OrNAN(missing) = %v, want NaN", got)
	}
}

func TestGetInt64(t *testing.T) {
	pm := GetParamsManager(map[string]any{
		"int": int64(99),
		"str": "7",
	})

	i, err := pm.GetInt64("int")
	if err != nil || i != 99 {
		t.Errorf("GetInt64(int) = %v, %v; want 99, nil", i, err)
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
