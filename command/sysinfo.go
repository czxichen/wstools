package command

import (
	"fmt"
	"runtime"

	"github.com/czxichen/wstools/common"
)

var SysInfo = &Command{
	UsageLine: `sysinfo`,
	Run:       sysinfo,
	Short:     "查看系统信息",
	Long: `
`,
}

func sysinfo(cmd *Command, args []string) bool {
	fmt.Printf("开机时长:%s\n", common.GetStartTime())
	fmt.Printf("当前用户:%s\n", common.GetUserName())
	fmt.Printf("当前系统:%s\n", runtime.GOOS)
	fmt.Printf("系统版本:%s\n", common.GetSystemVersion())
	fmt.Printf("Bios:\t%s\n", common.GetBiosInfo())
	fmt.Printf("Motherboard:\t%s\n", common.GetMotherboardInfo())

	fmt.Printf("CPU:\t%s\n", common.GetCpuInfo())
	fmt.Printf("Memory:\t%s\n", common.GetMemory())
	fmt.Printf("Disk:\n")
	infos := common.GetDiskInfo()
	for _, v := range infos {
		fmt.Printf("Path:%s\tTotal:%d\tFree:%d\n", v.Path, v.Total, v.Free)
	}
	fmt.Printf("Interfaces:\n")
	intfs := common.GetIntfs()
	for _, v := range intfs {
		fmt.Printf("Name:%s\tIpv4:%s\tIpv6:%s\n", v.Name, v.Ipv4, v.Ipv6)
	}
	return true
}
