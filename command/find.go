package command

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	findCFG = &findinfo{}
	Find    = &Command{
		UsageLine: `find -d "./" -b "20160101" -l 10 -s ".go"`,
		Run:       find,
		Short:     "根据条件查找文件",
		Long: `所有条件都是and关系,比如-b "20160101" -l 10 -s ".go"表示必须同时满足这个三个
条件,才算匹配成功
`,
	}
)

func init() {
	Find.Flag.StringVar(&findCFG.dir, "d", "", `-d="dirpath" 指定查找的目录`)
	Find.Flag.StringVar(&findCFG.prefix, "p", "", `-p="prefix" 搜索条件,表示查找此字符串开头的文件`)
	Find.Flag.StringVar(&findCFG.suffix, "s", "", `-s="suffix" 搜索条件,表示查找此字符串结尾的文件`)
	Find.Flag.StringVar(&findCFG.after, "a", "", `-a="20160101" 搜索条件,表示文件修改日期在这个时间之后`)
	Find.Flag.StringVar(&findCFG.befer, "b", "", `-b="20161231" 搜索条件,表示文件修改日期在这个时间之前`)
	Find.Flag.Int64Var(&findCFG.ltsize, "l", 0, `-l=1024 搜索条件,表示文件大小小于1024k`)
	Find.Flag.Int64Var(&findCFG.gtsize, "g", 0, `-g=1024 搜索条件,表示文件大小大于1k`)
	Find.Flag.BoolVar(&findCFG.all, "A", true, "-A=false 搜索条件,表示是否搜索目录下的所有子目录")
	Find.Flag.StringVar(&findCFG.output, "o", "", `-o="log" 指定结果文件`)
}

func find(cmd *Command, args []string) bool {
	if findCFG.dir == "" {
		return false
	}
	var w = os.Stdout
	defer w.Close()
	if findCFG.output != "" {
		var err error
		w, err = os.Create(findCFG.output)
		if err != nil {
			log.Printf("创建输出文件失败:", err.Error())
			return true
		}
	}
	err := findCFG.Find(w)
	if err != nil {
		log.Println(err)
	}
	return true
}

type findinfo struct {
	output          string
	fix, date, size bool
	all             bool
	dir             string
	btime           int64
	befer           string
	atime           int64
	after           string
	gtsize          int64
	ltsize          int64
	suffix          string
	prefix          string
}

func (find *findinfo) init() error {
	dir := strings.Replace(find.dir, "\\", "/", -1)
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	if find.befer != "" {
		t, err := time.Parse("20060102", find.befer)
		if err != nil {
			return errors.New("date format invalid")
		}
		find.btime = t.Unix()
		find.date = true
	}

	if find.after != "" {
		t, err := time.Parse("20060102", find.after)
		if err != nil {
			return errors.New("date format invalid")
		}
		find.atime = t.Unix()
		find.date = true
	}

	if find.gtsize > 0 || find.ltsize > 0 {
		find.gtsize *= 1024
		find.ltsize *= 1024
		find.size = true
	}

	if find.prefix != "" || find.suffix != "" {
		find.fix = true
	}
	return nil
}

func (find *findinfo) Find(w io.Writer) error {
	if err := find.init(); err != nil {
		return err
	}
	if find.all {
		find.walkdir(w)
	} else {
		find.dirs(w)
	}
	return nil
}

func (find *findinfo) math(file os.FileInfo) bool {
	switch {
	case find.fix:
		if find.suffix != "" {
			if !strings.HasSuffix(file.Name(), find.suffix) {
				return false
			}
		}
		if find.prefix != "" {
			if !strings.HasPrefix(file.Name(), find.prefix) {
				return false
			}
		}
		fallthrough
	case find.date:
		t := file.ModTime().Unix()
		if find.atime > 0 {
			if t < find.atime {
				return false
			}
		}
		if find.btime > 0 {
			if t > find.btime {
				return false
			}
		}
		fallthrough
	case find.size:
		if find.gtsize > 0 {
			if find.gtsize > file.Size() {
				return false
			}
		}
		if find.ltsize > 0 {
			if find.ltsize < file.Size() {
				return false
			}
		}
	}
	return true
}

func (find *findinfo) dirs(w io.Writer) {
	list, err := ioutil.ReadDir(find.dir)
	if err != nil {
		log.Println(err)
		return
	}
	for _, info := range list {
		if info.IsDir() {
			continue
		}
		if find.math(info) {
			fmt.Fprintln(w, info.Name())
		}
	}
}

func (find *findinfo) walkdir(w io.Writer) {
	err := filepath.Walk(find.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if find.math(info) {
			fmt.Fprintln(w, path)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
