package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/czxichen/command"
	"github.com/spf13/cobra"
)

var Compare = &cobra.Command{
	Use: "compare",
	Example: `	比较uuid和uuid_new目录,如果存在差异文件,则拷贝到diff目录中
	-s uuid -d uuid_new -c diff`,
	Short: "文件目录比较",
	Long:  "文件目录进行比较,支持取出差异包",
	RunE:  compare_run,
}

type compare struct {
	verbose     bool
	copyto      string
	source      string
	destination string
}

var _compare compare

func init() {
	Compare.PersistentFlags().StringVarP(&_compare.source, "source", "s", "", "指定原始路径,不能为空")
	Compare.PersistentFlags().StringVarP(&_compare.destination, "destination", "d", "", "指定目标路径,不能为空")
	Compare.PersistentFlags().StringVarP(&_compare.copyto, "copy", "c", "", "差异文件拷贝到此路径下")
	Compare.PersistentFlags().BoolVarP(&_compare.verbose, "verbose", "v", true, "输出详情")
}

func compare_run(cmd *cobra.Command, args []string) error {
	if _compare.source == "" || _compare.destination == "" {
		return fmt.Errorf("必须指定-s和-d参数")
	}

	if _compare.copyto != "" {
		_compare.copyto = filepath.Clean(_compare.copyto) + "/"
	}

	var handler = func(add bool, src, path string) error {
		if _compare.verbose {
			if add {
				fmt.Printf("增加文件:%s\n", path)
			} else {
				fmt.Printf("文件修改:%s\n", path)
			}
		}

		if _compare.copyto != "" {
			err := command.Copy(src+path, _compare.copyto+path)
			if err != nil {
				return err
			}
		}
		return nil
	}
	err := compare_path(_compare.source, _compare.destination, handler)
	if err != nil {
		fmt.Printf("对比失败:%s\n", err.Error())
	}
	return nil
}

func compare_path(spath, dpath string, handler func(add bool, src, path string) error) error {
	spath = filepath.Clean(spath)
	dpath = filepath.Clean(dpath)

	spath += string(filepath.Separator)
	dpath += string(filepath.Separator)

	return filepath.Walk(spath, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base := strings.TrimPrefix(root, spath)
		match, err := comparefile(root, dpath+base)
		if err != nil {
			if os.IsNotExist(err) && handler != nil {
				return handler(true, spath, base)
			}
			return err
		}
		if !match && handler != nil {
			return handler(match, spath, base)
		}
		return nil
	})
}

func comparefile(spath, dpath string) (bool, error) {
	sFile, err := os.Open(spath)
	if err != nil {
		return false, err
	}
	defer sFile.Close()

	dFile, err := os.Open(dpath)
	if err != nil {
		return false, err
	}
	defer dFile.Close()

	sInfo, err := sFile.Stat()
	if err != nil {
		return false, err
	}

	dInfo, err := dFile.Stat()
	if err != nil {
		return false, err
	}

	if dInfo.IsDir() && sInfo.IsDir() {
		return true, nil
	}

	if sInfo.Size() != dInfo.Size() {
		return false, nil
	}

	return comparebyte(sFile, dFile), nil
}

//使用字节码比较数据流
func comparebyte(sfile io.Reader, dfile io.Reader) bool {
	var (
		sint, dint int
		serr, derr error
		sbyte      []byte = make([]byte, 512)
		dbyte      []byte = make([]byte, 512)
	)
	for {
		sint, serr = sfile.Read(sbyte)
		dint, derr = dfile.Read(dbyte)
		if serr != nil || derr != nil {
			if serr == io.EOF && derr == io.EOF {
				return true
			}
			return false
		}
		if sint == dint && bytes.Equal(sbyte, dbyte) {
			continue
		}
		return false
	}
}
