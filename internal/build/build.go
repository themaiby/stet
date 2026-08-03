// Package build runs the generators that turn upstream data into rule files,
// and records how far it got.
//
// Sources are fetched at run time and never vendored. LanguageTool is LGPL-2.1
// and dict_uk is GPL-3.0, and keeping their data out of the repository keeps
// their licences out of it too.
package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/themaiby/stet/internal/fetch"
	"github.com/themaiby/stet/internal/generate"
	"github.com/themaiby/stet/internal/paths"
	"github.com/themaiby/stet/internal/warmup"
)

// Upstream coordinates. Pinned rather than tracked at latest, because slop
// rules decay and a refresh should be a deliberate change with a readable diff.
const (
	defaultDictVersion = "6.8.0"
	defaultLTRef       = "master"
)

const ltBase = "https://raw.githubusercontent.com/languagetool-org/languagetool/%s/languagetool-language-modules/%s/src/main/resources/org/languagetool/rules/%s"

// Runner holds what every builder needs.
type Runner struct {
	Layout paths.Layout
	Client *fetch.Client
	Log    io.Writer
}

// task is one builder: a name for the stamp file, a hint for the progress line,
// and the work.
type task struct {
	Name string
	Hint string
	Run  func(*Runner) error
}

// New returns a runner writing progress to w.
func New(layout paths.Layout, w io.Writer) *Runner {
	return &Runner{Layout: layout, Client: fetch.New(), Log: w}
}

// Tasks lists the builders for these language codes, skipping any whose output
// is already newer than this binary's idea of it.
func (r *Runner) Tasks(codes []string) []task {
	var out []task
	for _, code := range codes {
		for _, t := range byLanguage[code] {
			if r.stamped(t.Name) {
				continue
			}
			out = append(out, t)
		}
	}
	return out
}

// Run builds everything the codes ask for, writing the warm-up state as it goes.
func (r *Runner) Run(codes []string) error {
	tasks := r.Tasks(codes)
	if len(tasks) == 0 {
		return r.WriteState(warmup.State{Phase: warmup.Ready, Message: "already built"})
	}
	if err := os.MkdirAll(r.Layout.Dictionaries(), 0o755); err != nil {
		return err
	}

	for i, t := range tasks {
		step := i + 1
		r.WriteState(warmup.State{
			Phase: warmup.Building, Step: step, Total: len(tasks),
			Message: fmt.Sprintf("%s (%s)", t.Name, t.Hint),
		})
		fmt.Fprintf(r.Log, "stet: %d/%d building %s\n", step, len(tasks), t.Name)

		if err := t.Run(r); err != nil {
			// Silence here would let a rule check nothing at all, and the report
			// would read as clean because it was blind. Record the failure so
			// that the next lint run announces itself as partial.
			r.WriteState(warmup.State{
				Phase: warmup.Failed, Step: step, Total: len(tasks),
				Message: t.Name + " failed; rules from it are inactive",
			})
			return fmt.Errorf("stet: %s: %w", t.Name, err)
		}
		r.stamp(t.Name)
	}
	return r.WriteState(warmup.State{
		Phase: warmup.Ready, Step: len(tasks), Total: len(tasks),
		Message: fmt.Sprintf("built %d of %d", len(tasks), len(tasks)),
	})
}

// State reads the warm-up state.
func (r *Runner) State() warmup.State {
	data, err := os.ReadFile(r.Layout.State())
	if err != nil {
		return warmup.Missing()
	}
	return warmup.Parse(string(data))
}

// WriteState records the warm-up state where any process can read it.
func (r *Runner) WriteState(s warmup.State) error {
	if err := os.MkdirAll(r.Layout.Data, 0o755); err != nil {
		return err
	}
	return os.WriteFile(r.Layout.State(), []byte(s.String()+"\n"), 0o644)
}

func (r *Runner) stampPath(name string) string {
	return filepath.Join(r.Layout.Dictionaries(), "."+name+".built")
}

