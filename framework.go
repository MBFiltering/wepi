package wepi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validatorSingleton = validator.New()

// Run processes incoming HTTP requests through the wepi routing system.
func (w *WepiController) Run(pathHead string, req *http.Request, wr http.ResponseWriter) (bool, error) {
	path := strings.TrimPrefix(req.URL.Path, pathHead)

	if req.Method == http.MethodPut {
		req.Method = http.MethodPost
	}

	if w.optionsInterceptor(path, wr, req) {
		return true, nil
	}

	path, route, pathParams := w.loadRouteFromRequest(path, req.Method)

	if path == "" {
		return false, nil
	}

	if req.Method != route.method {
		log.Println("route " + route.route + " not same method " + req.Method)
		return false, errors.New("route " + route.route + " not same method " + req.Method)
	}

	handlerFunc, stType, err := validateAndExtractRouteFunc(route)
	if err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		return true, fmt.Errorf("error on route: "+route.route+", on path "+path+":", err)
	}

	values, structValue, err := readRequestValues(req, stType)
	if err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		if w.ShowErrors() {
			wr.Write([]byte(err.Error()))
		}
		return true, err
	}

	hasStructBody := structValue.IsValid()

	var args []reflect.Value
	var stValue reflect.Value

	if values == nil {
		values = make(map[string]any)
	}

	params := GetParamsManager(values)

	if pathParams != nil {
		for k, v := range pathParams {
			params.data[k] = v
		}
	}

	// First element in function is ParamsManager meaning simple function type
	if stType == reflect.TypeOf((*ParamsManager)(nil)).Elem() {
		if hasStructBody {
			log.Println(errors.New("this request doesnt contain a params manager"))
			wr.WriteHeader(http.StatusInternalServerError)
			return true, errors.New("this request doesnt contain a params manager")
		}

		stValue = reflect.ValueOf(&params)
		args = []reflect.Value{
			stValue.Elem(),       // The paramsManager
			reflect.ValueOf(req), // *http.Request
		}
	} else {
		stValue = structValue

		validatorSingle := validatorSingleton
		validateValue := stValue.Elem()

		if validateValue.Kind() == reflect.Pointer {
			validateValue = validateValue.Elem()
		}

		if validateValue.Kind() == reflect.Struct {
			err = validatorSingle.Struct(validateValue.Interface())
			if err != nil {
				log.Println("Validator Error ", err)
				wr.WriteHeader(http.StatusUnprocessableEntity)
				msg := fmt.Sprint("Error parsing data: ", err)
				json, _ := JsonifyPretty(map[string]any{"error": msg}, "", " ")
				if ve, ok := err.(validator.ValidationErrors); ok {
					list := make([]string, 0)
					for _, fe := range ve {
						list = append(list, getValidationError(fe, validateValue.Interface()))
					}
					json, _ = JsonifyPretty(map[string]any{"error": "validation errors", "list": list}, "", " ")
				}

				wr.Write([]byte(json))
				return true, fmt.Errorf("validator Error: %v", msg)
			}
		}

		args = []reflect.Value{
			stValue.Elem(),                  // The struct value
			reflect.ValueOf(&params).Elem(), // The paramsManager
			reflect.ValueOf(req),            // *http.Request
		}
	}

	// Set CORS headers
	if len(w.cors) > 0 {
		if w.isOriginAllowed(req.Header.Get("Origin")) {
			wr.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		}
	}

	// Handle middlewares
	for _, middleware := range route.Middlewares {
		if middleware != nil {
			c, err := middleware(stValue.Elem(), params, req)
			if err != nil {
				wr.WriteHeader(http.StatusInternalServerError)
				return true, err
			}

			if c != nil {
				if c.headers != nil {
					copyHeader(wr.Header(), c.headers)
				}
				if c.status != 0 {
					wr.WriteHeader(c.status)
				}
				if len(c.body) > 0 {
					wr.Write(c.body)
				}
				return true, nil
			}
		}
	}

	// Call the handler function
	results := handlerFunc.Call(args)

	// Check for error
	if len(results) > 2 && !results[2].IsNil() {
		err := results[2].Interface().(error)
		log.Println("Handler returned error:", err)
		wr.WriteHeader(http.StatusInternalServerError)
		wr.Write([]byte(err.Error()))
		if w.ShowErrors() {
			wr.Write([]byte(err.Error()))
		}
		return true, fmt.Errorf("handler returned error: %v", err)
	}

	// Get custom response from handler function
	var custom *CustomResponse
	if len(results) > 1 && !results[1].IsNil() {
		custom = results[1].Interface().(*CustomResponse)
	} else {
		custom = nil
	}

	resultInterface := results[0].Interface()
	resultValue := reflect.ValueOf(resultInterface)

	if resultValue.Kind() == reflect.Ptr {
		resultValue = resultValue.Elem()
	}

	var js []byte

	if !resultValue.IsValid() {
		if custom == nil {
			log.Println("Validator Error: ", errors.New("no data found on route return"))
			wr.WriteHeader(http.StatusInternalServerError)
			return true, errors.New("no data found on route return")
		}
	} else if resultValue.Kind() == reflect.String {
		wr.Header().Add("Content-Type", "text/html")
		js = []byte(resultValue.String())
	} else if _, ok := resultInterface.(io.Reader); ok {
		// io.Reader handled below
	} else {
		js, err = json.Marshal(resultValue.Interface())
		if err != nil {
			log.Println("Error writing data: ", err)
			wr.Write([]byte(fmt.Sprint("error writing data: ", err)))
			wr.WriteHeader(http.StatusBadRequest)
			return true, fmt.Errorf("error writing data: %v", err)
		}
		wr.Header().Add("Content-Type", "application/json")
	}

	body := []byte(js)
	status := 200

	if custom != nil {
		if custom.headers != nil {
			wr.Header().Del("Content-Type")
			copyHeader(wr.Header(), custom.headers)
		}
		if custom.status != 0 {
			wr.WriteHeader(custom.status)
			status = custom.status
		}
		if len(custom.body) > 0 {
			body = custom.body
			_, err = wr.Write(body)
			return true, err
		}
	}

	if r, ok := resultInterface.(io.Reader); ok {
		if rc, ok := resultInterface.(io.Closer); ok {
			defer rc.Close()
		}

		if custom.headers == nil {
			wr.Header().Set("Content-Disposition", `attachment; filename="file"`)
		}

		io.Copy(wr, r)
		return true, nil
	}

	log.Printf(req.URL.Path+" response status code: %v", status)

	wr.Write(body)

	return true, nil
}

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

	// The first input parameter is of type T for struct or ParamsManager for simple
	structType = handlerType.In(0)

	return handlerFunc, structType, nil
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func readRequestValues(req *http.Request, structType reflect.Type) (map[string]any, reflect.Value, error) {
	if req.Method == http.MethodGet {
		return GetURLQuery(req.URL.Query()), reflect.Value{}, nil
	}

	if req.Header.Get("Content-Type") == "application/json" {
		stValue := reflect.New(structType)
		err := json.NewDecoder(req.Body).Decode(stValue.Interface())
		if err != nil {
			return nil, reflect.Value{}, err
		}
		return nil, stValue, nil
	}

	values, err := GetPostFormValues(req)
	if err != nil {
		return nil, reflect.Value{}, err
	}

	// Generate map value if struct type is expected
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

func (wep *WepiController) optionsInterceptor(path string, w http.ResponseWriter, req *http.Request) bool {
	if len(wep.cors) == 0 {
		return false
	}

	if req.Method != http.MethodOptions {
		return false
	}

	returnCors := false
	pathFound, _, _ := wep.loadRouteFromRequest(path, http.MethodGet)
	if pathFound != "" {
		returnCors = true
	}
	if !returnCors {
		pathFound, _, _ = wep.loadRouteFromRequest(path, http.MethodPost)
		if pathFound != "" {
			returnCors = true
		}
	}

	if returnCors {
		if wep.isOriginAllowed(req.Header.Get("Origin")) {
			w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.WriteHeader(http.StatusNoContent)
			return true
		}
	}

	return false
}
