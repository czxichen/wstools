package command

import (
	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

// Ftp ftp  命令
var Ftp = &cobra.Command{
	Use: "ftp",
	Example: `	下载远程/tmp/main.go文件到./目录下
	-D -u root -p toor -d /server/Res/scenes.ini -s nima -H  127.0.0.1:21
	上传文件到/Server/main.go
	-u root -p toor -s main.go -d /Server/main.go -H 127.0.0.1:21`,
	Short: "FTP上传下载",
	Long:  "使用简单的FTP协议实现文件的上传下载",
	Run:   ftpRun,
}

var ftpConfig FTPConfig

func init() {
	Ftp.PersistentFlags().StringVarP(&ftpConfig.Host, "host", "H", "", "指定ftp地址端口,不能为空")
	Ftp.PersistentFlags().StringVarP(&ftpConfig.User, "user", "u", "", "指定登录的用户名,不能为空")
	Ftp.PersistentFlags().StringVarP(&ftpConfig.Passwd, "passwd", "p", "", "指定登录的用户密码,不能为空")
	Ftp.PersistentFlags().StringVarP(&ftpConfig.Source, "source", "s", "", "指定原始文件路径,不能为空")
	Ftp.PersistentFlags().StringVarP(&ftpConfig.Destination, "destination", "d", "", "指定目标文件路径,不能为空")
	Ftp.PersistentFlags().BoolVarP(&ftpConfig.Download, "download", "D", false, "从ftp上下载文件")
}

func ftpRun(cmd *cobra.Command, args []string) {
	if ftpConfig.Host == "" || ftpConfig.User == "" || ftpConfig.Passwd == "" || ftpConfig.Source == "" || ftpConfig.Destination == "" {
		cli.FatalOutput(1, "参数错误\n")
	}

	conn, err := cli.NewFTP(ftpConfig.Host, ftpConfig.User, ftpConfig.Passwd)
	if err != nil {
		cli.FatalOutput(1, "登录失败:%s\n", err.Error())
	}

	if ftpConfig.Download {
		err = conn.GetFile(ftpConfig.Source, ftpConfig.Destination)
	} else {
		err = conn.PutFile(ftpConfig.Source, ftpConfig.Destination)
	}
	if err != nil {
		cli.FatalOutput(1, "文件传输失败:%s\n", err.Error())
	}
}

// FTPConfig ftp 配置
type FTPConfig struct {
	Host        string
	User        string
	Passwd      string
	Source      string
	Destination string
	Download    bool
}
