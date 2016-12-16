package command

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type compareConfig struct {
	dpath   string
	spath   string
	quick   bool
	diff    string
	output  string
	split   string
	md5list string
}

var (
	compareCFG compareConfig
	Compare    = &Command{
		UsageLine: `compare -s server -d server_back`,
		Run:       compare,
		Short:     "对文件或者目录经进行比较",
		Long: `对目录进行比较的时候,只会比较源目录存在的文件,文件比较分为两种,快速模式是对内容进行比较,还有一种是对文
件的md5进行比较,验证指定路径文件的md5的时候,md5列表的文件内容格式235dc44e85afc9697e3da6c660ef4fe2   passwd
	wstools compare -d="./" -F="md5list"
	wstools compare -s command -d command_new -c diff
`,
	}
)

func init() {
	Compare.Flag.StringVar(&compareCFG.spath, "s", "", `-s="server" 指定源目录或文件`)
	Compare.Flag.StringVar(&compareCFG.dpath, "d", "", `-d="server_back" 指定目标目录或文件`)
	Compare.Flag.StringVar(&compareCFG.diff, "c", "", `-c="diff" 把源目录不匹配的文件提出来,为空则不拷贝`)
	Compare.Flag.StringVar(&compareCFG.output, "o", "", `指定匹配的结果输出文件,不指定则输出到标准输出`)
	Compare.Flag.StringVar(&compareCFG.split, "S", "	", `-S=" " 验证指定路径文件的md5值的时候指定md5和路径的分隔符`)
	Compare.Flag.StringVar(&compareCFG.md5list, "F", "", `-F="md5list" 验证指定路径文件的md5值,每行一条数据,md5码和文件路径`)
	Compare.Flag.BoolVar(&compareCFG.quick, "q", true, `-q=false 快速模式即用文件内容判断文件是否一致,如果为false则使用md5比较`)
}

func compare(cmd *Command, args []string) bool {
	var w = os.Stdout
	if compareCFG.output != "" {
		var err error
		w, err = os.Create(compareCFG.output)
		if err != nil {
			log.Println(err)
			return true
		}
	}

	if compareCFG.spath != "" && compareCFG.dpath != "" {
		comparePath(compareCFG.spath, compareCFG.dpath, compareCFG.diff, w, compareCFG.quick)
		return true
	}

	if compareCFG.md5list != "" {
		compareFromFile(compareCFG.md5list, compareCFG.dpath, compareCFG.split, w)
		return true
	}
	return false
}

func compareFromFile(path string, dir, split string, w io.Writer) {
	dir = formatSeparator(dir)
	File, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return
	}
	defer File.Close()
	buf := bufio.NewReader(File)
	h := md5.New()
	var m, dfile, md string
	var count int = 0
	var dirsuccess bool = true
	for {
		count += 1
		line, _, err := buf.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		list := bytes.Split(line, []byte(split))
		if len(list) != 2 {
			log.Printf("第%d行数据无效,确认分隔符\n", count)
			continue
		}
		dfile = string(bytes.TrimSpace(list[1]))
		md, err = getmd5(h, dir+dfile)
		if err != nil {
			fmt.Fprintf(w, "文件不存在:\t%s\n", dfile)
			if dirsuccess {
				dirsuccess = false
			}
			continue
		}

		m = string(bytes.TrimSpace(list[0]))
		if md != m {
			fmt.Fprintf(w, "内容不一致:\t%s\n", dfile)
			if dirsuccess {
				dirsuccess = false
			}
			continue
		}
	}
	if dirsuccess {
		fmt.Fprintln(w, "目录文件一致")
	}
}

