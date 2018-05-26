package command

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

type fsnotify_config struct {
	Dir    string
	Script string
	Path   string
	Debug  bool
}

var (
	_fsnotify = fsnotify_config{}
	Fsnotify  = &cobra.Command{
		Use: "fsnotify",
		Example: "	-d tools -s scripts.bat",
		RunE:  fsnotify_run,
		Short: "可以用来监控文件或者目录的变化",
		Long:  "为了不重复执行脚本,十秒内的改变只会执行一次脚本,脚本路径为空则不执行任何操作",
	}
)

func init() {
	Fsnotify.PersistentFlags().StringVarP(&_fsnotify.Dir, "dir", "d", "./", "指定要监控的目录和-s结合使用,当指定-f的时候此参数不生效")
	Fsnotify.PersistentFlags().StringVarP(&_fsnotify.Script, "script", "s", "", "指定当目录发生改变的时候调用此脚本,为空则不做任何操作")
	Fsnotify.PersistentFlags().BoolVarP(&_fsnotify.Debug, "debug", "D", false, "是否打印详细的变化信息")
	Fsnotify.PersistentFlags().StringVarP(&_fsnotify.Path, "config", "c", "", `从文件中读取配置,可以同时多个目录,每行一个,目录和脚本用','隔开`)
}

func fsnotify_run(cmd *cobra.Command, arg []string) error {
	var m = make(map[string]string)
	if _fsnotify.Path != "" {
		File, err := os.Open(_fsnotify.Path)
		if err != nil {
			fmt.Printf("读取配置文件失败:%s\n", err.Error())
			return nil
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
				fmt.Printf("读取数据失败:%s", err.Error())
				return nil
			}
			list = bytes.Split(line, []byte(","))
			if len(list) == 2 {
				m[filepath.Clean(string(bytes.TrimSpace(list[0])))] = filepath.Clean(string(bytes.TrimSpace(list[1])))
			} else {
				if len(list) != 1 {
					fmt.Printf("无效数据:%s\n", string(line))
					return nil
				}
				m[filepath.Clean(string(bytes.TrimSpace(list[0])))] = ""
			}
			if len(m) == 0 {
				fmt.Printf("未发现有效的数据,文件格式:/path/dir,script/path")
				return nil
			}
		}
	} else {
		if _fsnotify.Dir != "" {
			m[_fsnotify.Dir] = _fsnotify.Script
		} else {
			return fmt.Errorf("必须指定-c或者-d参数")
		}
	}
	if _fsnotify.Debug {
		fmt.Println(m)
	}
	err := watchfs(m)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func watchfs(env map[string]string) error {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watch.Close()

	for k, _ := range env {
		err = watch.Add(k)
		fmt.Printf("Add watch:%s\n", k)
		if err != nil {
			return err
		}
	}

	var ch = make(chan string, 1)
	go execute(ch, env)
	for {
		select {
		case event := <-watch.Events:
			if _fsnotify.Debug {
				fmt.Println(event.String())
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

			fmt.Printf("Path %s changed\n", path)
			m.lock.Lock()
			m.event[path] = true
			m.lock.Unlock()

			if script != "" {
				go func() {
					<-time.After(10 * time.Second)
					fmt.Printf("Start run %s\n", script)
					cmd := exec.Command(script)
					buf, err := cmd.Output()
					if err != nil {
						fmt.Printf("Run script faild,%s\n", err.Error())
					} else {
						fmt.Printf("Script %s result:%s\n", script, buf)
					}
					delete(m.event, path)
				}()
			}
		}
	}
}
