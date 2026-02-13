# wepi

A lightweight Go HTTP routing framework with generics, struct validation, CORS, middleware, and path parameters.

## Install

```bash
go get github.com/MBFiltering/wepi
```

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/MBFiltering/wepi"
)

func main() {
    w := wepi.Get()

    // GET route returning JSON
    wepi.AddGET[map[string]string](w, "/hello", func(params wepi.ParamsManager, req *http.Request) (map[string]string, *wepi.CustomResponse, error) {
        name := params.GetString("name", "world")
        return map[string]string{"message": "hello " + name}, nil, nil
    })

    http.HandleFunc("/", func(wr http.ResponseWriter, req *http.Request) {
        w.Run("", req, wr)
    })

    http.ListenAndServe(":8080", nil)
}
```

```bash
curl http://localhost:8080/hello?name=alice
# {"message":"hello alice"}
```

## Registering Routes

### GET routes

```go
wepi.AddGET[ResponseType](controller, "/path", handler, middlewares...)
```

The handler receives a `ParamsManager` for accessing query parameters:

```go
wepi.AddGET[map[string]any](w, "/users", func(params wepi.ParamsManager, req *http.Request) (map[string]any, *wepi.CustomResponse, error) {
    page, _ := params.GetInt64("page")
    return map[string]any{"page": page}, nil, nil
})
```

### POST routes with JSON body

```go
wepi.AddJsonPOST[InputStruct, ResponseType](controller, "/path", handler, middlewares...)
```

The request body is automatically deserialized into `InputStruct` and validated using `go-playground/validator` tags:

```go
type CreateUser struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

wepi.AddJsonPOST[CreateUser, map[string]string](w, "/users", func(user CreateUser, params wepi.ParamsManager, req *http.Request) (map[string]string, *wepi.CustomResponse, error) {
    // user is already validated — Name and Email are guaranteed non-empty
    return map[string]string{"created": user.Name}, nil, nil
})
```

If validation fails, wepi automatically returns `422 Unprocessable Entity` with a JSON error body listing each field violation.

### POST routes with form data

```go
wepi.AddFormPost[ResponseType](controller, "/path", handler, middlewares...)
```

Form values are accessed through `ParamsManager`:

```go
wepi.AddFormPost[string](w, "/login", func(params wepi.ParamsManager, req *http.Request) (string, *wepi.CustomResponse, error) {
    username := params.GetString("username", "")
    return "welcome " + username, nil, nil
})
```

## Path Parameters

Use `{param}` placeholders in route paths. Values are available via `ParamsManager`:

```go
wepi.AddGET[map[string]string](w, "/users/{id}/posts/{postId}", func(params wepi.ParamsManager, req *http.Request) (map[string]string, *wepi.CustomResponse, error) {
    return map[string]string{
        "userId": params.GetString("id", ""),
        "postId": params.GetString("postId", ""),
    }, nil, nil
})
```

## Response Types

Handlers return `(T, *CustomResponse, error)`. The response type is determined by `T`:

| Type | Behavior |
|---|---|
| `struct` / `map` | Serialized as JSON with `application/json` |
| `string` | Written as `text/html` |
| `io.Reader` | Streamed to client (file download) |

## Custom Responses

Use `CustomResponse` to override status codes, headers, or the body:

```go
wepi.AddGET[string](w, "/created", func(params wepi.ParamsManager, req *http.Request) (string, *wepi.CustomResponse, error) {
    return "", wepi.Custom().
        SetStatus(http.StatusCreated).
        SetBodyString(`{"id": 1}`).
        SetHeader("Content-Type", "application/json"), nil
})
```

Available methods: `SetStatus(int)`, `SetBody([]byte)`, `SetBodyString(string)`, `AddHeader(k, v)`, `SetHeader(k, v)`. All methods return `*CustomResponse` for chaining.

## Middleware

Middlewares run before the handler. Return `(*CustomResponse, error)`:
- Return `(nil, nil)` to pass through to the next middleware / handler
- Return `(*CustomResponse, nil)` to short-circuit and send that response
- Return `(nil, error)` to abort with 500

```go
authMiddleware := func(value any, params wepi.ParamsManager, req *http.Request) (*wepi.CustomResponse, error) {
    token := req.Header.Get("Authorization")
    if token == "" {
        return wepi.Custom().SetStatus(401).SetBodyString("unauthorized"), nil
    }
    // Store data for the handler to use
    params.SetAdditionalData("user", "alice")
    return nil, nil
}

wepi.AddGET[string](w, "/protected", func(params wepi.ParamsManager, req *http.Request) (string, *wepi.CustomResponse, error) {
    user := params.GetAdditionalData("user").(string)
    return "hello " + user, nil, nil
}, authMiddleware)
```

## CORS

```go
w := wepi.Get()

// Allow a specific origin
w.AddAllowedCORS("https://example.com")

// Or allow all origins
w.AddAllowedCORS("*")
```

When configured, wepi automatically handles `OPTIONS` preflight requests for any registered route, responding with `204 No Content` and the appropriate `Access-Control-Allow-*` headers.

## ParamsManager

`ParamsManager` provides typed access to query, form, and path parameters:

```go
params.GetString("key", "default")       // string with default
params.GetFloat64("key")                 // (float64, error)
params.GetFloat64OrNAN("key")            // float64 (NaN if missing/invalid)
params.GetInt64("key")                   // (int64, error)
params.GetBool("key")                    // bool (supports true/false and "true"/"false")
params.HasKey("key")                     // bool
params.GetDataMap()                      // map[string]any (raw data)
params.SetAdditionalData("key", value)   // store extra data (e.g. from middleware)
params.GetAdditionalData("key")          // retrieve extra data
```

## Error Handling

- **Validation errors** return `422` with a JSON body listing field-level errors
- **Handler errors** (third return value) return `500`
- Call `w.SetShowErrors()` to include error messages in response bodies (useful for development)

## Route Prefix

Strip a path prefix before route matching:

```go
// Routes are registered as "/users", but the full URL is "/api/v1/users"
http.HandleFunc("/api/v1/", func(wr http.ResponseWriter, req *http.Request) {
    w.Run("/api/v1", req, wr)
})
```

## PUT Handling

`PUT` requests are automatically treated as `POST` — register your route with `AddJsonPOST` or `AddFormPost` and it will handle both `POST` and `PUT`.

## Testing

```bash
go test -v ./...
```

## Project Structure

```
wepi.go             WepiController struct, constructor, configuration
handler.go          Run() — main request handling loop
request.go          Request parsing (JSON, form, query)
validation.go       Route handler extraction and struct validation
cors.go             CORS preflight and origin checking
composers.go        Route registration (AddGET, AddJsonPOST, AddFormPost)
customresponse.go   CustomResponse builder
paramsmanager.go    ParamsManager and type conversion
pathreader.go       URL path template matching
```
