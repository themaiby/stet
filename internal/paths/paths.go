// Package paths answers where everything lives.
//
// The plugin root holds what the repository commits: rules, registries and
// measurements. The data directory holds what is downloaded or generated, and
// uninstall empties it. Keeping the two apart is what lets uninstall promise
// that nothing outside them is touched.
package paths

import (
	"errors"
	"os"
	"path/filepath"
)

// Layout is the resolved pair of roots.
type Layout struct {
	Root   string
	Data   string
	styles string
}

// ErrNoRoot is returned when the plugin root cannot be found. It means the
// binary was moved away from the rules without being told where they went.
var ErrNoRoot = errors.New("stet: cannot find the plugin root; set STET_ROOT")

// Discover resolves both roots from the environment and the running binary.
func Discover() (Layout, error) {
	root, err := findRoot()
	if err != nil {
		return Layout{}, err
	}
	l := Layout{Root: root, Data: dataDir()}
	if styles := os.Getenv("STET_STYLES"); styles != "" {
		l.styles = styles
	} else {
		l.styles = filepath.Join(root, "styles")
	}
	return l, nil
}

// Styles is the directory Vale resolves style names against.
func (l Layout) Styles() string { return l.styles }

// Dictionaries holds the generated word lists and the build stamps beside them.
func (l Layout) Dictionaries() string {
	return filepath.Join(l.styles, "config", "dictionaries")
}

// ConfigCache holds generated .vale.ini files, keyed by what they were asked
// for.
func (l Layout) ConfigCache() string { return filepath.Join(l.Root, "configs", ".cache") }

// Bin holds downloaded executables.
func (l Layout) Bin() string { return filepath.Join(l.Data, "bin") }

// State is the warm-up state file.
func (l Layout) State() string { return filepath.Join(l.Data, "warmup.state") }

// Log is where a detached warm-up writes.
func (l Layout) Log() string { return filepath.Join(l.Data, "warmup.log") }

// PID records a running detached warm-up.
func (l Layout) PID() string { return filepath.Join(l.Data, "warmup.pid") }

// Languages is the language registry.
func (l Layout) Languages() string { return filepath.Join(l.Root, "languages.conf") }

// Presets is the preset registry, which the preset generator writes.
func (l Layout) Presets() string { return filepath.Join(l.Root, "presets.conf") }

// DataFile names a committed measurement file.
func (l Layout) DataFile(name string) string { return filepath.Join(l.Root, "data", name) }

// Style names a directory under the styles root.
func (l Layout) Style(parts ...string) string {
	return filepath.Join(append([]string{l.styles}, parts...)...)
}

func findRoot() (string, error) {
	for _, candidate := range []string{os.Getenv("STET_ROOT"), os.Getenv("CLAUDE_PLUGIN_ROOT")} {
		if candidate != "" && isRoot(candidate) {
			return candidate, nil
		}
	}
	exe, err := os.Executable()
	if err == nil {
		if exe, err = filepath.EvalSymlinks(exe); err == nil {
			dir := filepath.Dir(exe)
			for _, candidate := range []string{dir, filepath.Dir(dir)} {
				if isRoot(candidate) {
					return candidate, nil
				}
			}
		}
	}
	if wd, err := os.Getwd(); err == nil && isRoot(wd) {
		return wd, nil
	}
	return "", ErrNoRoot
}

// isRoot recognises the plugin root by the registry every command needs. A
// directory without it cannot serve a lint run whatever else it holds.
func isRoot(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "languages.conf"))
	return err == nil
}

func dataDir() string {
	if dir := os.Getenv("CLAUDE_PLUGIN_DATA"); dir != "" {
		return dir
	}
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "stet")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "stet")
	}
	return filepath.Join(home, ".local", "share", "stet")
}
