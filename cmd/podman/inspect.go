package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/containers/buildah/pkg/formats"
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/containers/libpod/pkg/adapter"
	"github.com/containers/libpod/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	inspectTypeContainer = "container"
	inspectTypeImage     = "image"
	inspectAll           = "all"
)

var (
	inspectCommand cliconfig.InspectValues

	inspectDescription = `This displays the low-level information on containers and images identified by name or ID.

  If given a name that matches both a container and an image, this command inspects the container.  By default, this will render all results in a JSON array.`
	_inspectCommand = cobra.Command{
		Use:   "inspect [flags] CONTAINER | IMAGE",
		Short: "Display the configuration of a container or image",
		Long:  inspectDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			inspectCommand.InputArgs = args
			inspectCommand.GlobalFlags = MainGlobalOpts
			inspectCommand.Remote = remoteclient
			return inspectCmd(&inspectCommand)
		},
		Example: `podman inspect alpine
  podman inspect --format "imageId: {{.Id}} size: {{.Size}}" alpine
  podman inspect --format "image: {{.ImageName}} driver: {{.Driver}}" myctr`,
	 }
	 inspectjsonarray formats.JSONStructArray
	 inspectblockoutput bool
)

//Restfulinspectinit init command function for api server
func Restfulinspectinit() *cliconfig.InspectValues{
	var restfulinspectCommand     cliconfig.InspectValues
	restfulinspectCommand.PodmanCommand.Command = &cobra.Command{
		Use:   _inspectCommand.Use,
		Short: _inspectCommand.Short,
		Long:  inspectDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			restfulinspectCommand.InputArgs = args
			restfulinspectCommand.GlobalFlags = MainGlobalOpts
			restfulinspectCommand.Remote = remoteclient
			return inspectCmd(&restfulinspectCommand)
		},
		Example: _inspectCommand.Example,
	}
	inspectInit(&restfulinspectCommand)
	return &restfulinspectCommand
}

// Getinspectcommandfunc Generate cobra.command struct for restful api
func Getinspectcommandfunc() func()*cliconfig.InspectValues{
	return Restfulinspectinit
}

// InspectCmd Called from restfulAPI to execute create command
func InspectCmd(c *cliconfig.InspectValues) (formats.JSONStructArray,error) {
	err:=inspectCmd(c)
	return inspectjsonarray,err
}

func inspectInit(command *cliconfig.InspectValues) {
	command.SetHelpTemplate(HelpTemplate())
	command.SetUsageTemplate(UsageTemplate())
	flags := command.Flags()
	flags.StringVarP(&command.Format, "format", "f", "", "Change the output format to a Go template")

	// -t flag applicable only to 'podman inspect', not 'image/container inspect'
	ambiguous := strings.Contains(command.Use, "|")
	if ambiguous {
		flags.StringVarP(&command.TypeObject, "type", "t", inspectAll, "Return JSON for specified type, (image or container)")
	}

	if strings.Contains(command.Use, "CONTAINER") {
		containers_only := " (containers only)"
		if !ambiguous {
			containers_only = ""
			command.TypeObject = inspectTypeContainer
		}
		flags.BoolVarP(&command.Latest, "latest", "l", false, "Act on the latest container podman is aware of"+containers_only)
		flags.BoolVarP(&command.Size, "size", "s", false, "Display total file size"+containers_only)
		markFlagHiddenForRemoteClient("latest", flags)
	} else {
		command.TypeObject = inspectTypeImage
	}
}
func init() {
	inspectCommand.Command = &_inspectCommand
	inspectblockoutput=false
	inspectInit(&inspectCommand)
}

