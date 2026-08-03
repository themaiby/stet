package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/themaiby/stet/internal/tool"
)

// runDoctor reports one line per capability rather than a single pass or fail,
// because most of what can be missing costs a rule set rather than the tool.
func runDoctor(e *env, args []string) int {
	fmt.Fprintf(e.Out, "stet %s on %s %s\n\n", Version, runtime.GOOS, runtime.GOARCH)

	line := func(name, state string) { fmt.Fprintf(e.Out, "  %-22s %s\n", name, state) }

	for _, t := range []tool.Spec{tool.Vale, tool.Dprint, tool.Harper} {
		path, onPath, ok := tool.Present(t, e.Layout)
		switch {
		case ok && onPath:
			line(t.Name, "on PATH, "+path)
		case ok:
			line(t.Name, "downloaded earlier")
		default:
			line(t.Name, "will download on first use")
		}
	}

	if languages, err := loadLanguages(e); err == nil {
		line("languages", fmt.Sprintf("%d registered: %v", len(languages), languages.Codes()))
	} else {
		line("languages", "registry missing, nothing can be linted")
	}

	if presets, err := os.Open(e.Layout.Presets()); err == nil {
		presets.Close()
		if rows, err := loadPresets(e); err == nil {
			line("presets", fmt.Sprintf("%d shipped, no build needed", len(rows)))
		}
	} else {
		line("presets", "registry missing, it builds from committed data on first use")
	}

	if _, err := os.Stat(e.Layout.Style("config", "dictionaries", "uk_UA.dic")); err == nil {
		line("Ukrainian dictionary", "built")
	} else {
		line("Ukrainian dictionary", "not built yet, the first lint fetches it")
	}

	state := buildState(e)
	line("warm-up", fmt.Sprintf("%s, %s", state.Phase, state.Message))
	return 0
}
