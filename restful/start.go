package restful

import(
	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/gorilla/mux"
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

func processStartQueryParameters(r *http.Request){
	if tmp:=r.URL.Query().Get("detachkeys");tmp!=""{
		startCommand.Flags().Lookup("detach-keys").Value.Set(tmp)
		startCommand.DetachKeys=tmp
	}
}

//Startcontainer handle func called by restful server
func Startcontainer(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	key:=vars["id"]
	startCommand=*reinitstartCommand()
	AddGlobal(startCommand.PodmanCommand.Flags())
	processStartQueryParameters(r)

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