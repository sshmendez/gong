package main

import (
	// "os"
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



///////////////////////////////////////////////////////



type Handler = http.Handler

type MuxConfig struct {
	hosts []HostConfig
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
	for _, val := range v["hosts"]{
		hostconfig := HostConfig{}
		hostconfig.Apply(&GenericHostConfig{
			val.(map[string]interface{})}, )
		
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
		server = rp.Build()
	default:
		
	}

	return server
}


///////////////////////////////////////////////


func (hc *HostConfig) Apply(gc *GenericHostConfig){

	hc.Hostname, _ = gc.config["hostname"].(string)
	hc.ServerType = gc.config["type"].(string)
	hc.Path = gc.config["path"].(string)
	hc.config = gc.config

	delete(gc.config, "path")
	delete(gc.config,"hostname")
	delete(gc.config,"type")

}

func (hc *HostConfig) Build() Handler{
	echoserver := RegexpHandler{}

	echoserver.HandleFunc(regexp.MustCompile(hc.Hostname+"/.*"), func(res http.ResponseWriter, req *http.Request){
		fmt.Fprint(res, req.URL.String())
	})
	

	return echoserver
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

func (rp *ReverseProxy) Build() Handler{
	
	host,_ := url.Parse(fmt.Sprintf("http://%s:%s/%s",rp.Remote, strconv.Itoa(rp.Port), rp.Path))
	host, _ = url.Parse("http://shanemendez.com")
	return  httputil.NewSingleHostReverseProxy(host)
}


///////////////////////////////////////////////








///////////////////////////////////////////////

func buildMux(muxconf MuxConfig) *http.ServeMux{
	mux := http.NewServeMux()

	for i := range muxconf.hosts {
		host := muxconf.hosts[i]
		handler := GetHandler(&host)
		if handler == nil{
			continue
		}
		log.Println("Adding :", host.Hostname)
		mux.Handle("shanemendez.com/", handler)
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

	muxconf.BuildMuxConfig(v)

	fmt.Println(muxconf.hosts)



	
	go func(){

		echoserver := RegexpHandler{}

		echoserver.HandleFunc(regexp.MustCompile("/.*"), func(res http.ResponseWriter, req *http.Request){
			log.Print("Getting echo")
			fmt.Fprint(res, req.URL.String())
		})
		
		log.Printf("Starting Echo Server")
		log.Fatal(http.ListenAndServe(":3000", echoserver))

	}()

	mux := buildMux(muxconf)
	log.Fatal(http.ListenAndServe(":9000", mux))
	log.Print("Starting Server")

	

}