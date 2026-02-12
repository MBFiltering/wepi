package wepi

import "sync"

// WepiController manages routes, path matching, and CORS configuration.
type WepiController struct {
	routes     sync.Map
	pathRegexs []*PathReader
	pathsMutex sync.Mutex
	header     string
	showErrors bool
	cors       map[string]bool
}

// Get creates a new WepiController instance which can be used to add routes.
func Get() *WepiController {
	return &WepiController{
		pathRegexs: make([]*PathReader, 0),
		cors:       make(map[string]bool),
	}
}

func (w *WepiController) AddRoutesHeader(header string) {
	w.header = header
}

func (w *WepiController) AddAllowedCORS(cors string) {
	w.cors[cors] = true
}

func (w *WepiController) isOriginAllowed(origin string) bool {
	ok := w.cors[origin]
	if !ok && w.cors["*"] {
		return true
	}
	return ok
}

// WepiComposedRoute holds a composed route to add to the controller.
type WepiComposedRoute struct {
	path   string
	route  *Route
	method string
}

func (w *WepiController) addRoute(converter *WepiComposedRoute) {
	w.addPattern(converter.path)
	w.routes.Store(converter.path+converter.method, converter.route)
}

func (w *WepiController) ShowErrors() bool {
	return w.showErrors
}

func (w *WepiController) SetShowErrors() {
	w.showErrors = true
}
