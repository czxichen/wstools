// +build windows linux

package common

// DiskUsage 磁盘使用率
type DiskUsage struct {
	Path  string `json:"path"`
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

// ListMode 列表模式
type ListMode uintptr

const (
	// LIST_MODULES_DEFAULT 查看默认的模块
	LIST_MODULES_DEFAULT ListMode = 0x0
	// LIST_MODULES_32BIT 查看32-bit的模块
	LIST_MODULES_32BIT = 0x01
	// LIST_MODULES_64BIT 查看64-bit的模块
	LIST_MODULES_64BIT = 0x02
	// LIST_MODULES_ALL 查看所有的模块
	LIST_MODULES_ALL = 0x03
)
