package command

import (
	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

// Net net 命令
var Net = &cobra.Command{
	Use: "net",
	Example: `	并发模式ping测试指定主机的网络连通性
	-a ping -i www.baidu.com,www.163.com -c 4 -q
	telnet测试指定端口是否开启
	-a telnet -i www.baidu.com:80`,
	Short: "网络连通性工具",
	Long:  "网络连通性工具,支持ping|telnet命令",
	Run:   netRun,
}

var netConfig cli.NetConfig

func init() {
	Net.PersistentFlags().StringVarP(&netConfig.Action, "action", "a", "ping", "指定动作只支持ping,telnet两种")
	Net.PersistentFlags().StringVarP(&netConfig.Host, "ip", "i", "", "指定目标地址,当-a为telnet的时候远程地址必须包含端口,多地址用','分割")
	Net.PersistentFlags().StringVarP(&netConfig.Hosts, "hosts", "H", "", "文件读取目标地址,按行解析,如果指定-h则此参数无效")
	Net.PersistentFlags().IntVarP(&netConfig.TimeOut, "timeout", "t", 5, "设置超时时间,使用telnet的时候此参数有效")
	Net.PersistentFlags().IntVarP(&netConfig.Count, "count", "c", 2, "指定发出ping的次数")
	Net.PersistentFlags().BoolVarP(&netConfig.Sum, "sum", "s", false, "以统计方式输出结果")
	Net.PersistentFlags().BoolVarP(&netConfig.Quick, "quick", "q", false, "使用并发模式")
}

func netRun(cmd *cobra.Command, args []string) {
	if err := cli.NetRun(&netConfig); err != nil {
		cli.FatalOutput(1, "net run error:%s\n", err.Error())
	}
}
