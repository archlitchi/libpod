package restful

import(
	"fmt"
	"strings"
	"strconv"
	"encoding/json"
//	"errors"
	"io/ioutil"
	"net/http"
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/gorilla/mux"
)

var (
	statsCommand     cliconfig.StatsValues
	reinitstatsCommand	func()*cliconfig.StatsValues
)

type errorstatsrespond struct{
	Message string `json:"message"`
}

//SetStatscommandfunc called by cmd/server to set the endpoint of stats command
func SetStatscommandfunc(f func()*cliconfig.StatsValues){
	reinitstatsCommand = f
}

func setstatserror(w http.ResponseWriter,code int,err error){
	respond:=&errorstatsrespond{
		Message:err.Error(),
	}
	str,_:=json.Marshal(respond)
	http.Error(w,string(str),code)
}

func processStatsQueryParameters(r *http.Request) error{
	var t string
	if tmp:=r.URL.Query().Get("stream");tmp!=""{
		if strings.Contains(tmp,"true"){
			t="false"
		}else{
			t="true"
		}
		statsCommand.Flags().Lookup("no-stream").Value.Set(t)
		statsCommand.NoStream,_ = strconv.ParseBool(t)
	}
	return nil
}

//Statscontainer handle func called by restful server
func Statscontainer(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	key:=vars["id"]
	fmt.Println("Statscontainer",key)
	statsCommand=*reinitstatsCommand()
	AddGlobal(statsCommand.PodmanCommand.Flags())
	err:=processStatsQueryParameters(r)
	if err!=nil{
		setstatserror(w,400,err)
		return
	}
	reqBody,_ := ioutil.ReadAll(r.Body)
	if len(reqBody) != 0{
		fmt.Println("statsreqbody=",reqBody)
	}
	statsCommand.InputArgs = []string{key}
	err=RestfulServer.Servercmd.Statscmd(&statsCommand,w)
	if err != nil{
		setstatserror(w,404,err)
		return
	}
	fmt.Println("After running!")
}