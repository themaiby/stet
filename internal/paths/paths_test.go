package paths

import (
	"os"
	"path/filepath"
	"testing"
)

// A downloaded binary sits in the data directory, nowhere near the rules, and a
// caller outside Claude Code has no CLAUDE_PLUGIN_ROOT to point it back. The
// recipe in the skill failed exactly that way until bootstrap.sh left the answer
// behind.
func TestRootComesFromTheFileBootstrapLeaves(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "languages.conf"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	data := t.TempDir()
	if err := os.WriteFile(filepath.Join(data, "root"), []byte(root+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("STET_ROOT", "")
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	t.Setenv("CLAUDE_PLUGIN_DATA", data)
	t.Chdir(t.TempDir())

	layout, err := Discover()
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if layout.Root != root {
		t.Errorf("Root = %q, want %q", layout.Root, root)
	}
}

func TestRootFailsWhenNothingPointsAtIt(t *testing.T) {
	t.Setenv("STET_ROOT", "")
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	t.Setenv("CLAUDE_PLUGIN_DATA", t.TempDir())
	t.Chdir(t.TempDir())

	if _, err := Discover(); err == nil {
		t.Error("Discover found a root that does not exist")
	}
}
