package restful

import(
	"fmt"
	"strconv"
	"encoding/json"
//	"errors"
	"io/ioutil"
	"net/http"
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/gorilla/mux"
)

var (
	inspectCommand     cliconfig.InspectValues
	reinitinspectCommand	func()*cliconfig.InspectValues
)

type errorinspectrespond struct{
	Message string `json:"message"`
}

//SetInspectcommandfunc called by cmd/server to set the endpoint of inspect command
func SetInspectcommandfunc(f func()*cliconfig.InspectValues){
	reinitinspectCommand = f
}

func setinspecterror(w http.ResponseWriter,code int,err error){
	respond:=&errorinspectrespond{
		Message:err.Error(),
	}
	str,_:=json.Marshal(respond)
	http.Error(w,string(str),code)
}


func processInspectQueryParameters(r *http.Request) error{
	if tmp:=r.URL.Query().Get("size");tmp!=""{
		inspectCommand.Flags().Lookup("size").Value.Set(tmp)
		inspectCommand.Size,_ = strconv.ParseBool(tmp)
	}
	return nil
}

//Inspectcontainer handle func called by restful server
func Inspectcontainer(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	key:=vars["id"]
	fmt.Println("Inspectcontainer",key)
	inspectCommand=*reinitinspectCommand()
	AddGlobal(inspectCommand.PodmanCommand.Flags())
	err:=processInspectQueryParameters(r)
	if err!=nil{
		setremoveerror(w,400,err)
		return
	}
	reqBody,_ := ioutil.ReadAll(r.Body)
	if len(reqBody) != 0{
		fmt.Println("startreqbody=",reqBody)
	}
	inspectCommand.InputArgs = []string{key}
	err=RestfulServer.Servercmd.Inspectcmd(&inspectCommand)
	if err != nil{
		setinspecterror(w,404,err)
		return
	}
	fmt.Println("After running!")
}