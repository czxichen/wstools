package command

import (
	"github.com/spf13/cobra"
)

type size_config struct {
	Path string
}

var (
	Size = &cobra.Command{}
)
