package wepi

import (
	"errors"
	"math"
)

// ParamsManager provides convenient access to request parameters (query, form, path).
type ParamsManager struct {
	data       map[string]any
	additional map[string]any
}

// GetParamsManager creates a new ParamsManager from the given data map.
func GetParamsManager(data map[string]any) ParamsManager {
	return ParamsManager{
		data:       data,
		additional: make(map[string]any),
	}
}

func (p ParamsManager) HasKey(key string) bool {
	_, has := p.data[key]
	return has
}

func (p ParamsManager) GetFloat64OrNAN(s string) float64 {
	if !p.HasKey(s) {
		return math.NaN()
	}
	fl, err := getFloat(p.data[s])
	if err != nil {
		return math.NaN()
	}
	return fl
}

func (p ParamsManager) GetBool(key string) bool {
	return GetBool(p.data, key)
}

func (p ParamsManager) GetFloat64(s string) (float64, error) {
	if !p.HasKey(s) {
		return 0, errors.New("key not found")
	}
	return getFloat(p.data[s])
}

func (p ParamsManager) GetInt64(s string) (int64, error) {
	if !p.HasKey(s) {
		return 0, errors.New("key not found")
	}
	return getInt(p.data[s])
}

func (p ParamsManager) SetAdditionalData(key string, v any) {
	p.additional[key] = v
}

func (p ParamsManager) GetAdditionalData(key string) any {
	v, ok := p.additional[key]
	if ok {
		return v
	}
	return nil
}

// GetString returns the string value for key s, or def if not found or not a string.
func (p ParamsManager) GetString(s string, def string) string {
	if !p.HasKey(s) {
		return def
	}
	str, ok := p.data[s].(string)
	if !ok {
		return def
	}
	return str
}

func (p ParamsManager) GetDataMap() map[string]any {
	return p.data
}