func (r *Runner) stamped(name string) bool {
	_, err := os.Stat(r.stampPath(name))
	return err == nil
}

func (r *Runner) stamp(name string) {
	os.WriteFile(r.stampPath(name), nil, 0o644)
}

func (r *Runner) write(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func ltURL(ref, lang, file string) string {
	return fmt.Sprintf(ltBase, ref, lang, lang) + "/" + file
}

func env(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

var byLanguage = map[string][]task{
	"uk": {
		{Name: "uk", Hint: "word forms, the slow one, around 20s", Run: (*Runner).buildUkrainianDictionary},
		{Name: "uk-lt", Hint: "rule pairs, a second or two", Run: (*Runner).buildUkrainianSubstitutions},
		{Name: "uk-lt-xml", Hint: "pattern rules, a second or two", Run: (*Runner).buildUkrainianPatterns},
		{Name: "uk-presets", Hint: "register rule sets, instant", Run: (*Runner).buildUkrainianPresets},
	},
	"en": {
		{Name: "en-lt", Hint: "rule pairs, a second or two", Run: (*Runner).buildEnglishSubstitutions},
		{Name: "en-presets", Hint: "register rule sets, instant", Run: (*Runner).buildEnglishPresets},
	},
}

func (r *Runner) buildUkrainianDictionary() error {
	version := env("UK_DICT_VERSION", defaultDictVersion)
	url := fmt.Sprintf("https://github.com/brown-uk/dict_uk/releases/download/v%s/dict_corp_vis.txt.bz2", version)
	fmt.Fprintf(r.Log, "stet: fetching the Ukrainian dictionary %s\n", version)

	corpus, err := r.Client.Bzip2(url)
	if err != nil {
		return err
	}
	defer corpus.Close()

	dest := r.Layout.Dictionaries()
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	dic, err := os.Create(filepath.Join(dest, "uk_UA.dic"))
	if err != nil {
		return err
	}
	defer dic.Close()
	lemmas, err := os.Create(filepath.Join(dest, "uk_lemmas.tsv"))
	if err != nil {
		return err
	}
	defer lemmas.Close()

	stats, err := generate.Dictionary(corpus, dic, lemmas)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dest, "uk_UA.aff"), []byte(generate.Affix), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(r.Log, "stet: %d word forms, %d lemmas indexed\n", stats.Forms, stats.Lemmas)
	return nil
}

func (r *Runner) buildUkrainianSubstitutions() error {
	ref := env("LT_REF", defaultLTRef)
	fmt.Fprintf(r.Log, "stet: fetching the LanguageTool Ukrainian rules (%s)\n", ref)

	files := []struct {
		Source string
		Out    string
		Opts   generate.SubstitutionOptions
	}{
		{"replace_soft.txt", "Preferred.yml", generate.SubstitutionOptions{
			Note:       "Softer preferences: the flagged form is not wrong, only weaker.",
			Provenance: "Generated by stet from LanguageTool (LGPL-2.1). Do not edit.",
			Message:    "Краще: '%s' замість '%s'.",
			Level:      "suggestion", BareKeysOnly: true, DropContext: true,
		}},
		{"replace.txt", "Barbarism.yml", generate.SubstitutionOptions{
			Note:       "Calques and borrowings with a settled native form.",
			Provenance: "Generated by stet from LanguageTool (LGPL-2.1). Do not edit.",
			Message:    "Калька або барбаризм: '%s'. Українською: '%s'.",
			Level:      "warning", BareKeysOnly: true, DropContext: true,
		}},
	}
	for _, f := range files {
		body, err := r.Client.Stream(ltURL(ref, "uk", f.Source))
		if err != nil {
			return err
		}
		content, pairs, err := generate.Substitution(body, f.Opts)
		body.Close()
		if err != nil {
			return err
		}
		if err := r.write(r.Layout.Style("ProseUK", f.Out), content); err != nil {
			return err
		}
		fmt.Fprintf(r.Log, "stet: %d pairs -> %s\n", pairs, f.Out)
	}
	return nil
}

func (r *Runner) buildEnglishSubstitutions() error {
	ref := env("LT_REF", defaultLTRef)
	fmt.Fprintf(r.Log, "stet: fetching the LanguageTool English rules (%s)\n", ref)

	body, err := r.Client.Stream(ltURL(ref, "en", "replace.txt"))
	if err != nil {
		return err
	}
	defer body.Close()

	content, pairs, err := generate.Substitution(body, generate.SubstitutionOptions{
		Provenance: "Generated by stet from LanguageTool (LGPL-2.1). Do not edit.",
		Message:    `Use '%s' instead of '%s'.`,
		Level:      "warning",
	})
	if err != nil {
		return err
	}
	if err := r.write(r.Layout.Style("ProseEN", "Plain.yml"), content); err != nil {
		return err
	}
	fmt.Fprintf(r.Log, "stet: %d pairs -> Plain.yml\n", pairs)
	return nil
}

func (r *Runner) buildUkrainianPatterns() error {
	ref := env("LT_REF", defaultLTRef)
	fmt.Fprintf(r.Log, "stet: fetching the LanguageTool pattern rules (%s)\n", ref)

	var documents [][]byte
	for _, name := range []string{"grammar-barbarism.xml", "grammar-style.xml"} {
		data, err := r.Client.Bytes(ltURL(ref, "uk", name))
		if err != nil {
			return err
		}
		documents = append(documents, data)
	}

	index := generate.LemmaIndex{}
	file, err := os.Open(filepath.Join(r.Layout.Dictionaries(), "uk_lemmas.tsv"))
	if err == nil {
		defer file.Close()
		// A missing index is not fatal: it costs the rules whose tokens are
		// inflected, and the literal ones still convert.
		if index, err = generate.LoadLemmas(file); err != nil {
			return err
		}
	}

	result := generate.Patterns(documents, index, generate.PatternOptions{
		Provenance: "Generated by stet from LanguageTool (LGPL-2.1). Do not edit.",
		Message:    "Калька або стилістична хиба: '%s'",
		Level:      "warning",
	})
	if err := r.write(r.Layout.Style("ProseUK", "Calque.yml"), result.Content); err != nil {
		return err
	}
	fmt.Fprintf(r.Log, "stet: %d rules converted, %d skipped\n", result.Converted, result.Skipped)
	return nil
}

func (r *Runner) buildUkrainianPresets() error { return r.buildPresets("UK", "uk") }

func (r *Runner) buildEnglishPresets() error { return r.buildPresets("EN", "en") }

// buildPresets reads the committed measurement rather than the corpora, so it
// needs no network and no fifteen megabytes of text.
func (r *Runner) buildPresets(lang, code string) error {
	excessName := code + "-excess.tsv"
	patternFile, err := os.Open(r.Layout.DataFile(code + "-patterns.tsv"))
	if err != nil {
		return err
	}
	defer patternFile.Close()
	patterns, err := generate.ParsePatterns(patternFile)
	if err != nil {
		return err
	}

	excessFile, err := os.Open(r.Layout.DataFile(excessName))
	if err != nil {
		return err
	}
	defer excessFile.Close()
	excess, notes, err := generate.ParseExcess(excessFile)
	if err != nil {
		return err
	}

	set := generate.Presets(patterns, excess, notes, lang, excessName)
	for _, f := range set.Files {
		if err := r.write(r.Layout.Style(f.Dir, f.Name), f.Content); err != nil {
			return err
		}
		fmt.Fprintf(r.Log, "stet: %s/%s, %d constructions\n", f.Dir, f.Name, f.Count)
	}

	var existing []string
	if data, err := os.ReadFile(r.Layout.Presets()); err == nil {
		existing = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}
	return r.write(r.Layout.Presets(), generate.MergeRegistry(existing, lang, set.Rows))
}
