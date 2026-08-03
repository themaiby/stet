// Command stet lints prose with Vale in Ukrainian, English or both, and
// formats markdown once the words are settled.
package main

import (
	"os"

	"github.com/themaiby/stet/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
