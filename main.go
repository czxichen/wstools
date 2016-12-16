package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/czxichen/wstools/command"
)

var commands = command.Commands

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	if args[0] == "help" {
		help(args[1:])
		for _, cmd := range commands {
			if len(args) >= 2 && cmd.Name() == args[1] && cmd.Run != nil {
				fmt.Fprintf(os.Stderr, "Default Parameters:\n")
				cmd.Flag.PrintDefaults()
			}
		}
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			if !cmd.Run(cmd, args) {
				fmt.Fprintf(os.Stderr, "\n")
				cmd.Flag.Usage()
				fmt.Fprintf(os.Stderr, "Default Parameters:\n")
				cmd.Flag.PrintDefaults()
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "wstools: unknown subcommand %q\nRun 'wstools help' for usage.\n", args[0])
}

var usageTemplate = `wstools Usage:
	wstools [arguments]
The commands are:{{range .}}{{if .Runnable}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}
	
Use "wstools help [command]" for more information about a command.
`
var helpTemplate = `{{if .Runnable}}Usage: wstools {{.UsageLine}}
{{end}}
  {{.Long}}
`

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace, "capitalize": capitalize})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, commands)
}

func usage() {
	printUsage(os.Stderr)
	//	fmt.Fprintf(os.Stderr, "For Logging, use \"wstools [logging_options] [command]\". The logging options are:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return
	}

	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: wstools help command\n\nToo many arguments given.\n")
		os.Exit(2)
	}

	arg := args[0]
	for _, cmd := range commands {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run 'wstools help'.\n", arg)
	os.Exit(2)
}
