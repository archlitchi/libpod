package cliconfig

import (
//	"net"
	"github.com/containers/libpod/pkg/rootless"
	"github.com/spf13/pflag"
	"github.com/sirupsen/logrus"
//	"github.com/spf13/cobra"
)

type RestfulServer struct{
	Servercmd	struct{
		Createcmd	func(c *CreateValues) (string,error)
		Startcmd 	func(c *StartValues) error
		Removecmd	func(c *RmValues) error
		Inspectcmd 	func(c *InspectValues) error
	}
	MainGlobalOpts	*MainFlags	
}

func (r *RestfulServer) InitRestfulServer() {
	r.MainGlobalOpts = nil 
}

//SetContainerCreatecmd called by cmd/server to set the create endpoint
func (r *RestfulServer) SetContainerCreatecmd(f func(c *CreateValues) (string,error)){
	r.Servercmd.Createcmd = f
}

//SetContainerStartcmd called by cmd/server to set the start endpoint
func (r *RestfulServer) SetContainerStartcmd(f func(c *StartValues) error){
	r.Servercmd.Startcmd = f
}

//SetContainerRemovecmd called by cmd/server to set the remove endpoint
func (r *RestfulServer) SetContainerRemovecmd(f func(c *RmValues) error){
	r.Servercmd.Removecmd = f
}

//SetContainerInspectcmd called by cmd/server to set the remove endpoint
func (r *RestfulServer) SetContainerInspectcmd(f func(c *InspectValues) error){
	r.Servercmd.Inspectcmd = f
}

//SetMainGlobalOpts called by cmd/server to set default podman configs
func (r *RestfulServer) SetMainGlobalOpts(m *MainFlags){
	r.MainGlobalOpts = m
}

func (r *RestfulServer) GetDefaultNetwork() string {
	if rootless.IsRootless() {
		return "slirp4netns"
	}
	return "bridge"
}

// markFlagHidden is a helper function to log an error if marking
// a flag as hidden happens to fail
func (r *RestfulServer) MarkFlagHidden(flags *pflag.FlagSet, flag string) {
	if err := flags.MarkHidden(flag); err != nil {
		logrus.Errorf("unable to mark flag '%s' as hidden: %q", flag, err)
	}
}

