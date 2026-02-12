package wepi

import (
	"fmt"
	"reflect"
	"strconv"
)

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
