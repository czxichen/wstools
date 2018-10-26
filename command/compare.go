package command

import (
	"fmt"
	"path/filepath"

	"github.com/czxichen/command"
	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

// Compare 文件比较命令
var Compare = &cobra.Command{
	Use: "compare",
	Example: `	比较uuid和uuid_new目录,如果存在差异文件,则拷贝到diff目录中
	-s uuid -d uuid_new -c diff`,
	Short: "文件目录比较",
	Long:  "文件目录进行比较,支持取出差异包",
	Run:   compareRun,
}

type compare struct {
	verbose     bool
	copyto      string
	source      string
	destination string
}

var compareConfig compare

func init() {
	Compare.PersistentFlags().StringVarP(&compareConfig.source, "source", "s", "", "指定原始路径,不能为空")
	Compare.PersistentFlags().StringVarP(&compareConfig.destination, "destination", "d", "", "指定目标路径,不能为空")
	Compare.PersistentFlags().StringVarP(&compareConfig.copyto, "copy", "c", "", "差异文件拷贝到此路径下")
	Compare.PersistentFlags().BoolVarP(&compareConfig.verbose, "verbose", "v", true, "输出详情")
}

func compareRun(cmd *cobra.Command, args []string) {
	if compareConfig.source == "" || compareConfig.destination == "" {
		cli.FatalOutput(1, "必须指定-s和-d参数")
	}

	if compareConfig.copyto != "" {
		compareConfig.copyto = filepath.Clean(compareConfig.copyto) + "/"
	}

	var handler = func(add bool, src, path string) error {
		if compareConfig.verbose {
			if add {
				fmt.Printf("增加文件:%s\n", path)
			} else {
				fmt.Printf("文件修改:%s\n", path)
			}
		}

		if compareConfig.copyto != "" {
			err := command.Copy(src+path, compareConfig.copyto+path)
			if err != nil {
				return err
			}
		}
		return nil
	}
	err := cli.ComparePath(compareConfig.source, compareConfig.destination, handler)
	if err != nil {
		cli.FatalOutput(1, "对比失败:%s\n", err.Error())
	}
}
