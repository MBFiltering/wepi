package wepi

import "net/http"

// optionsInterceptor handles CORS preflight (OPTIONS) requests.
// It checks if the requested path matches any registered GET or POST route,
// and if the origin is allowed, responds with the appropriate CORS headers
// and a 204 No Content status. Returns true if the request was handled.
func (wep *WepiController) optionsInterceptor(path string, w http.ResponseWriter, req *http.Request) bool {
	// CORS not configured — skip
	if len(wep.cors) == 0 {
		return false
	}

	// Only handle OPTIONS method (preflight)
	if req.Method != http.MethodOptions {
		return false
	}

	// Check if this path has a registered GET or POST route.
	// We need at least one match to confirm this is a valid endpoint worth
	// sending CORS headers for.
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

// isOriginAllowed checks if the given origin is in the CORS allow list.
// Returns true if the origin is explicitly listed, or if "*" (wildcard) is configured.
func (w *WepiController) isOriginAllowed(origin string) bool {
	ok := w.cors[origin]
	if !ok && w.cors["*"] {
		return true
	}
	return ok
}
