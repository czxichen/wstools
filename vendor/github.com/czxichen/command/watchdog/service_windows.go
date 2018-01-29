package watchdog

import (
	"os"
	"syscall"
)

const logDir = "./watchdog"

func newProc(svc *Service, null, pw *os.File) *os.ProcAttr {
	return &os.ProcAttr{Dir: svc.path, Files: []*os.File{null, pw, pw}}
}

func setPriority(pid, priority uintptr) syscall.Errno {
	return 0
}
