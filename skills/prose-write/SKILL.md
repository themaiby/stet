---
name: prose-write
description: Write or rewrite prose a person will read, free of machine phrasing. Use before drafting a README, release notes, a report, documentation, a ticket, a commit message, a pull request body or a source comment, in Ukrainian or English. Apply while drafting, then run the check it ends with.
---

# Prose write

The linter reads what you produced. This is about producing less for it to find,
and about the part it can never reach.

Hold these while drafting. Do not draft first and repair after: a sentence
written to be checked reads differently from one repaired into passing.

## What no rule reaches

Nothing below can be caught by matching text, so nothing else will catch it.

- **Write what exists.** A rejected alternative, an absent feature, a plan for
  later: none of it belongs in the text. State what is. A prohibition earns a
  sentence when the prohibition is itself the rule.
- **The text never describes itself.** No `this document covers`, no
  `in one
  place`, no `as mentioned above`. The title and the structure carry
  that already.
- **Point instead of repeating.** Content that lives somewhere else gets a
  reference. A summary beside the original goes stale on the day the original
  changes.
- **Real numbers, real names, real dates.** Never invent a figure to sound
  precise. "Around 30 seconds" beats a fabricated "28 seconds", and a
  measurement you have beats both.
- **Name the specific thing**, whether that is a tool, an endpoint or a config
  key. A sentence that works for any project describes none.
- **Templates over borrowed examples**, so `<type>/<slug>` and not a real branch
  name someone will copy.

## What the rules do reach, worth holding anyway

Sentence length varies, and three sentences of the same length in a row happen
to nobody writing by hand. A list of three appears when the content holds three
items, and a position gets stated once without hedging from both sides
afterwards. Em dashes stay out. An opener that announces the text before it
starts carries no content and goes.

Register decides the rest, and the same construction is a machine tell in one
register and ordinary writing in another. When the register is unclear, ask
before drafting: the `prose-lint` skill lists what exists.

## Then the check runs

**When the text reaches a file, run the full cycle. Always, without exception.**
Writing carefully lowers what the linter finds; it never lowers it to nothing,
and the parts above are exactly the parts no linter reports.

Invoke the `prose-lint` skill, which lints, works the findings and formats last.
Do not run the tool by hand instead: the skill holds the register question, the
three ways of answering a finding, and the ordering that keeps formatting after
the edits.

Text that never reaches a file, a chat reply or a commit message typed inline,
still gets everything above. It simply has nothing to run the tool against.
