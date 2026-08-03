package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/themaiby/stet/internal/valeconf"
)

// generated names the artefacts init does not copy. The dictionary alone is
// 96 MB, and a copy would rot with no way to rebuild it, so the project rebuilds
// them in place instead.
var generated = []string{
	"config/dictionaries",
	"ai-tells",
	"write-good",
	"ProseUK/Barbarism.yml",
	"ProseUK/Preferred.yml",
	"ProseUK/Calque.yml",
	"ProseEN/Plain.yml",
}

// runInit leaves the project self-contained: after this it lints with plain
// vale, and stet is needed only to refresh the generated rules.
func runInit(e *env, args []string) int {
	codes := "all"
	force := false
	target := "."

	for i := 0; i < len(args); {
		if args[i] == "--force" {
			force = true
			i++
			continue
		}
		if value, used, ok := flagValue(args, i, "lang"); ok {
			codes = value
			i += used
			continue
		}
		target = args[i]
		i++
	}

	languages, err := loadLanguages(e)
	if err != nil {
		fmt.Fprintln(e.Err, err)
		return 1
	}
	requested := languages.Codes()
	if codes != "all" {
		requested = strings.Split(codes, ",")
	}

	options := valeconf.Options{StylesPath: ".vale/styles"}
	for _, code := range requested {
		lang, ok := languages.Find(strings.TrimSpace(code))
		if !ok {
			fmt.Fprintf(e.Err, "stet: unknown language %q. Registered: %s\n",
				code, strings.Join(languages.Codes(), " "))
			return 1
		}
		options.Languages = append(options.Languages, lang)
	}
	if english, ok := languages.Find("en"); ok {
		options.English = &english
	}

	config := filepath.Join(target, ".vale.ini")
	if _, err := os.Stat(config); err == nil && !force {
		fmt.Fprintf(e.Err, "%s already exists. Re-run with --force to replace it.\n", config)
		return 1
	}

	if err := copyStyles(e.Layout.Styles(), filepath.Join(target, ".vale", "styles")); err != nil {
		fmt.Fprintln(e.Err, err)
		return 1
	}
	if err := os.WriteFile(config, []byte(valeconf.Render(options)), 0o644); err != nil {
		fmt.Fprintln(e.Err, err)
		return 1
	}

	fmt.Fprintf(e.Err, `Scaffolded in %s  (languages: %s)

  .vale.ini                                     policy: which styles, which severity
  .vale/styles/                                 the rules themselves
  .vale/styles/config/vocabularies/Project/     accept.txt: terms this project allows

Tune the policy, not the rules:
  ProseUK.Morphology = suggestion               demote
  ProseCore.Typography = NO                     switch off
  echo 'сорсинг.*' >> .vale/styles/config/vocabularies/Project/accept.txt

Generated data is not copied. Build it here:
  STET_STYLES=.vale/styles stet build

Packages, if any, are fetched by: vale sync
`, target, strings.Join(requested, ","))
	return 0
}

func copyStyles(from, to string) error {
	return filepath.WalkDir(from, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return os.MkdirAll(to, 0o755)
		}
		if isGenerated(relative) {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		destination := filepath.Join(to, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		return copyFile(path, destination)
	})
}

func isGenerated(relative string) bool {
	slashed := filepath.ToSlash(relative)
	for _, name := range generated {
		if slashed == name || strings.HasPrefix(slashed, name+"/") {
			return true
		}
	}
	// A register style is generated too, and its name comes from the generator.
	return isPresetStyle(strings.SplitN(slashed, "/", 2)[0])
}

func copyFile(from, to string) error {
	source, err := os.Open(from)
	if err != nil {
		return err
	}
	defer source.Close()
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	destination, err := os.Create(to)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}
