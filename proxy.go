package main

import (
	// "os"
	"log"
	"net/url"
	"net/http"
	// "encoding/json"
	"net/http/httputil"
	// "./ServerTypes"
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

func buildMux(config MuxConfig) http.ServeMux{
	mux := http.NewServeMux()
	remote, _ := url.Parse("http://localhost:3000")
	proxy := httputil.NewSingleHostReverseProxy(remote)


	mux.HandleFunc("shanemendez.com/", func(res http.ResponseWriter, req *http.Request){
		proxy.ServeHTTP(res,req)
	})

	return *mux
}

func main(){

	// config, _ := os.Open("config.json")
	
	

	log.Print("Starting Server")

	mux := buildMux(MuxConfig{})
	log.Fatal(http.ListenAndServe(":9000", &mux))
	
}