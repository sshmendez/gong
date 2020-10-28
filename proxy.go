package main

import (
	// "os"
	"errors"
	"strconv"
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

func EchoServer(port int){

	echoserver := RegexpHandler{}
	sport := strconv.Itoa(port)

	echoserver.HandleFunc(regexp.MustCompile("/.*"), func(res http.ResponseWriter, req *http.Request){
		log.Print("Getting echo")
		fmt.Fprint(res, req.URL.String())
	})
	
	log.Printf("Starting Echo Server %s", sport)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s",sport), echoserver))

}


func CorsServer(port int){
	handler := RegexpHandler{}
	sport := strconv.Itoa(port)
	handler.HandleFunc(regexp.MustCompile("/.*"), func(res http.ResponseWriter, req *http.Request){
		res.Header().Add("Access-Control-Allow-Origin","*")
		res.Write([]byte("Hitting Cors Server"))
	})
	
	log.Printf("Starting Cors Server %s", sport)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s",sport), handler))
}


///////////////////////////////////////////////////////



type Handler = http.Handler

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
	Path string
	GenericHostConfig
}

type ReverseProxy struct{
	HostConfig
	Port int
	Remote string

}

type FileServer struct{
	HostConfig
	Root string
	Index string
}


func (muxconf *MuxConfig) BuildMuxConfig(v map[string][]interface{}) *MuxConfig{
	if muxconf.handlers == nil{
		muxconf.handlers = map[string]Handler{}
	}
	for _, val := range v["hosts"]{
		hostconfig := HostConfig{}
		hostconfig.Apply(&GenericHostConfig{
			val.(map[string]interface{})}, )

		
		muxconf.AddHost(&hostconfig)
	}
	return muxconf
}

///////////////////////////////////////////////


func (muxconf *MuxConfig) Add(host *HostConfig, handler Handler) error{
	muxconf.hosts = append(muxconf.hosts, *host)
	muxconf.handlers[host.Hostname] = handler
	return nil
}

func (muxconf *MuxConfig) AddHost(host *HostConfig) error{
	handler, err := host.Build()
	if err != nil{
		return err
	}
	return muxconf.Add(host, handler)
}


///////////////////////////////////////////////




///////////////////////////////////////////////

func (hc *HostConfig) Build() (Handler, error){
	var server Handler
	var err error
	switch hc.ServerType {
	case "ReverseProxy":
		rp := ReverseProxy{}
		rp.Apply(hc)
		server, err = rp.Build()
	default:
		err = errors.New("Handler could not be constructed")
	}

	return server, err
}

func (hc *HostConfig) Apply(gc *GenericHostConfig){

	hc.Hostname, _ = gc.config["hostname"].(string)
	hc.ServerType = gc.config["type"].(string)
	hc.Path = gc.config["path"].(string)
	hc.config = gc.config

	delete(gc.config, "path")
	delete(gc.config,"hostname")
	delete(gc.config,"type")

}


///////////////////////////////////////////////


func (rp *ReverseProxy) Apply(hc *HostConfig) *ReverseProxy{
	config := hc.config["config"].(map[string]interface{})
	rp.Port = int(config["port"].(float64))

	remote := config["remote"]
	if remote == nil{
		remote = hc.Hostname
	}

	rp.Remote = remote.(string)

	delete(config,"port")
	delete(config, "remote")

	return rp
}

func (rp *ReverseProxy) Build() (Handler, error){
	
	host,_ := url.Parse(fmt.Sprintf("http://%s:%s",rp.Remote, strconv.Itoa(rp.Port)))

	return  httputil.NewSingleHostReverseProxy(host), nil
}


///////////////////////////////////////////////








///////////////////////////////////////////////

func buildMux(muxconf MuxConfig) *http.ServeMux{
	mux := http.NewServeMux()

	for i := range muxconf.hosts {
		host := muxconf.hosts[i]
		handler := muxconf.handlers[host.Hostname]
		log.Println("Adding :", host.Hostname)
		mux.Handle(host.Hostname+"/", handler)
	}


	return mux
}
//////////////////////////////////////////////////



func main(){

	config, _ := os.Open("config.json")
	// var confbytes []byte
	// config.Read(confbytes)
	// fmt.Println(confbytes)

	var v map[string][]interface{}
	decoder := json.NewDecoder(config)
	decoder.Decode(&v)


	muxconf := MuxConfig{}
	muxconf.BuildMuxConfig(v)


	go CorsServer(3001)
	go EchoServer(3002)

	mux := buildMux(muxconf)
	log.Fatal(http.ListenAndServe(":9000", mux))
	log.Print("Starting Server")

	

}