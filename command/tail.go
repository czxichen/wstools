package command

import (
	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

// Tail tail command
var Tail = &cobra.Command{
	Use: `tail`,
	Run: tailRun,
	Example: `	-f main.go -l 10 -n 5 -o tmp.txt`,
	Short: "从文件尾部操作文件",
	Long: `从文件结尾或指定位置读取内容,可以按行读取,也可以按大小读取,-i 和 -l同时使用的话-i生效,-s 与 -n 
同时使用的话-s生效
`}

var tailConfig cli.TailConfig

func init() {
	Tail.PersistentFlags().StringVarP(&tailConfig.Output, "Output", "o", "", "-o 指定输出的路径,不指定则输出到标准输出")
	Tail.PersistentFlags().IntVarP(&tailConfig.Line, "Line", "l", 0, "-l 指定从倒数第几行开始读取")
	Tail.PersistentFlags().IntVarP(&tailConfig.Lines, "number", "n", 0, "-n 指定读取的行数")
	Tail.PersistentFlags().StringVarP(&tailConfig.Offset, "index", "i", "", "-i 指定开始读取的位置,单位:b,kb,mb,默认单位:b")
	Tail.PersistentFlags().StringVarP(&tailConfig.Size, "Size", "s", "", "-s 指定读取的大小,单位:b,kb,mb,默认单位:b")
	Tail.PersistentFlags().StringVarP(&tailConfig.File, "File", "f", "", "-f 指定要查看的文件路径")
}

// TODO:
func tailRun(cmd *cobra.Command, args []string) {
	cli.RunTail(&tailConfig)
}
