package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main(){
	origin, _ := url.Parse("http://localhost:3000")
	director := func(req *http.Request){
		// log.Printf("%s",req.Header.Values("X-Forwarded-For"))
		req.Header.Del("X-Forwarded-For")
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
	}
	proxy := &httputil.ReverseProxy{Director: director}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		proxy.ServeHTTP(w,r)
	})
	log.Fatal(http.ListenAndServe(":2999", nil))
}