// Package tool resolves the external binaries this plugin drives: the one
// already on the machine, otherwise a download into the plugin data directory.
package tool

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/themaiby/stet/internal/fetch"
	"github.com/themaiby/stet/internal/paths"
)

// Spec describes a downloadable binary.
type Spec struct {
	// Name is the executable name, without any platform suffix.
	Name string
	// Version is the pinned release.
	Version string
	// Assets maps "goos/goarch" to the release asset holding the binary.
	Assets map[string]string
	// URL renders the download address for one asset.
	URL func(version, asset string) string
}

// Vale is the linter itself.
var Vale = Spec{
	Name:    "vale",
	Version: "3.16.0",
	Assets: map[string]string{
		"linux/amd64":   "vale_%s_Linux_64-bit.tar.gz",
		"linux/arm64":   "vale_%s_Linux_arm64.tar.gz",
		"darwin/amd64":  "vale_%s_macOS_64-bit.tar.gz",
		"darwin/arm64":  "vale_%s_macOS_arm64.tar.gz",
		"windows/amd64": "vale_%s_Windows_64-bit.zip",
		"windows/arm64": "vale_%s_Windows_arm64.zip",
	},
	// Releases moved from errata-ai to vale-cli. The old address still
	// redirects, and this is the one that will not need a redirect.
	URL: func(version, asset string) string {
		return "https://github.com/vale-cli/vale/releases/download/v" + version + "/" + asset
	},
}

// Dprint formats markdown. Linting reports what is wrong with the words, and
// this settles what the layout looks like once the words are fixed.
var Dprint = Spec{
	Name:    "dprint",
	Version: "0.55.2",
	Assets: map[string]string{
		"linux/amd64":   "dprint-x86_64-unknown-linux-gnu.zip",
		"linux/arm64":   "dprint-aarch64-unknown-linux-gnu.zip",
		"darwin/amd64":  "dprint-x86_64-apple-darwin.zip",
		"darwin/arm64":  "dprint-aarch64-apple-darwin.zip",
		"windows/amd64": "dprint-x86_64-pc-windows-msvc.zip",
		"windows/arm64": "dprint-aarch64-pc-windows-msvc.zip",
	},
	URL: func(version, asset string) string {
		return "https://github.com/dprint/dprint/releases/download/" + version + "/" + asset
	},
}

// Resolve returns a path to a usable binary, downloading one only if the
// machine has none.
func Resolve(spec Spec, layout paths.Layout, client *fetch.Client, log io.Writer) (string, error) {
	if path, err := exec.LookPath(spec.Name); err == nil {
		return path, nil
	}
	cached := filepath.Join(layout.Bin(), executable(spec.Name))
	if info, err := os.Stat(cached); err == nil && !info.IsDir() {
		return cached, nil
	}

	platform := runtime.GOOS + "/" + runtime.GOARCH
	pattern, ok := spec.Assets[platform]
	if !ok {
		return "", fmt.Errorf("stet: no prebuilt %s for %s; install it and keep it on PATH", spec.Name, platform)
	}
	asset := pattern
	if n := countVerbs(pattern); n > 0 {
		asset = fmt.Sprintf(pattern, spec.Version)
	}

	fmt.Fprintf(log, "stet: downloading %s\n", asset)
	if err := client.Binary(spec.URL(spec.Version, asset), executable(spec.Name), cached); err != nil {
		return "", err
	}
	return cached, nil
}

// Present reports where a binary is without fetching one, which is what doctor
// needs: it describes the machine and must not change it.
func Present(spec Spec, layout paths.Layout) (path string, onPath bool, ok bool) {
	if p, err := exec.LookPath(spec.Name); err == nil {
		return p, true, true
	}
	cached := filepath.Join(layout.Bin(), executable(spec.Name))
	if info, err := os.Stat(cached); err == nil && !info.IsDir() {
		return cached, false, true
	}
	return "", false, false
}

func executable(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func countVerbs(pattern string) int {
	count := 0
	for i := 0; i+1 < len(pattern); i++ {
		if pattern[i] == '%' && pattern[i+1] == 's' {
			count++
		}
	}
	return count
}
