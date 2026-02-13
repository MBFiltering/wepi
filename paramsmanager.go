package wepi

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
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

// GetBool returns true if the key exists and is a boolean true (or string "true").
func GetBool(m map[string]any, s string) bool {
	b, ok := m[s].(bool)
	if ok {
		return b
	}
	if str, ok := m[s].(string); ok {
		return str == "true"
	}
	return false
}

var floatType = reflect.TypeOf(float64(0))
var stringType = reflect.TypeOf("")
var intType = reflect.TypeOf(int64(0))

func getFloat(unk any) (float64, error) {
	switch i := unk.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int16:
		return float64(i), nil
	case int8:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint16:
		return float64(i), nil
	case uint8:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		r, err := strconv.ParseInt(i, 0, 64)
		if err != nil {
			r, err := strconv.ParseFloat(i, 64)
			if err != nil {
				return 0, err
			}
			return r, nil
		}
		return getFloat(r)
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			r, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0, err
			}
			return r, nil
		}
		return 0, fmt.Errorf("value %v not convertible to float", i)
	}
}

func getInt(unk any) (int64, error) {
	switch i := unk.(type) {
	case float64:
		return int64(i), nil
	case float32:
		return int64(i), nil
	case int64:
		return i, nil
	case int32:
		return int64(i), nil
	case int16:
		return int64(i), nil
	case int8:
		return int64(i), nil
	case int:
		return int64(i), nil
	case uint64:
		return int64(i), nil
	case uint32:
		return int64(i), nil
	case uint16:
		return int64(i), nil
	case uint8:
		return int64(i), nil
	case uint:
		return int64(i), nil
	case string:
		r, err := strconv.ParseInt(i, 0, 64)
		if err != nil {
			return 0, err
		}
		return r, nil
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(intType) {
			fv := v.Convert(intType)
			return fv.Int(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			r, err := strconv.ParseInt(s, 0, 64)
			if err != nil {
				return 0, err
			}
			return r, nil
		}
		return 0, fmt.Errorf("value %v not convertible to int", i)
	}
}
