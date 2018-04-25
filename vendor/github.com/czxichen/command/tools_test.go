package command

import (
	"runtime"
	"testing"
)

func Test_exit(t *testing.T) {
	Exit.RegisterFunc(func() { t.Log("Exit func") })
	Exit.RegisterFunc(func() error {
		t.Log("Return Error")
		return nil
	})
	Exit.RegisterFunc(func(int) { t.Log("Error func") })
	Exit.Exec()
}

func Test_md5(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	t.Log(FileMd5(filename))
}

func Test_rand(t *testing.T) {
	t.Log(NewID())
}
