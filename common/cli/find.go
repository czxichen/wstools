package cli

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// Find 文件查找
func Find(cfg *FindConfig, handler func(path string) error) error {
	if cfg.Path == "" {
		return fmt.Errorf("必须指定查找的路径")
	}

	var (
		err   error
		name  *regexp.Regexp
		size  *matchSize
		mtime *matchTime
	)

	if cfg.Name != "" {
		name, err = regexp.Compile(cfg.Name)
		if err != nil {
			return err
		}
	}

	if cfg.Size != "" {
		size, err = parseSize(cfg.Size)
		if err != nil {
			return err
		}
	}

	if cfg.Mtime != "" {
		mtime, err = parseMtime(cfg.Mtime)
		if err != nil {
			return err
		}
	}

	return filepath.Walk(cfg.Path, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && !cfg.Dir {
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
		return handler(root)
	})
}

// FindConfig 文件查找
type FindConfig struct {
	Dir   bool
	Path  string
	Name  string
	Size  string
	Mtime string
}

// matchTime 时间匹配器
type matchTime struct {
	Less bool
	Time time.Time
}

// Match Match
func (t *matchTime) Match(mtime time.Time) bool {
	if t.Less {
		return t.Time.Before(mtime)
	}
	return t.Time.After(mtime)
}

// parseMtime parseMtime
func parseMtime(mtime string) (*matchTime, error) {
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
	var t = &matchTime{Less: false, Time: time.Now()}
	mtime = mtime[:len(mtime)-1]
	nctime, err := strconv.Atoi(mtime)
	if err != nil {
		return t, err
	}
	t.Less = nctime < 0
	t.Time = time.Now().Add(^(time.Duration(math.Abs(float64(nctime))) * count))
	return t, nil
}

// matchSize matchSize
type matchSize struct {
	Less bool
	Size int64
}

// Match Match
func (s *matchSize) Match(size int64) bool {
	if s.Less {
		return size <= s.Size
	}
	return size >= s.Size
}

// parseSize 解析大小
func parseSize(size string) (*matchSize, error) {
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

	var match matchSize
	size = size[:len(size)-1]
	nsize, err := strconv.Atoi(size)
	if err != nil {
		return &match, err
	}

	if nsize < 0 {
		match.Less = true
		match.Size = int64(math.Abs(float64(nsize))) * count
	} else {
		match.Less = false
		match.Size = int64(nsize) * count
	}
	return &match, nil
}
