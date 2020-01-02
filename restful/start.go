package restful

import(
	"fmt"
	"encoding/json"
//	"errors"
//	"strings"
//	"log"
//	"net"
	"io/ioutil"
//	"strings"
	"net/http"
//	"encoding/json"
//	"github.com/gorilla/mux"
//	"github.com/docker/docker/runconfig"
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/gorilla/mux"
//	"github.com/docker/docker/api/types/container"
//	"github.com/spf13/pflag"
)

var (
	startCommand     cliconfig.StartValues
	reinitstartCommand	func()*cliconfig.StartValues
)

type errorstartrespond struct{
	Message string `json:"message"`
}

//SetStartcommandfunc called by cmd/server to set the endpoint of start command
func SetStartcommandfunc(f func()*cliconfig.StartValues){
	reinitstartCommand = f
}

//Startcontainer handle func called by restful server
func Startcontainer(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	key:=vars["id"]
	fmt.Println("Startcontainer",key)
	startCommand=*reinitstartCommand()
	AddGlobal(startCommand.PodmanCommand.Flags())

	reqBody,_ := ioutil.ReadAll(r.Body)
	if len(reqBody) != 0{
		fmt.Println("startreqbody=",reqBody)
	}

	startCommand.InputArgs = []string{key}
	err:=RestfulServer.Servercmd.Startcmd(&startCommand)
	if err != nil{
		respond:=&errorstartrespond{
			Message:err.Error(),
		}
		str,_:=json.Marshal(respond)
		http.Error(w,string(str),404)
		return
	}
	fmt.Println("After running!")
	}