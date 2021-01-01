package gong

import (
	// "os"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	. "github.com/sshmendez/gong/servers" 
	. "github.com/sshmendez/gong/types"
)


// type Gong struct {
// 	Resolvers []ServerBuilder
// }

func Resolve(hc *GenericHostConfig)  (ServerBuilder, error) {
	var err error
	var builder ServerBuilder
	switch hc.ServerType {
	case "ReverseProxy":
		return NewReverseProxy(hc), nil
	case "FileServer":
		return NewFileServer(hc), nil
	default:
		err = errors.New("Handler could not be constructed")
	}

	return builder, err
}

func AddEndPoint(mux *http.ServeMux) func(GenericHostConfig, http.Handler){
	return func (host GenericHostConfig, handler http.Handler){
		log.Println("Adding :", host.Hostname)
		mux.HandleFunc(host.Hostname+"/", func(res http.ResponseWriter, req *http.Request) {
			log.Println(fmt.Sprintf("Calling %s", host.Hostname))
			handler.ServeHTTP(res, req)
		})
	}
	
}
func BuildMux(muxconf MuxConfig) *http.ServeMux {
	mux := http.NewServeMux()
	muxconf.Map(AddEndPoint(mux))
	return mux
}



func BuildMuxConfigFromFile(filename string) (MuxConfig, error) {
	var muxconf MuxConfig
	config, err := os.Open(filename)

	if err != nil {
		return muxconf, err
	}
	// var confbytes []byte
	// config.Read(confbytes)
	// fmt.Println(confbytes)

	var v map[string]interface{}
	fmt.Println(v)
	decoder := json.NewDecoder(config)
	decoder.Decode(&v)

	muxconf = MuxConfig{}
	muxconf.BuildMuxConfig(Resolve, v)

	return muxconf, nil
}
