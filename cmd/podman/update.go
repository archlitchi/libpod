package main

import (
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"fmt"
	"github.com/containers/libpod/pkg/adapter"
	"github.com/containers/libpod/pkg/rootless"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	updateCommand     cliconfig.UpdateValues
	updateDescription = `
   podman container update

   Updates cpu or memory configuration on  one or more running containers. The container name or ID can be used.
`
	_updateCommand = &cobra.Command{
		Use:   "update [flags] CONTAINER [CONTAINER...]",
		Short: "Update one or more configs on containers",
		Long:  updateDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			updateCommand.InputArgs = args
			updateCommand.GlobalFlags = MainGlobalOpts
			updateCommand.Remote = remoteclient
			return updateCmd(&updateCommand)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			return checkAllLatestAndCIDFile(cmd, args, false, false)
		},
		Example: `podman container checkpoint --keep ctrID
  podman container checkpoint --all
  podman container checkpoint --leave-running --latest`,
	}
)

func init() {
	updateCommand.Command = _updateCommand
	updateCommand.SetHelpTemplate(HelpTemplate())
	updateCommand.SetUsageTemplate(UsageTemplate())

	flags := updateCommand.Flags()
	flags.Uint16Var(&updateCommand.Blkio, "blkio-weight", 0, " [X] Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)")
	flags.IntVar(&updateCommand.Cpu_period, "cpu-period", 0, "[X] Limit CPU CFS (Completely Fair Scheduler) period")
	flags.IntVar(&updateCommand.Cpu_quota, "cpu-quota", 0, "[X] Limit CPU CFS (Completely Fair Scheduler) quota")
	flags.IntVar(&updateCommand.Cpu_rt_period, "cpu-rt-period", 0, "[x]Limit the CPU real-time period in microseconds")
	flags.IntVar(&updateCommand.Cpu_rt_runtime, "cpu-rt-runtime", 0, "[x]Limit the CPU real-time runtime in microseconds")
	flags.IntVarP(&updateCommand.Cpu_shares, "cpu-shares", "c" , -1, "[x]CPU shares (relative weight)")
	flags.IntVar(&updateCommand.Cpus, "cpus", -1, "[x]Number of CPUs")
	flags.StringVar(&updateCommand.Cpuset_cpus, "cpuset-cpus", "", "[x]CPUs in which to allow execution (0-3, 0,1)")
	flags.StringVar(&updateCommand.Cpuset_mems, "cpuset-mems", "", "[x]MEMs in which to allow execution (0-3, 0,1)")
	flags.IntVar(&updateCommand.Kernel_memory, "kernel_memory", -1, "[x]kernel memory limit")
	flags.IntVarP(&updateCommand.Memory, "memory", "m",-1, "[x]Memory limit")
	flags.IntVar(&updateCommand.Memory_reservation, "memory_reservation", -1, "[x]Memory soft limit")
	flags.IntVar(&updateCommand.Memory_swap, "memory_swap", 0, "[x]Swap limit equal to memory plus swap: '-1' to enable unlimited swap")
	flags.StringVar(&updateCommand.Restart, "restart","", "[x]Restart policy to apply when a container exits")
	flags.BoolVarP(&updateCommand.All, "all", "a", false, "Checkpoint all running containers")
	flags.BoolVarP(&updateCommand.Latest, "latest", "l", false, "Act on the latest container podman is aware of")
	markFlagHiddenForRemoteClient("latest", flags)
}

func updateCmd(c *cliconfig.UpdateValues) error {
	fmt.Println("into the update Cmd export=",c.Memory_reservation)
	if rootless.IsRootless() {
		return errors.New("checkpointing a container requires root")
	}

	runtime, err := adapter.GetRuntime(getContext(), &c.PodmanCommand)
	if err != nil {
		return errors.Wrapf(err, "could not get runtime")
	}

	defer runtime.DeferredShutdown(false)
	return nil;
	//	return runtime.Checkpoint(c)
}
