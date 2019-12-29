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
		Createcmd	func(c *CreateValues) error
	}
	MainGlobalOpts	*MainFlags	
}

func (r *RestfulServer) InitRestfulServer() {
	r.MainGlobalOpts = nil 
}

func (r *RestfulServer) SetContainerCreatecmd(f func(c *CreateValues) error){
	r.Servercmd.Createcmd = f
}

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

