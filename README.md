# stet

*stet*, the proofreader's mark: let it stand. Which is what `accept.txt` does
for every finding this tool gets wrong.

Prose linting for Claude Code and Cowork, on top of [Vale](https://vale.sh).
Ukrainian and English out of the box.

Ask your assistant to check a document. It runs the linter, works the findings
with you, and tidies the markdown when the words are settled. Vale never
rewrites anything on its own: the findings set the scope, judgement sets the
fix.

## Install

```bash
claude plugin marketplace add https://github.com/themaiby/stet
claude plugin install stet
```

That is the whole installation. The plugin fetches its own binary, its own copy
of Vale and its own word lists the first time it runs, and keeps them in one
directory that `stet uninstall` empties again.

On Cowork, zip the repository and load it through Plugins, then Upload Plugin.

## Use

Ask for it in words. "Check this document", "proofread the README", "clean up
the Ukrainian in this export" all reach the same place.

The assistant will ask which register the text belongs to, because the answer
changes the result. An em dash runs 17x above human rate in Ukrainian technical
documentation and 1.9x in informal writing, so a rule that is right for a manual
is wrong for a chat reply.

It reads markdown, plain text, CSV and TSV, and it reads the comments out of
source files while leaving the code alone.

To keep a tree out of both the linter and the formatter, name it in a
`.stetignore` at the project root, one glob per line:

```text
vendor/
*.gen.md
```

## Disagreeing with the linter

Expected, and the reason the layers exist. In order of preference:

| the finding is                 | do this                           |
| ------------------------------ | --------------------------------- |
| a term this project allows     | add a line to `accept.txt`        |
| a class that matters less here | `ProseUK.Morphology = suggestion` |
| a class that does not apply    | `ProseCore.Typography = NO`       |

`accept.txt` takes regular expressions, so one line covers every inflected form:

```text
сорсинг.*
[Ff]eedback
```

Entries there reach every rule that matches whole words. They do not reach a
rule that matches below the word level, which is most of the punctuation rules
and half of `ai-tells`. Ask `stet rule <Style>.<Name>` which of the two a rule
is, and silence the second kind in `.vale.ini`.

Do not edit the bundled rules to quiet a finding. A rule states a claim about
the text. What to do about the claim is yours.

## How the rules got here

Nothing under `styles/` was typed from memory except the punctuation and script
mechanisms. Everything else is fetched from a maintained source or measured
against a human control corpus:

| source                                  | what it gives                                        |
| --------------------------------------- | ---------------------------------------------------- |
| LanguageTool                            | 9544 substitution pairs, 136 converted pattern rules |
| brown-uk/dict_uk                        | 4.07M word forms, so unknown words need no list      |
| brown-uk/corpus, UA-GEC, python-docs-uk | control corpora, 12M characters                      |
| `data/uk-excess.tsv`                    | measured excess per construction per register        |

A first set of Ukrainian rules was written from intuition and scored well on the
document behind it. Against a control corpus most of it collapsed, and two rules
turned out to be backwards: `не X, а Y` reads like a machine habit and runs at
1.1x, while `не просто X, а Y` is ten times commoner in human prose, so a rule
against it would have corrected people towards the machine.

Word lists also rot. `delve` dropped sharply once researchers named it a marker,
so every rule set carries a date and the packages are pinned rather than tracked
at `latest`. Structure decays slowest and deserves the most weight; vocabulary
decays fastest. Six months is a reasonable interval for a review.

## For anyone working on the plugin

```text
stet lint --lang uk --preset docs PATH...   check a document
stet lint --list-presets                    what registers exist
stet rule <Style>.<Name>                    which lever reaches a rule
stet vocab [PATH]                           which vocabulary a lint would read
stet fmt PATH...                            tidy markdown, after the edits
stet init --lang uk,en DIR                  give a project its own copy
stet doctor                                 what this machine can do
stet uninstall --dry-run                    what removing it would take
```

```text
languages.conf     the registry: code, style directory, packages
data/              measured excess ratios and the constructions behind them
styles/ProseCore/  language-neutral: dashes, spacing, mixed scripts
styles/ProseUK/    Ukrainian: dictionary inversion, morphology, generated pairs
styles/ProseEN/    English: generated pairs, ai-tells carries the rest
skills/prose-lint/ the skill and its per-language references
cmd/, internal/    the binary: generators, config, linting, warm-up
```

To add a language, create `styles/Prose<CODE>/` with one file per claim, add a
row to `languages.conf`, and write
`skills/prose-lint/references/<code>/policy.md` for what that language gets
wrong. Nothing in the plugin counts languages.

Leave `nonword` unset so tokens match whole words: that is what makes
`accept.txt` work, and it stops a rule for `skil` firing inside `skilful`. Regex
inside a token handles inflection, `sourcing\w*` covering every ending at once.
Reach for `nonword: true` only below the word level, for punctuation and script
mixing, and say in a comment that `accept.txt` cannot reach it.

Word lists are fetched from a source somebody else maintains, never assembled by
hand: a list typed from memory scores well only on the text behind it.
