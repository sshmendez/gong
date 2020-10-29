package gong

import (
	// "os"
	"errors"
	"log"
	"fmt"
	"net/url"
	"net/http"
	"encoding/json"
	"net/http/httputil"
	"os"
)

///////////////////////////////////////////////////////



type Handler = http.Handler

type MuxConfig struct {
	Port int
	hosts map[string]GenericHostConfig
	handlers map[string]Handler
}

type GenericHostConfig struct {
	Hostname string
	ServerType string
	Path string
	config map[string]interface{}
}
type ServerBuilder interface {
	Build() (Handler, error)
	Apply(*GenericHostConfig) ServerBuilder
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


func (muxconf *MuxConfig) BuildMuxConfig(v map[string]interface{}) *MuxConfig{
	if muxconf.handlers == nil{
		muxconf.handlers = map[string]Handler{}
	}
	if muxconf.hosts == nil{
		muxconf.hosts = map[string]GenericHostConfig{}
	}
	muxconf.port = int(v["port"].(float64))
	
	for _, val := range v["hosts"].([]interface{}){		
		genericconf := NewGenericConfig(val.(map[string]interface{}))
		serverconf, err := genericconf.Resolve()
		if  err != nil{

		}
		

		muxconf.Add(&genericconf, serverconf)
	}
	return muxconf
}

func BuildMux(muxconf MuxConfig) *http.ServeMux{
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

func BuildMuxConfigFromFile(filename string) (MuxConfig, error) {
	var muxconf MuxConfig
	config, err := os.Open(filename)

	if err != nil {
		return muxconf,err
	}
	// var confbytes []byte
	// config.Read(confbytes)
	// fmt.Println(confbytes)

	var v map[string]interface{}
	fmt.Println(v)
	decoder := json.NewDecoder(config)
	decoder.Decode(&v)

	
	muxconf = MuxConfig{}
	muxconf.BuildMuxConfig(v)

	return muxconf, nil
}

///////////////////////////////////////////////


func (muxconf *MuxConfig) Add(host *GenericHostConfig, hc ServerBuilder) error{
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


func (hc *GenericHostConfig) Resolve() (ServerBuilder, error){
	var err error
	var config ServerBuilder
	switch hc.ServerType {
	case "ReverseProxy":
		return ReverseProxy{}.Apply(hc), nil
	case "FileServer":
		return FileServer{}.Apply(hc), nil
	default:
		err = errors.New("Handler could not be constructed")
	}

	return config, err
}


func NewGenericConfig (gc map[string]interface{}) GenericHostConfig{
	hc := GenericHostConfig{}
	config := gc
	hc.Hostname, _ = config["hostname"].(string)
	hc.ServerType = config["type"].(string)
	hc.Path = config["path"].(string)
	hc.config = config

	
	return hc
}
///////////////////////////////////////////////


func (rp ReverseProxy) Apply(hc *GenericHostConfig) ServerBuilder{
	config := hc.config["config"].(map[string]interface{})
	rp.Port = int(config["port"].(float64))

	remote := config["remote"]
	if remote == nil{
		remote = hc.Hostname
	}

	rp.Remote = remote.(string)

	return rp
}

func (rp ReverseProxy) Build() (Handler, error){
	
	surl := fmt.Sprintf("http://%s:%d",rp.Remote, rp.Port)
	fmt.Println(surl)
	host,_ := url.Parse(surl)

	return  httputil.NewSingleHostReverseProxy(host), nil
}


///////////////////////////////////////////////



func (fs FileServer) Apply(hc *GenericHostConfig) ServerBuilder{
	config := hc.config["config"].(map[string]interface{})
	fmt.Println(hc)
	fs.Root = config["root"].(string)
	index := config["index"]

	if index == nil{
		index = ""
	}
	fs.Index = index.(string)

	return fs
}


func (fs FileServer) Build() (Handler, error){
	fmt.Println("root is",fs.Root)
	return http.FileServer(http.Dir(fs.Root)),nil
}

///////////////////////////////////////////////

//////////////////////////////////////////////////

