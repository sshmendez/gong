package main

import (
	// "os"
	"regexp"
	"log"
	"fmt"
	"net/url"
	"net/http"
	"encoding/json"
	"net/http/httputil"
	"os"
)
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



///////////////////////////////////////////////////////



type Handler = func(res http.ResponseWriter, req *http.Request)


type Service struct{
	Host string
	Port int
	Handlers []Handler
}

type MuxConfig struct {
	hosts []HostConfig
}

type GenericHostConfig struct {
	config map[string]interface{}
}
type HostConfig struct {
	Hostname string
	ServerType string
}

type ReverseProxy struct{
	HostConfig
	Port int

}

type FileServer struct{
	HostConfig
	Root string
	Index string
}


func parseHosts(v map[string][]map[string]interface{}) *MuxConfig{
	muxconf := &MuxConfig{}
	for _, val := range v["hosts"]{
		conf := parseHost(&GenericHostConfig{val})
		muxconf.hosts = append(muxconf.hosts, conf)
	}
	return muxconf
}

func parseHost(v *GenericHostConfig) HostConfig{
	var hostconfig HostConfig
	
	hostconfig.apply(v)

	return hostconfig


}
///////////////////////////////////////////////


func (hc *HostConfig) apply(gc *GenericHostConfig) GenericHostConfig{
	hc.Hostname, _ = gc.config["hostname"].(string)
	hc.ServerType = gc.config["type"].(string)

	delete(gc.config,"hostname")
	delete(gc.config,"type")

	return *gc
}

///////////////////////////////////////////////
func buildMux(config MuxConfig) *http.ServeMux{
	mux := http.NewServeMux()
	remote, _ := url.Parse("http://localhost:3000")
	proxy := httputil.NewSingleHostReverseProxy(remote)



	mux.HandleFunc("shanemendez.com/", func(res http.ResponseWriter, req *http.Request){
		proxy.ServeHTTP(res,req)
	})

	return mux
}

func main(){

	config, _ := os.Open("config.json")
	// var confbytes []byte
	// config.Read(confbytes)
	// fmt.Println(confbytes)

	var v map[string][]map[string]interface{}
	decoder := json.NewDecoder(config)
	decoder.Decode(&v)

	// hostconf := &GenericHostConfig{v["hosts"][0].(map[string]interface{})}

	
	muxconf := parseHosts(v)

	fmt.Println(muxconf)

	

	return
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