package watchdog

import (
	"os"
	"syscall"
)

const logDir = "/var/log/watchdog"

func newProc(svc *Service, null, pw *os.File) *os.ProcAttr {
	return &os.ProcAttr{
		Dir:   svc.path,
		Files: []*os.File{null, pw, pw},
		Sys: &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: svc.uid,
				Gid: svc.gid,
			},
			Setpgid: true,
		},
	}
}

func setPriority(pid, priority uintptr) syscall.Errno {
	_, _, err := syscall.Syscall(syscall.SYS_SETPRIORITY, uintptr(prioProcess), pid, priority)
	return err
}
