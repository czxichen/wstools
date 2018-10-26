package command

import (
	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

// HTTP Http
var HTTP = &cobra.Command{
	Use: "http",
	Example: `	开启目录访问uuid,并设置baseauth用户名和密码
	-d uuid -u root -p toor -i
	下载文件,并保存
	-w -H http://127.0.0.1:1789/type.proto -s t.proto -u root -p toor`,
	Short: "使用简单的http协议通讯",
	Long:  "使用http协议进行内容传输,支持文件上传下载",
	Run:   httpRun,
}

var httpConfig cli.HTTPConfig

func init() {
	HTTP.PersistentFlags().StringVarP(&httpConfig.Host, "host", "H", ":1789", "指定监听的地址端口,或者要访问的url")
	HTTP.PersistentFlags().StringVarP(&httpConfig.User, "user", "u", "", "指定BaseAuth的用户名,可以为空")
	HTTP.PersistentFlags().StringVarP(&httpConfig.Passwd, "passwd", "p", "", "指定BaseAuth的密码,可以为空")
	HTTP.PersistentFlags().StringVarP(&httpConfig.Crt, "crt", "c", "", "指定TLS的Crt文件,可以为空")
	HTTP.PersistentFlags().StringVarP(&httpConfig.Key, "key", "k", "", "指定TLS的Key文件,可以为空")
	HTTP.PersistentFlags().StringVarP(&httpConfig.Dir, "dir", "d", "", "指定共享目录,当server启动的时候不能为空")
	HTTP.PersistentFlags().StringVarP(&httpConfig.Save, "save", "s", "", "使用下载的时候,文件保存路径,为空则保存在当前目录")
	HTTP.PersistentFlags().BoolVarP(&httpConfig.Wget, "wget", "w", false, "从指定的host下载文件")
	HTTP.PersistentFlags().BoolVarP(&httpConfig.Quic, "quic", "q", false, "使用quic协议,默认会监听tcp,udp上")
	HTTP.PersistentFlags().BoolVarP(&httpConfig.OnlyQuic, "onlyquic", "o", false, "仅启动quic协议,只监听在udp")
	HTTP.PersistentFlags().BoolVarP(&httpConfig.Index, "index", "i", false, "启用目录索引,允许目录浏览")
	HTTP.PersistentFlags().BoolVarP(&httpConfig.Verbose, "verbose", "v", true, "关闭后台访问输出")
}

func httpRun(cmd *cobra.Command, args []string) {
	if err := cli.HTTPRun(&httpConfig); err != nil {
		cli.FatalOutput(1, "http run error:%s\n", err.Error())
	}
}
