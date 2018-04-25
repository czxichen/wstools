package command

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

var Replace = &cobra.Command{
	Use: "replace",
	Example: `	替换指定路径下,后缀为.go的文件内容
	-p sourcepath -o "oldstr" -n "newstr"  -s .go`,
	Short: "文本替换工具",
	Long:  "文本批量替换工具,支持正则匹配",
	RunE:  replace_run,
}

var _replace replace_config

func init() {
	Replace.PersistentFlags().StringVarP(&_replace.oldb, "old", "o", "", "被替换的内容,当是正则表达式的时候必须结合-r使用")
	Replace.PersistentFlags().StringVarP(&_replace.newb, "new", "n", "", "替换的内容")
	Replace.PersistentFlags().StringVarP(&_replace.path, "path", "p", "", "指定要操作的文件或目录")
	Replace.PersistentFlags().StringVarP(&_replace.suffix, "suffix", "s", "", "指定要匹配的文件后缀")
	Replace.PersistentFlags().StringVarP(&_replace.match, "match", "m", "", "使用正则匹配文件名称,如果和-s同时使用,则-s生效")
	Replace.PersistentFlags().BoolVarP(&_replace.usereg, "regexp", "r", false, "使用正比表达式匹配要替换的内容")
	Replace.PersistentFlags().BoolVarP(&_replace.quick, "quick", "q", true, "文件内容全部加载到内存中替换")
}

func replace_run(cmd *cobra.Command, args []string) error {
	if _replace.oldb == "" || _replace.newb == "" || _replace.path == "" {
		return fmt.Errorf("参数错误")
	}

	_replace.path = filepath.Clean(_replace.path)

	var err error

	if _replace.match != "" {
		_replace.namereg, err = regexp.Compile(_replace.match)
		if err != nil {
			fmt.Printf("解析正则失败:%s\n", err.Error())
			return nil
		}
	}

	if _replace.usereg {
		_replace.reg, err = regexp.Compile(_replace.oldb)
		if err != nil {
			fmt.Printf("解析正则失败:%s\n", err.Error())
			return nil
		}
	}
	err = walkFile(_replace.path, _replace, _replace.quick)
	if err != nil {
		fmt.Printf("文件替换出错:%s\n", err.Error())
	}
	return nil
}

type replace_config struct {
	path    string
	oldb    string
	newb    string
	suffix  string
	match   string
	usereg  bool
	quick   bool
	reg     *regexp.Regexp
	namereg *regexp.Regexp
}

func (r replace_config) Count(p []byte) int {
	if r.reg != nil {
		return len(r.reg.FindAll(p, -1))
	} else {
		return bytes.Count(p, []byte(r.oldb))
	}
}

func (r replace_config) Replace(p []byte) []byte {
	if r.reg != nil {
		return r.reg.ReplaceAll(p, []byte(r.newb))
	} else {
		return bytes.Replace(p, []byte(r.oldb), []byte(r.newb), -1)
	}
}

func (r replace_config) Match(info os.FileInfo) bool {
	if len(r.suffix) == 0 {
		if r.namereg != nil {
			return r.namereg.MatchString(info.Name())
		} else {
			return true
		}
	} else {
		return bytes.HasSuffix([]byte(info.Name()), []byte(r.suffix))
	}
}

func walkFile(dir string, _replace replace_config, quick bool) error {
	return filepath.Walk(dir, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if _replace.Match(info) {
			return replace(root, _replace, quick)
		}
		return nil
	})
}

func replace(path string, _replace replace_config, quick bool) error {
	if quick {
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if i := _replace.Count(buf); i > 0 {
			fmt.Printf("文件:%s 替换:%d处\n", path, i)
		} else {
			return nil
		}

		buf = _replace.Replace(buf)
		File, err := os.Create(path)
		if err != nil {
			return err
		}
		File.Write(buf)
		File.Close()
	} else {
		tmppath := filepath.Base(path) + ".tmp"
		tmp, err := os.Create(tmppath)
		if err != nil {
			return fmt.Errorf("创建临时文件失败:%s", err.Error())
		}

		File, err := os.Open(path)
		if err != nil {
			tmp.Close()
			os.Remove(tmppath)
			return fmt.Errorf("打开文件失败:", err.Error())
		}

		var line []byte
		buf := bufio.NewReader(File)
		var count int = 0
		for {
			line, _, err = buf.ReadLine()
			if err != nil {
				break
			}
			if i := _replace.Count(line); i > 0 {
				count += i
				line = _replace.Replace(line)
			}
			tmp.Write(line)
			tmp.Write([]byte("\r\n"))
		}

		tmp.Close()
		File.Close()
		if err != nil {
			if err != io.EOF {
				os.Remove(tmppath)
				return fmt.Errorf("读取:%s数据失败:%s\n", path, err)
			}
		}
		if count > 0 {
			os.Remove(path)
			os.Rename(tmppath, path)
			fmt.Printf("文件:%s 替换:%d处\n", path, count)
		} else {
			os.Remove(tmppath)
		}
	}
	return nil
}
