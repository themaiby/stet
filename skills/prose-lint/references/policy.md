# Policy, the part that holds in any language

What each language does with its own borrowings and its own padding lives in
`references/<code>/policy.md`. This file is what stays true regardless.

## Never touch

**Product and tool names.** LinkedIn, Claude, n8n, Notion, Looker, Figma. A name
is a name.

**Settled technical terms with no short native form.** Boolean, ATS, STAR, OKR,
API, SQL. Translating these makes the text heavier, not lighter.

**Identifiers.** Column headers, field names, keys, section or competency titles
that other documents point at. Translating an identifier breaks the reference. A
string cited elsewhere as a name counts as an identifier.

**Quotes and reported speech.** The person said it the way they said it.

**Code, paths, commands and their output.**

## Structure before prose

When the text sits in a table, a spreadsheet or an export, edit cell by cell and
verify afterwards that the column count, the header and the line endings
survived. A document that lints clean but no longer parses is a failure.

## Judgement calls

The shorter form wins when it loses no meaning. When the native equivalent runs
longer and vaguer than the borrowing, keep the borrowing and say so out loud.

A term that has become part of the team's working language belongs in
`accept.txt`, not in an edit. That is a policy decision, so surface it rather
than making it silently.

Fix what the linter found, not what it did not. Rewriting a sentence that no
rule touched is scope you were not given.

## Typography

Decorative dashes go; where grammar requires a dash, it stays. Normalise
apostrophes and quotation marks to one character each across the document.
Collapse double spaces.
