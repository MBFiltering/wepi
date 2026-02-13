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

// validateAndExtractRouteFunc uses reflection to extract the Handler function from a
// RouteHandler (either RouteHandlerWithStruct or RouteHandlerSimple). It returns:
//   - handlerFunc: the reflect.Value of the Handler field, ready to be called
//   - structType: the type of the first parameter (either a user struct T or ParamsManager),
//     which determines how the request body will be parsed
//   - err: if the RouteHandler is nil, not a struct, or missing a Handler field
func validateAndExtractRouteFunc(route *Route) (handlerFunc reflect.Value, structType reflect.Type, err error) {
	if route.RouteHandler == nil {
		return reflect.Value{}, nil, errors.New("handler object is nil")
	}

	// Unwrap pointer to get the underlying struct (RouteHandlerWithStruct or RouteHandlerSimple)
	rhValue := reflect.ValueOf(route.RouteHandler)

	if rhValue.Kind() == reflect.Ptr {
		rhValue = rhValue.Elem()
	}

	if rhValue.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("invalid RouteHandler type %v", rhValue.Kind())
	}

	// Look up the "Handler" field by name — this is the user-provided function
	handlerFunc = rhValue.FieldByName("Handler")

	if !handlerFunc.IsValid() {
		return reflect.Value{}, nil, errors.New("handler not found in RouteHandler")
	}

	handlerType := handlerFunc.Type()

	if handlerType.NumIn() < 1 {
		return reflect.Value{}, nil, errors.New("handler function has insufficient parameters")
	}

	// The first input parameter determines the route type:
	// - ParamsManager -> simple route (form/query only)
	// - any other type T -> struct route (JSON body decoded into T)
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
// validation error. It walks the struct namespace (e.g. "Outer.Inner.Field") using
// reflection to find each field's `json` tag, building a dot-separated JSON path
// like "outer.inner.field". Falls back to the Go field name if no json tag exists
// or if reflection panics (e.g. due to an embedded/anonymous type).
func GetJSONFieldName(e validator.FieldError, mainStruct any) (res string) {
	// Recover from any reflection panics and fall back to the Go field name
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			res = e.Field()
		}
	}()

	// Split "OuterStruct.InnerStruct.FieldName" into segments
	ns := strings.Split(e.StructNamespace(), ".")
	if len(ns) == 1 {
		return e.Field()
	}

	// Skip the root struct name (first segment) — we only need the field path
	ns = ns[1:]
	t := reflect.TypeOf(mainStruct)

	// Walk through each nested field, collecting json tag names
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
