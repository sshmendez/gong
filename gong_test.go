package gong

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"sabbo/gong/servers"
)


func TestMux(t *testing.T){


	muxconf, err := BuildMuxConfigFromFile("test_config.json")
	mux := BuildMux(muxconf)

	req, err := http.NewRequest("GET", "http://echo.com/proxy", nil)
    if err != nil {
        t.Fatal(err)
    }
	
	genericconf  := muxconf.hosts["echo.com"]
	server, err := genericconf.Resolve()
	
	if err != nil {

	}

	echoconf := server.(ReverseProxy)
	go gong.EchoServer(echoconf.Port)



	  // We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	  rr := httptest.NewRecorder()
	  handler := mux
  
	  // Our handlers satisfy http.Handler, so we can call their ServeHTTP method 
	  // directly and pass in our Request and ResponseRecorder.
	  handler.ServeHTTP(rr, req)
  
	  // Check the status code is what we expect.
	  if status := rr.Code; status != http.StatusOK {
		  t.Errorf("handler returned wrong status code: got %v want %v",
			  status, http.StatusOK)
	  }
  
	  // Check the response body is what we expect.
	  expected := `/proxy`
	  if rr.Body.String() != expected {
		  t.Errorf("handler returned unexpected body: got %v want %v",
			  rr.Body.String(), expected)
	  }




	

	


	

}