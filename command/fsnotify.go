package command

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	fsnotify "github.com/go-fsnotify/fsnotify"
)

type notifycfg struct {
	dir    string
	script string
	file   string
	debug  bool
}

var (
	notify   = notifycfg{}
	Fsnotify = &Command{
		UsageLine: `fsnotify -d tools -s scripts.bat`,
		Run:       watch,
		Short:     "可以用来监控文件或者目录的变化",
		Long: `为了不重复执行脚本,十秒内的改变只会执行一次脚本,脚本路径为空则不执行任何操作
`,
	}
)

func init() {
	Fsnotify.Flag.StringVar(&notify.dir, "d", "./", `-d="./" 指定要监控的目录和-s结合使用,当指定-f的时候此参数不生效`)
	Fsnotify.Flag.StringVar(&notify.script, "s", "", "-s scriptpath 指定当目录发生改变的时候调用此脚本,为空则不做任何操作")
	Fsnotify.Flag.BoolVar(&notify.debug, "D", false, "-D=true 是否打印详细的变化信息")
	Fsnotify.Flag.StringVar(&notify.file, "f", "", `-f configpath 从文件中读取配置,可以同时多个目录,每行一个,目录和脚本用","隔开`)
}

func watch(cmd *Command, arg []string) bool {
	var m = make(map[string]string)
	if notify.file != "" {
		File, err := os.Open(notify.file)
		if err != nil {
			log.Println(err.Error())
			return true
		}
		defer File.Close()
		r := bufio.NewReader(File)
		var list [][]byte
		for {
			line, _, err := r.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println(err.Error())
				return true
			}
			list = bytes.Split(line, []byte(","))
			if len(list) == 2 {
				m[filepath.Clean(string(bytes.TrimSpace(list[0])))] = filepath.Clean(string(bytes.TrimSpace(list[1])))
			} else {
				if len(list) != 1 {
					log.Printf("无效数据:%s\n", string(line))
					return false
				}
				m[filepath.Clean(string(bytes.TrimSpace(list[0])))] = ""
			}
			if len(m) == 0 {
				log.Println("未发现有效的数据,文件格式:/path/dir,script/path")
				return true
			}
		}
	} else {
		if notify.dir != "" {
			m[notify.dir] = notify.script
		} else {
			return false
		}
	}
	if notify.debug {
		log.Println(m)
	}
	err := watchfs(m)
	if err != nil {
		log.Println(err)
	}

	return true
}

func watchfs(env map[string]string) error {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watch.Close()

	for k, _ := range env {
		err = watch.Add(k)
		log.Printf("Add watch:%s\n", k)
		if err != nil {
			return err
		}
	}

	var ch = make(chan string, 1)
	go execute(ch, env)
	for {
		select {
		case event := <-watch.Events:
			if notify.debug {
				log.Println(event.String())
			}
			ch <- event.Name
		case err = <-watch.Errors:
			return err
		}
	}
}

func execute(ch chan string, env map[string]string) {
	var m = struct {
		event map[string]bool
		lock  sync.RWMutex
	}{event: make(map[string]bool)}
	var path string
	for {
		path = <-ch
		m.lock.RLock()
		ok := m.event[path]
		m.lock.RUnlock()
		if !ok {
			script, ok := env[path]
			if !ok {
				var tem string
				for tem, script = range env {
					if strings.HasPrefix(path, tem) {
						break
					}
				}
			}

			log.Printf("Path %s changed\n", path)
			m.lock.Lock()
			m.event[path] = true
			m.lock.Unlock()

			if script != "" {
				go func() {
					<-time.After(10 * time.Second)
					log.Printf("Start run %s\n", script)
					cmd := exec.Command(script)
					buf, err := cmd.Output()
					if err != nil {
						log.Printf("Run script faild,%s\n", err.Error())
					} else {
						log.Printf("Script %s result:%s\n", script, buf)
					}
					delete(m.event, path)
				}()
			}
		}
	}
}
