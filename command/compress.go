package command

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/czxichen/command/zip"
	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

// Compress 压缩命令
var Compress = &cobra.Command{
	Use: "compressConfig",
	Example: `	使用zip协议压缩目录uuid为uuid.zip
	-c -s uuid -d uuid.zip
	使用zip协议解压uuid.zip文件,到uuid目录
	-u -s uuid.zip -d /tmp`,
	Short: "文件的解压缩操作",
	Long:  "文件解压缩操作,支持zip|gz两种格式",
	Run:   compressRun,
}

type compressConfig struct {
	create      bool
	verbose     bool
	uncompress  bool
	ftype       string
	source      string
	destination string
}

var compress compressConfig

func init() {
	Compress.PersistentFlags().BoolVarP(&compress.verbose, "verbose ", "b", false, "输出详细内容")
	Compress.PersistentFlags().BoolVarP(&compress.create, "create", "c", false, "创建压缩文件")
	Compress.PersistentFlags().BoolVarP(&compress.uncompress, "uncompress", "u", false, "解压压缩文件")
	Compress.PersistentFlags().StringVarP(&compress.ftype, "type", "t", "zip", "指定文件压缩文件格式,只支持zip|gzip")
	Compress.PersistentFlags().StringVarP(&compress.source, "source", "s", "", "指定要操作的源路径")
	Compress.PersistentFlags().StringVarP(&compress.destination, "destination", "d", "", "指定要操作的目标路径")
}

func compressRun(cmd *cobra.Command, args []string) {
	if compress.ftype != "gzip" && compress.ftype != "zip" {
		cli.FatalOutput(1, "不支持的格式:%s", compress.ftype)
	}

	_, err := os.Lstat(compress.source)
	if err != nil {
		cli.FatalOutput(1, "查看文件失败:%s\n", err.Error())
	}

	if compress.create {
		//如果目标路径为空,则默认在当前路径
		if compress.destination == "" {
			compress.destination = fmt.Sprintf("./%s.%s", filepath.Base(compress.destination), compress.ftype)
		}

		//查看目标文件目录是否存在,如果不存在则直接返回
		_, err := os.Lstat(filepath.Dir(compress.destination))
		if err != nil {
			cli.FatalOutput(1, "查看文件状态失败:%s\n", err.Error())
		}

		File, err := os.Create(compress.destination)
		if err != nil {
			cli.FatalOutput(1, "创建文件状态失败:%s\n", err.Error())
		}

		defer File.Close()
		var write zip.Compress

		switch compress.ftype {
		case "zip":
			write = zip.NewZipWriter(File)
		case "gzip":
			write = zip.NewTgzWirter(File)
		}
		defer write.Close()
		write.Walk(compress.source)
	} else if compress.uncompress {
		if compress.destination == "" {
			compress.destination = "./"
		}
		compress.destination = filepath.Clean(compress.destination)
		var logger func(string, ...interface{})
		if compress.verbose {
			logger = log.Printf
		}

		switch compress.ftype {
		case "zip":
			err = zip.Unzip(compress.source, compress.destination, logger)
		case "gzip":
			err = zip.Ungzip(compress.source, compress.destination, logger)
		}
		if err != nil {
			cli.FatalOutput(1, "解压文件失败:%s\n", err.Error())
		}
	}
}