func inspectCmd(c *cliconfig.InspectValues) error {
	fmt.Println("inspect:",c.InputArgs,"size=",c.Bool("size"))
	args := c.InputArgs
	inspectType := c.TypeObject
	latestContainer := c.Latest
	if len(args) == 0 && !latestContainer {
		return errors.Errorf("container or image name must be specified: podman inspect [options [...]] name")
	}

	if len(args) > 0 && latestContainer {
		return errors.Errorf("you cannot provide additional arguments with --latest")
	}

	runtime, err := adapter.GetRuntime(getContext(), &c.PodmanCommand)
	if err != nil {
		return errors.Wrapf(err, "error creating libpod runtime")
	}
	defer runtime.DeferredShutdown(false)

	if !util.StringInSlice(inspectType, []string{inspectTypeContainer, inspectTypeImage, inspectAll}) {
		return errors.Errorf("the only recognized types are %q, %q, and %q", inspectTypeContainer, inspectTypeImage, inspectAll)
	}

	outputFormat := c.Format
	if strings.Contains(outputFormat, "{{.Id}}") {
		outputFormat = strings.Replace(outputFormat, "{{.Id}}", formats.IDString, -1)
	}
	// These fields were renamed, so we need to provide backward compat for
	// the old names.
	if strings.Contains(outputFormat, ".Src") {
		outputFormat = strings.Replace(outputFormat, ".Src", ".Source", -1)
	}
	if strings.Contains(outputFormat, ".Dst") {
		outputFormat = strings.Replace(outputFormat, ".Dst", ".Destination", -1)
	}
	if strings.Contains(outputFormat, ".ImageID") {
		outputFormat = strings.Replace(outputFormat, ".ImageID", ".Image", -1)
	}
	if latestContainer {
		lc, err := runtime.GetLatestContainer()
		if err != nil {
			return err
		}
		args = append(args, lc.ID())
		inspectType = inspectTypeContainer
	}

	inspectedObjects, iterateErr := iterateInput(getContext(), c.Size, args, runtime, inspectType)
	if iterateErr != nil {
		return iterateErr
	}

	var out formats.Writer
	if outputFormat != "" && outputFormat != formats.JSONString {
		//template
		out = formats.StdoutTemplateArray{Output: inspectedObjects, Template: outputFormat}
	} else {
		// default is json output
		out = inspectjsonarray
	}
	if !inspectblockoutput{
		err = out.Out()
	}else{
		inspectjsonarray = formats.JSONStructArray{Output: inspectedObjects}
	}
	return err
}

// func iterateInput iterates the images|containers the user has requested and returns the inspect data and error
func iterateInput(ctx context.Context, size bool, args []string, runtime *adapter.LocalRuntime, inspectType string) ([]interface{}, error) {
	var (
		data           interface{}
		inspectedItems []interface{}
		inspectError   error
	)

	for _, input := range args {
		switch inspectType {
		case inspectTypeContainer:
			ctr, err := runtime.LookupContainer(input)
			if err != nil {
				inspectError = errors.Wrapf(err, "error looking up container %q", input)
				break
			}
			data, err = ctr.Inspect(size)
			if err != nil {
				inspectError = errors.Wrapf(err, "error inspecting container %s", ctr.ID())
				break
			}
		case inspectTypeImage:
			image, err := runtime.NewImageFromLocal(input)
			if err != nil {
				inspectError = errors.Wrapf(err, "error getting image %q", input)
				break
			}
			data, err = image.Inspect(ctx)
			if err != nil {
				inspectError = errors.Wrapf(err, "error parsing image data %q", image.ID())
				break
			}
		case inspectAll:
			ctr, err := runtime.LookupContainer(input)
			if err != nil {
				image, err := runtime.NewImageFromLocal(input)
				if err != nil {
					inspectError = errors.Wrapf(err, "error getting image %q", input)
					break
				}
				data, err = image.Inspect(ctx)
				if err != nil {
					inspectError = errors.Wrapf(err, "error parsing image data %q", image.ID())
					break
				}
			} else {
				data, err = ctr.Inspect(size)
				if err != nil {
					inspectError = errors.Wrapf(err, "error inspecting container %s", ctr.ID())
					break
				}
			}
		}
		if inspectError == nil {
			inspectedItems = append(inspectedItems, data)
		}
	}
	return inspectedItems, inspectError
}
