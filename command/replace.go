package command

import (
	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

// Replace  替换命令
var Replace = &cobra.Command{
	Use: "replace",
	Example: `	替换指定路径下,后缀为.go的文件内容
	-p sourcepath -o "oldstr" -n "newstr"  -s .go`,
	Short: "文本替换工具",
	Long:  "文本批量替换工具,支持正则匹配",
	Run:   replaceRun,
}

var replaceCfg cli.ReplaceConfig

func init() {
	Replace.PersistentFlags().StringVarP(&replaceCfg.OldB, "old", "o", "", "被替换的内容,当是正则表达式的时候必须结合-r使用")
	Replace.PersistentFlags().StringVarP(&replaceCfg.NewB, "new", "n", "", "替换的内容")
	Replace.PersistentFlags().StringVarP(&replaceCfg.Path, "path", "p", "", "指定要操作的文件或目录")
	Replace.PersistentFlags().StringVarP(&replaceCfg.Suffix, "suffix", "s", "", "指定要匹配的文件后缀")
	Replace.PersistentFlags().StringVarP(&replaceCfg.MatchStr, "match", "m", "", "使用正则匹配文件名称,如果和-s同时使用,则-s生效")
	Replace.PersistentFlags().BoolVarP(&replaceCfg.UseReg, "regexp", "r", false, "使用正比表达式匹配要替换的内容")
	Replace.PersistentFlags().BoolVarP(&replaceCfg.Quick, "quick", "q", true, "文件内容全部加载到内存中替换")
}

func replaceRun(cmd *cobra.Command, args []string) {
	if err := cli.ReplaceRun(&replaceCfg); err != nil {
		cli.FatalOutput(1, "Replace run error:%s\n", err.Error())
	}
}
