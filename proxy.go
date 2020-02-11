package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

func skylinkProxy(
	w http.ResponseWriter,
	r *http.Request,
	siaPassword string,
	target string,
	copyRequestHeaders []string,
	copyResponseHeaders []string,
	replaceSkylinks bool,
) {
	rq, err := http.NewRequest(r.Method, target, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to prepare request: " + err.Error()))
		return
	}

	// Copy headers from the client to the server
	for _, headerName := range copyRequestHeaders {
		if headerValue := r.Header.Get(headerName); headerValue != "" {
			rq.Header.Set(headerName, headerValue)
		}
	}

	// Set the Sia user agent header
	rq.Header.Set("User-Agent", "Sia-Agent")

	// Add the Sia authentication
	rq.SetBasicAuth("", siaPassword)

	// Execute the request
	resp, err := http.DefaultClient.Do(rq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to execute request: " + err.Error()))
		return
	}
	defer resp.Body.Close()

	// Copy headers from the server to the client
	for _, headerName := range copyResponseHeaders {
		if headerValue := resp.Header.Get(headerName); headerValue != "" {
			w.Header().Set(headerName, headerValue)
		}
	}

	// If the attachment query parameter is set we overwrite the
	// Content-Disposition header with our own, allowing the file to be
	// downloaded directly in a web browser
	if _, ok := r.URL.Query()["attachment"]; ok {
		var name string
		_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
		if err != nil {
			name = "skynet_file"
		} else {
			name = params["filename"]
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(name))
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Range")
	w.Header().Set("Access-Control-Expose-Headers", "Accept-Ranges, Content-Encoding, Content-Length, Content-Range")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	w.WriteHeader(resp.StatusCode)

	length, _ := strconv.ParseUint(resp.Header.Get("Content-Length"), 10, 64)
	ctype := resp.Header.Get("Content-Type")

	// If the file is less than 16 MiB and it's a text file we'll replace all
	// skylinks with normal links pointing at this portal
	if replaceSkylinks && length != 0 && length < 1<<24 &&
		(strings.HasPrefix(ctype, "text/plain") ||
			strings.HasPrefix(ctype, "text/html") ||
			strings.HasPrefix(ctype, "text/css")) {

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to read response: " + err.Error()))
			return
		}

		w.Write(bytes.ReplaceAll(content, []byte("sia://"), []byte("/")))
	} else {
		// Copy the response from the server to the client
		io.Copy(w, resp.Body)
	}
}
