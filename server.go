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
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Or single handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// log requests
		fmt.Printf("%s - %s - %s\n", r.Method, time.Now().String(), r.URL.Path)

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
		name := r.URL.Path
		if jsonName, ok := data["uuid"].(string); ok {
			name = jsonName
		}
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "please provide the name of the zip as the url path\n")
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

		// if the json doc specifies a url value, grab it from the internet
		if jsonUrl, ok := data["url"].(string); ok {
			res, err := http.Get(jsonUrl)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, fmt.Sprintf("error fetching url '%s': %s", jsonUrl, err.Error()))
				return
			}

			htmlFile, err := zw.Create(fmt.Sprintf("%s.html", name))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, err.Error())
				return
			}
			defer res.Body.Close()

			io.Copy(htmlFile, res.Body)
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
