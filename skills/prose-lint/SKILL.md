---
name: prose-lint
description: Check, proofread and clean up prose with Vale, in Ukrainian, English or both at once. Use when the user asks to lint, check, proofread or tidy text in documents, spreadsheets, CSV exports, tickets or source comments, or to strip loanwords and machine phrasing from it.
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

## The run

```bash
stet=$("${CLAUDE_PLUGIN_ROOT}"/scripts/bootstrap.sh)
"$stet" lint --lang uk --preset docs PATH...
```

The first line prints the path to the binary and does nothing else. Everything
the tool needs, Vale included, it fetches itself. In Claude Code a session hook
has normally run it already, and a second run costs nothing.

`--lang` takes one code, a list, or `all`. Registered codes live in
`languages.conf`; read it rather than assuming which languages exist. Drop the
flag when the project has its own `.vale.ini`, which the tool finds by walking
up from the file. Pick the language from the text rather than from the
conversation: a document written in one language with terms borrowed from
another wants both codes.

`--output=JSON` when you need to count or group. `--fail` returns a non-zero
status when the text carries an error, which is what a gate wants and a report
does not.

Vale reads CSV and TSV directly, every field, newlines inside them included.
Source files work too: it pulls the comments out and leaves the code alone, so
`.ts`, `.py` and `.go` go in the same way. Comments are checked as English
whatever `--lang` says, and one written in another language is itself a finding.

A `.stetignore` at the project root holds the paths this and the formatter both
skip. A finding you expected and did not get is worth checking against it.

**Under `--lang en` alone, a grammar checker runs beside Vale**, and its
findings arrive named `harper.<rule>` in the same report. It parses the
sentence, so it reaches subject-verb agreement, `its` against `it's`, `their`
against `there` and repeated words, none of which pattern matching can see. It
stays off for every other language combination: an English grammar checker has
nothing to say about a Ukrainian document, and a mixed one is a document in
another language.

### Ask for the register first

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

Skip the question when the user already named the register, or when the document
is unmistakable, say a `.po` file of software strings or a signed official
letter. Then name your reading in one line so they can correct you.

Without `--preset` only the register-independent rules run. That is the safe
fallback when nobody answers, not the default to reach for.

## Working the findings

Read `references/policy.md` first: names, identifiers and structure, the part
that holds in any language. Then `references/<code>/policy.md` for what that
language gets wrong. Replacements come from the findings themselves, since most
rules are substitutions and carry their own suggestion.

Where a `references/<code>/glossary.md` exists it holds judgement notes rather
than word lists: the calls the rules cannot make. Read it, do not apply it
mechanically.

Group `--output=JSON` by matched term rather than walking hit by hit. One term
recurs across a document with the same answer every time, so decide once and
apply everywhere. A large report is normally a short list of decisions.

Each term falls into one of three buckets:

- **Replace.** The finding suggests a native word that says the same thing and
  fits here. Apply it to every occurrence, adjusting the grammar around it.
- **Keep.** The borrowing carries a meaning or a brevity the native word loses,
  or the term is a name, an identifier or a quote. Note the reason rather than
  skipping in silence.
- **Allow.** It belongs to this project's working vocabulary, which is a policy
  decision rather than a wording one. **Invoke the `prose-policy` skill**, which
  owns where that gets written and which findings it can reach.

Severity sets how much doubt you may leave. An `error` gets fixed or allowed,
never skipped. A `warning` or `suggestion` may stand when the text reads better
unchanged, and letting one stand is a decision you report.

Structure comes first. When the text sits in a table, edit cell by cell and
check afterwards that the column count, the header and the line endings
survived. A document that lints clean but no longer parses is a failure.

Report all three buckets at the end, not just the replacements. What you chose
to keep is the part the user most needs to see, because that is where you
overrode the linter.

When a finding is simply wrong, or a whole class of them is, that is policy
rather than wording. Invoke the `prose-policy` skill.

## Formatting, last

When the file is markdown, by its name or by what is inside it, finish with:

```bash
"$stet" fmt PATH...
```

**Run this after every edit is in place.** Editing is what breaks the layout: a
replaced word overflows the line, a rewritten cell knocks a table out of
alignment, a shortened sentence leaves the wrapping ragged. Formatting first
would only be undone by the next edit. It moves layout and never words, so it
cannot undo a decision made above. `--check` reports what would move without
moving it.

## When the project has no setup

A project with no `.vale.ini` of its own, or one that needs paths kept out:
invoke the `prose-setup` skill.
