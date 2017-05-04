package command

import (
	"fmt"
	"os"

	"github.com/czxichen/configmanage/client"
	"github.com/czxichen/configmanage/server"
)

var Deploy = &Command{
	UsageLine: `deploy server|client -h`,
	Run:       deploy,
	Short:     "快速搭建服务器",
	Long: `部署文件打包好,根据规则写好模版文件放在服务端,然后客户端执行相应的命令,即可快速部署服务器
`,
}

func deploy(cmd *Command, args []string) bool {
	if len(args) < 1 {
		return false
	}
	switch args[0] {
	case "server":
		if !server.Server(args[1:]) {
			fmt.Fprintln(os.Stderr, "Example: wstools.exe deploy server -c cfg.json")
			server.FlagSet.PrintDefaults()
			fmt.Fprintf(os.Stderr, "Description:\n\t%s\n", cmd.Long)
			os.Exit(1)
		}
	case "client":
		if !client.Client(args[1:]) {
			fmt.Fprintln(os.Stderr, "Example: wstools.exe deploy client -a install")
			client.FlagSet.PrintDefaults()
			fmt.Fprintf(os.Stderr, "Description:\n\t%s\n", cmd.Long)
			os.Exit(1)
		}
	default:
		return false
	}
	return true
}
