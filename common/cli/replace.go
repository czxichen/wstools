package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

// ReplaceRun 指定替换
func ReplaceRun(replace *ReplaceConfig) error {
	if replace.OldB == "" || replace.NewB == "" || replace.Path == "" {
		return fmt.Errorf("参数错误")
	}

	replace.Path = filepath.Clean(replace.Path)

	var err error
	if replace.MatchStr != "" {
		replace.namereg, err = regexp.Compile(replace.MatchStr)
		if err != nil {
			return err
		}
	}

	if replace.UseReg {
		replace.reg, err = regexp.Compile(replace.OldB)
		if err != nil {
			fmt.Printf("解析正则失败:%s\n", err.Error())
			return nil
		}
	}
	return walkFile(replace.Path, replace, replace.Quick)
}

// ReplaceConfig 配置
type ReplaceConfig struct {
	Path     string
	OldB     string
	NewB     string
	Suffix   string
	MatchStr string
	UseReg   bool
	Quick    bool
	reg      *regexp.Regexp
	namereg  *regexp.Regexp
}

// Count 统计
func (r ReplaceConfig) Count(p []byte) int {
	if r.reg != nil {
		return len(r.reg.FindAll(p, -1))
	}
	return bytes.Count(p, []byte(r.OldB))
}

// Replace 替换
func (r ReplaceConfig) Replace(p []byte) []byte {
	if r.reg != nil {
		return r.reg.ReplaceAll(p, []byte(r.NewB))
	}
	return bytes.Replace(p, []byte(r.OldB), []byte(r.NewB), -1)
}

// Match 匹配
func (r ReplaceConfig) Match(info os.FileInfo) bool {
	if len(r.Suffix) == 0 {
		if r.namereg != nil {
			return r.namereg.MatchString(info.Name())
		}
		return true
	}
	return bytes.HasSuffix([]byte(info.Name()), []byte(r.Suffix))
}

func walkFile(dir string, replace *ReplaceConfig, quick bool) error {
	return filepath.Walk(dir, func(root string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if replace.Match(info) {
			return replaceRun(root, replace, quick)
		}
		return nil
	})
}

func replaceRun(path string, replace *ReplaceConfig, quick bool) error {
	if quick {
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if i := replace.Count(buf); i > 0 {
			fmt.Printf("文件:%s 替换:%d处\n", path, i)
		} else {
			return nil
		}

		buf = replace.Replace(buf)
		if File, err := os.Create(path); err == nil {
			File.Write(buf)
			File.Close()
		} else {
			return err
		}
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
			return err
		}

		var line []byte
		var count int

		buf := bufio.NewReader(File)
		for {
			line, _, err = buf.ReadLine()
			if err != nil {
				break
			}
			if i := replace.Count(line); i > 0 {
				count += i
				line = replace.Replace(line)
			}
			tmp.Write(line)
			tmp.Write([]byte("\r\n"))
		}

		tmp.Close()
		File.Close()
		if err != nil && err != io.EOF {
			os.Remove(tmppath)
			return fmt.Errorf("读取:%s数据失败:%s", path, err)
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
