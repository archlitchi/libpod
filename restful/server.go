package restful

import (
	"fmt"
	"log"
	"net"
//	"io/ioutil"
	"net/http"
//	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
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

func AddGlobal(p *pflag.FlagSet){
	p.String("cgroup-manager",RestfulServer.MainGlobalOpts.CGroupManager,"cgrouphelp")

	p.String("cpu-profile",RestfulServer.MainGlobalOpts.CpuProfile,"Path for the cpu profiling results")
	p.String("config",RestfulServer.MainGlobalOpts.Config,"Path of a libpod config file detailing container server configuration options")		
	p.String("conmon",RestfulServer.MainGlobalOpts.ConmonPath,"Path of the conmon binary")
	p.String("network-cmd-path",RestfulServer.MainGlobalOpts.NetworkCmdPath,"Path to the command for configuring the network")
	p.String("default-mounts-file",RestfulServer.MainGlobalOpts.CniConfigDir,"Path of the configuration directory for CNI networks")
	p.String("events-backend",RestfulServer.MainGlobalOpts.EventsBackend,"Events backend to use")
	
//	p.Bool("help",false,"Help for podman")
	p.StringSlice("hooks-dir",RestfulServer.MainGlobalOpts.HooksDir,"Set the OCI hooks directory path (may be set multiple times)")
	p.String("log-level",RestfulServer.MainGlobalOpts.LogLevel,"Log messages above specified level: debug, info, warn, error, fatal or panic")
	p.Int("max-workers",RestfulServer.MainGlobalOpts.MaxWorks,"The maximum number of workers for parallel operations")

	p.String("namespace",RestfulServer.MainGlobalOpts.Namespace,"Set the libpod namespace, used to create separate views of the containers and pods on the system")
	p.String("root",RestfulServer.MainGlobalOpts.Root,"Path to the root directory in which data, including images, is stored")
	p.String("runroot",RestfulServer.MainGlobalOpts.Runroot,"Path to the 'run directory' where all state information is stored")
	p.String("runtime",RestfulServer.MainGlobalOpts.Runtime,"Path to the OCI-compatible binary used to run containers, default is /usr/bin/runc")

	// -s is deprecated due to conflict with -s on subcommands
	p.String("storage-driver",RestfulServer.MainGlobalOpts.StorageDriver,"Select which storage driver is used to manage storage of images and containers (default is overlay)")

	_,err:=p.GetStringSlice("storage-opt")
	if err!=nil{
		fmt.Println("err=",err)
		p.StringArray("storage-opt",RestfulServer.MainGlobalOpts.StorageOpts,"Used to pass an option to the storage driver")
	}
	p.Bool("syslog",RestfulServer.MainGlobalOpts.Syslog,"Output logging information to syslog as well as the console")
	p.String("tmpdir",RestfulServer.MainGlobalOpts.TmpDir, "Path to the tmp directory for libpod state content.\n\nNote: use the environment variable 'TMPDIR' to change the temporary storage location for container images, '/var/tmp'.\n")
	p.Bool("trace",RestfulServer.MainGlobalOpts.Trace,"trace usage")

}

func (s* HTTPServer) HandleRequests(){
	if podrouter == nil {
		podrouter = mux.NewRouter().StrictSlash(true)
	}
	podrouter.HandleFunc("/podman/container/create",Createcontainer).Methods("POST")
	podrouter.HandleFunc("/podman/container/{id}/start",Startcontainer).Methods("POST")
	log.Fatal(http.Serve(s.l,podrouter))
}

func (s *HTTPServer) Close() error{
	return s.l.Close()
}




