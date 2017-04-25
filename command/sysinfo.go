package command

import (
	"fmt"
	"runtime"
)

var SysInfo = &Command{
	UsageLine: `sysinfo`,
	Run:       sysinfo,
	Short:     "查看系统信息",
	Long: `
`,
}

func sysinfo(cmd *Command, args []string) bool {
	fmt.Printf("开机时长:%s\n", GetStartTime())
	fmt.Printf("当前用户:%s\n", GetUserName())
	fmt.Printf("当前系统:%s\n", runtime.GOOS)
	fmt.Printf("系统版本:%s\n", GetSystemVersion())
	fmt.Printf("Bios:\t%s\n", GetBiosInfo())
	fmt.Printf("Motherboard:\t%s\n", GetMotherboardInfo())

	fmt.Printf("CPU:\t%s\n", GetCpuInfo())
	fmt.Printf("Memory:\t%s\n", GetMemory())
	fmt.Printf("Disk:\n")
	infos := GetDiskInfo()
	for _, v := range infos {
		fmt.Printf("Path:%s\tTotal:%d\tFree:%d\n", v.Path, v.Total, v.Free)
	}
	fmt.Printf("Interfaces:\n")
	intfs := GetIntfs()
	for _, v := range intfs {
		fmt.Printf("Name:%s\tIpv4:%s\tIpv6:%s\n", v.Name, v.Ipv4, v.Ipv6)
	}
	return true
}

type diskusage struct {
	Path  string `json:"path"`
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}
