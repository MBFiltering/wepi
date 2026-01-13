package wepi

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func GetURLQuery(values url.Values) map[string]any {
	result := make(map[string]any)
	for key, value := range values {
		if len(value) > 0 {
			result[key] = value[0]
		}
	}
	return result
}

func GetPostFormValues(req *http.Request) (map[string]any, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, err
	}
	result := GetURLQuery(req.PostForm)
	return result, nil
}

func GetReqJson(req *http.Request) (map[string]any, error) {
	var dat map[string]any
	err := json.NewDecoder(req.Body).Decode(&dat)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// GetBool return a true bool if it exists or is true otherwise false.
func GetBool(m map[string]any, s string) bool {
	b, ok := m[s].(bool)
	if ok {
		return b
	} else {
		//check for bad encoded bool
		if str, ok := m[s].(string); ok {
			return str == "true"
		}
		return false
	}
}

func Jsonify(mp map[string]any) (string, error) {
	r, err := json.Marshal(mp)

	if err != nil {
		return "", err
	}

	return string(r), nil
}

func JsonifyPretty(mp map[string]any, preffix string, indent string) (string, error) {
	r, err := json.MarshalIndent(mp, preffix, indent)

	if err != nil {
		return "", err
	}

	return string(r), nil
}
