package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	listen := flag.String("listen", ":8082", "The address which the server will listen on")
	resourceDir := flag.String("res", "res", "Path of the resources directory")
	siadURL := flag.String("siad-url", "http://127.0.0.1:9980", "URL of the siad API")
	flag.Parse()

	apipass, err := ioutil.ReadFile(os.Getenv("HOME") + "/.sia/apipassword")
	if err != nil {
		panic(err)
	}
	siaPassword := strings.TrimSpace(string(apipass))

	mux := http.NewServeMux()
	p := proxy{siaPassword: siaPassword, siadURL: *siadURL}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) == 47 {
			p.getSkylinkProxy(w, r, strings.TrimPrefix(r.URL.Path, "/"))
			return
		}
		http.ServeFile(w, r, *resourceDir+"/index.html")
	})
	mux.HandleFunc("/file/", func(w http.ResponseWriter, r *http.Request) {
		p.getSkylinkProxy(w, r, strings.TrimPrefix(r.URL.Path, "/file/")+"?attachment=true")
	})
	mux.HandleFunc("/skynet/skyfile", p.postSkylinkProxy)
	mux.HandleFunc("/skynet/skyfile/", p.postSkylinkProxy)
	mux.HandleFunc("/res/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, *resourceDir+"/"+strings.TrimPrefix(r.URL.Path, "/res/"))
	})

	fmt.Println("Serving on " + *listen)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		panic(err)
	}
}

type proxy struct {
	siaPassword string
	siadURL     string
}

// Add CORS headers so the API can be used from javascript
func enableCORS(w http.ResponseWriter, method string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", method)
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
	for headerName := range r.Header {
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
	for headerName := range resp.Header {
		if headerValue := resp.Header.Get(headerName); headerValue != "" {
			w.Header().Set(headerName, headerValue)
		}
	}

	return resp, nil
}

func (p proxy) getSkylinkProxy(w http.ResponseWriter, r *http.Request, skylink string) {
	resp, err := p.siadRequest(w, r, "GET", "/skynet/skylink/"+skylink, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to execute request: " + err.Error()))
		return
	}
	defer resp.Body.Close()

	enableCORS(w, "GET")

	// Copy the response from the server to the client
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (p proxy) postSkylinkProxy(w http.ResponseWriter, r *http.Request) {
	resp, err := p.siadRequest(
		w, r, "POST",
		"/skynet/skyfile/uploads/"+time.Now().Format("2006-01-02_15:04:05.000000000"),
		r.Body,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to execute request: " + err.Error()))
		return
	}
	defer resp.Body.Close()

	enableCORS(w, "POST")

	// Copy the response from the server to the client
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
