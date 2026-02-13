package wepi

import (
	"log"
	"regexp"
	"strings"
)

// matcher is the regex group used to capture a single path segment (anything except '/')
var matcher = "([^/]+)"

// PathReader holds a compiled regex pattern for matching URL path templates.
type PathReader struct {
	keys    []string
	pattern string
	regex   *regexp.Regexp
}

func (w *WepiController) addPattern(path string) {
	w.pathsMutex.Lock()
	defer w.pathsMutex.Unlock()
	reg, keys := buildRegexFromTemplate(path)
	if len(keys) > 0 {
		w.addRegex(reg, keys, path)
	}
}

func (w *WepiController) addRegex(regex *regexp.Regexp, keys []string, pattern string) {
	w.pathRegexs = append(w.pathRegexs, &PathReader{regex: regex, keys: keys, pattern: pattern})
}

func (w *WepiController) checkPatternsForPath(path string) (map[string]string, string) {
	for _, pReader := range w.pathRegexs {
		m := extractPatternValues(pReader.regex, pReader.keys, path)
		if m != nil {
			return m, pReader.pattern
		}
	}
	return nil, ""
}

// loadRouteFromRequest tries to find a registered route for the given path and method.
// It first checks if the path matches any registered template patterns (e.g. "/users/{id}"),
// and if so, swaps the path for the template string used as the route key.
// Returns the resolved path, the Route, and any extracted path parameters.
func (w *WepiController) loadRouteFromRequest(path string, method string) (newPath string, _ *Route, pathPatternParams map[string]string) {
	pathPatternParams, foundPatternPath := w.checkPatternsForPath(path)

	if foundPatternPath != "" {
		// Use the template pattern as the route lookup key (e.g. "/users/{id}")
		path = foundPatternPath
	} else {
		pathPatternParams = nil
	}

	// Routes are stored as path+method (e.g. "/users/{id}GET")
	r, ok := w.routes.Load(path + method)
	if !ok {
		return "", nil, nil
	}

	return path, r.(*Route), pathPatternParams
}

// buildRegexFromTemplate converts a path template like "/users/{id}/posts/{postId}"
// into a compiled regex like "^/users/([^/]+)/posts/([^/]+)$" and returns the
// ordered list of parameter names ["id", "postId"].
// Returns (nil, nil) if the template has no placeholders or is invalid.
func buildRegexFromTemplate(template string) (*regexp.Regexp, []string) {
	re := regexp.MustCompile(`\{([^{}]+)\}`)
	var keys []string

	// Replace each {param} with a capturing group
	pattern := re.ReplaceAllStringFunc(template, func(s string) string {
		m := re.FindStringSubmatch(s)
		keys = append(keys, m[1])
		return matcher
	})

	// Anchor the pattern to match the full path
	pattern = "^" + pattern + "$"

	if len(keys) == 0 {
		return nil, nil
	}

	// Reject patterns with consecutive captures (e.g. "{a}{b}") — they're ambiguous
	if strings.Contains(pattern, matcher+matcher) {
		log.Println("not valid pattern: " + pattern + " for path: " + template)
		return nil, nil
	}

	compiledRe, err := regexp.Compile(pattern)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	return compiledRe, keys
}

// extractPatternValues matches a path against a compiled regex and returns a map
// of parameter names to their captured values. Returns nil if the path doesn't match.
func extractPatternValues(re *regexp.Regexp, keys []string, path string) map[string]string {
	match := re.FindStringSubmatch(path)
	if match == nil {
		return nil
	}
	values := make(map[string]string)
	for i, key := range keys {
		values[key] = match[i+1]
	}
	return values
}
