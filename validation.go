package wepi

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validatorSingleton = validator.New()

// validateAndExtractRouteFunc extracts the Handler function from a RouteHandler via reflection.
func validateAndExtractRouteFunc(route *Route) (handlerFunc reflect.Value, structType reflect.Type, err error) {
	if route.RouteHandler == nil {
		return reflect.Value{}, nil, errors.New("handler object is nil")
	}

	rhValue := reflect.ValueOf(route.RouteHandler)

	if rhValue.Kind() == reflect.Ptr {
		rhValue = rhValue.Elem()
	}

	if rhValue.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("invalid RouteHandler type %v", rhValue.Kind())
	}

	handlerFunc = rhValue.FieldByName("Handler")

	if !handlerFunc.IsValid() {
		return reflect.Value{}, nil, errors.New("handler not found in RouteHandler")
	}

	handlerType := handlerFunc.Type()

	if handlerType.NumIn() < 1 {
		return reflect.Value{}, nil, errors.New("handler function has insufficient parameters")
	}

	structType = handlerType.In(0)

	return handlerFunc, structType, nil
}

func getValidationError(er validator.FieldError, mainStruct any) string {
	switch er.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required", GetJSONFieldName(er, mainStruct))
	default:
		return fmt.Sprintf(
			"Field '%s', requires '%s' = '%s'",
			GetJSONFieldName(er, mainStruct), er.Tag(), er.Param(),
		)
	}
}

// GetJSONFieldName returns the JSON tag name for the struct field referenced by a
// validation error. Falls back to the Go field name if no json tag exists.
func GetJSONFieldName(e validator.FieldError, mainStruct any) (res string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			res = e.Field()
		}
	}()

	ns := strings.Split(e.StructNamespace(), ".")
	if len(ns) == 1 {
		return e.Field()
	}

	ns = ns[1:] // skip root struct name
	t := reflect.TypeOf(mainStruct)

	allNames := ""
	for _, name := range ns {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		field, ok := t.FieldByName(name)
		if !ok {
			return e.Field()
		}

		tag := field.Tag.Get("json")
		if tag == "" {
			tag = e.Field()
		}
		if allNames != "" {
			allNames += "."
		}
		allNames += strings.Split(tag, ",")[0]
		if name == ns[len(ns)-1] {
			return allNames
		}

		t = field.Type
		if t.Kind() != reflect.Struct {
			return e.Field()
		}
	}

	return e.Field()
}
