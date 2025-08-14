package wepi

import (
	"errors"
	"math"
	"github.com/MBFiltering/go-helpers/maphelper"
)

type ParamsManager struct {
	data       map[string]any
	additional map[string]any
}

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
	if p.HasKey(s) {
		fl, err := getFloat(p.data[s])
		if err != nil {
			return math.NaN()
		}
		return fl
	} else {
		return math.NaN()
	}

}

func (p ParamsManager) GetBool(key string) bool {
	return maphelper.GetBool(p.data, key)
}

func (p ParamsManager) GetFloat64(s string) (float64, error) {
	if p.HasKey(s) {
		return getFloat(p.data[s])
	} else {
		return 0, errors.New("key not found")
	}
}

func (p ParamsManager) GetInt64(s string) (int64, error) {
	if p.HasKey(s) {
		return getInt(p.data[s])
	} else {
		return 0, errors.New("key not found")
	}
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

// GetString return a string if it exists and is a string otherwise returns default
func (p ParamsManager) GetString(s string, def string) string {
	if !p.HasKey(s) {
		return def
	}
	str, ok := p.data[s].(string)
	if ok {
		return str
	} else {
		return def
	}
}

func (p ParamsManager) GetDataMap() map[string]any {
	return p.data
}
