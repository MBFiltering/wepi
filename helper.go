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
