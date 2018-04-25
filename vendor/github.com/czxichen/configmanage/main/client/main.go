package main

import (
	"fmt"
	"os"

	"github.com/czxichen/configmanage/client"
)

func main() {
	if !client.Client(os.Args[1:]) {
		fmt.Fprintln(os.Stderr, "Usage:")
		client.FlagSet.PrintDefaults()
	}
}
