// Package cli is the driving side of the tool: it turns a command line into a
// call on one of the packages beneath it and turns the result into output.
//
// What lives here is the argument parsing and the wording of what a person
// sees. Rule generation, config text and state transitions belong to the
// packages beneath.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/themaiby/stet/internal/fetch"
	"github.com/themaiby/stet/internal/paths"
)

const usage = `stet: prose linting for Ukrainian and English, on top of Vale.

  stet lint [--lang CODES] [--preset NAME] [--config PATH] [--output FMT] PATH...
  stet lint --list-presets
  stet fmt [--check] [PATH...]        format markdown, after the edits are in
  stet build [--detach|--status] [CODES]
  stet init [--lang CODES] [--force] [DIR]
  stet doctor
  stet uninstall [--dry-run]

Without --lang, a .vale.ini above the target wins, then every registered
language.`

// env carries what every command needs, so that no command reaches for a
// global and every test can hand it somewhere else to write.
type env struct {
	Layout paths.Layout
	Client *fetch.Client
	Out    io.Writer
	Err    io.Writer
}

// Main runs one command and returns the process exit code.
func Main(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}

	command, rest := args[0], args[1:]
	switch command {
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stdout, usage)
		return 0
	case "version", "--version":
		fmt.Fprintln(os.Stdout, "stet "+Version)
		return 0
	}

	layout, err := paths.Discover()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	e := &env{Layout: layout, Client: fetch.New(), Out: os.Stdout, Err: os.Stderr}

	var run func(*env, []string) int
	switch command {
	case "lint":
		run = runLint
	case "fmt":
		run = runFormat
	case "build":
		run = runBuild
	case "init":
		run = runInit
	case "doctor":
		run = runDoctor
	case "uninstall":
		run = runUninstall
	default:
		fmt.Fprintf(os.Stderr, "stet: unknown command %q\n\n%s\n", command, usage)
		return 2
	}
	return run(e, rest)
}

// Version is set at build time. An unset value says the binary was built
// outside the release process, which is worth seeing in a bug report.
var Version = "dev"

// exitCode unwraps the status of a child that ran and failed, apart from the
// errors that mean it never ran at all. A linter that found something and a
// linter that could not start need different reporting.
func exitCode(err error) (int, bool) {
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), true
	}
	return 0, false
}

// flagValue reads "--name value" and "--name=value" from args at position i,
// returning the value and how many arguments it consumed.
func flagValue(args []string, i int, name string) (string, int, bool) {
	arg := args[i]
	if arg == "--"+name {
		if i+1 < len(args) {
			return args[i+1], 2, true
		}
		return "", 1, true
	}
	if len(arg) > len(name)+3 && arg[:len(name)+3] == "--"+name+"=" {
		return arg[len(name)+3:], 1, true
	}
	return "", 0, false
}
