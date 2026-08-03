// Package registry reads the two pipe-separated tables the rest of the tool
// steers by: which languages exist, and which presets were measured for them.
package registry

import (
	"bufio"
	"io"
	"strings"
)

// Language is one row of languages.conf:
//
//	code | style directory | packages | policy lines | grammar rules
type Language struct {
	Code     string
	Style    string
	Packages []string
	Policy   []string
	// Grammar names the rules a sentence-parsing checker runs beside Vale.
	Grammar []string
}

// Preset is one row of presets.conf:
//
//	lang | code | style directory | description | preliminary
type Preset struct {
	Lang        string
	Code        string
	Style       string
	Description string
	Preliminary bool
}

// Languages is the language table in file order.
type Languages []Language

// Presets is the preset table in file order.
type Presets []Preset

// ParseLanguages reads languages.conf. Rows short of a code are skipped rather
// than reported: a comment style nobody agreed on should not stop a lint run.
func ParseLanguages(r io.Reader) (Languages, error) {
	var out Languages
	err := eachRow(r, func(fields []string) {
		if fields[0] == "" {
			return
		}
		out = append(out, Language{
			Code:     fields[0],
			Style:    field(fields, 1),
			Packages: strings.Fields(field(fields, 2)),
			Policy:   policyLines(field(fields, 3)),
			Grammar:  splitList(field(fields, 4)),
		})
	})
	return out, err
}

// ParsePresets reads presets.conf.
func ParsePresets(r io.Reader) (Presets, error) {
	var out Presets
	err := eachRow(r, func(fields []string) {
		if len(fields) < 3 || fields[0] == "" || fields[1] == "" {
			return
		}
		out = append(out, Preset{
			Lang:        fields[0],
			Code:        fields[1],
			Style:       fields[2],
			Description: field(fields, 3),
			Preliminary: field(fields, 4) == "preliminary",
		})
	})
	return out, err
}

// PresetPolicy holds the policy lines a preset adds, keyed by language and code.
// presets.conf is generated from measurement and states no opinions, so the
// opinions live apart from it.
type PresetPolicy map[string][]string

// ParsePresetPolicy reads presets-policy.conf.
func ParsePresetPolicy(r io.Reader) (PresetPolicy, error) {
	out := PresetPolicy{}
	err := eachRow(r, func(fields []string) {
		if len(fields) < 3 || fields[0] == "" || fields[1] == "" {
			return
		}
		if lines := policyLines(fields[2]); len(lines) > 0 {
			out[fields[0]+"|"+fields[1]] = lines
		}
	})
	return out, err
}

// For returns the policy a preset adds, and nothing when it adds none.
func (p PresetPolicy) For(lang, code string) []string { return p[lang+"|"+code] }

// Find returns the row for a language code.
func (l Languages) Find(code string) (Language, bool) {
	for _, lang := range l {
		if lang.Code == code {
			return lang, true
		}
	}
	return Language{}, false
}

// Codes lists every registered language in file order.
func (l Languages) Codes() []string {
	out := make([]string, 0, len(l))
	for _, lang := range l {
		out = append(out, lang.Code)
	}
	return out
}

// Find returns the preset for a language and code. A code such as "docs"
// exists under more than one language, so the pair is the key. Looking up the
// code alone picked the wrong language.
func (p Presets) Find(lang, code string) (Preset, bool) {
	for _, preset := range p {
		if preset.Lang == lang && preset.Code == code {
			return preset, true
		}
	}
	return Preset{}, false
}

// FindByCode returns the first preset with this code under any language, which
// is what the refusal message needs to name.
func (p Presets) FindByCode(code string) (Preset, bool) {
	for _, preset := range p {
		if preset.Code == code {
			return preset, true
		}
	}
	return Preset{}, false
}

func eachRow(r io.Reader, fn func(fields []string)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fn(strings.Split(line, "|"))
	}
	return scanner.Err()
}

func field(fields []string, i int) string {
	if i < len(fields) {
		return strings.TrimSpace(fields[i])
	}
	return ""
}

func splitList(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func policyLines(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ";") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
