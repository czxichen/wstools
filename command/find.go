package command

import (
	"fmt"

	"github.com/czxichen/wstools/common/cli"

	"github.com/spf13/cobra"
)

// Find 查找命令
var Find = &cobra.Command{
	Use: "find",
	Example: `	查找./下小于1MB的文件
	-p ./ -s "-1M"
	查找./目录下,修改时间在一天内且大于1MB的文件
	-p ./ -s "1M" -m "-1d"`,
	Short: "文件查找",
	Long:  "按名称,时间,大小查找文件",
	Run:   findRun,
}

var findConfig cli.FindConfig

func init() {
	Find.PersistentFlags().BoolVarP(&findConfig.Dir, "dir", "d", false, "是否启用查找目录")
	Find.PersistentFlags().StringVarP(&findConfig.Path, "path", "p", "", "查找的路径")
	Find.PersistentFlags().StringVarP(&findConfig.Name, "name", "n", "", "按照名称正则查找文件")
	Find.PersistentFlags().StringVarP(&findConfig.Size, "size", "s", "", "按照大小查找文件,单位(K,M,G不分大小写),不带单位默认字节")
	Find.PersistentFlags().StringVarP(&findConfig.Mtime, "mtime", "m", "", "按照修改时间查找文件,单位(M,H,d,m,y区分大小写),不带单位默认为秒")
}

func findRun(cmd *cobra.Command, args []string) {
	cli.Find(&findConfig, func(path string) error {
		fmt.Println(path)
		return nil
	})
}
