package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runVocab reports which vocabulary a lint of this path would read. A project
// without its own .vale.ini falls back to the plugin's copy, which every other
// project on the machine shares and uninstall removes, so a term written there
// is neither private nor kept.
func runVocab(e *env, args []string) int {
	target := "."
	if len(args) > 0 {
		target = args[0]
	}

	config, local := findProjectConfig(target)
	dir := e.Layout.Style("config", "vocabularies", "Project")
	if local {
		base := filepath.Dir(config)
		styles, vocab := readStyles(config)
		dir = filepath.Join(base, filepath.FromSlash(styles), "config", "vocabularies", vocab)
	}

	fmt.Fprintln(e.Out, dir)
	for _, name := range []string{"accept.txt", "reject.txt"} {
		fmt.Fprintf(e.Out, "  %-11s %s\n", name, countTerms(filepath.Join(dir, name)))
	}
	if local {
		fmt.Fprintf(e.Out, "  %-11s this project, through %s\n", "scope", config)
		return 0
	}
	fmt.Fprintf(e.Out, "  %-11s the plugin, shared by every project and removed by uninstall\n", "scope")
	fmt.Fprintf(e.Out, "  %-11s run 'stet init' in the project to give it one of its own\n", "fix")
	return 0
}

func readStyles(config string) (styles, vocab string) {
	styles, vocab = "styles", "Project"
	file, err := os.Open(config)
	if err != nil {
		return styles, vocab
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "StylesPath":
			styles = strings.TrimSpace(value)
		case "Vocab":
			vocab = strings.TrimSpace(value)
		}
	}
	return styles, vocab
}

func countTerms(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "missing"
	}
	terms := 0
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			terms++
		}
	}
	return fmt.Sprintf("%d terms", terms)
}
