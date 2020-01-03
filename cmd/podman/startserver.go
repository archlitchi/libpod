package main

import (
	"fmt"

	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/containers/libpod/restful"
	"github.com/spf13/cobra"
//	"github.com/gorilla/mux"
)

var (
	startserverCommand  cliconfig.StartserverValues
	_startserverCommand = &cobra.Command{
		Use:   "startserver",
		Args:  noSubArgs,
		Short: "Start the podman RESTful server",
		RunE: func(cmd *cobra.Command, args []string) error {
			startserverCommand.InputArgs = args
			startserverCommand.GlobalFlags = MainGlobalOpts
			startserverCommand.Remote = remoteclient
			return startserverCmd(&startserverCommand)
		},
	}
)

func init() {
	startserverCommand.Command = _startserverCommand
	startserverCommand.SetUsageTemplate(UsageTemplate())
	flags := startserverCommand.Flags()
	flags.StringVarP(&startserverCommand.SockPath, "socketpath", "s", "", "The path of unix socket")
}

func startserverCmd(c *cliconfig.StartserverValues) error {
	startserver(c)
	return nil
}

func setinitCommand(){
	restful.SetCreatecommandfunc(Getcreatecommandfunc())
	restful.SetStartcommandfunc(Getstartcommandfunc())
	restful.SetRemovecommandfunc(Getremovecommandfunc())
	restful.SetInspectcommandfunc(Getinspectcommandfunc())
	restful.SetStopcommandfunc(Getstopcommandfunc())
	restful.SetStatscommandfunc(Getstatscommandfunc())
}

func startserver(c *cliconfig.StartserverValues){
	path:=c.String("socketpath")
	if path == ""{
		path="/home/limengxuan/docker.sock"
	}
	s:=restful.New(path)
	defer s.Close()
	fmt.Println("socket established!")
	restful.RestfulServer = new(cliconfig.RestfulServer)
	cmdv := restful.RestfulServer
	cmdv.InitRestfulServer()
	cmdv.SetContainerCreatecmd(CreateCmd)
	cmdv.SetContainerStartcmd(StartCmd)
	cmdv.SetContainerRemovecmd(RemoveCmd)
	cmdv.SetContainerInspectcmd(InspectCmd)
	cmdv.SetContainerStopcmd(StopCmd)
	cmdv.SetContainerStatscmd(StatsCmd)
	cmdv.SetMainGlobalOpts(&MainGlobalOpts)
	setinitCommand()
	s.HandleRequests()
}

