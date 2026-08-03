// Command stet lints prose with Vale in Ukrainian, English or both, and
// formats the markdown once the words are settled.
//
// One static binary, so the only thing a machine needs is the binary itself.
// Native Windows without Git for Windows is the case that forced it: there the
// host runs commands through PowerShell, and a shell script has nothing to run
// it.
package main

import (
	"os"

	"github.com/themaiby/stet/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
