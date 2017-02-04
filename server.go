// zip-starter is a single-route app that only accepts JSON POST requests
// It's job is to create a "Starter" zip archive for adding data to, properly
// configured according to the guide outlined here:
// 		https://github.com/datarefugephilly/workflow/tree/master/harvesting-toolkit#3-generate-html-json--directory
//
// the archive derives it's name from the path posted to, or the "uuid" paremeter of the json document
//
// if a "url" key is specified in the json document, the server will fetch the document in question and write it
// to [name].html
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	// get port from enviornment, default to port 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Or single handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// log requests
		fmt.Printf("%s - %s - %s\n", r.Method, time.Now().String(), r.URL.Path)

		// support CORS from, well, anywhere.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			return
		}

		// only allow POST requests
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "this server only accepts JSON POST requests\n")
			return
		}

		// refuse anything that doesn't specify a json header
		if r.Header.Get("Content-Type") != "application/json" {
			fmt.Printf("Bad Content-Type: '%s'\n", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "this server only accepts JSON POST requests\n")
			return
		}

		// Decode JSON Body
		data := map[string]interface{}{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		// Determine the name of the archive, favouring the "uuid" param
		// if it exists in the json document
		name := strings.Trim(r.URL.Path, "/")
		if jsonName, ok := data["uuid"].(string); ok {
			name = jsonName
		}
		if jsonName, ok := data["UUID"].(string); ok {
			name = jsonName
		}
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "please provide the name of the zip as either the path you're posting to or a 'uuid' value in the json document\n")
			return
		}

		// allocate a buffer to hold our data & zip writer
		buf := &bytes.Buffer{}
		zw := zip.NewWriter(buf)

		// write the actual json file
		jsonFile, err := zw.Create(fmt.Sprintf("%s.json", name))
		d, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}
		jsonFile.Write(d)

		if _, err := FetchUrlIfExists(name, data, zw); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, fmt.Sprintf("error fetching url '%s': %s", jsonFile, err.Error()))
			return
		}

		// add data directory
		_, err = zw.Create("/data/")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}

		// add tools directory
		_, err = zw.Create("/tools/")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}

		// flush the writer to the buffer & close
		if err := zw.Flush(); err != nil {
			panic(err)
		}
		zw.Close()

		// write the finished zip archive as a response
		w.Header().Add("Content-Type", "application/zip, application/octet-stream")
		w.Header().Add("Content-Disposition", fmt.Sprintf("inline; filename=\"%s.zip\"", name))
		w.Write(buf.Bytes())
	})

	// spin up the server
	fmt.Println("zip-starter server starting on port ", port)
	http.ListenAndServe(":"+port, nil)
}

// FetchUrlIfExists looks at the "url" entry for the passed-in json. if it's a string, it checks to see
// if it's a valid url. If so it'll grab the url & qrite it to the passed-in zip writer at the specified
// name value with a .html extension
func FetchUrlIfExists(name string, data map[string]interface{}, zw *zip.Writer) (bool, error) {
	// if the json doc specifies a url value, grab it from the internet
	if jsonUrl, ok := data["url"].(string); ok {
		parsed, err := url.Parse(jsonUrl)
		if err != nil {
			return false, err
		}
		if !(parsed.Scheme == "http" || parsed.Scheme == "https") {
			return false, nil
		}

		res, err := http.Get(parsed.String())
		if err != nil {
			return false, err
		}

		htmlFile, err := zw.Create(fmt.Sprintf("%s.html", name))
		if err != nil {
			return false, err
		}
		defer res.Body.Close()

		_, err = io.Copy(htmlFile, res.Body)
		return true, err
	}

	return false, nil
}
