---
name: prose-policy
description: Change what the prose linter accepts in a project. Use when a finding is wrong, when a term should be allowed or banned from now on, when a rule should be demoted or switched off, or when the user asks why a finding keeps coming back.
---

# Prose policy

A rule states a claim about the text. What a project does about the claim is its
own decision, and this is where that decision gets written down.

Never edit the bundled rules to silence a finding. They are generated, so the
next build overwrites the edit, and the claim was never the problem.

## Where it goes

**Ask before writing a word:**

```bash
stet=$("${CLAUDE_PLUGIN_ROOT}"/scripts/bootstrap.sh)
"$stet" vocab PATH
```

A project scaffolded by `stet init` has its own vocabulary, under
`.vale/styles/config/vocabularies/Project/`. A project without one falls back to
the copy inside the plugin, which every other project on the machine shares and
`stet uninstall` deletes. Writing there is worse than not writing at all, so
when `vocab` reports that scope, invoke the `prose-setup` skill first.

`accept.txt` holds terms this project allows, `reject.txt` terms it bans that no
rule catches yet. Lines are regular expressions, so one line covers every
inflected form: `сорсинг.*` clears every ending at once.

## Which lever reaches the finding

In order of preference:

1. **The term is fine here.** One line in `accept.txt`.
2. **The class matters less here.** Demote it in the project's `.vale.ini`:
   `ProseUK.Morphology = suggestion`.
3. **The class does not apply.** Switch it off: `ProseCore.Typography = NO`.

**Lever 1 does not reach every rule, and reaching for it anyway looks like it
worked while changing nothing.** A rule matching below the word level never
reads the vocabulary, and that covers most punctuation rules and 36 of the 76 in
`ai-tells`. Ask:

```bash
"$stet" rule ai-tells.ShipOveruse
```

When it reports that `accept.txt` does not reach the rule, only levers 2 and 3
are left. **Both are the user's call, not yours**: say which rule you would
demote or disable and why, and let them decide. Silencing a class of findings is
a bigger decision than any single finding.

## Deciding

Per term, not per finding, because one term recurs with the same answer:

- **Allow** when it recurs, the project has settled on it, and no native word
  carries the same meaning: a product name, an API, a domain term, a spelling
  this project chose. Write the line as you decide, not later. A term you meant
  to keep and did not record comes back as a finding on every run.
- **Fix** when a native word says the same thing here. Change the text and
  record nothing.
- **Skip once** when the term is a quotation, an identifier, or the subject the
  sentence is discussing. Change nothing, write nothing. A one-off does not earn
  a line, and a vocabulary of one-offs stops meaning anything.

Preset rules carry their excess ratio in a comment beside the pattern. Near 2x
the signal is weak and the text may well be fine; above 8x it hardens, and
overriding it wants a reason. Read that number before deciding.

Report what you wrote, and where, with the reason. Policy the user never saw is
policy they cannot disagree with.
