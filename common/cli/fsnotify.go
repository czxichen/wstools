package cli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

/*
var m = make(map[string]string)
	if fsnotify.Path != "" {
		File, err := os.Open(fsnotify.Path)
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
		if fsnotify.Dir != "" {
			m[fsnotify.Dir] = fsnotify.Script
		} else {
			return fmt.Errorf("必须指定-c或者-d参数")
		}
	}
	if fsnotify.Debug {
		fmt.Println(m)
	}
	err := watchfs(m)
	if err != nil {
		fmt.Println(err)
	}
	return nil
*/

// FsnotifyConfig 目录监控配置
type FsnotifyConfig struct {
	Dir    string
	Script string
	Path   string
	Debug  bool
}

// ParseNotifyConfig 解析配置
func ParseNotifyConfig(cfg *FsnotifyConfig) (map[string]string, error) {
	pathMap := make(map[string]string)
	if cfg.Path != "" {
		File, err := os.Open(cfg.Path)
		if err != nil {
			return nil, err
		}
		defer File.Close()
		fileBuf := bufio.NewReader(File)
		var lineIndex = 1
		for {
			line, _, err := fileBuf.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
			list := bytes.Split(line, []byte(","))
			if len(list) == 2 {
				pathMap[filepath.Clean(string(bytes.TrimSpace(list[0])))] = filepath.Clean(string(bytes.TrimSpace(list[1])))
			} else {
				if len(list) != 1 {
					return nil, fmt.Errorf("line %d invalid line data:%s", lineIndex, line)
				}
				pathMap[filepath.Clean(string(bytes.TrimSpace(list[0])))] = ""
			}
			lineIndex++
		}
	}

	if cfg.Dir != "" {
		pathMap[cfg.Dir] = cfg.Script
	}

	if len(pathMap) == 0 {
		return nil, fmt.Errorf("未发现有效的数据,文件格式:/path/dir,script/path")
	}
	return pathMap, nil
}

// FsnotifyRun 执行监控
func FsnotifyRun(ctx context.Context, pathMap map[string]string, eventHook func(event fsnotify.Event)) error {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watch.Close()

	for path := range pathMap {
		if err = watch.Add(path); err != nil {
			return err
		}
	}
	if eventHook == nil {
		eventMap := make(map[string]struct{})
		eventHook = func(event fsnotify.Event) {
			_, exist := eventMap[event.Name]
			if exist {
				return
			}
			eventMap[event.Name] = struct{}{}
			script := pathMap[event.Name]
			if script != "" {
				go func() {
					<-time.After(10 * time.Second)
					buf, err := exec.Command(script).Output()
					if err == nil {
						fmt.Printf("Execute script:%s success:%s\n", script, buf)
					} else {
						fmt.Printf("Execute script:%s error:%s\n", script, err.Error())
					}
					delete(eventMap, event.Name)
				}()
			}
		}
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-watch.Events:
			eventHook(event)
		case err = <-watch.Errors:
			return err
		}
	}
}
