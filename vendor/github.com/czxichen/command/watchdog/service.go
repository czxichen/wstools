package watchdog

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

//+build windows,linux

//默认的优先级为0
const prioProcess = 0

//定义服务的类型.
type Service struct {
	name   string
	binary string
	path   string
	args   []string

	uid      uint32
	gid      uint32
	priority int

	dependencies map[string]*Service
	dependents   map[string]*Service

	termTimeout time.Duration

	lock    sync.Mutex
	process *os.Process

	done     chan bool
	shutdown chan bool
	started  chan bool
	stopped  chan bool

	failures uint64
	restarts uint64

	lastFailure time.Time
	lastRestart time.Time
}

//初始化一个Service.
func newService(name, binary string) *Service {
	return &Service{
		name:         name,
		binary:       binary,
		args:         make([]string, 0),
		dependencies: make(map[string]*Service),
		dependents:   make(map[string]*Service),

		done:     make(chan bool),
		shutdown: make(chan bool, 1),
		started:  make(chan bool, 1),
		stopped:  make(chan bool, 1),

		termTimeout: 5 * time.Second,
	}
}

//给这个服务添加依赖.
func (svc *Service) AddDependency(name string) {
	svc.dependencies[name] = nil
}

//为服务添加启动参数.
func (svc *Service) AddArgs(args string) {
	svc.args = strings.Fields(args)
}

//为进程设置优先级,Windows下面无效.
func (svc *Service) SetPriority(priority int) error {
	if priority < -20 || priority > 19 {
		return fmt.Errorf("Invalid priority %d - must be between -20 and 19", priority)
	}
	svc.priority = priority
	return nil
}

func (svc *Service) SetTermTimeout(tt time.Duration) {
	svc.termTimeout = tt
}

func (svc *Service) SetUser(username string) error {
	u, err := user.Lookup(username)
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return err
	}
	svc.uid = uint32(uid)
	svc.gid = uint32(gid)
	return nil
}

func (svc *Service) run() {
	//如果存在依赖,要等依赖全部启动完毕之后才会自动自身.
	for _, dep := range svc.dependencies {
		log.Printf("Service %s waiting for %s to start", svc.name, dep.name)
		select {
		case started := <-dep.started:
			dep.started <- started
		case <-svc.shutdown:
			goto done
		}
	}

	for {
		//如果启动失败,怎等待时间会延长,最大不超过restartBackoffMax时间
		//程序启动必须是阻塞的,不然会重复运行
		if svc.failures > 0 {
			delay := time.Duration(svc.failures) * restartBackoff
			if delay > restartBackoffMax {
				delay = restartBackoffMax
			}
			log.Printf("Service %s has failed %d times - delaying %s before restart",
				svc.name, svc.failures, delay)

			select {
			case <-time.After(delay):
			case <-svc.shutdown:
				goto done
			}
		}

		svc.restarts++
		svc.lastRestart = time.Now()
		svc.runOnce()

		select {
		case <-time.After(restartDelay):
		case <-svc.shutdown:
			goto done
		}
	}
done:
	svc.done <- true
}

//为服务创建日志文件
func (svc *Service) logFile() (*os.File, error) {
	logName := svc.name + ".log"

	if err := os.MkdirAll(logDir, 0666); err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}
	f, err := os.OpenFile(path.Join(logDir, logName), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(f, "Log file for %s (stdout/stderr)\n", svc.name)
	fmt.Fprintf(f, "Created at: %s\n", time.Now().Format("2006/01/02 15:04:05"))
	return f, nil
}

//运行程序
func (svc *Service) runOnce() {
	args := make([]string, len(svc.args)+1)
	args[0] = svc.name
	copy(args[1:], svc.args)

	null, err := os.Open(os.DevNull)
	if err != nil {
		log.Printf("Service %s - failed to open %s: %v", svc.name, os.DevNull, err)
		return
	}

	lfile, err := svc.logFile()
	if err != nil {
		log.Printf("Service %s - failed to create log file: %v", svc.name, err)
		null.Close()
		return
	}

	attr := newProc(svc, null, lfile)

	log.Printf("Starting service %s...", svc.name)
	proc, err := os.StartProcess(svc.binary, args, attr)
	if err != nil {
		log.Printf("Service %s failed to start: %v", svc.name, err)
		svc.lastFailure = time.Now()
		svc.failures++
		null.Close()
		return
	}

	null.Close()
	lfile.Close()
	svc.lock.Lock()
	svc.process = proc
	svc.lock.Unlock()

	if err := setPriority(uintptr(proc.Pid), uintptr(svc.priority)); err != 0 {
		log.Printf("Failed to set priority to %d for service %s: %v", svc.priority, svc.name, err)
	}
	select {
	case svc.started <- true:
	default:
	}

	state, err := svc.process.Wait()
	if err != nil {
		log.Printf("Service %s wait failed with %v", svc.name, err)
		svc.lastFailure = time.Now()
		svc.failures++
		return
	}
	if !state.Success() {
		log.Printf("Service %s exited with %v", svc.name, state)
		svc.lastFailure = time.Now()
		svc.failures++
		return
	}

	svc.failures = 0
	log.Printf("Service %s exited normally.", svc.name)
}

//给进程发送信号
func (svc *Service) signal(sig os.Signal) error {
	svc.lock.Lock()
	defer svc.lock.Unlock()
	if svc.process == nil {
		return nil
	}
	return svc.process.Signal(sig)
}

//停止服务
func (svc *Service) stop() {
	log.Printf("Stopping service %s...", svc.name)
	//等待依赖它的进程退出完毕之后再退出自己.
	for _, dep := range svc.dependents {
		log.Printf("Service %s waiting for %s to stop", svc.name, dep.name)
		stopped := <-dep.stopped
		dep.stopped <- stopped
	}

	svc.shutdown <- true
	//首先给进程发送退出信号,如果超时没有退出,则直接发送Kill信号.
	svc.signal(syscall.SIGTERM)
	select {
	case <-svc.done:
	case <-time.After(svc.termTimeout):
		svc.signal(syscall.SIGKILL)
		<-svc.done
	}
	log.Printf("Service %s stopped", svc.name)
	svc.stopped <- true
}
