// Package plugin holds no code. The test beside this file guards the plugin
// manifest, which no compiler and no `claude plugin validate` reads.
package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// The standard hooks/hooks.json loads by convention. Naming it in the manifest
// as well makes the host refuse the whole plugin, skills and all, and neither
// the validator nor the installer says so: the failure appears only when a
// session loads the hooks.
func TestManifestDoesNotNameTheStandardHooksFile(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}

	standard := filepath.Join(root, "hooks", "hooks.json")
	if _, err := os.Stat(standard); err != nil {
		t.Skip("no hooks/hooks.json to conflict with")
	}
	if hooks, named := manifest["hooks"]; named {
		t.Errorf("manifest names hooks %v while hooks/hooks.json exists; "+
			"the host loads that file by convention and refuses the plugin for the duplicate", hooks)
	}
}

// A manifest without a name fails with a schema error at install time.
func TestManifestCarriesWhatTheHostRequires(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name == "" || manifest.Version == "" || manifest.Description == "" {
		t.Errorf("manifest is missing a required field: %+v", manifest)
	}
}
