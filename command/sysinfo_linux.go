package command

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func GetMotherboardInfo() string { return "" }

func GetBiosInfo() string { return "" }

func GetDiskInfo() (infos []diskusage) {
	fs := getFileSystems()
	if len(fs) == 0 {
		return
	}
	var (
		all        bool
		mounts     []string
		StringsHas = func(target []string, src string) bool {
			for _, t := range target {
				if strings.TrimSpace(t) == src {
					return true
				}
			}
			return false
		}
	)

	lines := getLines("/etc/mtab")
	for _, line := range lines {
		fields := strings.Fields(line)
		if all == false {
			if "none" == fields[0] || !StringsHas(fs, fields[2]) {
				continue
			}
		}
		mounts = append(mounts, fields[1])
	}

	for _, path := range mounts {
		var usage = diskusage{Path: path}
		var info = new(syscall.Statfs_t)
		if err := syscall.Statfs(path, info); err == nil {
			usage.Total = info.Blocks * uint64(info.Bsize)
			usage.Free = info.Bavail * uint64(info.Bsize)
			infos = append(infos, usage)
		}
	}
	return
}

func getLines(path string) (list []string) {
	File, err := os.Open(path)
	if err != nil {
		return
	}
	defer File.Close()
	buf := bufio.NewReader(File)

	//最多只读取一百行
	for i := 0; i < 100; i++ {
		line, _, err := buf.ReadLine()
		if err != nil {
			break
		}
		list = append(list, strings.TrimSpace(string(line)))
	}
	return
}

func getFileSystems() (ret []string) {
	lines := getLines("/proc/filesystems")
	for _, line := range lines {
		if !strings.HasPrefix(line, "nodev") {
			ret = append(ret, strings.TrimSpace(line))
			continue
		}
		t := strings.Split(line, "\t")
		if len(t) != 2 || t[1] != "zfs" {
			continue
		}
		ret = append(ret, strings.TrimSpace(t[1]))
	}
	return
}

func GetCpuInfo() string {
	return fmt.Sprintf("Num:%d Arch:%s", runtime.NumCPU(), runtime.GOARCH)
}

func GetMemory() string {
	buf, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return ""
	}
	lines := bytes.Split(buf, []byte{10})
	if len(lines) < 14 {
		return ""
	}
	MemTotal := bytes.TrimSpace(bytes.TrimPrefix(lines[0], []byte("MemTotal:")))
	MemTotal = MemTotal[:len(MemTotal)-3]
	SwapTotal := bytes.TrimSpace(bytes.TrimPrefix(lines[13], []byte("SwapTotal:")))
	SwapTotal = SwapTotal[:len(SwapTotal)-3]
	mem, _ := strconv.ParseUint(string(MemTotal), 10, 64)
	swap, _ := strconv.ParseUint(string(SwapTotal), 10, 64)
	return fmt.Sprintf("Phsical:%d Swap:%d", mem, swap)
}

type intfInfo struct {
	Name string
	Ipv4 []string
	Ipv6 []string
}

//网卡信息
func GetIntfs() []intfInfo {
	intf, err := net.Interfaces()
	if err != nil {
		return []intfInfo{}
	}
	var is = make([]intfInfo, len(intf))
	for i, v := range intf {
		ips, err := v.Addrs()
		if err != nil {
			continue
		}
		is[i].Name = v.Name
		for _, ip := range ips {
			if strings.Contains(ip.String(), ":") {
				is[i].Ipv6 = append(is[i].Ipv6, ip.String())
			} else {
				is[i].Ipv4 = append(is[i].Ipv4, ip.String())
			}
		}
	}
	return is
}

func GetStartTime() string {
	buf, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return ""
	}
	bufs := bytes.Split(bytes.TrimSpace(buf), []byte{32})
	if len(bufs) != 2 {
		return ""
	}
	f, err := strconv.ParseFloat(string(bufs[0]), 0)
	if err != nil {
		return ""
	}
	f = f * float64(time.Second)
	return time.Duration(f).String()
}

func getUserName(uid int) string {
	if uid < 0 {
		return ""
	}
	if uid == 0 {
		return "root"
	}
	File, err := os.Open("/etc/passwd")
	if err != nil {
		return ""
	}
	defer File.Close()
	buf := bufio.NewReader(File)
	u := []byte(strconv.Itoa(uid))
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			return ""
		}
		lines := bytes.Split(line, []byte(":"))
		if len(lines) != 7 {
			continue
		}
		if bytes.Equal(lines[2], u) {
			return string(lines[0])
		}
	}
}

func GetUserName() string {
	user := getUserName(syscall.Getuid())
	buf, err := ioutil.ReadFile("/proc/sys/kernel/hostname")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s@%s", user, string(bytes.TrimSpace(buf)))
}

func GetSystemVersion() string {
	buf, err := ioutil.ReadFile("/proc/version")
	if err != nil {
		return ""
	}
	bufs := strings.Split(string(buf), " ")
	if len(bufs) < 3 {
		return ""
	}
	return strings.TrimSpace(bufs[2])
}
