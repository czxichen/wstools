package command

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/czxichen/command"
	"github.com/spf13/cobra"
)

var Md5sum = &cobra.Command{
	Use: "md5sum",
	Example: `	
	计算sourcepath下的符合计算规则的文件md5
	-s sourcepath -o .*\.go`,
	Short: "计算文件的md5值",
	Long:  "用来计算文件或者个目录下所有文件的md5值",
	RunE:  md5sum_run,
}

type md5sum_config struct {
	source string
	only   string
	invert string
}

var _md5sum md5sum_config

func init() {
	Md5sum.PersistentFlags().StringVarP(&_md5sum.source, "source", "s", "", "指定要计算的路径")
	Md5sum.PersistentFlags().StringVarP(&_md5sum.only, "only", "o", "", "正则表达式,表示只计算匹配到的文件")
	Md5sum.PersistentFlags().StringVarP(&_md5sum.invert, "invert", "v", "", "正则表达式,表示不计算匹配到的文件")
}

func md5sum_run(cmd *cobra.Command, args []string) error {
	_, err := os.Lstat(_md5sum.source)
	if err != nil {
		return err
	}

	var (
		only   *regexp.Regexp
		invert *regexp.Regexp
	)

	if _md5sum.only != "" {
		only, err = regexp.Compile(_md5sum.only)
		if err != nil {
			return err
		}

	}

	if _md5sum.invert != "" {
		invert, err = regexp.Compile(_md5sum.invert)
		if err != nil {
			return err
		}
	}

	err = filepath.Walk(_md5sum.source, func(path string, info os.FileInfo, err error) error {
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
	if err != nil {
		fmt.Printf("计算MD5失败:%s\n", err.Error())
	}
	return nil
}
