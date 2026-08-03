package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/themaiby/stet/internal/build"
	"github.com/themaiby/stet/internal/warmup"
)

// runUninstall removes everything this plugin downloaded or generated. Nothing
// outside the plugin directory and its own data directory is touched. Removing
// the plugin itself is the package manager's job.
func runUninstall(e *env, args []string) int {
	dry := len(args) > 0 && args[0] == "--dry-run"

	targets := []string{
		e.Layout.Data,
		e.Layout.ConfigCache(),
		e.Layout.Style("config", "dictionaries"),
		e.Layout.Style("ai-tells"),
		e.Layout.Style("write-good"),
		e.Layout.Style("ProseUK", "Barbarism.yml"),
		e.Layout.Style("ProseUK", "Preferred.yml"),
		e.Layout.Style("ProseUK", "Calque.yml"),
		e.Layout.Style("ProseEN", "Plain.yml"),
		e.Layout.Presets(),
	}
	// Preset directories are named by the generator, so they are found rather
	// than listed: pinning the names here would duplicate the generator.
	if entries, err := os.ReadDir(e.Layout.Styles()); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() && isPresetStyle(name) {
				targets = append(targets, e.Layout.Style(name))
			}
		}
	}

	var freed int64
	for _, target := range targets {
		if _, err := os.Lstat(target); err != nil {
			continue
		}
		freed += diskUsage(target)
		relative := strings.TrimPrefix(target, e.Layout.Root+string(filepath.Separator))
		if dry {
			fmt.Fprintf(e.Out, "  would remove  %s\n", relative)
			continue
		}
		if err := os.RemoveAll(target); err != nil {
			fmt.Fprintf(e.Err, "stet: cannot remove %s: %v\n", relative, err)
			continue
		}
		fmt.Fprintf(e.Out, "  removed  %s\n", relative)
	}

	if freed == 0 {
		fmt.Fprintln(e.Out, "stet: nothing to remove.")
		return 0
	}
	verb := "freed"
	if dry {
		verb = "would free"
	}
	fmt.Fprintf(e.Out, "stet: %s %d MB. Run the linter again to rebuild it.\n", verb, freed/(1024*1024))
	return 0
}

// isPresetStyle recognises a generated register style by its language prefix.
// The hand-written styles are ProseCore, ProseUK and ProseEN, which this leaves
// alone.
func isPresetStyle(name string) bool {
	for _, prefix := range []string{"UK", "EN"} {
		if strings.HasPrefix(name, prefix) && !strings.HasPrefix(name, "Prose") {
			return true
		}
	}
	return false
}

func diskUsage(path string) int64 {
	var total int64
	filepath.WalkDir(path, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if info, err := entry.Info(); err == nil && !entry.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

func buildState(e *env) warmup.State {
	return build.New(e.Layout, e.Err).State()
}
