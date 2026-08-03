package registry

import (
	"reflect"
	"strings"
	"testing"
)

const languagesFile = `# a comment

uk|ProseUK||
en|ProseEN|https://a.test/one.zip https://a.test/two.zip|write-good.E-Prime = NO; write-good.Passive = suggestion
`

func TestParseLanguages(t *testing.T) {
	got, err := ParseLanguages(strings.NewReader(languagesFile))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("rows = %d, want 2", len(got))
	}
	if !reflect.DeepEqual(got.Codes(), []string{"uk", "en"}) {
		t.Errorf("codes = %v", got.Codes())
	}

	en, ok := got.Find("en")
	if !ok {
		t.Fatal("en is missing")
	}
	if len(en.Packages) != 2 {
		t.Errorf("packages = %v", en.Packages)
	}
	want := []string{"write-good.E-Prime = NO", "write-good.Passive = suggestion"}
	if !reflect.DeepEqual(en.Policy, want) {
		t.Errorf("policy = %v, want %v", en.Policy, want)
	}

	uk, _ := got.Find("uk")
	if len(uk.Packages) != 0 || len(uk.Policy) != 0 {
		t.Errorf("empty fields became entries: %+v", uk)
	}
}

const presetsFile = `# lang|code|style|description
en|docs|ENDocs|documentation
uk|docs|UKDocs|документація
uk|fiction|UKFiction|художня література|preliminary
`

func TestPresetsAreFoundByLanguageAndCode(t *testing.T) {
	// A code such as "docs" exists under more than one language, so the pair is
	// the key. Looking up the code alone picked the wrong language.
	got, err := ParsePresets(strings.NewReader(presetsFile))
	if err != nil {
		t.Fatal(err)
	}

	uk, ok := got.Find("uk", "docs")
	if !ok || uk.Style != "UKDocs" {
		t.Errorf("uk docs = %+v", uk)
	}
	en, ok := got.Find("en", "docs")
	if !ok || en.Style != "ENDocs" {
		t.Errorf("en docs = %+v", en)
	}
	if _, ok := got.Find("en", "fiction"); ok {
		t.Error("a preset was found under a language it was not measured on")
	}
}

func TestFindByCodeNamesTheMeasuredLanguage(t *testing.T) {
	// This is what the refusal message needs: which language the preset does
	// belong to.
	got, _ := ParsePresets(strings.NewReader(presetsFile))
	preset, ok := got.FindByCode("fiction")
	if !ok || preset.Lang != "uk" {
		t.Errorf("FindByCode = %+v", preset)
	}
}

func TestPreliminaryIsCarried(t *testing.T) {
	got, _ := ParsePresets(strings.NewReader(presetsFile))
	fiction, _ := got.Find("uk", "fiction")
	if !fiction.Preliminary {
		t.Error("the preliminary marker was lost")
	}
	docs, _ := got.Find("uk", "docs")
	if docs.Preliminary {
		t.Error("a settled preset was marked preliminary")
	}
}
