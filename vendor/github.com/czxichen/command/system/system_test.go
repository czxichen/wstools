package system

import (
	"testing"
)

func Test_GetSystemInfo(t *testing.T) {
	info := GetSystemInfo()
	t.Logf("HostName: %s\n", info.PlatForm)
	t.Logf("HostName: %s\n", info.Sysversion)
	t.Logf("HostName: %s\n\n", info.HostName)
	for _, v := range info.Interfaces {
		if v.IPv4 == "" {
			continue
		}
		t.Logf("Interface Name: %s\n", v.Name)
		t.Logf("Interface Mac: %s\n", v.Mac)
		t.Logf("Interface IPV4: %s\n", v.IPv4)
		t.Logf("Interface IPV6: %s\n", v.IPv6)
	}
}
