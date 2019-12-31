package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/containers/libpod/pkg/adapter"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	createCommand     cliconfig.CreateValues
	createDescription = `Creates a new container from the given image or storage and prepares it for running the specified command.

  The container ID is then printed to stdout. You can then start it at any time with the podman start <container_id> command. The container will be created with the initial state 'created'.`
	_createCommand= &cobra.Command{
		Use:   "create [flags] IMAGE [COMMAND [ARG...]]",
		Short: "Create but do not start a container",
		Long:  createDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			createCommand.InputArgs = args
			createCommand.GlobalFlags = MainGlobalOpts
			createCommand.Remote = remoteclient
			return createCmd(&createCommand)
		},
		Example: `podman create alpine ls
  podman create --annotation HELLO=WORLD alpine ls
  podman create -t -i --name myctr alpine ls`,
	}
	retvalue	string
)

func init() {
	fmt.Println("create______:init")
	createCommand.PodmanCommand.Command = _createCommand
	createCommand.SetHelpTemplate(HelpTemplate())
	createCommand.SetUsageTemplate(UsageTemplate())

	getCreateFlags(&createCommand.PodmanCommand)
	flags := createCommand.Flags()
	flags.SetInterspersed(false)
	flags.SetNormalizeFunc(aliasFlags)
}

//Restfulinit init command function for api server
func Restfulinit() *cliconfig.CreateValues{
	var restfulCreatecommand cliconfig.CreateValues
	restfulCreatecommand.PodmanCommand.Command = &cobra.Command{
		Use:   "create [flags] IMAGE [COMMAND [ARG...]]",
		Short: "Create but do not start a container",
		Long:  createDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			restfulCreatecommand.InputArgs = args
			restfulCreatecommand.GlobalFlags = MainGlobalOpts
			restfulCreatecommand.Remote = remoteclient
			return createCmd(&restfulCreatecommand)
		},
		Example: `podman create alpine ls
  		podman create --annotation HELLO=WORLD alpine ls
  		podman create -t -i --name myctr alpine ls`,
	}
	restfulCreatecommand.SetHelpTemplate(HelpTemplate())
	restfulCreatecommand.SetUsageTemplate(UsageTemplate())

	getCreateFlags(&restfulCreatecommand.PodmanCommand)
	flags := restfulCreatecommand.Flags()
	flags.SetInterspersed(false)
	flags.SetNormalizeFunc(aliasFlags)
	return &restfulCreatecommand
}

// Getcreatecommand Generate cobra.command struct for restful api
func Getcreatecommandfunc() func()*cliconfig.CreateValues{
	return Restfulinit
}

// CreateCmd Called from restfulAPI to execute create command
func CreateCmd(c *cliconfig.CreateValues,s *string) error {
	return createCmd(c)
}

func createCmd(c *cliconfig.CreateValues) error {
	if c.Bool("trace") {
		span, _ := opentracing.StartSpanFromContext(Ctx, "createCmd")
		defer span.Finish()
	}

	if c.String("authfile") != "" {
		if _, err := os.Stat(c.String("authfile")); err != nil {
			return errors.Wrapf(err, "error getting authfile %s", c.String("authfile"))
		}
	}

	if err := createInit(&c.PodmanCommand); err != nil {
		return err
	}

	runtime, err := adapter.GetRuntime(getContext(), &c.PodmanCommand)
	if err != nil {
		return errors.Wrapf(err, "error creating libpod runtime")
	}
	defer runtime.DeferredShutdown(false)

	cid, err := runtime.CreateContainer(getContext(), c)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", cid)
	retvalue = cid
	return nil
}

func createInit(c *cliconfig.PodmanCommand) error {
	fmt.Println("createInit:",c,"args:",c.InputArgs,
	"label=",len(c.StringArray("label"))," ",c.StringArray("label"),"memory=",c.String("memory"),
	"env=",len(c.StringArray("env"))," ",c.StringArray("env"),
	"hostname=",c.String("hostname"),
	"attach=",len(c.StringSlice("attach"))," ",c.StringSlice("attach"),
	"binds=",len(c.StringArray("volume"))," ",c.StringArray("volume"),
	"networkmode=",c.String("network"),
	"ipc=",c.String("ipc"),
	"init=",c.Bool("init"),
	"memory=",c.String("memory"),
	"memoryswap=",c.String("memory-swap"),
	"memoryreservation=",c.String("memory-reservation"),
	"kernelmemory=",c.String("kernel-memory"),
	"cpus",c.Float64("cpus"),
	"cpushare",c.Uint64("cpu-shares"),
	"cpuperiod",c.Uint64("cpu-period"),
	"cpuquota",c.Int64("cpu-quota"),
	"cpusetcpus",c.String("cpuset-cpus"),
	"cpusetmems",c.String("cpuset-mems"),
	"memoryswapness",c.Int64("memory-swappiness"),
	"oomkilldisable",c.Bool("oom-kill-disable"),
	"pidslimit",c.Int64("pids-limit"),
	"cgroupparent",c.String("cgroup-parent"))

	fmt.Println("aaaa ",c.Command.Flag("label").Value.String())

	if c.IsSet("memory"){
		fmt.Println("memory:",c.String("memory"))
	}
	if !c.IsSet("memory"){
		fmt.Println("memory not set")
	}	
//	fmt.Println("create_podcommand=",c.Command)	
	if !remote && c.Bool("trace") {
		span, _ := opentracing.StartSpanFromContext(Ctx, "createInit")
		defer span.Finish()
	}

	if c.IsSet("privileged") && c.IsSet("security-opt") {
		logrus.Warn("setting security options with --privileged has no effect")
	}

	var setNet string
	if c.IsSet("network") {
		setNet = c.String("network")
	} else if c.IsSet("net") {
		setNet = c.String("net")
	}
	if (c.IsSet("dns") || c.IsSet("dns-opt") || c.IsSet("dns-search")) && (setNet == "none" || strings.HasPrefix(setNet, "container:")) {
		return errors.Errorf("conflicting options: dns and the network mode.")
	}

	// Docker-compatibility: the "-h" flag for run/create is reserved for
	// the hostname (see https://github.com/containers/libpod/issues/1367).

	if len(c.InputArgs) < 1 {
		return errors.Errorf("image name or ID is required")
	}

	return nil
}
