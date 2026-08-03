package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/themaiby/stet/internal/ignore"
	"github.com/themaiby/stet/internal/tool"
)

// runFormat settles the layout of markdown. It runs after the edits, because
// editing is what pushes tables and wrapping out of shape.
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

	dprintArgs := []string{action, "--config", config}
	scope := "."
	if len(targets) > 0 {
		scope = targets[0]
	}
	if patterns, from := ignore.Load(scope); len(patterns) > 0 {
		dprintArgs = append(dprintArgs, patterns.DprintExcludes()...)
		fmt.Fprintf(e.Err, "stet: %d path patterns ignored, from %s\n", len(patterns), from)
	}
	// A variadic --excludes would swallow the paths that follow it.
	if len(targets) > 0 {
		dprintArgs = append(dprintArgs, "--")
	}
	cmd := exec.Command(dprint, append(dprintArgs, targets...)...)
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
