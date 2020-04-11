package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"

	"github.com/jvramirez/urlshort"
)

func main() {

	yamlFile := flag.String("yml", "", "yml file mapping reroutes initial path to url")
	jsonFile := flag.String("json", "", "json file mapping reroutes initial path to url")
	flag.Parse()

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	handler := urlshort.MapHandler(pathsToUrls, mux)

	// Build the YAMLHandler using the mapHandler as the fallback
	if *yamlFile != "" {
		fmt.Println("..reading in additional routes from ", *yamlFile)
		ymlHandler, err := createYMLHandler(*yamlFile, handler)
		if err != nil {
			panic(err)
		}

		handler = ymlHandler
	}

	// Build the JSONHandler using the mapHandler as the fallback
	if *jsonFile != "" {
		fmt.Println("..reading in additional routes from ", *jsonFile)
		jsonHandler, err := createJSONHandler(*jsonFile, handler)
		if err != nil {
			panic(err)
		}

		handler = jsonHandler
	}

	setupBoltDB()
	// Decorate handler with routes from BoltDB
	dbHandler, err := createBoltDBHandler(handler)
	if err != nil {
		panic(err)
	}
	handler = dbHandler

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", handler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func createYMLHandler(ymlFile string, fallback http.Handler) (http.HandlerFunc, error) {
	// 1. read yml file
	yml, err := ioutil.ReadFile(ymlFile)
	if err != nil {
		return nil, fmt.Errorf("Error: could not open file %s", ymlFile)
	}

	// 2. create yml handler
	yamlHandler, err := urlshort.YAMLHandler(yml, fallback)
	if err != nil {
		return nil, err
	}

	return yamlHandler, nil
}

func createJSONHandler(jsonFile string, fallback http.Handler) (http.HandlerFunc, error) {
	// 1. read json file
	json, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Error: could not open file %s", jsonFile)
	}

	// 2. create json handler
	jsonHandler, err := urlshort.JSONHandler(json, fallback)
	if err != nil {
		return nil, err
	}

	return jsonHandler, nil
}

func setupBoltDB() error {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open("routes.db", 0600, nil)
	defer db.Close()
	if err != nil {
		return err
	}

	// fill with data
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("MyRoutes"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		err = b.Put([]byte("/bolt-github"), []byte("https://github.com/boltdb/bolt"))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func createBoltDBHandler(fallback http.Handler) (http.HandlerFunc, error) {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open("routes.db", 0600, nil)
	defer db.Close()
	if err != nil {
		return nil, err
	}

	pathsToUrls := make(map[string]string)
	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("MyRoutes"))
		// take the (key,value) and put it in the map
		b.ForEach(func(key []byte, value []byte) error {
			pathsToUrls[string(key)] = string(value)
			return nil
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	// use maphandler to make db handler
	dbHandler := urlshort.MapHandler(pathsToUrls, fallback)

	return dbHandler, nil
}
