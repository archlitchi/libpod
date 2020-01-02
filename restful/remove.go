package restful

import(
	"fmt"
	"strconv"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/gorilla/mux"
)

var (
	removeCommand     cliconfig.RmValues
	reinitremoveCommand	func()*cliconfig.RmValues
)

type errorremoverespond struct{
	Message string `json:"message"`
}

//SetRemovecommandfunc called by cmd/server to set the endpoint of remove command
func SetRemovecommandfunc(f func()*cliconfig.RmValues){
	reinitremoveCommand = f
}

func setremoveerror(w http.ResponseWriter,code int,err error){
	respond:=&errorremoverespond{
		Message:err.Error(),
	}
	str,_:=json.Marshal(respond)
	http.Error(w,string(str),code)
}


func processRemoveQueryParameters(r *http.Request) error{
	if tmp:=r.URL.Query().Get("force");tmp!=""{
		removeCommand.Flags().Lookup("force").Value.Set(tmp)
		removeCommand.Force,_ = strconv.ParseBool(tmp)
	}
	if tmp:=r.URL.Query().Get("v");tmp!=""{
		removeCommand.Flags().Lookup("volumes").Value.Set(tmp)
		removeCommand.Volumes,_ = strconv.ParseBool(tmp)
	}
	if tmp:=r.URL.Query().Get("link");tmp!=""{
		return errors.New("link parameter not supported")
	}
	return nil
}

//Removecontainer handle func called by restful server
func Removecontainer(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	key:=vars["id"]
	fmt.Println("Removecontainer",key)
	removeCommand=*reinitremoveCommand()
	AddGlobal(removeCommand.PodmanCommand.Flags())
	err:=processRemoveQueryParameters(r)
	if err!=nil{
		setremoveerror(w,400,err)
		return
	}

	reqBody,_ := ioutil.ReadAll(r.Body)
	if len(reqBody) != 0{
		fmt.Println("startreqbody=",reqBody)
	}

	removeCommand.InputArgs = []string{key}
	err=RestfulServer.Servercmd.Removecmd(&removeCommand)
	if err != nil{
		setremoveerror(w,404,err)
		return
	}
	fmt.Println("After running!")
}