package restful

import (
	"fmt"
	"log"
	"net"
//	"io/ioutil"
	"net/http"
//	"encoding/json"
	"github.com/gorilla/mux"
//	"github.com/docker/docker/runconfig"
//	"github.com/docker/docker/api/types/container"
)
type HTTPServer struct {
	srv *http.Server
	l   net.Listener
}

var podrouter *mux.Router

func New(sockaddr string) *HTTPServer{
	listener,_ :=net.Listen("unix",sockaddr)
	fmt.Println("restful:server new",sockaddr)
	return &HTTPServer{
		srv: nil,
		l: listener,
	}
}

/*func restful_createcontainer(w http.ResponseWriter, r *http.Request){
//	reqBody,_ := ioutil.ReadAll(r.Body)
	decoder := &runconfig.ContainerDecoder{}
	config,_,_,err:=decoder.DecodeConfig(r.Body);
	if err != nil {
		fmt.Fprintln(w,"config error!",err)
	}
//	fmt.Fprintf(w,"%x+v",string(reqBody))
	fmt.Fprintln(w,"configg=",config)
//	json.NewEncoder(w).Encode()

}*/

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




