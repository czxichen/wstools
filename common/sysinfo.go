//+build windws linux
package common

type diskusage struct {
	Path  string `json:"path"`
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

type ListMode uintptr

const (
	LIST_MODULES_DEFAULT ListMode = 0x0  //查看默认的模块
	LIST_MODULES_32BIT            = 0x01 //查看32-bit的模块
	LIST_MODULES_64BIT            = 0x02 //查看64-bit的模块
	LIST_MODULES_ALL              = 0x03 //查看所有的模块
)
