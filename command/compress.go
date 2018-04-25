package command

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/czxichen/command/zip"
	"github.com/spf13/cobra"
)

var Compress = &cobra.Command{
	Use: "compress",
	Example: `	使用zip协议压缩目录uuid为uuid.zip
	-c -s uuid -d uuid.zip
	使用zip协议解压uuid.zip文件,到uuid目录
	-u -s uuid.zip -d /tmp`,
	Short: "文件的解压缩操作",
	Long:  "文件解压缩操作,支持zip|gz两种格式",
	RunE:  compress_run,
}

type compress struct {
	create      bool
	verbose     bool
	uncompress  bool
	ftype       string
	source      string
	destination string
}

var _compress compress

func init() {
	Compress.PersistentFlags().BoolVarP(&_compress.verbose, "verbose ", "b", false, "输出详细内容")
	Compress.PersistentFlags().BoolVarP(&_compress.create, "create", "c", false, "创建压缩文件")
	Compress.PersistentFlags().BoolVarP(&_compress.uncompress, "uncompress", "u", false, "解压压缩文件")
	Compress.PersistentFlags().StringVarP(&_compress.ftype, "type", "t", "zip", "指定文件压缩文件格式,只支持zip|gzip")
	Compress.PersistentFlags().StringVarP(&_compress.source, "source", "s", "", "指定要操作的源路径")
	Compress.PersistentFlags().StringVarP(&_compress.destination, "destination", "d", "", "指定要操作的目标路径")
}

func compress_run(cmd *cobra.Command, args []string) error {
	if _compress.ftype != "gzip" && _compress.ftype != "zip" {
		return fmt.Errorf("Unsupported types:%s", _compress.ftype)
	}

	_, err := os.Lstat(_compress.source)
	if err != nil {
		fmt.Printf("查看文件失败:%s\n", err.Error())
		return nil
	}

	if _compress.create {
		//如果目标路径为空,则默认在当前路径
		if _compress.destination == "" {
			_compress.destination = fmt.Sprintf("./%s.%s", filepath.Base(_compress.destination), _compress.ftype)
		}

		//查看目标文件目录是否存在,如果不存在则直接返回
		_, err := os.Lstat(filepath.Dir(_compress.destination))
		if err != nil {
			return err
		}

		File, err := os.Create(_compress.destination)
		if err != nil {
			return err
		}

		defer File.Close()
		var write zip.Compress

		switch _compress.ftype {
		case "zip":
			write = zip.NewZipWriter(File)
		case "gzip":
			write = zip.NewTgzWirter(File)
		}
		defer write.Close()
		return write.Walk(_compress.source)
	}

	if _compress.uncompress {
		if _compress.destination == "" {
			_compress.destination = "./"
		}
		_compress.destination = filepath.Clean(_compress.destination)
		var logger func(string, ...interface{})
		if _compress.verbose {
			logger = log.Printf
		}

		switch _compress.ftype {
		case "zip":
			return zip.Unzip(_compress.source, _compress.destination, logger)
		case "gzip":
			return zip.Ungzip(_compress.source, _compress.destination, logger)
		}
	}
	return fmt.Errorf("必须指定-c或者个-u参数")
}
