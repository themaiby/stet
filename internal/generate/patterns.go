package generate

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// Limits that keep out rules Vale would run but nobody would want.
const (
	maxForms       = 40   // a lemma with more forms than this makes an unreadable regex
	minLiteral     = 6    // below this a mandatory match is too broad to mean anything
	minAlternative = 3    // single-letter alternatives match half the language
	maxTokens      = 6    // longer patterns lean on the tagger this cannot ask
	maxExpression  = 4000 // past this the rule is unreadable when it fires
)

// LemmaIndex maps a lemma to its inflected forms, stored unsplit. Splitting on
// lookup keeps the 174 MB Ukrainian index near its file size in memory.
type LemmaIndex map[string]string

// Forms returns the inflected forms of a lemma.
func (l LemmaIndex) Forms(lemma string) []string {
	raw, ok := l[lemma]
	if !ok {
		return nil
	}
	return strings.Fields(raw)
}

// PatternOptions describes the existence rule built from LanguageTool pattern
// XML.
type PatternOptions struct {
	Provenance string
	Message    string
	Level      string
}

// PatternResult reports what the conversion produced.
type PatternResult struct {
	Content   string
	Converted int
	Skipped   int
}

// plainStem matches a lemma written without any regex syntax. Anything else
// cannot be looked up in the index, because the index holds words.
var plainStem = regexp.MustCompile(`^[^\\\[\](){}*+?.^$|]+$`)

// openEndedWildcard matches ".*", ".+" and ".{", which the euphony rules use to
// mean "any neighbouring word". Outside LanguageTool's tokeniser that matches
// whole sentences.
var openEndedWildcard = regexp.MustCompile(`\.[*+{]`)

// backReference matches the "\1" placeholders a suggestion uses to quote the
// text it matched.
var backReference = regexp.MustCompile(`\\\d+`)

// Patterns converts LanguageTool pattern rules into a Vale existence rule. A
// rule survives when every token is literal text, a regex, or an inflected form
// the index can expand; anything needing a tagger is counted and skipped.
func Patterns(documents [][]byte, lemmas LemmaIndex, o PatternOptions) PatternResult {
	type converted struct {
		Expr  string
		Hints []string
	}
	var rules []converted
	skipped := 0

	for _, doc := range documents {
		root, err := parseXML(doc)
		if err != nil || root == nil {
			// One unreadable upstream file must not take the other rules down.
			continue
		}
		root.descendants("rule", func(rule *xnode) {
			expr, hints, ok := convertRule(rule, lemmas)
			if !ok {
				skipped++
				return
			}
			rules = append(rules, converted{expr, hints})
		})
	}

	seen := map[string]bool{}
	var lines []string
	for _, r := range rules {
		if seen[r.Expr] {
			continue
		}
		seen[r.Expr] = true
		if len(r.Hints) > 0 {
			lines = append(lines, "  # "+strings.Join(r.Hints, " / "))
		}
		lines = append(lines, "  - '"+strings.ReplaceAll(r.Expr, "'", "''")+"'")
	}

	var b strings.Builder
	b.WriteString("# " + o.Provenance + "\n")
	fmt.Fprintf(&b, "# %d rules converted, %d skipped as needing a tagger.\n", len(seen), skipped)
	b.WriteString("extends: existence\n")
	b.WriteString("message: \"" + o.Message + "\"\n")
	b.WriteString("level: " + o.Level + "\n")
	b.WriteString("ignorecase: true\n")
	b.WriteString("nonword: true\n")
	b.WriteString("tokens:\n")
	b.WriteString(strings.Join(lines, "\n") + "\n")

	return PatternResult{Content: b.String(), Converted: len(seen), Skipped: skipped}
}

func convertRule(rule *xnode, lemmas LemmaIndex) (string, []string, bool) {
	// Antipatterns carve exceptions out of a match, and dropping them would turn
	// a careful rule into a blunt one.
	if rule.find("antipattern") != nil {
		return "", nil, false
	}
	pattern := rule.find("pattern")
	if pattern == nil {
		return "", nil, false
	}
	tokens := pattern.findAll("token")
	if len(tokens) == 0 || len(tokens) > maxTokens {
		return "", nil, false
	}

	pieces := make([]string, 0, len(tokens))
	mandatory, literal := 0, 0
	for _, token := range tokens {
		piece, required, shortest, ok := tokenRegex(token, lemmas)
		if !ok {
			return "", nil, false
		}
		pieces = append(pieces, piece)
		if required {
			mandatory++
			literal += shortest
		}
	}
	if mandatory < 2 || literal < minLiteral {
		return "", nil, false
	}

	expr := strings.ReplaceAll(strings.Join(pieces, `[ \t]+`), `?[ \t]*[ \t]+`, `?[ \t]*`)
	if utf8.RuneCountInString(expr) > maxExpression {
		return "", nil, false
	}

	var hints []string
	if message := rule.find("message"); message != nil {
		message.descendants("suggestion", func(s *xnode) {
			text := strings.TrimSpace(backReference.ReplaceAllString(s.innerText(), "…"))
			if text != "" && len(hints) < 3 {
				hints = append(hints, text)
			}
		})
	}
	return expr, hints, true
}

// tokenRegex renders one token. It reports the expression, whether the token
// has to match, and the length of its shortest alternative.
func tokenRegex(token *xnode, lemmas LemmaIndex) (string, bool, int, bool) {
	if token.Attr["postag"] != "" || token.Attr["postag_regexp"] != "" {
		return "", false, 0, false
	}
	for _, child := range token.Children {
		if child.Name == "exception" || child.Name == "match" {
			return "", false, 0, false
		}
	}

	text := strings.TrimSpace(token.Text)
	if text == "" {
		return "", false, 0, false
	}

	var body string
	switch {
	case token.Attr["inflected"] == "yes":
		stems := []string{text}
		if token.Attr["regexp"] == "yes" {
			stems = strings.Split(text, "|")
		}
		var forms []string
		for _, stem := range stems {
			if !plainStem.MatchString(stem) {
				return "", false, 0, false
			}
			got := lemmas.Forms(stem)
			if len(got) == 0 {
				return "", false, 0, false
			}
			forms = append(forms, got...)
		}
		forms = sortedUnique(forms)
		if len(forms) == 0 || len(forms) > maxForms {
			return "", false, 0, false
		}
		escaped := make([]string, 0, len(forms))
		for _, f := range forms {
			escaped = append(escaped, escapeRegex(f))
		}
		body = strings.Join(escaped, "|")
	case token.Attr["regexp"] == "yes":
		if openEndedWildcard.MatchString(text) {
			return "", false, 0, false
		}
		body = text
	default:
		body = escapeRegex(text)
	}

	shortest := 0
	found := false
	for _, alt := range strings.Split(body, "|") {
		if alt == "" {
			continue
		}
		if n := utf8.RuneCountInString(alt); !found || n < shortest {
			shortest, found = n, true
		}
	}
	if !found || shortest < minAlternative {
		return "", false, 0, false
	}

	piece := "(?:" + body + ")"
	if token.Attr["min"] == "0" {
		return piece + `?[ \t]*`, false, 0, true
	}
	return piece, true, shortest, true
}

// regexSpecials is wider than what regexp.QuoteMeta escapes, and the difference
// shows up in the generated file.
const regexSpecials = "()[]{}?*+-|^$\\.&~# \t\n\r\v\f"

func escapeRegex(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r < utf8.RuneSelf && strings.ContainsRune(regexSpecials, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func sortedUnique(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
