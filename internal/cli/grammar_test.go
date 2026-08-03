package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/themaiby/stet/internal/registry"
)

var registered = registry.Languages{
	{Code: "uk", Style: "ProseUK"},
	{Code: "en", Style: "ProseEN"},
}

func configWith(t *testing.T, styles string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".vale.ini")
	body := "StylesPath = styles\n\n[*.md]\nBasedOnStyles = ProseCore, " + styles + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLanguagesInConfig(t *testing.T) {
	cases := map[string][]string{
		"ProseEN":                    {"en"},
		"ProseUK":                    {"uk"},
		"ProseUK, ProseEN":           {"uk", "en"},
		"ai-tells, write-good":       nil,
		"ProseEN, ai-tells, ProseEN": {"en"},
	}
	for styles, want := range cases {
		got := languagesInConfig(configWith(t, styles), registered)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("styles %q gave %v, want %v", styles, got, want)
		}
	}
}

func TestLanguagesInConfigIgnoresAnUnreadableFile(t *testing.T) {
	if got := languagesInConfig("does/not/exist", registered); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestGrammarRunsForEnglishAlone(t *testing.T) {
	// An English grammar checker has nothing useful to say about a document in
	// another language, and a mixed document is one of those.
	english := registry.Language{Code: "en", Style: "ProseEN", Grammar: []string{"RepeatedWords"}}
	languages := registry.Languages{{Code: "uk", Style: "ProseUK"}, english}

	cases := []struct {
		Codes []string
		Want  bool
	}{
		{[]string{"en"}, true},
		{[]string{"uk"}, false},
		{[]string{"uk", "en"}, false},
		{nil, false},
	}
	for _, c := range cases {
		if got := len(grammarRules(languages, c.Codes)) > 0; got != c.Want {
			t.Errorf("codes %v ran grammar = %v, want %v", c.Codes, got, c.Want)
		}
	}
}

func TestGrammarStaysOffWhenTheLanguageDeclaresNoRules(t *testing.T) {
	languages := registry.Languages{{Code: "en", Style: "ProseEN"}}
	if got := grammarRules(languages, []string{"en"}); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

// Vale reads a path it cannot find as stdin and reports nothing wrong with it,
// so an unchecked target comes back as a clean document. That is the worst way
// for a linter to fail, and it happened: a file list that zsh passed as one
// argument reported clean over 36 files.
func TestMissingTargetsAreReported(t *testing.T) {
	present := filepath.Join(t.TempDir(), "doc.md")
	if err := os.WriteFile(present, []byte("# T\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := missingTargets([]string{present, "no-such-file.md", present + "\nalso-not-here.md"})
	if len(got) != 2 {
		t.Fatalf("missing = %v, want two entries", got)
	}
	if missingTargets([]string{present}) != nil {
		t.Error("a target that exists was reported missing")
	}
}
