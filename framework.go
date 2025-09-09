package wepi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MBFiltering/go-helpers/maphelper"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validatorSingleton = validator.New()

// run the wepi handlers
func (w *WepiController) Run(pathHead string, req *http.Request, wr http.ResponseWriter) (bool, error) {

	//get path without prefix
	path := strings.TrimPrefix(req.URL.Path, pathHead)

	if req.Method == http.MethodPut {
		req.Method = http.MethodPost
	}

	if w.optionsInterceptor(path, wr, req) {
		return true, nil
	}

	//load route from pathreaders
	path, route, pathParams := w.loadRouteFromRequest(path, req.Method)

	//path unavailable
	if path == "" {
		return false, nil
	}

	//check for method, this must never fail
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

	//log.Println(stType)

	var args []reflect.Value
	var stValue reflect.Value

	if values == nil {
		values = make(map[string]any)
	}

	params := GetParamsManager(values)

	//add path parameters into data
	if pathParams != nil {
		for k, v := range pathParams {
			params.data[k] = v
		}
	}

	//first element in function is Param Manager meaning simple function type
	if stType == reflect.TypeOf((*ParamsManager)(nil)).Elem() {

		if hasStructBody {
			log.Println(errors.New("this request doesnt contain a params manager"))
			wr.WriteHeader(http.StatusInternalServerError)
			return true, errors.New("this request doesnt contain a params manager")
		}

		stValue = reflect.ValueOf(&params)
		// Prepare arguments
		args = []reflect.Value{
			stValue.Elem(),       //The paramsManager
			reflect.ValueOf(req), // *http.Request
		}

	} else {

		/*if !isBody {
			log.Println(errors.New("this request doesnt allow body"))
			wr.WriteHeader(http.StatusInternalServerError)
			return true
		}*/

		//struct value return from getParameters
		stValue = structValue

		//validate values
		validatorSingle := validatorSingleton
		validateValue := stValue.Elem()

		//check if value is pointer and get value
		if validateValue.Kind() == reflect.Pointer {
			validateValue = validateValue.Elem()
		}

		//check if value is struct to validate it
		if validateValue.Kind() == reflect.Struct {
			err = validatorSingle.Struct(validateValue.Interface())
			if err != nil {
				log.Println("Validator Error ", err)
				wr.WriteHeader(http.StatusUnprocessableEntity)
				msg := fmt.Sprint("Error parsing data: ", err)
				json, _ := maphelper.JsonifyPretty(map[string]any{"error": msg}, "", " ")
				if ve, ok := err.(validator.ValidationErrors); ok {
					list := make([]string, 0)
					for _, fe := range ve {
						list = append(list, getValidationError(fe, validateValue.Interface()))
					}
					json, _ = maphelper.JsonifyPretty(map[string]any{"error": "validation errors", "list": list}, "", " ")
				}

				wr.Write([]byte(json))
				return true, fmt.Errorf("validator Error: %v", msg)
			}
		}

		// Prepare arguments
		args = []reflect.Value{
			stValue.Elem(),                  // The struct value
			reflect.ValueOf(&params).Elem(), //the paramsManager
			reflect.ValueOf(req),            // *http.Request
		}

	}

	//prepare CORS

	if len(w.cors) > 0 {
		if w.isOriginAllowed(req.Header.Get("Origin")) {
			wr.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
			//wr.Header().Set("Access-Control-Allow-Origin", "*")
		}
	}

	//handle middlewares
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

	//log.Println(stValue)

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

	//get custom response from handler function
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

	if resultValue.IsValid() {
		//log.Println(resultValue.Interface())
	}

	var js []byte

	if !resultValue.IsValid() {
		if custom == nil {
			log.Println("Validator Error: ", errors.New("no data found on route return"))
			wr.WriteHeader(http.StatusInternalServerError)
			//wr.Write([]byte(fmt.Sprint("Error parsing data: ", err)))
			return true, errors.New("no data found on route return")
		}
	} else if resultValue.Kind() == reflect.String {
		wr.Header().Add("Content-Type", "text/html")
		js = []byte(resultValue.String())
	} else if _, ok := resultInterface.(io.Reader); ok {
		// just to jump else below
	} else {
		//we are not validating output fo now
		// if resultValue.Kind() == reflect.Struct {
		// 	validator := validatorSingleton
		// 	err = validator.Struct(resultValue.Interface())
		// 	if err != nil {
		// 		log.Println("Validator Error: ", err)
		// 		wr.WriteHeader(http.StatusInternalServerError)
		// 		//wr.Write([]byte(fmt.Sprint("Error parsing data: ", err)))
		// 		return true, fmt.Errorf("validator Error: %v", err)
		// 	}
		// }

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
		if len(custom.body) > 0 {
			body = custom.body
		}
		if custom.status != 0 {
			wr.WriteHeader(custom.status)
			status = custom.status
		}
	}

	if r, ok := resultInterface.(io.Reader); ok && (custom == nil || len(custom.body) == 0) {
		if rc, ok := resultInterface.(io.Closer); ok {
			defer rc.Close()
		}

		if custom == nil || custom.headers == nil {
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

	// Retrieve the RouteHandler
	rhValue := reflect.ValueOf(route.RouteHandler)

	// If rhValue is a pointer, get the element
	if rhValue.Kind() == reflect.Ptr {
		rhValue = rhValue.Elem()
	}

	// Ensure that we have a struct
	if rhValue.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("invalid RouteHandler type %v", rhValue.Kind())
	}

	// Get the Handler function
	handlerFunc = rhValue.FieldByName("Handler")

	// Ensure handlerFunc is valid
	if !handlerFunc.IsValid() {
		return reflect.Value{}, nil, errors.New("handler not found in RouteHandler")
	}

	// Get the type of the Handler function
	handlerType := handlerFunc.Type()

	//log.Println(handlerType)

	//check if function has parameters
	if handlerType.NumIn() < 1 {
		return reflect.Value{}, nil, errors.New("handler function has insufficient parameters")
	}

	// The first input parameter of the Handler function is of type T for struct or paramsManager for simple
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
	} else {

		if req.Header.Get("Content-Type") == "application/json" {

			stValue := reflect.New(structType)

			err := json.NewDecoder(req.Body).Decode(stValue.Interface())

			if err != nil {

				return nil, reflect.Value{}, err
			}

			return nil, stValue, nil

		} else {

			values, err := GetPostFormValues(req)

			if err != nil {

				return nil, reflect.Value{}, err
			}

			//generate map value if struct type is expected
			if structType != reflect.TypeOf((*ParamsManager)(nil)).Elem() {

				jsonstr, err := maphelper.Jsonify(values)
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

	}
}

/*func (wep *WepiController) optionsInterceptor(path string, w http.ResponseWriter, req *http.Request) bool {

	if len(wep.cors) == 0 {
		return false
	} else if req.Method == http.MethodOptions {
		// Always respond to OPTIONS requests with CORS headers
		// w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		w.WriteHeader(http.StatusNoContent)
		return true
	}

	return false
}*/

func (wep *WepiController) optionsInterceptor(path string, w http.ResponseWriter, req *http.Request) bool {
	if len(wep.cors) == 0 {
		return false
	} else if req.Method == http.MethodOptions {
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
				w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin")) // Change this to specific domains if needed
				//w.Header().Set("Access-Control-Allow-Origin", "*") // Change this to specific domains if needed
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

				w.WriteHeader(http.StatusNoContent)
				return true
			}
		}

	}

	return false

}
