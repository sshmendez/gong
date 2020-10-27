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


type MuxConfig struct {
	hosts []HostConfig
	handlers map[string]Handler
}

type GenericHostConfig struct {
	config map[string]interface{}
}
type HostConfig struct {
	Hostname string
	ServerType string
	GenericHostConfig
}

type ReverseProxy struct{
	HostConfig
	Port int
	Server *httputil.ReverseProxy

}

type FileServer struct{
	HostConfig
	Root string
	Index string
}


func BuildMuxConfig(v map[string][]interface{}, muxconf *MuxConfig) *MuxConfig{
	for _, val := range v["hosts"]{
		hostconfig := HostConfig{}
		hostconfig.Apply(&GenericHostConfig{
			val.(map[string]interface{})}, )
		
		handler := GetHandler(&hostconfig)
		if handler == nil{
			continue
		} 
		muxconf.handlers[hostconfig.Hostname] = handler
		muxconf.hosts = append(muxconf.hosts, hostconfig)
	}
	return muxconf
}
///////////////////////////////////////////////


func GetHandler(hc *HostConfig) Handler{
	var server Handler

	switch hc.ServerType {
	case "ReverseProxy":
		rp := ReverseProxy{}
		rp.Apply(hc)
		server = rp.ServeHTTP
	default:
		
	}

	return server
}


///////////////////////////////////////////////


func (hc *HostConfig) Apply(gc *GenericHostConfig){
	hc.Hostname, _ = gc.config["hostname"].(string)
	hc.ServerType = gc.config["type"].(string)
	hc.config = gc.config
	delete(gc.config,"hostname")
	delete(gc.config,"type")

}

func (hc *HostConfig) Build(){

}
func (hc *HostConfig) ServeHTTP(){

}


///////////////////////////////////////////////


func (rp *ReverseProxy) Apply(hc *HostConfig) *ReverseProxy{
	config := hc.config["config"].(map[string]interface{})
	rp.Port = int(config["port"].(float64))

	delete(config,"port")

	return rp
}

func (rp ReverseProxy) Build(){
	host,_ := url.Parse(fmt.Sprint(rp.Hostname,":",rp.Port))
	rp.Server = httputil.NewSingleHostReverseProxy(host)
	
}
func (rp *ReverseProxy) ServeHTTP(res http.ResponseWriter, req *http.Request){
	fmt.Println("calling handler!")
	rp.Server.ServeHTTP(res,req)
}

///////////////////////////////////////////////








///////////////////////////////////////////////
func buildMux(muxconf MuxConfig) *http.ServeMux{
	mux := http.NewServeMux()

	for hostname, handler := range muxconf.handlers {
		log.Println("Adding :", hostname)
		mux.HandleFunc("shanemendez.com", handler)
	}


	return mux
}

func main(){

	config, _ := os.Open("config.json")
	// var confbytes []byte
	// config.Read(confbytes)
	// fmt.Println(confbytes)

	var v map[string][]interface{}
	decoder := json.NewDecoder(config)
	decoder.Decode(&v)

	// hostconf := &GenericHostConfig{v["hosts"][0].(map[string]interface{})}

	muxconf := MuxConfig{}
	muxconf.handlers = make(map[string]Handler)

	BuildMuxConfig(v, &muxconf)
	buildMux(muxconf)



	
	go func(){

		echoserver := RegexpHandler{}

		echoserver.HandleFunc(regexp.MustCompile("/.*"), func(res http.ResponseWriter, req *http.Request){
			log.Print("Getting echo")
			fmt.Fprint(res, req.URL.String())
		})
		
		log.Printf("Starting Echo Server")
		log.Fatal(http.ListenAndServe(":3000", echoserver))

	}()

	mux := buildMux(MuxConfig{})
	log.Fatal(http.ListenAndServe(":9000", mux))
	log.Print("Starting Server")

	

}