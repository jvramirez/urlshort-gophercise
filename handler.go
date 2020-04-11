package urlshort

import (
	"encoding/json"
	"net/http"

	"gopkg.in/yaml.v2"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		redirect, exist := pathsToUrls[r.URL.Path]

		if exist {
			http.Redirect(w, r, redirect, http.StatusFound)
			return
		}

		fallback.ServeHTTP(w, r)
	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(ymlBytes []byte, fallback http.Handler) (http.HandlerFunc, error) {
	ymlData, err := parseYml(ymlBytes)
	if err != nil {
		return nil, err
	}

	pathMap := createMap(ymlData)

	return MapHandler(pathMap, fallback), nil
}

func parseYml(ymlBytes []byte) ([]Entry, error) {
	var ymlData []Entry

	err := yaml.Unmarshal(ymlBytes, &ymlData)
	if err != nil {
		return nil, err
	}

	return ymlData, nil
}

// JSONHandler parses []byte of json data and creates redirect routers for included data
// otherwise defaults route to fallback handler
func JSONHandler(jsonBytes []byte, fallback http.Handler) (http.HandlerFunc, error) {

	jsonData, err := parseJSON(jsonBytes)
	if err != nil {
		return nil, err
	}

	pathMap := createMap(jsonData)

	return MapHandler(pathMap, fallback), nil
}

func parseJSON(jsonBytes []byte) ([]Entry, error) {
	var jsonData []Entry

	err := json.Unmarshal(jsonBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func createMap(entryData []Entry) map[string]string {
	pathMap := make(map[string]string)
	for _, item := range entryData {
		pathMap[item.Path] = item.URL
	}
	return pathMap
}

// Entry expected format of filereading entries
type Entry struct {
	Path string `yml:"path" json:"path"`
	URL  string `yml:"url" json:"url"`
}
