package command

import (
	"github.com/spf13/cobra"
)

// HelpFunc HelpFunc
func HelpFunc(c *cobra.Command) {
	c.SetHelpCommand(&cobra.Command{
		Use:   "help [command]",
		Short: "获取命令帮助信息",
		Long: `应用程序中的命令提供帮助信息.
例如 ` + c.Name() + ` help [path to command] 来获取详细的帮助内容.`,

		Run: func(c *cobra.Command, args []string) {
			cmd, _, e := c.Root().Find(args)
			if cmd == nil || e != nil {
				c.Printf("未知的 help topic %#q\n", args)
				c.Root().Usage()
			} else {
				cmd.InitDefaultHelpFlag() // make possible 'help' flag to be shown
				cmd.Help()
			}
		},
	})
}
