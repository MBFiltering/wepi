package wepi

import (
	"encoding/json"
	"net/http"
	"net/url"
)

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

// GetBool returns true if the key exists and is a boolean true (or string "true"), otherwise false.
func GetBool(m map[string]any, s string) bool {
	b, ok := m[s].(bool)
	if ok {
		return b
	}
	// Check for string-encoded bool
	if str, ok := m[s].(string); ok {
		return str == "true"
	}
	return false
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
