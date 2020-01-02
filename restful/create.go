package restful

import (
	"fmt"
	"errors"
	"strings"
//	"log"
//	"net"
//	"io/ioutil"
//	"strings"
	"net/http"
	"encoding/json"
//	"github.com/gorilla/mux"
	"github.com/docker/docker/runconfig"
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/docker/docker/api/types/container"
	"github.com/spf13/pflag"
)

var (
	createCommand     cliconfig.CreateValues
	reinitcreateCommand	func()*cliconfig.CreateValues
)

type acceptcreaterespond struct{
	ID string `json:"ID"`
	Warnings string `json:"warnings"`
}

type errorcreaterespond struct{
	Message string `json:"message"`
}

const (
	idTruncLength      = 12
	sizeWithUnitFormat = "(format: `<number>[<unit>]`, where unit = b (bytes), k (kilobytes), m (megabytes), or g (gigabytes))"
)

//SetCreatecommandfunc Set the init function for cobra.command in create option
func SetCreatecommandfunc(c func()*cliconfig.CreateValues){
	reinitcreateCommand = c
}

func setFlagsFromConfig(f *pflag.FlagSet, c *container.Config,h *container.HostConfig) error {
	
	var temp string

	f.Lookup("hostname").Value.Set(c.Hostname)
	
	// DomainName not implemented!
	if c.Domainname!="" {
		return errors.New("domain name not implemented")
	}

	f.Lookup("user").Value.Set(c.User)

	attach:=[]string{}
	if c.AttachStdin{
		attach=append(attach,"STDIN")
	}
	if c.AttachStdout{
		attach=append(attach,"STDOUT")
	}
	if c.AttachStderr{
		attach=append(attach,"STDERR")
	}
	if len(attach) != 0{
		temp=strings.Trim(fmt.Sprint(attach),"[]")
		temp=strings.ReplaceAll(temp," ",",")
		fmt.Println("Get c.attach=",temp)
		f.Lookup("attach").Value.Set(temp)
	} 	

	//Exposed ports Not implemented
    if len(c.ExposedPorts)!=0{
		return errors.New("exposedports not implemented")
	}
	
	f.Lookup("tty").Value.Set(fmt.Sprint(c.Tty))

	f.Lookup("interactive").Value.Set(fmt.Sprint(c.OpenStdin))

	//StdinOnce not implemnented default to false
	if c.StdinOnce == true{
		return errors.New("stdinonce not implemented")
	}
	if len(c.Env) != 0{
		for _,value:=range(c.Env){
			f.Lookup("env").Value.Set(value)
		}
	}

	//HealthCheck not implemented
	if c.Healthcheck!=nil{
		return errors.New("healthcheck not implemented")
	}

	//ArgsEscaped not implemented
	if c.ArgsEscaped!=false{
		return errors.New("argsescaped not impelmented")
	}
	//volumes not implemented, use hostconfig.bind
	//f.Lookup("volume").Value.Set(fmt.Sprint())
	if len(c.Volumes)!=0{
		return errors.New("volumes not implemented.Use hostname:bind")
	}

	f.Lookup("workdir").Value.Set(c.WorkingDir)

	//EntryPoint not implemented
	//f.Lookup("entrypoint").Value.Set(c.Entrypoint)
	if len(c.Entrypoint)!=0{
		return errors.New("Entrypoint not implemented.Use cmd")
	}

	//NetworkDisabled not implemented default to false
	if c.NetworkDisabled == true{
		return errors.New("Networkdisabled not supported")
	}

	f.Lookup("mac-address").Value.Set(c.MacAddress)

	//Onbuild not implemented
	if len(c.OnBuild)!=0{
		return errors.New("onbuild not implemented")
	}

	temp = ""
	if len(c.Labels) != 0{
		for key,value:=range(c.Labels){
			temp=key+":"+value
			f.Lookup("label").Value.Set(temp)
		}
	}

	f.Lookup("stop-signal").Value.Set(c.StopSignal)

	f.Lookup("stop-timeout").Value.Set(fmt.Sprint(c.StopTimeout))

	if h != nil{
	if len(h.Binds) != 0{
		for _,value:=range(h.Binds){
			f.Lookup("volume").Value.Set(value)
		}
	}

	if h.NetworkMode.IsNone(){
		f.Lookup("network").Value.Set("host")
	}else{
		f.Lookup("network").Value.Set(fmt.Sprint(h.NetworkMode))
	}

	f.Lookup("ipc").Value.Set(fmt.Sprint(h.IpcMode))

	f.Lookup("init").Value.Set(fmt.Sprint(*h.Init))

	//Resources:
	f.Lookup("cpu-shares").Value.Set(fmt.Sprint(h.CPUShares))

	if h.Memory!=0{
		f.Lookup("memory").Value.Set(fmt.Sprint(h.Memory,"b"))
	}
	
	if h.NanoCPUs!=0{
		f.Lookup("cpus").Value.Set(fmt.Sprint(h.NanoCPUs))
	}

	f.Lookup("cgroup-parent").Value.Set(h.CgroupParent)

	if h.BlkioWeight!=0{
		f.Lookup("blkio-weigut").Value.Set(fmt.Sprint(h.BlkioWeight))
	}

	f.Lookup("cpu-period").Value.Set(fmt.Sprint(h.CPUPeriod))

	f.Lookup("cpu-quota").Value.Set(fmt.Sprint(h.CPUQuota))

	f.Lookup("cpuset-cpus").Value.Set(h.CpusetCpus)

	f.Lookup("cpuset-mems").Value.Set(h.CpusetMems)

	f.Lookup("kernel-memory").Value.Set(fmt.Sprint(h.KernelMemory,"b"))

	f.Lookup("memory-reservation").Value.Set(fmt.Sprint(h.MemoryReservation,"b"))

	f.Lookup("memory-swap").Value.Set(fmt.Sprint(h.MemorySwap,"b"))

	f.Lookup("memory-swappiness").Value.Set(fmt.Sprint(*h.MemorySwappiness))

	f.Lookup("oom-kill-disable").Value.Set(fmt.Sprint(*h.OomKillDisable))

	f.Lookup("pids-limit").Value.Set(fmt.Sprint(*h.PidsLimit))
	}
	return nil
}

