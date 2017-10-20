package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		return
	}
	if os.Args[1] == "time" {
		fmt.Print(time.Now().Format("2006-01-02 15:04:05"))
	} else {
		os.RemoveAll(os.Args[1])
	}
}
