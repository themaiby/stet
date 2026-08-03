package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/themaiby/stet/internal/build"
	"github.com/themaiby/stet/internal/grammar"
	"github.com/themaiby/stet/internal/ignore"
	"github.com/themaiby/stet/internal/registry"
	"github.com/themaiby/stet/internal/tool"
	"github.com/themaiby/stet/internal/valeconf"
	"github.com/themaiby/stet/internal/warmup"
)

// waitForWarmup bounds how long a lint run will sit behind a build another
// process started. Past this it reports rather than hangs.
const waitForWarmup = 3 * time.Minute

type lintFlags struct {
	Languages   string
	Preset      string
	Config      string
	Output      string
	ListPresets bool
	Fail        bool
	Targets     []string
}

func runLint(e *env, args []string) int {
	var f lintFlags
	f.Output = "line"

	for i := 0; i < len(args); {
		switch {
		case args[i] == "--list-presets":
			f.ListPresets = true
			i++
		case args[i] == "--fail":
			f.Fail = true
			i++
		case args[i] == "--":
			f.Targets = append(f.Targets, args[i+1:]...)
			i = len(args)
		default:
			matched := false
			for _, flag := range []struct {
				Name   string
				Target *string
			}{
				{"lang", &f.Languages}, {"preset", &f.Preset},
				{"config", &f.Config}, {"output", &f.Output},
			} {
				if value, used, ok := flagValue(args, i, flag.Name); ok {
					*flag.Target = value
					i += used
					matched = true
					break
				}
			}
			if !matched {
				f.Targets = append(f.Targets, args[i])
				i++
			}
		}
	}

	if f.ListPresets {
		return listPresets(e)
	}
	if len(f.Targets) == 0 {
		fmt.Fprintln(e.Err, usage)
		return 2
	}

	// Vale reads a path it cannot find as stdin and reports nothing wrong with
	// it, so a mistyped or unsplit argument comes back as a clean document.
	if missing := missingTargets(f.Targets); len(missing) > 0 {
		for _, target := range missing {
			fmt.Fprintf(e.Err, "stet: no such file or directory: %s\n", target)
		}
		fmt.Fprintln(e.Err, "stet: nothing was checked.")
		return 2
	}

	vale, err := tool.Resolve(tool.Vale, e.Layout, e.Client, e.Err)
	if err != nil {
		fmt.Fprintln(e.Err, err)
		return 1
	}

	state := settleWarmup(e, f.Languages)
	if state.Partial() {
		// A failed build must not pass as a clean report: the reader would
		// trust an empty result from a rule that checked nothing.
		fmt.Fprintf(e.Err, "stet: WARNING, %s\n", state.Message)
		fmt.Fprintln(e.Err, "stet: this report is PARTIAL. Run 'stet build' to see the error.")
	}

	languages, err := loadLanguages(e)
	if err != nil {
		fmt.Fprintln(e.Err, err)
		return 1
	}
	codes := requestedCodes(languages, f.Languages)

	config := f.Config
	if config == "" {
		if config, err = chooseConfig(e, f); err != nil {
			fmt.Fprintln(e.Err, err)
			return 1
		}
	}
	if _, err := os.Stat(config); err != nil {
		fmt.Fprintf(e.Err, "stet: no such config: %s\n", config)
		return 1
	}
	syncPackages(e, vale, config)

	// Reporting is the normal job, so findings do not fail the run. A gate asks
	// for --fail and gets Vale's own exit code, which an error trips and a
	// warning does not.
	valeArgs := []string{"--config=" + config, "--output=" + f.Output}
	if !f.Fail {
		valeArgs = append([]string{"--no-exit"}, valeArgs...)
	}
	if patterns, from := ignore.Load(f.Targets[0]); len(patterns) > 0 {
		valeArgs = append(valeArgs, "--glob="+patterns.ValeGlob())
		fmt.Fprintf(e.Err, "stet: %d path patterns ignored, from %s\n", len(patterns), from)
	}
	// Without --lang the config settles what the text is, whether the project
	// wrote it or this run generated it.
	if f.Languages == "" {
		if fromConfig := languagesInConfig(config, languages); len(fromConfig) > 0 {
			codes = fromConfig
		}
	}

	findings := checkGrammar(e, grammarRules(languages, codes), f.Targets)
	if patterns, _ := ignore.Load(f.Targets[0]); len(patterns) > 0 {
		kept := findings[:0]
		for _, finding := range findings {
			if !patterns.Match(finding.File) {
				kept = append(kept, finding)
			}
		}
		findings = kept
	}

	cmd := exec.Command(vale, append(valeArgs, f.Targets...)...)
	cmd.Stderr, cmd.Stdin = e.Err, os.Stdin
	var captured bytes.Buffer
	if len(findings) > 0 && strings.EqualFold(f.Output, "JSON") {
		cmd.Stdout = &captured
	} else {
		cmd.Stdout = e.Out
	}

	status := 0
	if err := cmd.Run(); err != nil {
		code, ok := exitCode(err)
		if !ok {
			fmt.Fprintln(e.Err, err)
			return 1
		}
		status = code
	}

	if len(findings) > 0 {
		if captured.Len() > 0 || strings.EqualFold(f.Output, "JSON") {
			merged, err := grammar.MergeJSON(captured.Bytes(), findings)
			if err != nil {
				fmt.Fprintln(e.Err, err)
				return 1
			}
			e.Out.Write(append(merged, '\n'))
		} else {
			fmt.Fprint(e.Out, grammar.Lines(findings))
		}
		if f.Fail && grammar.HasError(findings) {
			status = 1
		}
	}
	return status
}

