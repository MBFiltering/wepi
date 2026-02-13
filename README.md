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

var app = wepi.Get()

func main() {
    CreateRoutes()

    http.HandleFunc("/", func(wr http.ResponseWriter, req *http.Request) {
        app.Run("", req, wr)
    })
    http.ListenAndServe(":8080", nil)
}

func CreateRoutes() {
    app.AddAllowedCORS("*")

    wepi.AddGET(app, "/hello", HelloHandler, nil)
    wepi.AddJsonPOST(app, "/hello", PostHelloHandler, nil)
}

// HelloHandler returns a greeting using the name query parameter
func HelloHandler(params wepi.ParamsManager, req *http.Request) (map[string]string, *wepi.CustomResponse, error) {
    name := params.GetString("name", "world")
    return map[string]string{"message": "hello " + name}, nil, nil
}

type HelloInput struct {
    Name string `json:"name" validate:"required"`
}

// PostHelloHandler creates a greeting from a JSON body
func PostHelloHandler(st HelloInput, params wepi.ParamsManager, req *http.Request) (map[string]string, *wepi.CustomResponse, error) {
    return map[string]string{"message": "hello " + st.Name}, nil, nil
}
```

```bash
curl http://localhost:8080/hello?name=alice
# {"message":"hello alice"}

curl -X POST http://localhost:8080/hello -H "Content-Type: application/json" -d '{"name":"bob"}'
# {"message":"hello bob"}
```

## Registering Routes

### GET routes

```go
wepi.AddGET(controller, "/path", handler, middlewares...)
```

The handler receives a `ParamsManager` for accessing query parameters:

```go
// GetUserList returns a paginated list of users
func GetUserList(params wepi.ParamsManager, req *http.Request) (map[string]any, *wepi.CustomResponse, error) {
    page, _ := params.GetInt64("page")
    return map[string]any{"page": page}, nil, nil
}

wepi.AddGET(app, "/users", GetUserList, authMiddleware)
```

### POST routes with JSON body

```go
wepi.AddJsonPOST(controller, "/path", handler, middlewares...)
```

The request body is automatically deserialized into the input struct and validated using `go-playground/validator` tags:

```go
type CreateUser struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

// PostCreateUser creates a new user from the JSON body
func PostCreateUser(user CreateUser, params wepi.ParamsManager, req *http.Request) (map[string]string, *wepi.CustomResponse, error) {
    // user is already validated — Name and Email are guaranteed non-empty
    return map[string]string{"created": user.Name}, nil, nil
}

wepi.AddJsonPOST(app, "/users", PostCreateUser, authMiddleware)
```

If validation fails, wepi automatically returns `422 Unprocessable Entity` with a JSON error body listing each field violation.

### POST routes with form data

```go
wepi.AddFormPost(controller, "/path", handler, middlewares...)
```

Form values are accessed through `ParamsManager`:

```go
// PostLogin handles form-based login
func PostLogin(params wepi.ParamsManager, req *http.Request) (string, *wepi.CustomResponse, error) {
    username := params.GetString("username", "")
    return "welcome " + username, nil, nil
}

wepi.AddFormPost(app, "/login", PostLogin, nil)
```

## Path Parameters

Use `{param}` placeholders in route paths. Values are available via `ParamsManager`:

```go
// GetDeviceDetails returns details for a specific device
func GetDeviceDetails(params wepi.ParamsManager, req *http.Request) (map[string]string, *wepi.CustomResponse, error) {
    return map[string]string{"id": params.GetString("id", "")}, nil, nil
}

// PostSyncDevice triggers a sync command for a device
func PostSyncDevice(st SyncInput, params wepi.ParamsManager, req *http.Request) (map[string]string, *wepi.CustomResponse, error) {
    deviceID := params.GetString("id", "")
    return map[string]string{"synced": deviceID}, nil, nil
}

wepi.AddGET(app, "/device/{id}", GetDeviceDetails, authMiddleware)
wepi.AddJsonPOST(app, "/device/{id}/sync", PostSyncDevice, authMiddleware)
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
// PostCreateItem creates an item and returns 201 with a custom JSON body
func PostCreateItem(st ItemInput, params wepi.ParamsManager, req *http.Request) (string, *wepi.CustomResponse, error) {
    return "", wepi.Custom().
        SetStatus(http.StatusCreated).
        SetBodyString(`{"id": 1}`).
        SetHeader("Content-Type", "application/json"), nil
}

wepi.AddJsonPOST(app, "/items", PostCreateItem, authMiddleware)
```

Available methods: `SetStatus(int)`, `SetBody([]byte)`, `SetBodyString(string)`, `AddHeader(k, v)`, `SetHeader(k, v)`. All methods return `*CustomResponse` for chaining.

## Middleware

Middlewares run before the handler. Return `(*CustomResponse, error)`:
- Return `(nil, nil)` to pass through to the next middleware / handler
- Return `(*CustomResponse, nil)` to short-circuit and send that response
- Return `(nil, error)` to abort with 500

Pass `nil` for public routes that don't need middleware.

```go
// authMiddleware validates the request token and attaches user data to the context
func authMiddleware(value any, params wepi.ParamsManager, req *http.Request) (*wepi.CustomResponse, error) {
    token := req.Header.Get("Authorization")
    if token == "" {
        return wepi.Custom().SetStatus(http.StatusUnauthorized).SetBodyString("unauthorized"), nil
    }
    // Store data for the handler to use
    params.SetAdditionalData("userID", "123")
    return nil, nil
}

// GetProfile returns the authenticated user's profile
func GetProfile(params wepi.ParamsManager, req *http.Request) (map[string]string, *wepi.CustomResponse, error) {
    userID := params.GetAdditionalData("userID").(string)
    return map[string]string{"id": userID}, nil, nil
}

// Protected route — requires auth
wepi.AddGET(app, "/profile", GetProfile, authMiddleware)

// Public route — no auth needed
wepi.AddJsonPOST(app, "/contact", PostContact, nil)
```

## CORS

```go
app := wepi.Get()

// Allow a specific origin
app.AddAllowedCORS("https://example.com")

// Or allow all origins
app.AddAllowedCORS("*")
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
- Call `app.SetShowErrors()` to include error messages in response bodies (useful for development)

## Route Prefix

Strip a path prefix before route matching:

```go
// Routes are registered as "/users", but the full URL is "/api/v1/users"
http.HandleFunc("/api/v1/", func(wr http.ResponseWriter, req *http.Request) {
    app.Run("/api/v1", req, wr)
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
