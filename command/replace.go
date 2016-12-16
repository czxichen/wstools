package command

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

type replacecfg struct {
	path     string
	newstr   string
	oldstr   string
	suffix   string
	quick    bool
	isrepexp bool
}

var (
	replaceCFG = replacecfg{}
	Replace    = &Command{
		UsageLine: `replace -l main.go -r /mnt/main.go`,
		Run:       replace,
		Short:     "替换文本内容",
		Long: `
	wstools replace -o "Hello world" -n "World Hello" -d ./ -s ".json" -q=true
	wstools replace -o "[0-9]+" -n "digit" -d ./ -s ".json" -e=true
`,
	}
)

func init() {
	Replace.Flag.StringVar(&replaceCFG.newstr, "n", "", `-n="newstr" 指定新的字符串`)
	Replace.Flag.StringVar(&replaceCFG.oldstr, "o", "", `-o="oldstr" 指定要替换的字符串`)
	Replace.Flag.StringVar(&replaceCFG.path, "d", "", `-d="dirpath" 指定替换的目录或者文件`)
	Replace.Flag.StringVar(&replaceCFG.suffix, "s", "", `-s=".xml" 只替换以此后缀结尾的文件`)
	Replace.Flag.BoolVar(&replaceCFG.quick, "q", false, `-q=true 把文件读入内存替换,超大文件不建议使用`)
	Replace.Flag.BoolVar(&replaceCFG.isrepexp, "e", false, `-e=true 使用正则表达式替换文本`)
}

func replace(cmd *Command, args []string) bool {
	if replaceCFG.oldstr == "" || replaceCFG.newstr == "" || replaceCFG.path == "" {
		return false
	}
	var r repl
	if replaceCFG.isrepexp {
		var err error
		r.reg, err = regexp.Compile(replaceCFG.oldstr)
		if err != nil {
			log.Println(err)
			return true
		}
	} else {
		r.oldb = []byte(replaceCFG.oldstr)
	}
	r.suffix = []byte(replaceCFG.suffix)
	r.newb = []byte(replaceCFG.newstr)
	walkFile(replaceCFG.path, r, replaceCFG.quick)
	return true
}

type repl struct {
	oldb   []byte
	newb   []byte
	suffix []byte
	reg    *regexp.Regexp
}

func (r repl) Count(p []byte) int {
	if r.reg != nil {
		return len(r.reg.FindAll(p, -1))
	} else {
		return bytes.Count(p, r.oldb)
	}
}

func (r repl) Replace(p []byte) []byte {
	if r.reg != nil {
		return r.reg.ReplaceAll(p, r.newb)
	} else {
		return bytes.Replace(p, r.oldb, r.newb, -1)
	}
}

func (r repl) Match(info os.FileInfo) bool {
	if len(r.suffix) == 0 {
		return true
	} else {
		return bytes.HasSuffix([]byte(info.Name()), r.suffix)
	}
}

func walkFile(dir string, replace repl, quick bool) {
	filepath.Walk(dir, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if replace.Match(info) {
			_replace(root, replace, quick)
		}
		return nil
	})
}

func _replace(path string, replace repl, quick bool) {
	if quick {
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return
		}

		if i := replace.Count(buf); i > 0 {
			log.Printf("文件:%s 替换:%d处\n", path, i)
		} else {
			return
		}
		buf = replace.Replace(buf)
		File, err := os.Create(path)
		if err != nil {
			log.Println(err)
			return
		}
		File.Write(buf)
		File.Close()
	} else {
		tmppath := filepath.Base(path) + ".tmp"
		tmp, err := os.Create(tmppath)
		if err != nil {
			log.Printf("创建临时文件失败,%s\n", err.Error())
			return
		}
		File, err := os.Open(path)
		if err != nil {
			tmp.Close()
			os.Remove(tmppath)
			log.Printf("打开文件失败:", err.Error())
			return
		}
		var line []byte
		buf := bufio.NewReader(File)
		var count int = 0
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
		if err != nil {
			if err != io.EOF {
				os.Remove(tmppath)
				log.Printf("读取:%s数据失败:%s\n", path, err)
				return
			}
		}
		if count > 0 {
			os.Remove(path)
			os.Rename(tmppath, path)
			log.Printf("文件:%s 替换:%d处\n", path, count)
		} else {
			os.Remove(tmppath)
		}
	}
}
