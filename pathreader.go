package wepi

import (
	"log"
	"regexp"
	"strings"
)

// matcher to find
var matcher = "([^/]+)"

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

func (w *WepiController) loadRouteFromRequest(path string, method string) (newPath string, _ *Route, pathPatternParams map[string]string) {

	pathPatternParams, foundPatternPath := w.checkPatternsForPath(path)

	if foundPatternPath != "" {
		path = foundPatternPath
	} else {
		pathPatternParams = nil
	}

	//load path from sync map
	r, ok := w.routes.Load(path + method)

	if !ok {
		return "", nil, nil
	}

	return path, r.(*Route), pathPatternParams

}

func buildRegexFromTemplate(template string) (*regexp.Regexp, []string) {
	re := regexp.MustCompile(`\{([^{}]+)\}`)
	var keys []string

	// Replace placeholders with regex groups
	pattern := re.ReplaceAllStringFunc(template, func(s string) string {
		m := re.FindStringSubmatch(s)
		keys = append(keys, m[1])
		return matcher
	})

	// Compile the pattern
	pattern = "^" + pattern + "$"

	if len(keys) == 0 {
		return nil, nil
	}

	if strings.Contains(pattern, matcher+matcher) {
		log.Println("not valid pattern: " + pattern + " for path: " + template)
		return nil, nil
	}

	// log.Printf("Adding Path:"+template+" pattern: "+pattern+" for keys: %v", keys)
	compiledRe, err := regexp.Compile(pattern)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	return compiledRe, keys
}

func extractPatternValues(re *regexp.Regexp, keys []string, path string) map[string]string {
	match := re.FindStringSubmatch(path)
	if match == nil {
		return nil // No match found
	}
	values := make(map[string]string)
	for i, key := range keys {
		values[key] = match[i+1] // match[0] is the full match
	}
	return values
}
