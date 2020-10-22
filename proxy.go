package main

import (
	// "os"
	"regexp"
	"log"
	"fmt"
	"net/url"
	"net/http"
	// "encoding/json"
	"net/http/httputil"
)
 
type Handler = func(res http.ResponseWriter, req *http.Request)


type Service struct{
	Host string
	Port int
	Handlers []Handler
}

type Config struct {

}

type MuxConfig struct {
	Config
	Hostname string
	Port int
	ServerType string
}



func parseConfig(config string) map[string]Service{
	origin , _ := url.Parse("http://shanemendez.com")
	port := 3000

	
	return map[string]Service{
		origin.Host : Service{Host: origin.Host, Port: port},
	}


}

func buildMux(config MuxConfig) *http.ServeMux{
	mux := http.NewServeMux()
	remote, _ := url.Parse("http://localhost:3000")
	proxy := httputil.NewSingleHostReverseProxy(remote)



	mux.HandleFunc("shanemendez.com/", func(res http.ResponseWriter, req *http.Request){
		proxy.ServeHTTP(res,req)
	})

	return mux
}
type route struct {
    pattern *regexp.Regexp
    handler http.Handler
}

type RegexpHandler struct {
    routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler http.Handler) {
    h.routes = append(h.routes, &route{pattern, handler})
}

func (h *RegexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request)) {
    h.routes = append(h.routes, &route{pattern, http.HandlerFunc(handler)})
}

func (h RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf(r.URL.Path)
    for _, route := range h.routes {
		log.Print("Matching routes")
        if route.pattern.MatchString(r.URL.Path) {
            route.handler.ServeHTTP(w, r)
            return
        }
    }
    // no pattern matched; send 404 response
    http.NotFound(w, r)
}

func main(){

	// config, _ := os.Open("config.json")
	

	go func(){

		echoserver := RegexpHandler{}

		echoserver.HandleFunc(regexp.MustCompile("/.*"), func(res http.ResponseWriter, req *http.Request){
			log.Print("Getting echo")
			fmt.Fprint(res, req.URL.String())
		})
		
		log.Printf("Starting Echo Server")
		log.Fatal(http.ListenAndServe(":3000", echoserver))

	}()

	log.Print("Starting Server")

	mux := buildMux(MuxConfig{})
	log.Fatal(http.ListenAndServe(":9000", mux))
	

}