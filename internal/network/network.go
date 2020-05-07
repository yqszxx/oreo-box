package network

import (
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"github.com/yqszxx/oreo-box/config"
	"github.com/yqszxx/oreo-box/internal"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)

var (
	drivers  = map[string]Driver{}
	networks = map[string]*Network{}
)

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	Network     *Network
	PortMapping []string
}

type Network struct {
	Name    string
	IpRange *net.IPNet
	Driver  string
}

type Driver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network Network, endpoint *Endpoint) error
}

func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dumpPath, 0755); err != nil {
				return fmt.Errorf("cannot make directory: %v", err)
			}
		} else {
			return fmt.Errorf("cannot stat the file: %v", err)
		}
	}

	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("cannot open file %v", err)

	}
	defer func() {
		if err := nwFile.Close(); err != nil {
			panic(err)
		}
	}()

	nwJson, err := json.Marshal(nw)
	if err != nil {
		return fmt.Errorf("cannot marshal %v", err)
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		return fmt.Errorf("cannot write %v", err)
	}
	return nil
}

func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}

func (nw *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := nwConfigFile.Close(); err != nil {
			panic(err)
		}
	}()
	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		return fmt.Errorf("cannot load nw info %v", err)
	}
	return nil
}

func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(config.NetworkPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(config.NetworkPath, 0755); err != nil {
				return fmt.Errorf("cannot make directory: %v", err)
			}
		} else {
			return fmt.Errorf("cannot stat the file: %v", err)
		}
	}

	err := filepath.Walk(config.NetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}

		if err := nw.load(nwPath); err != nil {
			return fmt.Errorf("cannot load network: %v", err)
		}

		networks[nwName] = nw
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func CreateNetwork(driver, subnet, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	ip, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = ip

	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}

	return nw.dump(config.NetworkPath)
}

func ListNetwork() error {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, nw := range networks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("cannot flush %v", err)
	}
	return nil
}

func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("cannot find network `%s`", networkName)
	}

	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("cannot remove gateway ip: %s", err)
	}

	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("cannot delete network driver: %s", err)
	}

	return nw.remove(config.NetworkPath)
}

func enterboxNetns(enLink *netlink.Link, cinfo *internal.BoxInfo) (func() error, error) {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot get box netns of pid `%s`: %v", cinfo.Pid, err)
	}

	nsFD := f.Fd()
	runtime.LockOSThread()

	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		return nil, fmt.Errorf("cannot set netns of link `%s`: %v", (*enLink).Attrs().Name, err)
	}

	originalNS, err := netns.Get()
	if err != nil {
		return nil, fmt.Errorf("cannot get current netns: %v", err)
	}

	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		return nil, fmt.Errorf("cannot set netns: %v", err)
	}
	return func() error {
		if err := netns.Set(originalNS); err != nil {
			return fmt.Errorf("cannot set netns back: %v", err)
		}
		if err := originalNS.Close(); err != nil {
			return fmt.Errorf("cannot close handle to original netns: %v", err)
		}
		runtime.UnlockOSThread()
		if err := f.Close(); err != nil {
			return fmt.Errorf("cannot close netns fd: %v", err)
		}
		return nil
	}, nil
}

func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *internal.BoxInfo) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("cannot config endpoint: %v", err)
	}

	leaveNS, err := enterboxNetns(&peerLink, cinfo)
	if err != nil {
		return fmt.Errorf("cannot enter box netns: %v", err)
	}
	defer func() {
		if err := leaveNS(); err != nil {
			panic(err)
		}
	}()

	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress

	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("cannot set ip of interface `%s`: %v", ep.Device.Name, err)
	}

	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return fmt.Errorf("cannot bring up interface `%s`: %v", ep.Device.PeerName, err)
	}

	if err = setInterfaceUP("lo"); err != nil {
		return fmt.Errorf("cannot bring up interface `lo`: %v", err)
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}

func configPortMapping(ep *Endpoint) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			return fmt.Errorf("port mapping format error, %v", pm)
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		//err := cmd.Run()
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("cannot create iptables, %v", output)
		}
	}
	return nil
}

func Connect(networkName string, cinfo *internal.BoxInfo) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("cannot find network `%s`", networkName)
	}

	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}

	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}

	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}

	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return err
	}

	if err := configPortMapping(ep); err != nil {
		return fmt.Errorf("cannot config port mapping: %v", err)
	}

	return nil
}
