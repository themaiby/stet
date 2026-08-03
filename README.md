# stet

*stet*, the proofreader's mark: let it stand. Which is what `accept.txt` does for
every finding this tool gets wrong.

Prose linting for Claude Code and Cowork, on top of [Vale](https://vale.sh).
Ukrainian and English out of the box, one directory and one row per further
language.

The skill runs Vale and works its findings; Vale never rewrites anything on its
own. The findings set the scope, judgement sets the fix.

The plugin provides rules. The project keeps the policy. Nothing here decides for
you which findings matter.

## Install

**Claude Code**

```bash
claude plugin marketplace add https://github.com/themaiby/stet
claude plugin install stet
```

A session hook resolves Vale and warms the generated data in the background: an
existing binary if the machine has one, a download if not.

**Cowork**

Zip the repository and load it through Plugins, then the plus button, then
Upload Plugin. Cowork does not run session hooks, so the skill warms up on first
use instead.

## Use

Ask for it in words, or call the wrapper:

```bash
scripts/prose-lint.sh --list-presets
scripts/prose-lint.sh --lang uk --preset docs  manual.md
scripts/prose-lint.sh --lang uk --preset press article.md
scripts/prose-lint.sh --lang uk,en             export.csv
```

`--preset` picks the register: `docs`, `official`, `press`, `academic`,
`informal`, `fiction`, and `docs` for English. A preset holds for the language behind
its measurement, and asking for one under another language is refused rather
than silently ignored. The choice changes results. An em dash runs 17x above
human rate in Ukrainian technical documentation and 1.9x in informal writing, so
one rule cannot serve both. Ratios come from `data/uk-excess.tsv`, measured
against human corpora rather than chosen.

Vale reads CSV and TSV directly, fields with embedded newlines included. Source
files work too: it extracts the comments and leaves the code alone, so `.ts`,
`.py` and `.go` go in the same way. Comments are checked as English whatever
`--lang` says, and one written in another language is itself a finding. Turn
`ProseCore.CommentLanguage` off if your project disagrees.

To give a project its own copy:

```bash
scripts/init-project.sh --lang uk,en /path/to/project
```

That writes `.vale.ini` and `.vale/styles/`, after which the project lints with
plain `vale` and needs nothing from the plugin.

## Disagreeing with the linter

Expected, and the reason the layers exist. In order of preference:

| the finding is | do this |
| --- | --- |
| a term this project allows | add a line to `accept.txt` |
| a class that matters less here | `ProseUK.Morphology = suggestion` |
| a class that does not apply | `ProseCore.Typography = NO` |

`accept.txt` takes regular expressions, so one line covers every inflected form:

```text
сорсинг.*
[Ff]eedback
```

Entries there become exceptions for every word-matching rule, in every language.
Punctuation rules work below the word level and ignore the file; silence those in
`.vale.ini`.

Do not edit the bundled rules to quiet a finding. A rule states a claim about the
text. What to do about the claim is yours.

## Layout

```text
languages.conf     the registry: code, style directory, packages
data/              measured excess ratios and the constructions behind them
styles/ProseCore/  language-neutral: dashes, spacing, mixed scripts
styles/ProseUK/    Ukrainian: dictionary inversion, morphology, generated pairs
styles/ProseEN/    English: generated pairs, ai-tells carries the rest
generators/        fetch upstream rule data and emit Vale styles
dictionaries/      fetch upstream word lists for the spelling rules
skills/prose-lint/ the skill and its per-language references
scripts/           ensure-vale, ensure-data, prose-lint, init-project, doctor, uninstall
```

## How the rules got here

Nothing under `styles/` was typed from memory except the punctuation and script
mechanisms. Everything else is fetched from a maintained source or measured
against a human control corpus:

| source | what it gives |
| --- | --- |
| LanguageTool | 9544 substitution pairs, 136 converted pattern rules |
| brown-uk/dict_uk | 4.07M word forms, so unknown words need no list |
| brown-uk/corpus, UA-GEC, python-docs-uk | control corpora, 12M characters |
| `data/uk-excess.tsv` | measured excess per construction per register |

### Why the numbers matter

A first set of Ukrainian rules was written from intuition and scored well on the
document behind it. Against a control corpus most of it collapsed, and two rules
turned out to be backwards: `не X, а Y` reads like a machine habit and runs at
1.1x, while `не просто X, а Y` is ten times commoner in human prose, so a rule
against it would have corrected people towards the machine.

Register decides as much as language. An em dash runs 755x above human rate in
English documentation and 1.9x in informal Ukrainian; a colon followed by a
capital is a marker in Ukrainian and three times commoner in human English
documentation than in machine output. One threshold cannot serve both.

Word lists also rot. `delve` dropped sharply once researchers named it a marker, so
every rule set carries a date and the packages are pinned rather than tracked at
`latest`. Structure decays slowest and deserves the most weight; vocabulary
decays fastest. Six months is a reasonable interval for a review.

## Adding a language

Create `styles/Prose<CODE>/` with one file per claim, add a row to
`languages.conf`, and write `skills/prose-lint/references/<code>/policy.md` for
what that language gets wrong. Nothing in the plugin counts languages.

Leave `nonword` unset so tokens match whole words: that is what makes
`accept.txt` work, and it stops a rule for `skil` firing inside `skilful`. Regex
inside a token handles inflection, `sourcing\w*` covering every ending at once.
Reach for `nonword: true` only below the word level, for punctuation and script
mixing, and say in a comment that `accept.txt` cannot reach it.

Word lists belong in `generators/<code>-*.sh`, fetched from a source somebody
else maintains. A list assembled by hand scores well only on the text behind it.

## Removing it

```bash
scripts/uninstall.sh --dry-run
scripts/uninstall.sh
```

Deletes the downloaded Vale, the word lists, the generated styles and the config
cache, around 260 MB. Nothing outside the plugin directory and its own data
directory is touched, and the next lint rebuilds whatever it needs.

## Requirements

Run `scripts/doctor.sh` and it will tell you what this machine can do. Nothing
below stops the tool outright except the first row.

| needed for | tool | absent |
| --- | --- | --- |
| everything | `bash`, `curl`, `tar` | nothing runs |
| the linter | Vale | downloaded on first use, 39 MB |
| 9700 substitution rules | none beyond the above | always available |
| the Ukrainian dictionary | `bunzip2` or `python3` | `ProseUK.NotUkrainian` stays off |
| 136 pattern rules | `python3` | those rules stay off |
| presets | none | they arrive built |

Presets used to be generated and now travel with the repository, because they
come from measurements this project owns rather than from anyone else's data.
That leaves `python3` optional: without it you lose 136 rules out of roughly ten
thousand.

| platform | state |
| --- | --- |
| macOS, arm64 and x86-64 | verified, including a download from an empty machine |
| Linux, arm64 and x86-64 | expected to work; `rsync` and `bunzip2` have fallbacks |
| Windows through WSL | reports itself as Linux and takes the Linux path |
| Windows through Git Bash | implemented, not tested; Vale arrives as a zip |

A shell is needed either way, so PowerShell alone will not do.
