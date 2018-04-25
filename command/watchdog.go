package command

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/czxichen/command/watchdog"
	conf "github.com/dlintw/goconf"
	"github.com/spf13/cobra"
)

var Watchdog = &cobra.Command{
	Use:     `watchdog`,
	Example: "-c watch.ini",
	RunE:    watchdog_run,
	Short:   "进程守护",
	Long:    `用来监控进程,可以带依赖模式监控`,
}

type watchdog_config struct {
	logpath    string
	configFile string
	createcfg  bool
}

var _watchdog watchdog_config

func init() {
	Watchdog.PersistentFlags().BoolVarP(&_watchdog.createcfg, "createcfg", "C", false, "创建配置样例文件")
	Watchdog.PersistentFlags().StringVarP(&_watchdog.logpath, "log", "l", "watchdog.log", "指定log输出到文件")
	Watchdog.PersistentFlags().StringVarP(&_watchdog.configFile, "config", "c", "watchdog.ini", "指定watchdog的配置文件")
}

func watchdog_run(cmd *cobra.Command, args []string) error {
	logFile, err := os.Create(_watchdog.logpath)
	if err != nil {
		log.Fatalf("创建日志文件失败:%s\n", err.Error())
	}

	defer logFile.Close()
	log.SetOutput(logFile)

	if _watchdog.createcfg {
		File, err := os.Create(_watchdog.configFile)
		if err != nil {
			log.Fatalf("创建配置示例文件失败:%s\n", err.Error())
		}
		File.WriteString(`[Srv_01]
binary = binarypath
args = arg01
user = root
term_timeout = 10s
priority = -10 

[Srv_02]
binary = binarypath
args = arg01 arg02
user = root
term_timeout = 10s
priority = -10
dependency =  Srv_01`)
		File.Close()
		return nil
	}
	cfg, err := conf.ReadConfigFile(_watchdog.configFile)
	if err != nil {
		log.Fatalf("读取配置文件失败 %q: %v", _watchdog.configFile, err)
	}

	fido := watchdog.NewWatchdog()
	shutdownHandler(fido)
	for _, name := range cfg.GetSections() {
		if name == "default" {
			continue
		}
		binary := svcOpt(cfg, name, "binary", true)
		args := svcOpt(cfg, name, "args", false)

		svc, err := fido.AddService(name, binary)
		if err != nil {
			log.Fatalf("添加服务失败 %q: %v", name, err)
		}
		svc.AddArgs(args)
		if dep := svcOpt(cfg, name, "dependency", false); dep != "" {
			svc.AddDependency(dep)
		}
		if opt := svcOpt(cfg, name, "priority", false); opt != "" {
			prio, err := strconv.Atoi(opt)
			if err != nil {
				log.Fatalf("服务 %s 设置了无效的优先级 %q: %v", name, opt, err)
			}
			if err := svc.SetPriority(prio); err != nil {
				log.Fatalf("设置服务优先级失败 %s: %v", name, err)
			}
		}
		if opt := svcOpt(cfg, name, "term_timeout", false); opt != "" {
			tt, err := time.ParseDuration(opt)
			if err != nil {
				log.Fatalf("服务 %s 设置了无效的退出超时时间 %q: %v", name, opt, err)
			}
			svc.SetTermTimeout(tt)
		}

		if user := svcOpt(cfg, name, "user", false); user != "" {
			if err := svc.SetUser(user); err != nil {
				log.Fatalf("设置服务用户失败 %s: %v", name, err)
			}
		}
	}
	fido.Walk()
	return nil
}

func cfgOpt(cfg *conf.ConfigFile, section, option string) string {
	if !cfg.HasOption(section, option) {
		return ""
	}
	s, err := cfg.GetString(section, option)
	if err != nil {
		log.Fatalf("Failed to get %s for %s: %v", option, section, err)
	}
	return s
}

func svcOpt(cfg *conf.ConfigFile, service, option string, required bool) string {
	opt := cfgOpt(cfg, service, option)
	if opt == "" && required {
		log.Fatalf("Service %s has missing %s option", service, option)
	}
	return opt
}

var signalNames = map[syscall.Signal]string{
	syscall.SIGINT:  "SIGINT",
	syscall.SIGQUIT: "SIGQUIT",
	syscall.SIGTERM: "SIGTERM",
}

func signalName(s syscall.Signal) string {
	if name, ok := signalNames[s]; ok {
		return name
	}
	return fmt.Sprintf("SIG %d", s)
}

type Shutdowner interface {
	Shutdown()
}

func shutdownHandler(server Shutdowner) {
	sigc := make(chan os.Signal, 3)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		for s := range sigc {
			name := s.String()
			if sig, ok := s.(syscall.Signal); ok {
				name = signalName(sig)
			}
			log.Printf("Received %v, initiating shutdown...", name)
			server.Shutdown()
		}
	}()
}
