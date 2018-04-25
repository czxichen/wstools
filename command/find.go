package command

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var Find = &cobra.Command{
	Use: "find",
	Example: `	查找./下小于1MB的文件
	-p ./ -s "-1M"
	查找./目录下,修改时间在一天内且大于1MB的文件
	-p ./ -s "1M" -m "-1d"`,
	Short: "文件查找",
	Long:  "按名称,时间,大小查找文件",
	RunE:  find_run,
}

type find_config struct {
	dir   bool
	path  string
	name  string
	size  string
	mtime string
}

var _find find_config

func init() {
	Find.PersistentFlags().BoolVarP(&_find.dir, "dir", "d", false, "是否启用查找目录")
	Find.PersistentFlags().StringVarP(&_find.path, "path", "p", "", "查找的路径")
	Find.PersistentFlags().StringVarP(&_find.name, "name", "n", "", "按照名称正则查找文件")
	Find.PersistentFlags().StringVarP(&_find.size, "size", "s", "", "按照大小查找文件,单位(K,M,G不分大小写),不带单位默认字节")
	Find.PersistentFlags().StringVarP(&_find.mtime, "mtime", "m", "", "按照修改时间查找文件,单位(M,H,d,m,y区分大小写),不带单位默认为秒")
}

func find_run(cmd *cobra.Command, args []string) error {
	if _find.path == "" {
		return fmt.Errorf("必须指定查找的路径")
	}

	var (
		err   error
		name  *regexp.Regexp
		size  *size_match
		mtime *time_match
	)

	if _find.name != "" {
		name, err = regexp.Compile(_find.name)
		if err != nil {
			return err
		}
	}

	if _find.size != "" {
		size, err = parse_size(_find.size)
		if err != nil {
			return err
		}
	}

	if _find.mtime != "" {
		mtime, err = parse_mtime(_find.mtime)
		if err != nil {
			return err
		}
	}

	err = filepath.Walk(_find.path, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && !_find.dir {
			return nil
		}

		if name != nil {
			if !name.MatchString(info.Name()) {
				return nil
			}
		}

		if size != nil {
			if !size.Match(info.Size()) {
				return nil
			}
		}
		if mtime != nil {
			if !mtime.Match(info.ModTime()) {
				return nil
			}
		}
		fmt.Printf("%s\n", root)
		return nil
	})
	if err != nil {
		fmt.Printf("文件查找失败:%s\n", err.Error())
	}
	return nil
}

type time_match struct {
	less bool
	time time.Time
}

func (t *time_match) Match(mtime time.Time) bool {
	if t.less {
		return t.time.Before(mtime)
	} else {
		return t.time.After(mtime)
	}
}

func parse_mtime(mtime string) (*time_match, error) {
	var (
		count time.Duration
		unit  = mtime[len(mtime)-1]
	)

	switch unit {
	default:
		count = 1
		mtime += "s"
	case 'M':
		count = time.Minute
	case 'H':
		count = time.Hour
	case 'd':
		count = time.Hour * 24
	case 'm':
		count = time.Hour * 24 * 30
	case 'y':
		count = time.Hour * 24 * 30 * 365
	}
	var t = &time_match{less: false, time: time.Now()}
	mtime = mtime[:len(mtime)-1]
	nctime, err := strconv.Atoi(mtime)
	if err != nil {
		return t, err
	}
	t.less = nctime < 0
	t.time = time.Now().Add(^(time.Duration(math.Abs(float64(nctime))) * count))
	return t, nil
}

type size_match struct {
	less bool
	size int64
}

func (s *size_match) Match(size int64) bool {
	if s.less {
		return size <= s.size
	} else {
		return size >= s.size
	}
}

func parse_size(size string) (*size_match, error) {
	var (
		count int64
		unit  = size[len(size)-1]
	)

	switch unit {
	default:
		count = 1
		size += "b"
	case 'K', 'k':
		count = 1024
	case 'M', 'm':
		count = 1024 * 1024
	case 'G', 'g':
		count = 1024 * 1024 * 1024
	}

	var match size_match
	size = size[:len(size)-1]
	nsize, err := strconv.Atoi(size)
	if err != nil {
		return &match, err
	}

	if nsize < 0 {
		match.less = true
		match.size = int64(math.Abs(float64(nsize))) * count
	} else {
		match.less = false
		match.size = int64(nsize) * count
	}
	return &match, nil
}
