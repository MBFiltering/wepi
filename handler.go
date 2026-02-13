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

// Run processes incoming HTTP requests through the wepi routing system.
// It matches the request path and method to a registered route, parses the request
// body into the appropriate type (JSON struct, form values, or query params),
// runs validation and middlewares, then calls the handler and writes the response.
// Returns (true, nil) if the route was handled, (false, nil) if no route matched.
func (w *WepiController) Run(pathHead string, req *http.Request, wr http.ResponseWriter) (bool, error) {
	// Strip the base path prefix so route matching works on the relative path
	path := strings.TrimPrefix(req.URL.Path, pathHead)

	// Treat PUT as POST — wepi routes only distinguish GET vs POST
	if req.Method == http.MethodPut {
		req.Method = http.MethodPost
	}

	// Handle CORS preflight (OPTIONS) requests before route matching
	if w.optionsInterceptor(path, wr, req) {
		return true, nil
	}

	// Look up the route by path and method; also extracts path template params (e.g. {id})
	path, route, pathParams := w.loadRouteFromRequest(path, req.Method)

	// No matching route found — return false so the caller can try other handlers
	if path == "" {
		return false, nil
	}

	if req.Method != route.method {
		log.Println("route " + route.route + " not same method " + req.Method)
		return false, errors.New("route " + route.route + " not same method " + req.Method)
	}

	// Extract the handler function and its first parameter type via reflection.
	// stType will be either a struct type (for JSON routes) or ParamsManager (for simple routes).
	handlerFunc, stType, err := validateAndExtractRouteFunc(route)
	if err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		return true, fmt.Errorf("error on route: "+route.route+", on path "+path+":", err)
	}

	// Parse the request body based on Content-Type.
	// For GET: returns query params as map. For JSON POST: decodes into struct.
	// For form POST: parses form values, optionally unmarshals into struct.
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

	// Merge URL path template parameters (e.g. {id}) into the params manager
	if pathParams != nil {
		for k, v := range pathParams {
			params.data[k] = v
		}
	}

	// Build the argument list for the handler function.
	// Simple routes (ParamsManager as first arg) get: (ParamsManager, *http.Request)
	// Struct routes get: (T, ParamsManager, *http.Request) with struct validation.
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
		// Struct route: validate the deserialized struct using go-playground/validator tags
		stValue = structValue

		validatorSingle := validatorSingleton
		validateValue := stValue.Elem()

		// Unwrap pointer if the struct type is a pointer-to-struct
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

	// Run middlewares in order. If a middleware returns a CustomResponse,
	// short-circuit: write that response and stop (don't call the handler).
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

	// Call the handler function via reflection.
	// Handler returns: (result, *CustomResponse, error)
	results := handlerFunc.Call(args)

	// Check the third return value (error) first — if present, fail immediately
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

	// Extract the optional CustomResponse (second return value).
	// Handlers use this to override status codes, headers, or the entire body.
	var custom *CustomResponse
	if len(results) > 1 && !results[1].IsNil() {
		custom = results[1].Interface().(*CustomResponse)
	} else {
		custom = nil
	}

	// Determine the response type from the first return value.
	// Supported types: string (text/html), io.Reader (file download),
	// struct/map (JSON), or nil (only valid if CustomResponse provides the body).
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

	// Apply CustomResponse overrides (headers, status, body) if provided
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

	// If the handler returned an io.Reader, stream it directly to the client.
	// Default to attachment disposition unless custom headers were set.
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

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
