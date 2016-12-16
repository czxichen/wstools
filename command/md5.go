package command

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type md5cfg struct {
	path    string
	cpath   string
	opath   string
	exclude string
	suffix  string
}

var (
	md5CFG = md5cfg{}
	Md5    = &Command{
		UsageLine: `md5 -d ../tools -s=".go"`,
		Run:       md5sum,
		Short:     "计算指定路径的md5值,可以是目录",
		Long: `主要是用来计算文件的md5值,可以用-s指定文件文件类型,或者-e 排除哪些文件类型
	wstools md5 -d ./ -e .exe
`,
	}
)

func init() {
	Md5.Flag.StringVar(&md5CFG.path, "d", "./", `-d="./" 指定要计算的路径,当指定-c的时候此参数无效`)
	Md5.Flag.StringVar(&md5CFG.cpath, "c", "", `-c="filelist.txt" 从文件读取路径,按行分割,可以为目录路径`)
	Md5.Flag.StringVar(&md5CFG.opath, "o", "", `-o="md5.txt" 将计算的结果输出到文件`)
	Md5.Flag.StringVar(&md5CFG.exclude, "e", "", `-e=".exe" 指不对以此结尾的文件进行md5计算`)
	Md5.Flag.StringVar(&md5CFG.suffix, "s", "", `-s=".txt" 指仅对以此结尾的文件进行md5计算`)
}

func md5sum(cmd *Command, args []string) bool {
	var err error
	var output = os.Stdout
	defer output.Close()

	if md5CFG.opath != "" {
		output, err = os.Create(md5CFG.opath)
		if err != nil {
			log.Println(err)
			return true
		}
	}
	if md5CFG.cpath != "" {
		File, err := os.Open(md5CFG.cpath)
		if err != nil {
			log.Println(err)
			return true
		}
		defer File.Close()

		buf := bufio.NewReader(File)
		for {
			line, _, err := buf.ReadLine()
			if err != nil {
				if err != io.EOF {
					log.Println(err)
				}
				return true
			}
			err = md5walk(string(line), md5CFG, output)
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		err = md5walk(md5CFG.path, md5CFG, output)
		if err != nil {
			log.Println(err)
		}
	}
	return true
}

func md5walk(path string, cfg md5cfg, w io.Writer) error {
	path = filepath.Clean(path)
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	h := md5.New()
	if !info.IsDir() {
		m, err := getmd5(h, path)
		if err == nil {
			fmt.Fprintf(w, "%s\t%s", m, path)
		}
		return err
	}
	if !strings.HasSuffix(path, string(filepath.Separator)) {
		path += string(filepath.Separator)
	}
	return filepath.Walk(path, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		dir := strings.TrimPrefix(root, path)
		m, err := getmd5(h, root)
		if err != nil {
			log.Println("计算md5失败:", err.Error())
			return nil
		}
		if filepath.Separator == '\\' {
			dir = strings.Replace(dir, "\\", "/", -1)
		}

		if cfg.suffix != "" || cfg.exclude != "" {
			if !strings.HasSuffix(info.Name(), cfg.suffix) {
				return nil
			}
			if cfg.exclude != "" {
				if strings.HasSuffix(info.Name(), cfg.exclude) {
					return nil
				}
			}
		}
		fmt.Fprintf(w, "%s\t%s\n", m, dir)
		return nil
	})
}

func getmd5(md5hash hash.Hash, path string) (string, error) {
	File, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer File.Close()
	md5hash.Reset()
	_, err = io.Copy(md5hash, File)
	if err != nil {
		return "", err
	}
	result := make([]byte, 0, 32)
	result = md5hash.Sum(result)
	return hex.EncodeToString(result), nil
}
