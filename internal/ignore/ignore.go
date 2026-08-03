// Package ignore reads the paths a project keeps out of the linter and the
// formatter, and renders them for each.
package ignore

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileName is the file a project writes its exclusions in.
const FileName = ".stetignore"

// Patterns are glob patterns, in file order.
type Patterns []string

// Parse reads one pattern per line, skipping comments and blanks. A pattern
// ending in a slash covers the directory's contents, the way it reads.
func Parse(r io.Reader) Patterns {
	var out Patterns
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, "/") {
			line += "**"
		}
		out = append(out, line)
	}
	return out
}

// Load reads the file above start, the way a project config is found. It
// returns the patterns and the file they came from.
func Load(start string) (Patterns, string) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, ""
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	for {
		candidate := filepath.Join(dir, FileName)
		if file, err := os.Open(candidate); err == nil {
			defer file.Close()
			return Parse(file), candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ""
		}
		dir = parent
	}
}

// ValeGlob renders the patterns as the single negated glob Vale accepts. Vale
// takes one --glob and a second flag replaces the first, so several patterns
// have to arrive as one brace expression.
func (p Patterns) ValeGlob() string {
	switch len(p) {
	case 0:
		return ""
	case 1:
		return "!" + p[0]
	default:
		return "!{" + strings.Join(p, ",") + "}"
	}
}

// Match reports whether a path is one the project keeps out. It exists for the
// tools that take no exclusion flag of their own, whose findings have to be
// dropped after the fact.
func (p Patterns) Match(path string) bool {
	path = filepath.ToSlash(path)
	segments := strings.Split(path, "/")
	for _, pattern := range p {
		if prefix, ok := strings.CutSuffix(pattern, "/**"); ok {
			for _, segment := range segments[:max(len(segments)-1, 0)] {
				if segment == prefix {
					return true
				}
			}
			continue
		}
		if ok, _ := filepath.Match(pattern, filepath.Base(path)); ok {
			return true
		}
		if ok, _ := filepath.Match(pattern, path); ok {
			return true
		}
	}
	return false
}

// DprintExcludes renders the patterns as the arguments dprint adds to whatever
// its config already excludes. The flag is variadic and refuses to appear
// twice, so every pattern hangs off one of it.
func (p Patterns) DprintExcludes() []string {
	if len(p) == 0 {
		return nil
	}
	return append([]string{"--excludes"}, p...)
}
