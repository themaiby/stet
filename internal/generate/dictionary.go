package generate

import (
	"bufio"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Ukrainian spells the apostrophe in more than one way, and the speller
// compares characters, so a word is emitted under each.
const (
	apostropheModifier = "ʼ" // ʼ modifier letter
	apostropheAscii    = "'"
	apostropheRight    = "’" // ’ right single quote
)

// DictionaryStats reports what a dictionary build produced.
type DictionaryStats struct {
	Forms  int
	Lemmas int
}

// Affix is the whole affix file. Forms are expanded up front instead of
// declared here: Vale compiles Hunspell conditions to Go regexes, and 647 of
// the upstream rules need a lookbehind, which Go has not got.
const Affix = "SET UTF-8\n"

// Dictionary converts the dict_uk corpus into a Hunspell dictionary and the
// lemma index the pattern converter expands inflected tokens with.
//
// The corpus puts a lemma at column zero and its forms on indented lines, so
// stripping the indent first is mandatory or only lemmas survive.
func Dictionary(corpus io.Reader, dic, lemmas io.Writer) (DictionaryStats, error) {
	var forms []string
	seenLemma := map[string]bool{}

	lemmaOut := bufio.NewWriter(lemmas)
	stats := DictionaryStats{}

	var lemma string
	var group []string

	flush := func() error {
		if lemma == "" || seenLemma[lemma] {
			return nil
		}
		seenLemma[lemma] = true
		stats.Lemmas++
		if _, err := lemmaOut.WriteString(lemma + "\t" + strings.Join(group, " ") + "\n"); err != nil {
			return err
		}
		return nil
	}

	scanner := bufio.NewScanner(corpus)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		indented := strings.HasPrefix(line, " ")
		word := firstField(strings.TrimLeft(line, " "))
		if word == "" {
			continue
		}
		forms = append(forms, strings.ReplaceAll(word, apostropheAscii, apostropheModifier))

		if indented {
			if lemma != "" {
				group = append(group, word)
			}
			continue
		}
		if err := flush(); err != nil {
			return stats, err
		}
		lemma, group = word, []string{word}
	}
	if err := scanner.Err(); err != nil {
		return stats, err
	}
	if err := flush(); err != nil {
		return stats, err
	}
	if err := lemmaOut.Flush(); err != nil {
		return stats, err
	}

	all := expandApostrophes(forms)
	sort.Strings(all)
	unique := dedupeSorted(all)
	stats.Forms = len(unique)

	dicOut := bufio.NewWriter(dic)
	// Hunspell reads the first line as the entry count.
	if _, err := dicOut.WriteString(strconv.Itoa(len(unique)) + "\n"); err != nil {
		return stats, err
	}
	for _, form := range unique {
		if _, err := dicOut.WriteString(form + "\n"); err != nil {
			return stats, err
		}
	}
	return stats, dicOut.Flush()
}

// expandApostrophes adds the other two spellings for every word that carries
// one, which is few of them, so this costs far less than three copies.
func expandApostrophes(forms []string) []string {
	out := forms
	for _, form := range forms {
		if strings.Contains(form, apostropheModifier) {
			out = append(out,
				strings.ReplaceAll(form, apostropheModifier, apostropheAscii),
				strings.ReplaceAll(form, apostropheModifier, apostropheRight))
		}
	}
	return out
}

func dedupeSorted(sorted []string) []string {
	if len(sorted) == 0 {
		return sorted
	}
	out := sorted[:1]
	for _, s := range sorted[1:] {
		if s != out[len(out)-1] {
			out = append(out, s)
		}
	}
	return out
}

// firstField returns the text up to the first space, which is where the corpus
// keeps the word and after which it keeps the tags.
func firstField(line string) string {
	if i := strings.IndexByte(line, ' '); i >= 0 {
		return line[:i]
	}
	return line
}

// LoadLemmas reads the lemma index written by Dictionary.
func LoadLemmas(r io.Reader) (LemmaIndex, error) {
	index := LemmaIndex{}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		lemma, forms, _ := strings.Cut(scanner.Text(), "\t")
		if lemma != "" {
			index[lemma] = forms
		}
	}
	return index, scanner.Err()
}