// chooseConfig settles which policy applies: an explicit language, then a
// project's own .vale.ini above the target, then every registered language.
// requestedCodes resolves what --lang asked for into registered codes.
func requestedCodes(languages registry.Languages, requested string) []string {
	if requested == "" || requested == "all" {
		return languages.Codes()
	}
	codes := strings.Split(requested, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}
	return codes
}

func chooseConfig(e *env, f lintFlags) (string, error) {
	languages, err := loadLanguages(e)
	if err != nil {
		return "", err
	}

	if f.Languages == "" {
		// A preset has to be honoured or refused, never dropped because a
		// .vale.ini happened to sit above the target.
		if found, ok := findProjectConfig(f.Targets[0]); ok && f.Preset == "" {
			return found, nil
		}
		return generateConfig(e, languages, languages.Codes(), f.Preset)
	}

	codes := strings.Split(f.Languages, ",")
	if f.Languages == "all" {
		codes = languages.Codes()
	}
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}
	return generateConfig(e, languages, codes, f.Preset)
}

func generateConfig(e *env, languages registry.Languages, codes []string, presetName string) (string, error) {
	options := valeconf.Options{
		Header:     "Generated by stet. Edit languages.conf, not this file.",
		StylesPath: "../../styles",
	}
	for _, code := range codes {
		lang, ok := languages.Find(code)
		if !ok {
			return "", fmt.Errorf("stet: unknown language %q. Registered: %s",
				code, strings.Join(languages.Codes(), " "))
		}
		options.Languages = append(options.Languages, lang)
	}
	if english, ok := languages.Find("en"); ok {
		options.English = &english
	}

	if presetName != "" {
		preset, err := resolvePreset(e, codes, presetName)
		if err != nil {
			return "", err
		}
		options.Preset = preset
		options.PresetPolicy = presetPolicy(e, preset.Lang, preset.Code)
		// A language measured on one register has no base style to add.
		base := strings.ToUpper(preset.Lang) + "Base"
		if info, err := os.Stat(e.Layout.Style(base)); err == nil && info.IsDir() {
			options.BaseStyle = base
		}
	}

	if err := os.MkdirAll(e.Layout.ConfigCache(), 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(e.Layout.ConfigCache(), valeconf.CacheName(codes, presetName))
	return path, os.WriteFile(path, []byte(valeconf.Render(options)), 0o644)
}

// resolvePreset looks up by the pair, because a code such as "docs" exists
// under more than one language. A preset outside those languages is refused:
// its rules were measured on other text.
func resolvePreset(e *env, codes []string, name string) (*registry.Preset, error) {
	presets, err := loadPresets(e)
	if err != nil {
		return nil, err
	}
	for _, code := range codes {
		if preset, ok := presets.Find(code, name); ok {
			return &preset, nil
		}
	}
	if preset, ok := presets.FindByCode(name); ok {
		return nil, fmt.Errorf(
			"stet: preset %q was measured for %q, which --lang did not ask for.\n"+
				"stet: its rules would not match this text. Add --lang %s, or drop --preset.",
			name, preset.Lang, preset.Lang)
	}
	var lines []string
	for _, p := range presets {
		lines = append(lines, fmt.Sprintf("  %-10s %s", p.Code, p.Description))
	}
	return nil, fmt.Errorf("stet: unknown preset %q. Available:\n%s", name, strings.Join(lines, "\n"))
}

// presetPolicy reads what the chosen register turns off. A missing file means
// no register has an opinion, which is a fine state for a project to be in.
func presetPolicy(e *env, lang, code string) []string {
	file, err := os.Open(e.Layout.PresetPolicy())
	if err != nil {
		return nil
	}
	defer file.Close()
	policy, err := registry.ParsePresetPolicy(file)
	if err != nil {
		fmt.Fprintf(e.Err, "stet: cannot read the preset policy: %v\n", err)
		return nil
	}
	return policy.For(lang, code)
}

// missingTargets returns the paths that are not there. A linter that reports a
// clean document it never opened is worse than one that refuses to start.
func missingTargets(targets []string) []string {
	var missing []string
	for _, target := range targets {
		if _, err := os.Stat(target); err != nil {
			missing = append(missing, target)
		}
	}
	return missing
}

func findProjectConfig(target string) (string, bool) {
	dir, err := filepath.Abs(target)
	if err != nil {
		return "", false
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	for {
		candidate := filepath.Join(dir, ".vale.ini")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// syncPackages downloads the Vale packages a config names, once. The marker
// beside the config is what makes it once rather than every run.
func syncPackages(e *env, vale, config string) {
	data, err := os.ReadFile(config)
	if err != nil {
		return
	}
	text := string(data)
	if !strings.HasPrefix(text, "Packages") && !strings.Contains(text, "\nPackages") {
		return
	}
	// The marker belongs to this machine, not to the project, so it goes with
	// the rest of what stet downloads. It goes stale when the config changes its
	// package list, and a stale marker keeps the old rules in place with nothing
	// to show for it.
	sum := sha256.Sum256([]byte(config))
	marker := filepath.Join(e.Layout.Data, "synced", hex.EncodeToString(sum[:8]))
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		return
	}
	if synced, err := os.Stat(marker); err == nil {
		if written, err := os.Stat(config); err == nil && !written.ModTime().After(synced.ModTime()) {
			return
		}
	}
	cmd := exec.Command(vale, "--config="+config, "sync")
	if err := cmd.Run(); err == nil {
		os.WriteFile(marker, nil, 0o644)
	}
}

// settleWarmup reports the build rather than blocking on it in silence.
func settleWarmup(e *env, languages string) warmup.State {
	runner := build.New(e.Layout, e.Err)
	codes := warmupCodes(e, languages)

	switch state := runner.State(); state.Phase {
	case warmup.Cold:
		fmt.Fprintln(e.Err, "stet: first run, building rule data. This takes about 30 seconds.")
		runner.Run(codes)
	case warmup.Building:
		fmt.Fprintf(e.Err, "stet: warm-up running (%s). Waiting.\n", state.Message)
		deadline := time.Now().Add(waitForWarmup)
		for time.Now().Before(deadline) {
			time.Sleep(2 * time.Second)
			if runner.State().Phase != warmup.Building {
				break
			}
		}
	default:
		runner.Run(codes)
	}
	return runner.State()
}

func warmupCodes(e *env, languages string) []string {
	if languages != "" && languages != "all" {
		return strings.Split(languages, ",")
	}
	if registered, err := loadLanguages(e); err == nil {
		return registered.Codes()
	}
	return nil
}

func loadLanguages(e *env) (registry.Languages, error) {
	file, err := os.Open(e.Layout.Languages())
	if err != nil {
		return nil, fmt.Errorf("stet: cannot read the language registry: %w", err)
	}
	defer file.Close()
	return registry.ParseLanguages(file)
}

// loadPresets reads the preset registry, building it first on a fresh clone.
// The build reads committed data only and needs no network.
func loadPresets(e *env) (registry.Presets, error) {
	if _, err := os.Stat(e.Layout.Presets()); err != nil {
		runner := build.New(e.Layout, e.Err)
		for _, code := range []string{"uk", "en"} {
			runner.Run([]string{code})
		}
	}
	file, err := os.Open(e.Layout.Presets())
	if err != nil {
		return nil, fmt.Errorf("stet: cannot read the preset registry: %w", err)
	}
	defer file.Close()
	return registry.ParsePresets(file)
}

func listPresets(e *env) int {
	presets, err := loadPresets(e)
	if err != nil {
		fmt.Fprintln(e.Err, err)
		return 1
	}
	fmt.Fprintf(e.Out, "Available presets for --preset:\n\n")
	for _, p := range presets {
		note := ""
		if p.Preliminary {
			note = "  (preliminary)"
		}
		fmt.Fprintf(e.Out, "  %-10s %-4s %s%s\n", p.Code, p.Lang, p.Description, note)
	}
	fmt.Fprintf(e.Out, "\nA preset holds for the language it was measured on, in the second column.\n")
	fmt.Fprintf(e.Out, "A language without one is checked by its register-independent rules alone,\n")
	fmt.Fprintf(e.Out, "which is what running without --preset does.\n")
	return 0
}
