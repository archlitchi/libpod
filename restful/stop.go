package restful

import(
	"fmt"
	"strconv"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/gorilla/mux"
)

var (
	stopCommand     cliconfig.StopValues
	reinitstopCommand	func()*cliconfig.StopValues
)

type errorstoprespond struct{
	Message string `json:"message"`
}

//SetStopcommandfunc called by cmd/server to set the endpoint of remove command
func SetStopcommandfunc(f func()*cliconfig.StopValues){
	reinitstopCommand = f
}

func setstoperror(w http.ResponseWriter,code int,err error){
	respond:=&errorstoprespond{
		Message:err.Error(),
	}
	str,_:=json.Marshal(respond)
	http.Error(w,string(str),code)
}

func processStopQueryParameters(r *http.Request) error{
	if tmp:=r.URL.Query().Get("t");tmp!=""{
		stopCommand.Flags().Lookup("t").Value.Set(tmp)
		i,_:=strconv.Atoi(tmp)
		stopCommand.Timeout = uint(i)
	}
	return nil
}

//Stopcontainer handle func called by restful server
func Stopcontainer(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	key:=vars["id"]
	fmt.Println("Stopcontainer",key)
	stopCommand=*reinitstopCommand()
	AddGlobal(stopCommand.PodmanCommand.Flags())
	err:=processRemoveQueryParameters(r)
	if err!=nil{
		setstoperror(w,400,err)
		return
	}
	reqBody,_ := ioutil.ReadAll(r.Body)
	if len(reqBody) != 0{
		fmt.Println("startreqbody=",reqBody)
	}
	stopCommand.InputArgs = []string{key}
	err=RestfulServer.Servercmd.Stopcmd(&stopCommand)
	if err != nil{
		setstoperror(w,404,err)
		return
	}
	fmt.Println("After running!")
}