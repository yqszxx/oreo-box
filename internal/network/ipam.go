package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"

type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return fmt.Errorf("cannot stat the file: %v", err)
		}
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	if err != nil {
		return fmt.Errorf("cannot open %s: %v", ipam.SubnetAllocatorPath, err)
	}
	defer func() {
		if err := subnetConfigFile.Close(); err != nil {
			panic(err)
		}
	}()
	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return fmt.Errorf("cannot read: %v", err)
	}

	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		return fmt.Errorf("cannot unmarshal: %v", err)
	}
	return nil
}

func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(ipamConfigFileDir, 0755); err != nil {
				return fmt.Errorf("cannot make directory: %v", err)
			}
		} else {
			return fmt.Errorf("cannot stat the file: %v", err)
		}
	}
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %v", ipam.SubnetAllocatorPath, err)
	}
	defer func() {
		if err := subnetConfigFile.Close(); err != nil {
			panic(err)
		}
	}()

	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return fmt.Errorf("cannot marshal: %v", err)
	}

	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return fmt.Errorf("cannot write: %v", err)
	}

	return nil
}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 存放网段中地址分配信息的数组
	ipam.Subnets = &map[string]string{}

	// 从文件中加载已经分配的网段信息
	err = ipam.load()
	if err != nil {
		return nil, fmt.Errorf("cannot dump allocation info: %v", err)
	}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	one, size := subnet.Mask.Size()

	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}

	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			ip = subnet.IP
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3] += 1
			break
		}
	}

	if err := ipam.dump(); err != nil {
		return nil, fmt.Errorf("cannot dump: %v", err)
	}
	return
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		return fmt.Errorf("cannot dump allocation info: %v", err)
	}

	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	if err := ipam.dump(); err != nil {
		return fmt.Errorf("cannot dump: %v", err)
	}
	return nil
}
