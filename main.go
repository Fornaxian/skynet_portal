package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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
			p.getSkylinkProxy(w, r)
			return
		}
		http.ServeFile(w, r, *resourceDir+"/index.html")
	})
	mux.HandleFunc("/file/", p.getRawSkylinkProxy)
	mux.HandleFunc("/skynet/skyfile", p.postSkylinkProxy)
	mux.HandleFunc("/res/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, *resourceDir+"/"+strings.TrimPrefix(r.URL.Path, "/res/"))
	})

	fmt.Println("Serving on " + *listen)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		panic(err)
	}
}
