package system

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/czxichen/work-stacks/tools/parse"
)

const (
	IPC_RMID   = 0
	IPC_CREAT  = 00001000
	IPC_EXCL   = 00002000
	IPC_NOWAIT = 00004000
)

func Lock(NameMutex string) (uintptr, bool) {
	id, _, err := syscall.Syscall6(syscall.SYS_SHMGET, uintptr(unsafe.Pointer(&NameMutex)), 1, IPC_CREAT|IPC_EXCL, 0, 0, 0)
	if err.Error() != "errno 0" {
		return 0, false
	}
	return id, true
}

func UnLock(id uintptr) {
	syscall.Syscall6(syscall.SYS_SHMCTL, id, IPC_RMID, 0, 0, 0, 0)
}

var sysProcAttr = &syscall.SysProcAttr{Setsid: true, Credential: &syscall.Credential{Uid: 0, Gid: 0}}

//返回*syscall.SysProcAttr类型用来指定进程的属主属组
func GetSysProcAttr(user string) (*syscall.SysProcAttr, error) {
	var uid, gid int
	if user == "" || user == "root" {
		return sysProcAttr, nil
	} else {
		uid, gid = parse.GetId([]byte(user))
		if uid == -1 || gid == -1 {
			return nil, fmt.Errorf("User:%s is not exist\n", user)
		}
	}
	return &syscall.SysProcAttr{Setsid: true, Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}}, nil
}

//设置进程的组ID
func SetPgid(pid, pgid int) error {
	return syscall.Setpgid(pid, pgid)
}

func GetPPids(pid int) ([]int, error) {
	return []int{}, nil
}

//杀死自定义的进程列表
func Kill(pids []uint32) {
	for _, pid := range pids {
		syscall.Kill(int(pid), syscall.SIGKILL)
	}
}

//linux平台使用组ID杀死子进程
func KillAll(pid int) error {
	return syscall.Kill(pid-(pid*2), syscall.SIGKILL)
}

//从/proc/sys/kernel/osrelease读取Linux系统的内核版本
func GetSystemVersion() string {
	f, err := os.Open("/proc/sys/kernel/osrelease")
	if err != nil {
		return ""
	}
	defer f.Close()

	var buf [512]byte
	n, err := f.Read(buf[0:])
	if err != nil {
		return ""
	}

	if n > 0 && buf[n-1] == '\n' {
		n--
	}
	return string(buf[0:n])
}
