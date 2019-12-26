package restful

import (
	"fmt"
	"strings"
//	"log"
//	"net"
//	"io/ioutil"
//	"strings"
	"net/http"
//	"encoding/json"
//	"github.com/gorilla/mux"
	"github.com/docker/docker/runconfig"
	"github.com/docker/docker/api/types/container"
)
func createcmdfromconfig(w http.ResponseWriter,config *container.Config) string {
	var s string
	s = "podman container create"
	if config.Image != ""{
		s = fmt.Sprintf("%s %s ",s,config.Image)
	}

	if config.Cmd != nil {
		temp := fmt.Sprint(config.Cmd)
		s = fmt.Sprintln(s,strings.Trim(temp,"[]"))
	}
		//	if config.Cmd != nil
	fmt.Fprintln(w,"createdcmdline=",s)
	return s
}

func restful_createcontainer(w http.ResponseWriter, r *http.Request){
	//	reqBody,_ := ioutil.ReadAll(r.Body)
		decoder := &runconfig.ContainerDecoder{}
		config,_,_,err:=decoder.DecodeConfig(r.Body);
		if err != nil {
			fmt.Fprintln(w,"config error!",err)
		}
	//	fmt.Fprintf(w,"%x+v",string(reqBody))
		fmt.Fprintln(w,"configg=",config,"image=",config.Image,"cmd=",config.Cmd)
	//	json.NewEncoder(w).Encode()
		createcmdfromconfig(w,config)
	}
	