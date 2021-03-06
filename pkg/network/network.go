package network

import (
	"github.com/containers/libpod/pkg/util"
	"net"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/allocator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// SupportedNetworkDrivers describes the list of supported drivers
var SupportedNetworkDrivers = []string{"bridge"}

// IsSupportedDriver checks if the user provided driver is supported
func IsSupportedDriver(driver string) error {
	if util.StringInSlice(driver, SupportedNetworkDrivers) {
		return nil
	}
	return errors.Errorf("driver '%s' is not supported", driver)
}

// GetLiveNetworks returns a slice of networks representing what the system
// has defined as network interfaces
func GetLiveNetworks() ([]*net.IPNet, error) {
	var nets []*net.IPNet
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, address := range addrs {
		_, n, err := net.ParseCIDR(address.String())
		if err != nil {
			return nil, err
		}
		nets = append(nets, n)
	}
	return nets, nil
}

// GetLiveNetworkNames returns a list of network interfaces on the system
func GetLiveNetworkNames() ([]string, error) {
	var interfaceNames []string
	liveInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range liveInterfaces {
		interfaceNames = append(interfaceNames, i.Name)
	}
	return interfaceNames, nil
}

// GetFreeNetwork looks for a free network according to existing cni configuration
// files and network interfaces.
func GetFreeNetwork() (*net.IPNet, error) {
	networks, err := GetNetworksFromFilesystem()
	if err != nil {
		return nil, err
	}
	liveNetworks, err := GetLiveNetworks()
	if err != nil {
		return nil, err
	}
	nextNetwork, err := GetDefaultPodmanNetwork()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("default network is %s", nextNetwork.String())
	for {
		newNetwork, err := NextSubnet(nextNetwork)
		if err != nil {
			return nil, err
		}
		logrus.Debugf("checking if network %s intersects with other cni networks", nextNetwork.String())
		if intersectsConfig, _ := networkIntersectsWithNetworks(newNetwork, allocatorToIPNets(networks)); intersectsConfig {
			logrus.Debugf("network %s is already being used by a cni configuration", nextNetwork.String())
			nextNetwork = newNetwork
			continue
		}
		logrus.Debugf("checking if network %s intersects with any network interfaces", nextNetwork.String())
		if intersectsLive, _ := networkIntersectsWithNetworks(newNetwork, liveNetworks); !intersectsLive {
			break
		}
		logrus.Debugf("network %s is being used by a network interface", nextNetwork.String())
		nextNetwork = newNetwork
	}
	return nextNetwork, nil
}

func allocatorToIPNets(networks []*allocator.Net) []*net.IPNet {
	var nets []*net.IPNet
	for _, network := range networks {
		if len(network.IPAM.Ranges) > 0 {
			// this is the new IPAM range style
			// append each subnet from ipam the rangeset
			for _, r := range network.IPAM.Ranges[0] {
				nets = append(nets, newIPNetFromSubnet(r.Subnet))
			}
		} else {
			//	 looks like the old, deprecated style
			nets = append(nets, newIPNetFromSubnet(network.IPAM.Subnet))
		}
	}
	return nets
}

func newIPNetFromSubnet(subnet types.IPNet) *net.IPNet {
	n := net.IPNet{
		IP:   subnet.IP,
		Mask: subnet.Mask,
	}
	return &n
}

func networkIntersectsWithNetworks(n *net.IPNet, networklist []*net.IPNet) (bool, *net.IPNet) {
	for _, nw := range networklist {
		if networkIntersect(n, nw) {
			return true, nw
		}
	}
	return false, nil
}

func networkIntersect(n1, n2 *net.IPNet) bool {
	return n2.Contains(n1.IP) || n1.Contains(n2.IP)
}

// ValidateUserNetworkIsAvailable returns via an error if a network is available
// to be used
func ValidateUserNetworkIsAvailable(userNet *net.IPNet) error {
	networks, err := GetNetworksFromFilesystem()
	if err != nil {
		return err
	}
	liveNetworks, err := GetLiveNetworks()
	if err != nil {
		return err
	}
	logrus.Debugf("checking if network %s exists in cni networks", userNet.String())
	if intersectsConfig, _ := networkIntersectsWithNetworks(userNet, allocatorToIPNets(networks)); intersectsConfig {
		return errors.Errorf("network %s is already being used by a cni configuration", userNet.String())
	}
	logrus.Debugf("checking if network %s exists in any network interfaces", userNet.String())
	if intersectsLive, _ := networkIntersectsWithNetworks(userNet, liveNetworks); intersectsLive {
		return errors.Errorf("network %s is being used by a network interface", userNet.String())
	}
	return nil
}
