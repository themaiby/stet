# English: what to replace

Read `../policy.md` first. This file adds what is specific to English.

English has no transliterated-jargon problem. Its failure modes are inflation,
padding and the house style of language models.

## Latinate inflation

The long word standing where a short one works: `utilize` for `use`, `commence`
for `start`, `facilitate` for `help`, `ascertain` for `find out`.
`ProseEN.Plain` swaps these directly.

## Padding

Phrases that occupy space without adding meaning: `in order to`,
`due to the
fact that`, `at this point in time`, `it is worth noting that`. Cut
or shorten. `write-good.TooWordy` and `ai-tells.FillerPhrases` cover most of it.

## Machine phrasing

The register a language model reaches for by default: `delve into`,
`at its
core`, `a testament to`, `not just X, but Y`, `move the needle`. The
`ai-tells` package finds these. Rewrite rather than swap: the fix is to state
the point the phrase stood in for.

## Adjectives that promise more than they carry

`vibrant`, `pivotal`, `seamless`, `groundbreaking`, `transformative`,
`comprehensive` about your own output. Cut the adjective and state the fact
behind it.

## Hedging and weasels

`arguably`, `somewhat`, `a number of`, `many experts say`. Either name the
number and the source, or drop the qualifier. `write-good.Weasel` finds them.

## What English keeps

Passive voice where the actor does not matter, which in documentation is often.
`write-good.Passive` comes set to `suggestion` for that reason, and many
projects switch it off entirely.

Semicolons and colons. They carry structure, and the packages flag them only
because models overuse them.
