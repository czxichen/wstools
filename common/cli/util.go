package cli

import (
	"fmt"
	"os"
)

// FatalOutput 内容打印到错误输出,并退出
func FatalOutput(code int, format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(code)
}
