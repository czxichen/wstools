package main

import (
	"fmt"
	"os"

	"github.com/czxichen/configmanage/server"
)

func main() {
	if !server.Server(os.Args[1:]) {
		fmt.Fprintln(os.Stderr, "Usage:")
		server.FlagSet.PrintDefaults()
	}
}