func createcmdfromconfig(w http.ResponseWriter,config *container.Config,hostconfig *container.HostConfig) error {

	inargs := &createCommand.InputArgs

	if config.Image != ""{
		*inargs = append(*inargs,config.Image)
	}
	if config.Cmd != nil {
		temp := strings.Trim(fmt.Sprint(config.Cmd),"[]")
		for _,val := range(strings.Split(temp," ")){
			*inargs = append(*inargs,val)
		}
	}
	AddGlobal(createCommand.PodmanCommand.Flags())
	flags := createCommand.Flags()
	err:=setFlagsFromConfig(flags,config,hostconfig)
	if err!=nil{
		respond:=errorstartrespond{
			Message: err.Error(),
		}
		str,_:=json.Marshal(respond)
		http.Error(w,string(str),400)
		return err
	}

	var retval string
	RestfulServer.Servercmd.Createcmd(&createCommand,&retval)
	respond := &acceptcreaterespond{
		ID:retval,
		Warnings:"null",
	}
	str,_:=json.Marshal(*respond)
	http.Error(w,string(str),201)
	return nil
}

//Createcontainer Handler function for container create in restful server
func Createcontainer(w http.ResponseWriter, r *http.Request){
	//	reqBody,_ := ioutil.ReadAll(r.Body)
		decoder := &runconfig.ContainerDecoder{}
		config,hostconfig,_,err:=decoder.DecodeConfig(r.Body);
		if err != nil {
			fmt.Println("config error!",err)
		}
		if config == nil {
			fmt.Println("config error is null!",err)
		}
		fmt.Println("before createcmdfromconfig")
		createCommand=*reinitcreateCommand()
	//	fmt.Fprintln(w,"configg=",config,"image=",config.Image,"cmd=",config.Cmd)
		createcmdfromconfig(w,config,hostconfig)
	}
	