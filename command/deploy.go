package command

import (
	"github.com/czxichen/configmanage/client"
	"github.com/czxichen/configmanage/server"
	"github.com/spf13/cobra"
)

// Deploy 部署命令
var Deploy = &cobra.Command{
	Use:   `deploy`,
	Short: "快速搭建服务器",
	Long: `部署文件打包好,根据规则写好模版文件放在服务端,然后客户端执行相应的命令,即可快速部署服务器
`,
}

func init() {
	Deploy.AddCommand(server.DeployServer, client.DeployClient)
}
