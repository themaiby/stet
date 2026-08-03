// Package valeconf renders a .vale.ini from the language registry, for both the
// lint cache and the config `stet init` leaves in a project.
package valeconf

import (
	"path"
	"strings"

	"github.com/themaiby/stet/internal/registry"
)

// Prose is the glob for files that are prose all the way through.
const Prose = "[*.{md,txt,csv,tsv,html,rst,adoc}]"

// Code is the glob for files Vale reads comments out of. Shell is missing on
// purpose: Vale has no comment syntax for it and reads the whole script as
// prose, so every case arm turns into a finding.
const Code = "[*.{ts,tsx,js,jsx,mjs,cjs,go,py,rb,java,kt,cs,php,rs,c,h,cpp,hpp,swift,scala,lua}]"

// Options is everything the renderer needs. Resolving names to rows, and
// deciding whether a base style exists on disk, happens before this point.
type Options struct {
	// Header is an optional first line, written as a comment.
	Header string
	// StylesPath is what Vale should resolve style names against.
	StylesPath string
	// Languages are the rows to load, in the order the caller asked for them.
	Languages []registry.Language
	// English is the row whose rules apply to comments, whatever the languages
	// above are. Another language's rules there would bless comments that
	// should not be in that language.
	English *registry.Language
	// Preset is the register to add, if any.
	Preset *registry.Preset
	// BaseStyle is the register-independent style that accompanies a preset,
	// empty when the language has none.
	BaseStyle string
	// PresetPolicy is what the chosen register turns off or demotes. A rule can
	// be wrong for a manual and right for a chat reply.
	PresetPolicy []string
}

// Render returns the config text.
func Render(o Options) string {
	styles := []string{"ProseCore"}
	if o.Preset != nil {
		if o.BaseStyle != "" {
			styles = append(styles, o.BaseStyle)
		}
		styles = append(styles, o.Preset.Style)
	}

	var packages []string
	var policy []string
	for _, lang := range o.Languages {
		styles = append(styles, lang.Style)
		packages = append(packages, lang.Packages...)
		policy = append(policy, lang.Policy...)
	}

	// The style names come from the packages the prose section actually loads,
	// which is settled before the English packages join the download list below.
	proseStyles := append(append([]string{}, styles...), styleNames(packages)...)

	codeStyles := []string{"ProseCore"}
	var englishPackages []string
	if o.English != nil {
		codeStyles = append(codeStyles, o.English.Style)
		englishPackages = o.English.Packages
	}
	codeStyles = append(codeStyles, styleNames(englishPackages)...)

	// Comments are checked with the English rules, so their packages have to be
	// downloaded even when no requested language asked for them.
	download := append([]string{}, packages...)
	for _, p := range englishPackages {
		if !contains(download, p) {
			download = append(download, p)
		}
	}

	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	if o.Header != "" {
		line("# " + o.Header)
	}
	line("StylesPath = " + o.StylesPath)
	line("MinAlertLevel = suggestion")
	line("Vocab = Project")
	if len(download) > 0 {
		line("Packages = " + strings.Join(download, ", "))
	}
	line("")
	line("[formats]")
	line("csv = txt")
	line("tsv = txt")
	line("ts = js")
	line("tsx = js")
	line("mts = js")
	line("cts = js")
	line("")
	line(Prose)
	line("BasedOnStyles = " + strings.Join(proseStyles, ", "))
	line("ProseCore.CommentLanguage = NO")
	if len(download) > 0 {
		line("Vale.Terms = NO")
		// ai-tells owns the dash where it is loaded; keeping ours counts it twice.
		line("ProseCore.Typography = NO")
	}
	for _, p := range policy {
		line(p)
	}
	// The register goes last because it knows more about this text than the
	// language row does.
	for _, p := range o.PresetPolicy {
		line(p)
	}
	line("")
	line(Code)
	line("BasedOnStyles = " + strings.Join(codeStyles, ", "))
	// Vale leaves the leading "*" of a JSDoc block in the text and lints code
	// inside @example as prose. On a 240-file TypeScript project that was 545
	// of 547 findings, so everything keyed to layout is off here.
	line("ProseCore.Formatting = NO")
	line("ProseCore.Typography = NO")
	if len(englishPackages) > 0 {
		// Comments are terse by design, so the wordiness rules come off too.
		for _, rule := range []string{
			"Vale.Terms",
			"write-good.TooWordy",
			"write-good.Weasel",
			"write-good.Passive",
			"write-good.E-Prime",
			"ai-tells.FormalRegister",
			"ai-tells.SemicolonUsage",
			"ai-tells.ColonUsage",
			"ai-tells.EmDashUsage",
			"ai-tells.VerbTricolon",
		} {
			line(rule + " = NO")
		}
	}

	return b.String()
}

// CacheName is the cache key: two runs asking for the same languages and preset
// reuse one file.
func CacheName(codes []string, preset string) string {
	name := strings.Join(codes, "-")
	if preset != "" {
		name += "-" + preset
	}
	return name + ".ini"
}

func styleNames(packages []string) []string {
	out := make([]string, 0, len(packages))
	for _, p := range packages {
		out = append(out, strings.TrimSuffix(path.Base(p), ".zip"))
	}
	return out
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
