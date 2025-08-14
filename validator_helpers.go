package wepi

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

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

// GetJSONFieldName returns the JSON tag for the struct field in the validation error
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

	ns = ns[1:]

	t := reflect.TypeOf(mainStruct)

	// Walk through nested fields
	allNames := ""
	for _, name := range ns {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		field, ok := t.FieldByName(name)
		if !ok {
			return e.Field()
		}

		// This is the field you want
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

		t = field.Type // drill down

		if t.Kind() != reflect.Struct {
			return e.Field() // Cannot go deeper
		}

	}
	return e.Field() // fallbac
}
