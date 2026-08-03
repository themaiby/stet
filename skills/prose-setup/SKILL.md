---
name: prose-setup
description: Set prose linting up for a project, or change what it covers. Use when a project carries no Vale config yet, when the user asks to add prose checking to a repository, to exclude vendored or generated paths, or to add a language.
---

# Prose setup

Setting a project up leaves it self-contained: after this it lints with plain
`vale` and needs the plugin only to refresh the generated rules.

```bash
stet=$("${CLAUDE_PLUGIN_ROOT}"/scripts/bootstrap.sh)
"$stet" init --lang uk,en /path/to/project
```

That writes a starter `.vale.ini` and copies the rules into `.vale/styles`,
which is also what gives the project a vocabulary of its own. Without it, terms
the linter is told to accept go to the plugin's shared copy and disappear with
`stet uninstall`.

Pick the languages from what the project writes, not from the language of its
code. A codebase with Ukrainian documentation and English comments wants both.

Generated data is not copied, because the dictionary alone is 96 MB and a copy
would rot with no way to rebuild it. The project rebuilds it in place:

```bash
STET_STYLES=.vale/styles "$stet" build
```

## Keeping paths out

A project carries trees it did not write: vendored code, generated files,
fixtures, anything whose prose belongs to somebody else. Name them in a
`.stetignore` at the project root, one glob per line, a trailing slash covering
a directory:

```text
# Vendored and generated trees this project does not own.
vendor/
node_modules/
*.gen.md
CHANGELOG.md
```

Both the linter and the formatter obey it, and both obey it even when a path is
named on the command line. Write the reason as a comment: a bare list of paths
tells the next reader nothing about why any of them is there.

Reach for this when a whole tree is out of scope. When a single rule is wrong
rather than a whole path, that is the `prose-policy` skill instead.

## Grammar rules for a language

The fifth field of a `languages.conf` row names the grammar rules that run
beside Vale for that language. Only English has any, and a rule belongs there
when it catches a plain mistake: agreement, confused words, repetition. Anything
about house style stays with Vale, which knows the register.

Adding one is measured against the project's own writing before it goes in. Of
the checker's 823 rules, the set in use reports twice on this repository and
both are real; the full set reports 64, most of them about heading case and
spelling out numbers.

## What the machine can do

```bash
"$stet" doctor
```

One line per capability rather than a single verdict, because most of what can
be missing costs a rule set rather than the tool. Run it before blaming the
configuration for a missing finding.

## Adding a language

Create `styles/Prose<CODE>/` with one file per claim, add a row to
`languages.conf`, and write `skills/prose-lint/references/<code>/policy.md` for
what that language gets wrong. Nothing in the plugin counts languages.

Leave `nonword` unset so tokens match whole words. That is what lets a project
exempt a term, and it stops a rule for `skil` firing inside `skilful`. Regex
inside a token handles inflection, `sourcing\w*` covering every ending at once.
Reach for `nonword: true` only below the word level, for punctuation and script
mixing, and say in a comment that the vocabulary cannot reach it.

Word lists are fetched from a source somebody else maintains, never assembled by
hand: a list typed from memory scores well only on the text behind it.

## Removing it

```bash
"$stet" uninstall --dry-run
"$stet" uninstall
```

Deletes the downloaded binaries, the word lists, the generated styles and the
config cache, around 300 MB. Nothing outside the plugin directory and its own
data directory is touched, and the next lint rebuilds whatever it needs.
