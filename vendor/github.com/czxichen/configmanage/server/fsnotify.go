package server

import (
	"log"
	"os"
	"time"

	"github.com/howeyc/fsnotify"
)

func Notify(path string, interval int, run func()) {
	Env, err := notify(path)
	if err != nil {
		log.Printf("监控目录%s失败:%s\n", path, err)
		os.Exit(1)
	}
	var IsRunning bool
	for {
		v := <-Env
		log.Println(v)
		if IsRunning {
			continue
		}
		IsRunning = true
		go func() {
			time.Sleep(1e9 * time.Duration(interval))
			log.Printf("目录%s已改动,重新加载:", path)
			run()
			IsRunning = false
		}()
	}
}

func notify(path string) (chan *fsnotify.FileEvent, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = w.Watch(path)
	if err != nil {
		return nil, err
	}
	return w.Event, nil
}
