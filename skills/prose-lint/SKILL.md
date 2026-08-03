---
name: prose-lint
description: Lint and clean up prose with Vale, in Ukrainian, English or both at once. Use when the user asks to check, lint, proofread or normalise text in documents, spreadsheets, CSV exports, tickets or docs, to strip loanwords and machine phrasing, or to set prose linting up for a project.
---

# Prose lint

Vale is your instrument, not a step that happens before you. You run it, and its
findings are your work list. It says where to look; you decide what the text
should say instead.

That division is the whole point. A linter that rewrote on its own would flatten
every borrowing it recognised, including the ones carrying a meaning no native
word carries. A rewrite done without it would wander into text nobody asked you
to touch.

So: **the findings set the scope, and judgement sets the fix.** Work the list.
Do not rewrite lines the linter did not flag, and do not apply a replacement
just because the glossary offers one.

An empty report means nothing to do inside that scope. The text may still read
badly, and finding that out is a different job nobody asked for. Say the scope
was clear and stop.

## Setup

Run this first. It prints the path to the binary and does nothing else:

```bash
stet=$("${CLAUDE_PLUGIN_ROOT}"/scripts/bootstrap.sh)
```

The script finds an existing `stet`, downloads a release for this platform, or
builds one from source when a Go toolchain is present. In Claude Code a session
hook has normally run it already, and a second run costs nothing. Use `"$stet"`
for every command below.

Everything else the tool needs, Vale included, it fetches itself on first use.
Nothing here asks to be installed by hand, and no step needs handing to the
user.

## Linting

```bash
"$stet" lint --lang uk --preset docs PATH...
```

`--lang` takes one code, a list, or `all`: `--lang uk`, `--lang uk,en`,
`--lang all`. Registered codes live in `languages.conf`; read it rather than
assuming which languages exist. Drop the flag when the project has its own
`.vale.ini`: the tool walks up from the file and uses that instead.

Pick the language from the text, not from the conversation. A document written
in one language with terms borrowed from another wants both codes.

### Ask for the style first

**Before linting anything, ask which register the text belongs to.** Get the
list from the tool rather than from memory:

```bash
"$stet" lint --list-presets
```

Put the options to the user as a short question and say what the choice changes.
The same construction is a machine tell in one register and ordinary writing in
another: an em dash runs 17x above human rate in Ukrainian technical
documentation and 1.9x in informal writing, so a rule that is right for a manual
is wrong for a chat reply. Choosing wrongly means correcting someone towards the
machine, which is worse than not checking at all.

The question can be skipped in two cases. If the user already named the style,
use it. If the document is unmistakable, say a `.po` file of software strings or
a signed official letter, name your reading in one line and proceed, so they can
correct you.

Without `--preset` only the register-independent rules run. That is the safe
fallback when nobody answers, not the default to reach for.

`--output=JSON` when you need to count or group; the default line format is for
reading.

Vale reads CSV and TSV directly and checks every field, including the ones with
newlines inside them.

Source files work too: Vale pulls the comments out and leaves the code alone.
Pass `.ts`, `.py`, `.go` and the rest the same way.

Comments are checked as English whatever `--lang` says, and a comment in another
language is itself a finding. Presets never apply there, since they were
measured on registers of prose. A project that writes comments in its own
language turns `ProseCore.CommentLanguage` off in its `.vale.ini`.

## Working the findings

Read `references/policy.md` first: names, identifiers and structure, the part
that holds in any language. Then `references/<code>/policy.md` for what that
language gets wrong. Replacements come from the findings themselves, since most
rules are substitutions and carry their own suggestion.

Where a `references/<code>/glossary.md` exists it holds judgement notes, not
word lists: the calls the rules cannot make. Read it, do not apply it
mechanically.

Then group `--output=JSON` by matched term rather than walking hit by hit. One
term recurs across a document with the same answer every time, so decide once
and apply everywhere. A large report is normally a short list of decisions.

Each term falls into one of three buckets:

- **Replace.** The finding suggests a native word that says the same thing and
  fits here. Apply it to every occurrence, adjusting the grammar around it.
- **Keep.** The borrowing carries a meaning or a brevity the native word loses,
  or the term is a name, an identifier or a quote. Note the reason rather than
  skipping in silence.
- **Allow.** It belongs to this project's working vocabulary, which is a policy
  decision rather than a wording one. Propose an `accept.txt` line and let the
  user confirm before writing it.

Structure comes first. When the text sits in a table, edit cell by cell and
check afterwards that the column count, the header and the line endings
survived. A document that lints clean but no longer parses is a failure.

Report all three buckets at the end, not just the replacements. What you chose
to keep is the part the user most needs to see, because that is where you
overrode the linter.

## Formatting, once the edits are in

When the file is markdown, by its name or by what is inside it, finish with:

```bash
"$stet" fmt PATH...
```

**Run this last, after every edit is in place.** Editing is what breaks the
layout: a replaced word overflows the line, a rewritten cell knocks a table out
of alignment, and a shortened sentence leaves the wrapping ragged. Formatting
first would only be undone by the next edit.

It changes layout and never words, so it cannot undo a decision made above. Use
`--check` to see whether anything would move without moving it.

## When the linter is wrong

It often will be, and that is the design. Reach for these in order:

1. **The term is fine in this project.** Add it to
   `.vale/styles/config/vocabularies/Project/accept.txt`. Lines are regular
   expressions, so one line covers every inflected form. Reach here first for a
   documented list of permitted terms.
2. **The whole class matters less here.** Demote it in `.vale.ini`:
   `ProseUK.Morphology = suggestion`.
3. **The class does not apply at all.** Switch it off:
   `ProseCore.Typography = NO`.

Each preset rule states its excess ratio in a comment beside the pattern. Near
2x the signal stays weak, above 8x it hardens. Judge by that number before
overriding a finding.

Never edit the bundled rules to silence a finding. A rule states a claim about
the text; the project decides what to do about the claim.

Punctuation rules work below the word level and ignore `accept.txt`. Only levers
2 and 3 reach them.

## Setting a project up

```bash
"$stet" init --lang uk,en /path/to/project
```

Copies the rules into `.vale/styles` and writes a starter `.vale.ini`. After
that the project lints with plain `vale` and owes this plugin nothing.

To add a language, follow the section in the plugin README.
