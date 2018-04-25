package system

import (
	"errors"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	MAX_PATH           = 260
	TH32CS_SNAPPROCESS = 0x00000002
)

type ProcessInfo struct {
	Name string
	Pid  uint32
	PPid uint32
}

type PROCESSENTRY32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [MAX_PATH]uint16
}

type HANDLE uintptr

var (
	SysProcAttr                  *syscall.SysProcAttr
	modkernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot = modkernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = modkernel32.NewProc("Process32FirstW")
	procProcess32Next            = modkernel32.NewProc("Process32NextW")
	procCloseHandle              = modkernel32.NewProc("CloseHandle")
)

func Lock(NameMutex string) (uintptr, bool) {
	id, _, err := modkernel32.NewProc("CreateMutexA").Call(0, 1, uintptr(unsafe.Pointer(&NameMutex)))
	if err.Error() != "The operation completed successfully." {
		return 0, false
	}
	return id, true
}

func UnLock(id uintptr) {
	syscall.CloseHandle(syscall.Handle(id))
}

func GetSysProcAttr(user string) (*syscall.SysProcAttr, error) {
	return nil, errors.New("Platform unsupport")
}

//调用Kernel32的GetVersion方法获取windows系统的版本
func GetSystemVersion() string {
	v, _, _ := modkernel32.NewProc("GetVersion").Call()
	return strconv.Itoa(int(byte(v))) + "." + strconv.Itoa(int(uint8(v>>8)))
}

func SetPgid(pid, pgid int) error {
	return nil
}

//linux平台使用组ID杀死子进程
func KillAll(pid int) error {
	pids := Getppids(uint32(pid))
	Kill(pids)
	return nil
}

//杀死自定义的进程列表
func Kill(pids []uint32) {
	for _, pid := range pids {
		pro, err := os.FindProcess(int(pid))
		if err != nil {
			continue
		}
		pro.Kill()
	}
}

//通过系统调用遍历出所有进程,然后找出进程的所有子进程
func Getppids(pid uint32) []uint32 {
	infos, err := GetProcs()
	if err != nil {
		return []uint32{pid}
	}
	var pids []uint32 = make([]uint32, 0, len(infos))
	var index int = 0
	pids = append(pids, pid)

	var length int = len(pids)
	for index < length {
		for _, info := range infos {
			if info.PPid == pids[index] {
				pids = append(pids, info.Pid)
			}
		}
		index += 1
		length = len(pids)
	}
	return pids
}

//遍历系统的所有进程,然后返回[]ProcessInfo类型
func GetProcs() (procs []ProcessInfo, err error) {
	snap := createToolhelp32Snapshot(TH32CS_SNAPPROCESS, uint32(0))
	if snap == 0 {
		err = syscall.GetLastError()
		return
	}

	defer closeHandle(snap)

	var pe32 PROCESSENTRY32

	pe32.DwSize = uint32(unsafe.Sizeof(pe32))
	if process32First(snap, &pe32) == false {
		err = syscall.GetLastError()
		return
	}
	procs = append(procs, ProcessInfo{syscall.UTF16ToString(pe32.SzExeFile[:MAX_PATH]), pe32.Th32ProcessID, pe32.Th32ParentProcessID})
	for process32Next(snap, &pe32) {
		procs = append(procs, ProcessInfo{syscall.UTF16ToString(pe32.SzExeFile[:MAX_PATH]), pe32.Th32ProcessID, pe32.Th32ParentProcessID})
	}
	return
}

func createToolhelp32Snapshot(flags, processId uint32) HANDLE {
	ret, _, _ := procCreateToolhelp32Snapshot.Call(
		uintptr(flags),
		uintptr(processId))

	if ret <= 0 {
		return HANDLE(0)
	}
	return HANDLE(ret)
}

func process32First(snapshot HANDLE, pe *PROCESSENTRY32) bool {
	ret, _, _ := procProcess32First.Call(
		uintptr(snapshot),
		uintptr(unsafe.Pointer(pe)))

	return ret != 0
}

func process32Next(snapshot HANDLE, pe *PROCESSENTRY32) bool {
	ret, _, _ := procProcess32Next.Call(
		uintptr(snapshot),
		uintptr(unsafe.Pointer(pe)))

	return ret != 0
}

func closeHandle(object HANDLE) bool {
	ret, _, _ := procCloseHandle.Call(
		uintptr(object))
	return ret != 0
}
