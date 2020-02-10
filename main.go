package main

import (
	"flag"
	"fmt"
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

	// Headers which will be copied from the client to the server and from the
	// server to the client when proxying a request
	headers := []string{
		"Accept",
		"Range",
		"Accept-Ranges",
		"Content-Length",
		"Content-Range",
		"Content-Type",
		"Content-Disposition",
		"Date",
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) == 47 {
			skylinkProxy(
				w, r, siaPassword,
				*siadURL+"/skynet/skylink/"+strings.TrimPrefix(r.URL.Path, "/"),
				headers, headers, true,
			)
			return
		}
		http.ServeFile(w, r, *resourceDir+"/index.html")
	})
	mux.HandleFunc("/file/", func(w http.ResponseWriter, r *http.Request) {
		skylinkProxy(
			w, r, siaPassword,
			*siadURL+"/skynet/skylink/"+strings.TrimPrefix(r.URL.Path, "/file/"),
			headers, headers, false,
		)
	})
	mux.HandleFunc("/api/skyfile", func(w http.ResponseWriter, r *http.Request) {
		skylinkProxy(
			w, r, siaPassword,
			*siadURL+"/skynet/skyfile/uploads/"+time.Now().Format("2006-01-02_15:04:05.000000000"),
			headers, headers, false,
		)
	})
	mux.HandleFunc("/res/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, *resourceDir+"/"+strings.TrimPrefix(r.URL.Path, "/res/"))
	})

	fmt.Println("Serving on " + *listen)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		panic(err)
	}
}
