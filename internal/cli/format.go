package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/themaiby/stet/internal/tool"
)

// runFormat settles the layout of markdown once the words are settled.
//
// It runs last on purpose. Formatting before the edits would be undone by them:
// a rewritten sentence pushes a table out of shape and a replaced word breaks
// the wrapping, so the layout can only be fixed after the text stops moving.
func runFormat(e *env, args []string) int {
	action := "fmt"
	var targets []string
	for _, arg := range args {
		if arg == "--check" {
			action = "check"
			continue
		}
		targets = append(targets, arg)
	}

	dprint, err := tool.Resolve(tool.Dprint, e.Layout, e.Client, e.Err)
	if err != nil {
		fmt.Fprintln(e.Err, err)
		return 1
	}

	config := filepath.Join(e.Layout.Root, "dprint.json")
	if _, err := os.Stat(config); err != nil {
		fmt.Fprintf(e.Err, "stet: no formatter config at %s\n", config)
		return 1
	}

	cmd := exec.Command(dprint, append([]string{action, "--config", config}, targets...)...)
	cmd.Stdout, cmd.Stderr = e.Out, e.Err
	// The plugin is fetched rather than vendored, like every other third-party
	// artefact here, and its cache goes where uninstall will find it.
	cmd.Env = append(os.Environ(), "DPRINT_CACHE_DIR="+filepath.Join(e.Layout.Data, "dprint"))

	if err := cmd.Run(); err != nil {
		if code, ok := exitCode(err); ok {
			return code
		}
		fmt.Fprintln(e.Err, err)
		return 1
	}
	return 0
}
