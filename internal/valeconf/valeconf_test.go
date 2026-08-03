package valeconf

import (
	"strings"
	"testing"

	"github.com/themaiby/stet/internal/registry"
)

var (
	ukrainian = registry.Language{Code: "uk", Style: "ProseUK"}
	english   = registry.Language{
		Code:     "en",
		Style:    "ProseEN",
		Packages: []string{"https://example.test/ai-tells.zip"},
		Policy:   []string{"write-good.E-Prime = NO"},
	}
)

func section(config, header string) string {
	_, rest, ok := strings.Cut(config, header+"\n")
	if !ok {
		return ""
	}
	body, _, _ := strings.Cut(rest, "\n[")
	return body
}

func TestRenderChecksCommentsAsEnglishWhateverTheLanguage(t *testing.T) {
	// Comments are written in English by convention. Applying another
	// language's rules there would quietly bless comments that should not be in
	// that language.
	got := Render(Options{
		StylesPath: "styles",
		Languages:  []registry.Language{ukrainian},
		English:    &english,
	})

	prose := section(got, Prose)
	if !strings.Contains(prose, "BasedOnStyles = ProseCore, ProseUK\n") {
		t.Errorf("prose section:\n%s", prose)
	}
	code := section(got, Code)
	if !strings.Contains(code, "BasedOnStyles = ProseCore, ProseEN, ai-tells\n") {
		t.Errorf("code section:\n%s", code)
	}
}

func TestRenderDownloadsEnglishPackagesForCommentsAlone(t *testing.T) {
	// A Ukrainian-only run still checks comments with the English rules, so
	// their packages have to arrive even though no requested language asked.
	got := Render(Options{
		StylesPath: "styles",
		Languages:  []registry.Language{ukrainian},
		English:    &english,
	})
	if !strings.Contains(got, "Packages = https://example.test/ai-tells.zip\n") {
		t.Errorf("the English packages were not requested:\n%s", got)
	}
}

func TestRenderListsEachPackageOnce(t *testing.T) {
	got := Render(Options{
		StylesPath: "styles",
		Languages:  []registry.Language{ukrainian, english},
		English:    &english,
	})
	if strings.Count(got, "https://example.test/ai-tells.zip") != 1 {
		t.Errorf("a package was requested twice:\n%s", got)
	}
}

func TestRenderCarriesLanguagePolicy(t *testing.T) {
	got := Render(Options{
		StylesPath: "styles",
		Languages:  []registry.Language{english},
		English:    &english,
	})
	if !strings.Contains(section(got, Prose), "write-good.E-Prime = NO\n") {
		t.Errorf("the language's default policy was dropped:\n%s", got)
	}
}

func TestRenderKeepsFormattingRulesOffForSourceFiles(t *testing.T) {
	// Measured on a 240-file TypeScript project: of 547 findings, 545 came from
	// doc-comment formatting rather than the prose.
	got := section(Render(Options{
		StylesPath: "styles",
		Languages:  []registry.Language{english},
		English:    &english,
	}), Code)

	for _, want := range []string{
		"ProseCore.Formatting = NO",
		"ProseCore.Typography = NO",
		"write-good.TooWordy = NO",
		"ai-tells.EmDashUsage = NO",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in the code section:\n%s", want, got)
		}
	}
}

func TestRenderPlacesThePresetAheadOfTheLanguage(t *testing.T) {
	preset := registry.Preset{Lang: "uk", Code: "docs", Style: "UKDocs"}
	got := Render(Options{
		StylesPath: "styles",
		Languages:  []registry.Language{ukrainian},
		Preset:     &preset,
		BaseStyle:  "UKBase",
	})
	if !strings.Contains(got, "BasedOnStyles = ProseCore, UKBase, UKDocs, ProseUK\n") {
		t.Errorf("style order:\n%s", got)
	}
}

func TestRenderKeepsPresetsOutOfComments(t *testing.T) {
	// Presets were measured on registers of prose, which a comment is not.
	preset := registry.Preset{Lang: "uk", Code: "docs", Style: "UKDocs"}
	got := section(Render(Options{
		StylesPath: "styles",
		Languages:  []registry.Language{ukrainian},
		English:    &english,
		Preset:     &preset,
		BaseStyle:  "UKBase",
	}), Code)

	if strings.Contains(got, "UKDocs") || strings.Contains(got, "UKBase") {
		t.Errorf("a preset reached the code section:\n%s", got)
	}
}

func TestCacheNameKeysOnLanguagesAndPreset(t *testing.T) {
	if got := CacheName([]string{"uk", "en"}, ""); got != "uk-en.ini" {
		t.Errorf("CacheName = %q", got)
	}
	if got := CacheName([]string{"uk"}, "docs"); got != "uk-docs.ini" {
		t.Errorf("CacheName = %q", got)
	}
}
