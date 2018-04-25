package system

import (
	"net"
	"os"
	"runtime"
	"strings"
)

type Interf struct {
	Mac  string `json:"mac"`
	Name string `json:"name"`
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
}

type SysTemInfo struct {
	HostName   string
	PlatForm   string
	Sysversion string
	Interfaces []Interf
}

func GetHostName() string {
	name, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return name
}

func GetInterfaces() (interfs []Interf) {
	intfs, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, v := range intfs {
		i := Interf{}
		i.Name = v.Name
		i.Mac = v.HardwareAddr.String()
		ips, err := v.Addrs()
		if err != nil {
			continue
		}
		for _, ip := range ips {
			if strings.Contains(ip.String(), ":") {
				i.IPv6 = ip.String()
			} else {
				i.IPv4 = ip.String()
			}
		}
		interfs = append(interfs, i)
	}
	return
}

func GetSystemInfo() (sys SysTemInfo) {
	sys.HostName = GetHostName()
	sys.Sysversion = GetSystemVersion()
	sys.PlatForm = runtime.GOOS
	sys.Interfaces = GetInterfaces()
	return
}
