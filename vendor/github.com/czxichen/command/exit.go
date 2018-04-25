package command

import (
	"os"
	"reflect"
	"sync"
)

type exit struct {
	lock  sync.RWMutex
	funcs []interface{}
}

var Exit = &exit{funcs: make([]interface{}, 0, 5)}

//注只支持func()或者func()error类型的函数
//用来做清理用的
func (self *exit) RegisterFunc(f interface{}) {
	self.lock.Lock()
	defer self.lock.Unlock()

	typ := reflect.TypeOf(f).String()
	if typ != "func()" && typ != "func() error" {
		return
	}
	self.funcs = append(self.funcs, f)
}

//执行所有的函数,并清空self.funcs
func (self *exit) Exec() {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, f := range self.funcs {
		switch reflect.TypeOf(f).String() {
		case "func()":
			fn, ok := f.(func())
			if ok {
				fn()
			}
		case "func() error":
			fn, ok := f.(func() error)
			if ok {
				fn()
			}
		}
	}
	self.funcs = self.funcs[:0]
}

//执行所有的函数,并退出程序
func (self *exit) Exit(code int) {
	self.Exec()
	os.Exit(code)
}
