package gong

import (
	"testing"
	"net/http"
	"fmt"

)

func TestBuildMuxConfigFromFile(t *testing.T) {
	muxconf,err := BuildMuxConfigFromFile("./test/test_config.json")
	if(err != nil){

	}
	mux := BuildMux(muxconf)
	http.ListenAndServe(fmt.Sprintf(":%d", muxconf.Port), mux)
	

}
