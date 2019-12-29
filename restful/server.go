package restful

import (
	"fmt"
	"log"
	"net"
//	"io/ioutil"
	"net/http"
//	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/containers/libpod/cmd/podman/cliconfig"
//	"github.com/docker/docker/runconfig"
//	"github.com/docker/docker/api/types/container"
)
type HTTPServer struct {
	srv *http.Server
	l   net.Listener
}

var podrouter *mux.Router
var RestfulServer	*cliconfig.RestfulServer

func New(sockaddr string) *HTTPServer{
	listener,_ :=net.Listen("unix",sockaddr)
	fmt.Println("restful:server new",sockaddr)
	return &HTTPServer{
		srv: nil,
		l: listener,
	}
}

func (s *HTTPServer) HandleRequests(){
	if podrouter == nil {
		podrouter = mux.NewRouter().StrictSlash(true)
	}
	podrouter.HandleFunc("/podman/container/create",restful_createcontainer).Methods("POST")
	log.Fatal(http.Serve(s.l,podrouter))
}

func (s *HTTPServer) Close() error{
	return s.l.Close()
}




