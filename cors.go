package wepi

import "net/http"

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

func (w *WepiController) isOriginAllowed(origin string) bool {
	ok := w.cors[origin]
	if !ok && w.cors["*"] {
		return true
	}
	return ok
}