func comparePath(spath, dpath, diff string, w io.Writer, quick bool) {
	sinfo, err := os.Lstat(spath)
	if err != nil {
		log.Println(err)
		return
	}
	dinfo, err := os.Lstat(dpath)
	if err != nil {
		log.Println(err)
		return
	}

	if !sinfo.IsDir() && !dinfo.IsDir() {
		if !comparefile(spath, dpath, quick) {
			fmt.Fprintf(w, "内容不一致:\t%s\n", dinfo.Name())
		}
		return
	}

	if !(sinfo.IsDir() || dinfo.IsDir()) {
		log.Println("原路径和目标路径必须同时为文件或者同时为目录")
		return
	}

	if diff != "" {
		diff = formatSeparator(diff)
		os.RemoveAll(diff)
		os.MkdirAll(diff, 0666)
	}

	spath = formatSeparator(spath)
	dpath = formatSeparator(dpath)
	var dirsuccess bool = true
	filepath.Walk(spath, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		path := strings.TrimPrefix(root, spath)
		dfile := dpath + path
		dinfo, err := os.Lstat(dfile)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(w, "文件不存在:\t%s\n", path)
				if dirsuccess {
					dirsuccess = false
				}
				if diff != "" {
					err = copyFile(root, diff+path)
					if err != nil {
						log.Printf("拷贝文件错误:%s\n", err)
					} else {
						log.Printf("拷贝文件成功:%s\n", path)
					}
				}
			}
			return nil
		}

		if info.Size() == dinfo.Size() {
			if comparefile(root, dfile, quick) {
				return nil
			}
		}
		fmt.Fprintf(w, "内容不一致:\t%s\n", path)
		if dirsuccess {
			dirsuccess = false
		}
		if diff != "" {
			err = copyFile(root, diff+path)
			if err != nil {
				log.Printf("拷贝文件错误:%s\n", err)
			} else {
				log.Printf("拷贝文件成功:%s\n", path)
			}
		}
		return nil
	})

	if dirsuccess {
		fmt.Fprintln(w, "目录文件一致")
	}

	return
}
func comparefile(spath, dpath string, quick bool) bool {
	if !quick {
		h := md5.New()
		smd5, err := getmd5(h, spath)
		if err != nil {
			return false
		}
		dmd5, err := getmd5(h, dpath)
		if err != nil {
			return false
		}
		return smd5 == dmd5
	}

	sFile, err := os.Open(spath)
	if err != nil {
		return false
	}
	defer sFile.Close()
	dFile, err := os.Open(dpath)
	if err != nil {
		return false
	}
	defer dFile.Close()
	return comparebyte(sFile, dFile)
}

//下面可以代替md5比较.
func comparebyte(sfile io.Reader, dfile io.Reader) bool {
	var sbyte []byte = make([]byte, 512)
	var dbyte []byte = make([]byte, 512)
	var serr, derr error
	for {
		_, serr = sfile.Read(sbyte)
		_, derr = dfile.Read(dbyte)
		if serr != nil || derr != nil {
			if serr != derr {
				return false
			}
			if serr == io.EOF {
				break
			}
		}
		if bytes.Equal(sbyte, dbyte) {
			continue
		}
		return false
	}
	return true
}

func formatSeparator(path string) string {
	Separator := string(filepath.Separator)
	if Separator != "/" {
		path = strings.Replace(path, "/", Separator, -1)
	} else {
		path = strings.Replace(path, "\\", Separator, -1)
	}
	if !strings.HasSuffix(path, Separator) {
		path += Separator
	}
	return path
}

func copyFile(spath, dpath string) error {
	err := copyfile(spath, dpath)
	if err != nil {
		return err
	}
	info, err := os.Lstat(spath)
	if err != nil {
		return err
	}
	os.Chmod(dpath, info.Mode())
	os.Chtimes(dpath, info.ModTime(), info.ModTime())
	return nil
}

func copyfile(spath, dpath string) error {
	basedir := filepath.Dir(dpath)
	err := os.MkdirAll(basedir, 0666)
	if err != nil {
		return err
	}
	dFile, err := os.Create(dpath)
	if err != nil {
		return err
	}
	defer dFile.Close()
	sFile, err := os.Open(spath)
	if err != nil {
		return err
	}
	defer sFile.Close()
	_, err = io.Copy(dFile, sFile)
	return err
}
