package ignore

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	source := strings.Join([]string{
		"# what this project keeps out",
		"",
		"vendor/",
		"  build/**  ",
		"*.gen.md",
	}, "\n")

	got := Parse(strings.NewReader(source))
	want := Patterns{"vendor/**", "build/**", "*.gen.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %v, want %v", got, want)
	}
}

func TestValeGlobJoinsIntoOnePattern(t *testing.T) {
	// Vale takes one --glob, and a second flag replaces the first rather than
	// adding to it, so everything has to arrive in one expression.
	cases := []struct {
		Patterns Patterns
		Want     string
	}{
		{nil, ""},
		{Patterns{"vendor/**"}, "!vendor/**"},
		{Patterns{"vendor/**", "*.gen.md"}, "!{vendor/**,*.gen.md}"},
	}
	for _, c := range cases {
		if got := c.Patterns.ValeGlob(); got != c.Want {
			t.Errorf("ValeGlob(%v) = %q, want %q", c.Patterns, got, c.Want)
		}
	}
}

func TestDprintExcludesHangsEveryPatternOffOneFlag(t *testing.T) {
	// dprint refuses a repeated --excludes and takes the patterns variadically.
	got := Patterns{"vendor/**", "*.gen.md"}.DprintExcludes()
	want := []string{"--excludes", "vendor/**", "*.gen.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DprintExcludes = %v, want %v", got, want)
	}
	if Patterns(nil).DprintExcludes() != nil {
		t.Error("no patterns produced arguments")
	}
}
