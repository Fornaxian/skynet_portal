package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type proxy struct {
	siaPassword         string
	siadURL             string
	copyRequestHeaders  []string
	copyResponseHeaders []string
}

func (p proxy) siadRequest(
	w http.ResponseWriter,
	r *http.Request,
	method string,
	target string,
	body io.ReadCloser,
) (
	resp *http.Response,
	err error,
) {
	rq, err := http.NewRequest(method, p.siadURL+target, body)
	if err != nil {
		return nil, err
	}

	// Copy headers from the client to the server
	for _, headerName := range p.copyRequestHeaders {
		if headerValue := r.Header.Get(headerName); headerValue != "" {
			rq.Header.Set(headerName, headerValue)
		}
	}

	// Set the Sia user agent header
	rq.Header.Set("User-Agent", "Sia-Agent")

	// Add the Sia authentication
	rq.SetBasicAuth("", p.siaPassword)

	if resp, err = http.DefaultClient.Do(rq); err != nil {
		return nil, err
	}

	// Copy headers from the server to the client
	for _, headerName := range p.copyResponseHeaders {
		if headerValue := resp.Header.Get(headerName); headerValue != "" {
			w.Header().Set(headerName, headerValue)
		}
	}

	return resp, nil
}

func (p proxy) getSkylinkProxy(
	w http.ResponseWriter,
	r *http.Request,
	siaPassword string,
	target string,
	replaceSkylinks bool,
) {
	resp, err := p.siadRequest(w, r, r.Method, target, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to execute request: " + err.Error()))
		return
	}
	defer resp.Body.Close()

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

	// Add CORS headers so the API can be used from javascript
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

func (p proxy) postSkylinkProxy(
	w http.ResponseWriter,
	r *http.Request,
	siaPassword string,
	target string,
) {
	// Get the file from the multipart request
	mpr, err := r.MultipartReader()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No multipart headers found"))
		return
	}
	part, err := mpr.NextPart()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No multipart headers found"))
		return
	}
	defer part.Close()

	if part.FormName() != "file" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("First multipart should be file"))
		return
	}

	resp, err := p.siadRequest(w, r, r.Method, target+"?name="+url.QueryEscape(part.FileName()), part)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to execute request: " + err.Error()))
		return
	}
	defer resp.Body.Close()

	// Add CORS headers to the API can be used from javascript
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Range")
	w.Header().Set("Access-Control-Expose-Headers", "Accept-Ranges, Content-Encoding, Content-Length, Content-Range")
	w.Header().Set("Access-Control-Allow-Methods", "POST")

	// Copy the response from the server to the client
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
