package command

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/czxichen/wstools/common/cli"

	"github.com/czxichen/command"
	"github.com/spf13/cobra"
)

// Md5sum 计算md5
var Md5sum = &cobra.Command{
	Use: "md5sum",
	Example: `	
	计算sourcepath下的符合计算规则的文件md5
	-s sourcepath -o .*\.go`,
	Short: "计算文件的md5值",
	Long:  "用来计算文件或者个目录下所有文件的md5值",
	Run:   md5sumRun,
}

// Md5sumConfig 计算md5
type Md5sumConfig struct {
	source string
	only   string
	invert string
}

var md5sum Md5sumConfig

func init() {
	Md5sum.PersistentFlags().StringVarP(&md5sum.source, "source", "s", "", "指定要计算的路径")
	Md5sum.PersistentFlags().StringVarP(&md5sum.only, "only", "o", "", "正则表达式,表示只计算匹配到的文件")
	Md5sum.PersistentFlags().StringVarP(&md5sum.invert, "invert", "v", "", "正则表达式,表示不计算匹配到的文件")
}

func md5sumRun(cmd *cobra.Command, args []string) {
	_, err := os.Lstat(md5sum.source)
	if err != nil {
		cli.FatalOutput(1, "查看路径状态错误:%s\n", err.Error())
	}

	var (
		only   *regexp.Regexp
		invert *regexp.Regexp
	)

	if md5sum.only != "" {
		if only, err = regexp.Compile(md5sum.only); err != nil {
			cli.FatalOutput(1, "%s\n", err.Error())
		}
	}

	if md5sum.invert != "" {
		if invert, err = regexp.Compile(md5sum.invert); err != nil {
			cli.FatalOutput(1, "%s\n", err.Error())
		}
	}

	err = filepath.Walk(md5sum.source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if only != nil {
			if !only.MatchString(info.Name()) {
				return nil
			}
		}

		if invert != nil {
			if invert.MatchString(info.Name()) {
				return nil
			}
		}

		File, err := os.Open(path)
		if err != nil {
			return err
		}

		md5 := command.ReadMd5(File)
		File.Close()
		fmt.Printf("%s\t%s\n", md5, path)

		return nil
	})
	cli.FatalOutput(1, "计算MD5失败:%s\n", err.Error())
}
