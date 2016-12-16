package command

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	compressCFG compresserConfig
	Compress    = &Command{
		UsageLine: `compress -c -p /tmp -o tmp.zip`,
		Run:       compresser,
		Short:     "压缩解压文件",
		Long: `可以解压gzip和zip两种格式文件,压缩使用的是zip编码,-c和-x不能同时使用
	wstools compress -x -p tmp.zip -o ./
	wstools compress -c -p command -o tmp.zip
`,
	}
)

type compresserConfig struct {
	Create    bool
	Decomp    bool
	Printinfo bool
	Path      string
	OutName   string
}

func init() {
	Compress.Flag.BoolVar(&compressCFG.Create, "c", false, "-c=true 压缩文件")
	Compress.Flag.BoolVar(&compressCFG.Decomp, "x", false, "-x=true 解压文件")
	Compress.Flag.BoolVar(&compressCFG.Printinfo, "v", false, "-v=true 关闭详细输出")
	Compress.Flag.StringVar(&compressCFG.Path, "p", "", `-p="./test" 指定输入路径,不能为空`)
	Compress.Flag.StringVar(&compressCFG.OutName, "o", "", `-o="/tmp/test.zip" 指定输出路径,不能为空`)

}

func compresser(cmd *Command, args []string) bool {
	if !compressCFG.Create && !compressCFG.Decomp {
		return false
	}
	if compressCFG.Path == "" || compressCFG.OutName == "" {
		return false
	}
	if compressCFG.Create {
		File, err := os.Create(compressCFG.OutName)
		if err != nil {
			log.Println(err)
			return true
		}
		defer File.Close()
		cwrite := newzipWriter(File)
		err = cwrite.Walk(compressCFG.Path)
		cwrite.Close()
		if err != nil {
			log.Println(err)
		}
	} else {
		info, err := os.Lstat(compressCFG.OutName)
		if err != nil {
			log.Println(err)
			return true
		}
		if !info.IsDir() {
			log.Println("The destination path must be a directory")
			return true
		}
		if checkValidZip(compressCFG.Path) {
			err = unzip(compressCFG.Path, compressCFG.OutName)
		} else {
			err = ungzip(compressCFG.Path, compressCFG.OutName)
		}
		if err != nil {
			log.Println(err)
		}
	}
	return true
}

type compress interface {
	Close() error
	WriteHead(path string, info os.FileInfo) error
	Write(p []byte) (int, error)
}

func zipwalk(path string, compresser compress) error {
	_, err := os.Lstat(path)
	if err != nil {
		return err
	}

	var baseDir string
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.TrimSuffix(path, "/")
	if path != "./" && path != "" {
		path = strings.TrimPrefix(path, "./")
	}

	baseDir = filepath.Base(path)
	if baseDir == "./" || baseDir == "." {
		baseDir = ""
	}

	filepath.Walk(path, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		root = filepath.ToSlash(root)
		fileroot := root
		if root = strings.TrimPrefix(root, path); root == "" {
			root = baseDir
		} else {
			root = baseDir + root
		}

		if compressCFG.Printinfo {
			log.Println("compress : ", root)
		}

		err = compresser.WriteHead(root, info)
		if err != nil {
			return nil
		}
		F, err := os.Open(fileroot)
		if err != nil {
			return nil
		}
		io.Copy(compresser, F)
		F.Close()
		return nil
	})
	return nil
}

func newzipWriter(File io.Writer) *zipWrite {
	zipwrite := zip.NewWriter(File)
	return &zipWrite{zone: 8, zw: zipwrite, file: File}
}

type zipWrite struct {
	zone   int64
	zw     *zip.Writer
	writer io.Writer
	file   io.Writer
}

func (self *zipWrite) Close() error {
	return self.zw.Close()
}

func (self *zipWrite) WriteHead(path string, info os.FileInfo) error {
	if path == "." || path == ".." {
		return nil
	}
	head, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		head.Method = zip.Deflate
	}
	head.Name = path
	if info.IsDir() {
		head.Name += "/"
	}
	head.SetModTime(time.Unix(info.ModTime().Unix()+self.zone*60*60, 0))
	write, err := self.zw.CreateHeader(head)
	if err != nil {
		return err
	}
	self.writer = write
	return nil
}

func (self *zipWrite) Write(p []byte) (int, error) {
	return self.writer.Write(p)
}

func (self *zipWrite) Walk(source string) error {
	return zipwalk(source, self)
}

func unzip(filename, dir string) error {
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	File, err := zip.OpenReader(filename)
	if err != nil {
		return errors.New("Error Open zip faild: " + err.Error())
	}

	defer File.Close()
	for _, v := range File.File {
		err := createFile(v, dir)
		if err != nil {
			if compressCFG.Printinfo {
				log.Printf("unzip file err %v \n", err)
			}
			return err
		}
		os.Chtimes(v.Name, v.ModTime().Add(-8*time.Hour), v.ModTime().Add(-8*time.Hour))
		os.Chmod(v.Name, v.Mode())
		if compressCFG.Printinfo {
			log.Printf("unzip %s\n", v.Name)
		}
	}
	return nil
}

func createFile(v *zip.File, dscDir string) error {
	v.Name = dscDir + v.Name
	info := v.FileInfo()
	if info.IsDir() {
		err := os.MkdirAll(v.Name, v.Mode())
		if err != nil {
			return errors.New("Error Create direcotry" + v.Name + "faild: " + err.Error())
		}
		return nil
	}
	srcFile, err := v.Open()
	if err != nil {
		return errors.New("Error Read from zip faild: " + err.Error())
	}
	defer srcFile.Close()
	newFile, err := os.Create(v.Name)
	if err != nil {
		return errors.New("Error Create file faild: " + err.Error())
	}

	defer newFile.Close()
	io.Copy(newFile, srcFile)
	return nil
}

func ungzip(filepath, desdir string) error {
	File, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer File.Close()
	desdir = strings.Replace(desdir, "\\", "/", -1)
	if !strings.HasSuffix(desdir, "/") {
		desdir = desdir + "/"
	}
	gw, err := gzip.NewReader(File)
	if err != nil {
		return err
	}
	defer gw.Close()
	tw := tar.NewReader(gw)
	for {
		head, err := tw.Next()
		if err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}
		if head.FileInfo().IsDir() {
			err := os.MkdirAll(desdir+head.Name, os.FileMode(head.Mode))
			if err != nil {
				return err
			}
			if compressCFG.Printinfo {
				log.Printf("create directory: %s\n", desdir+head.Name)
			}
			continue
		}
		F, err := os.OpenFile(desdir+head.Name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(head.Mode))
		if err != nil {
			continue
		}
		io.Copy(F, tw)
		F.Close()
		os.Chtimes(desdir+head.Name, head.AccessTime, head.ModTime)
		if compressCFG.Printinfo {
			log.Printf("create file: %s\n", desdir+head.Name)
		}
	}
	return nil
}

func checkValidZip(path string) bool {
	z, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	z.Close()
	return true
}
