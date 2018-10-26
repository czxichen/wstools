package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

var (
	fsnotify = cli.FsnotifyConfig{}
	// Fsnotify 目录监控命令
	Fsnotify = &cobra.Command{
		Use: "fsnotify",
		Example: "	-d tools -s scripts.bat",
		Run:   fsnotifyRun,
		Short: "可以用来监控文件或者目录的变化",
		Long:  "为了不重复执行脚本,十秒内的改变只会执行一次脚本,脚本路径为空则不执行任何操作",
	}
)

func init() {
	Fsnotify.PersistentFlags().StringVarP(&fsnotify.Dir, "dir", "d", "./", "指定要监控的目录和-s结合使用,当指定-f的时候此参数不生效")
	Fsnotify.PersistentFlags().StringVarP(&fsnotify.Script, "script", "s", "", "指定当目录发生改变的时候调用此脚本,为空则不做任何操作")
	Fsnotify.PersistentFlags().BoolVarP(&fsnotify.Debug, "debug", "D", false, "是否打印详细的变化信息")
	Fsnotify.PersistentFlags().StringVarP(&fsnotify.Path, "config", "c", "", `从文件中读取配置,可以同时多个目录,每行一个,目录和脚本用','隔开`)
}

func fsnotifyRun(cmd *cobra.Command, arg []string) {
	ctx, cancel := context.WithCancel(context.Background())
	pathMap, err := cli.ParseNotifyConfig(&fsnotify)
	if err == nil {
		if fsnotify.Debug {
			fmt.Printf("Path and Script info:%v\n", pathMap)
		}
		err = cli.FsnotifyRun(ctx, pathMap, nil)
	}
	if err != nil {
		cli.FatalOutput(1, "Fsnotify 执行错误:%s\n", err.Error())
	}
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	select {
	case <-signalChan:
		cancel()
	}
}
