package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/themaiby/stet/internal/grammar"
	"github.com/themaiby/stet/internal/registry"
	"github.com/themaiby/stet/internal/tool"
)

// grammarRules returns the rules to run beside Vale, and nothing unless English
// is the only language asked for. A checker that parses English sentences has
// nothing useful to say about a document written in another language, and a
// mixed document is one of those.
func grammarRules(languages registry.Languages, codes []string) []string {
	if len(codes) != 1 || codes[0] != "en" {
		return nil
	}
	english, ok := languages.Find("en")
	if !ok {
		return nil
	}
	return english.Grammar
}

// checkGrammar runs the grammar checker over the targets and returns what it
// found. A checker that cannot be fetched costs those rules and nothing else,
// so the failure is reported and the lint carries on.
func checkGrammar(e *env, rules, targets []string) []grammar.Finding {
	if len(rules) == 0 {
		return nil
	}
	harper, err := tool.Resolve(tool.Harper, e.Layout, e.Client, e.Err)
	if err != nil {
		fmt.Fprintf(e.Err, "stet: no grammar check this run: %v\n", err)
		return nil
	}

	// One target per run. The checker shortens the paths it reports, so two
	// files of the same name under different directories come back
	// indistinguishable when they arrive together.
	var all []grammar.Finding
	for _, target := range targets {
		var out bytes.Buffer
		cmd := exec.Command(harper, "lint", "--quiet", "--no-color",
			"--format", "json", "--only", strings.Join(rules, ","), target)
		cmd.Stdout = &out
		// Findings make it exit non-zero, so only an empty report means a real
		// failure to run.
		if err := cmd.Run(); err != nil && out.Len() == 0 {
			fmt.Fprintf(e.Err, "stet: grammar check failed on %s: %v\n", target, err)
			continue
		}
		findings, err := grammar.Parse(&out)
		if err != nil {
			fmt.Fprintf(e.Err, "stet: cannot read the grammar report: %v\n", err)
			continue
		}
		for i := range findings {
			findings[i].File = resolveAgainst(target, findings[i].File)
		}
		all = append(all, findings...)
	}
	return all
}

// resolveAgainst restores the path the caller named. The checker reports a path
// relative to the parent of what it was given, which is enough to rebuild it.
func resolveAgainst(target, reported string) string {
	if filepath.IsAbs(reported) {
		return reported
	}
	candidate := filepath.Join(filepath.Dir(target), reported)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return reported
}
