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
	hosts map[string]GenericHostConfig
	handlers map[string]Handler
}

type GenericHostConfig struct {
	Hostname string
	ServerType string
	Path string
	config map[string]interface{}
}
type HostConfig interface {
	Build() (Handler, error)
	Apply(*GenericHostConfig) (HostConfig)
}

type ReverseProxy struct{
	GenericHostConfig
	Port int
	Remote string

}

type FileServer struct{
	GenericHostConfig
	Root string
	Index string
}


func (muxconf *MuxConfig) BuildMuxConfig(v map[string][]interface{}) *MuxConfig{
	if muxconf.handlers == nil{
		muxconf.handlers = map[string]Handler{}
	}
	if muxconf.hosts == nil{
		muxconf.hosts = map[string]GenericHostConfig{}
	}
	for _, val := range v["hosts"]{
		genericconf := GenericHostConfig{}
		
		serverconf := genericconf.Apply(val.(map[string]interface{}))
		
		
		muxconf.Add(&genericconf, *serverconf)
	}
	return muxconf
}

///////////////////////////////////////////////


func (muxconf *MuxConfig) Add(host *GenericHostConfig, hc HostConfig) error{
	handler, err := hc.Build()
	if err != nil{

	}

	muxconf.hosts[host.Hostname] = *host
	muxconf.handlers[host.Hostname] = handler
	return nil
}

// func (muxconf *MuxConfig) AddHost(host *GenericHostConfig) error{
// 	handler, err := host.Build()
// 	if err != nil{
// 		return err
// 	}
// 	return muxconf.Add(host, handler)
// }


///////////////////////////////////////////////




///////////////////////////////////////////////

func (hc *GenericHostConfig) Resolve() (*HostConfig, error){
	var config HostConfig
	var err error
	switch hc.ServerType {
	case "ReverseProxy":
		config = &ReverseProxy{}
	case "FileServer":
		config = &FileServer{}
	default:
		err = errors.New("Handler could not be constructed")
	}

	return &config, err
}
func (hc *GenericHostConfig) Build() (Handler, error){
	var server Handler
	var err error	

	return server, err
}

func (hc *GenericHostConfig) Apply(gc map[string]interface{}) *HostConfig{

	config := gc
	hc.Hostname, _ = config["hostname"].(string)
	hc.ServerType = config["type"].(string)
	hc.Path = config["path"].(string)
	hc.config = config

	serverconf,err := hc.Resolve()
	
	if err == nil{}



	delete(config, "path")
	delete(config,"hostname")
	delete(config,"type")

	return serverconf
}


///////////////////////////////////////////////


func (rp *ReverseProxy) Apply(hc *GenericHostConfig) HostConfig{
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



func (fs *FileServer) Apply(hc *GenericHostConfig) HostConfig{
	config := hc.config["config"].(map[string]interface{})
	fmt.Println(hc)
	fs.Root = config["root"].(string)
	index := config["index"]

	if index == nil{
		index = ""
	}
	fs.Index = index.(string)

	delete(config, "index")
	delete(config, "root")

	return fs
}


func (fs *FileServer) Build() (Handler, error){
	return http.FileServer(http.Dir(fs.Root)),nil
}

///////////////////////////////////////////////

func buildMux(muxconf MuxConfig) *http.ServeMux{
	mux := http.NewServeMux()

	for i := range muxconf.hosts {
		host := muxconf.hosts[i]
		handler := muxconf.handlers[host.Hostname]
		log.Println("Adding :", host.Hostname)
		mux.HandleFunc(host.Hostname+"/", func(res http.ResponseWriter, req *http.Request){
			log.Println(fmt.Sprintf("Calling %s", host.Hostname))
			handler.ServeHTTP(res,req)
		})
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

	host := muxconf.hosts["shanemendez.com"]
	handler,err  := host.Build()
	if err == nil{

	}
	go http.ListenAndServe(":3003", handler)
	mux := buildMux(muxconf)
	log.Fatal(http.ListenAndServe(":9000", mux))
	log.Print("Starting Server")

	

}