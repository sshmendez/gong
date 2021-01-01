package servers
 
import (
	"strconv"
	"regexp" 
	"net/http"
	"log"
	"fmt"
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

func EchoHandler() func(res http.ResponseWriter, req *http.Request){
	return func(res http.ResponseWriter, req *http.Request){
		log.Print("Getting echo")
		fmt.Fprint(res, req.URL.String())
	}
}
func EchoServer(port int){

	echoserver := RegexpHandler{}
	sport := strconv.Itoa(port)

	echoserver.HandleFunc(regexp.MustCompile("/.*"), EchoHandler())
	
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

