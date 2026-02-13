package wepi

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"reflect"
)

// readRequestValues parses the incoming request based on method and Content-Type.
// It returns a flat map of values (for ParamsManager), and optionally a reflect.Value
// containing the decoded struct (for typed routes).
//
// Routing logic:
//   - GET requests: parse query string into a flat map
//   - POST with Content-Type application/json: decode body into the expected struct type
//   - POST with form data: parse form values, then attempt to unmarshal into struct
//     by round-tripping through JSON (map -> JSON string -> struct)
func readRequestValues(req *http.Request, structType reflect.Type) (map[string]any, reflect.Value, error) {
	if req.Method == http.MethodGet {
		return GetURLQuery(req.URL.Query()), reflect.Value{}, nil
	}

	// JSON body: decode directly into the expected struct type via reflection
	if req.Header.Get("Content-Type") == "application/json" {
		stValue := reflect.New(structType)
		err := json.NewDecoder(req.Body).Decode(stValue.Interface())
		if err != nil {
			return nil, reflect.Value{}, err
		}
		return nil, stValue, nil
	}

	// Form-encoded body: parse into flat map
	values, err := GetPostFormValues(req)
	if err != nil {
		return nil, reflect.Value{}, err
	}

	// If the route expects a struct (not ParamsManager), try to convert the form
	// values map into the struct by marshaling to JSON and back. This allows form
	// data to populate typed structs with json tags.
	if structType != reflect.TypeOf((*ParamsManager)(nil)).Elem() {
		jsonstr, err := Jsonify(values)
		if err == nil {
			stValue := reflect.New(structType)
			err = json.Unmarshal([]byte(jsonstr), stValue.Interface())
			if err == nil {
				return values, stValue, nil
			}
		}
		if err != nil {
			log.Println(err)
		}
	}

	return values, reflect.Value{}, nil
}

// GetURLQuery converts url.Values into a flat map using the first value for each key.
func GetURLQuery(values url.Values) map[string]any {
	result := make(map[string]any)
	for key, value := range values {
		if len(value) > 0 {
			result[key] = value[0]
		}
	}
	return result
}

// GetPostFormValues parses the request form and returns values as a flat map.
func GetPostFormValues(req *http.Request) (map[string]any, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, err
	}
	return GetURLQuery(req.PostForm), nil
}

// GetReqJson decodes the request body as JSON into a map.
func GetReqJson(req *http.Request) (map[string]any, error) {
	var dat map[string]any
	err := json.NewDecoder(req.Body).Decode(&dat)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// Jsonify marshals a map to a JSON string.
func Jsonify(mp map[string]any) (string, error) {
	r, err := json.Marshal(mp)
	if err != nil {
		return "", err
	}
	return string(r), nil
}

// JsonifyPretty marshals a map to an indented JSON string.
func JsonifyPretty(mp map[string]any, preffix string, indent string) (string, error) {
	r, err := json.MarshalIndent(mp, preffix, indent)
	if err != nil {
		return "", err
	}
	return string(r), nil
}
