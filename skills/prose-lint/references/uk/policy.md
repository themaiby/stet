# Ukrainian: what to replace

Read `../policy.md` first. This file adds what is specific to Ukrainian.

## The thick layer: transliterated jargon

The dominant problem in Ukrainian workplace writing, and the biggest win.
English stems carried over with Ukrainian endings, where a native verb already
exists: `аутріч`, `челенджити`, `пітчити`, `фолоапити`, `овнити`, `пушити`,
`шерити`, `апрувити`.

Replace these by default. `ProseUK.Anglicism` finds them; `glossary.md` says what
to write instead.

## Whole English sentences left in

`Never gives up`, `Creates momentum`, `Designs interview frameworks` sitting
inside Ukrainian text. Not terminology, just an unfinished translation. Translate
them fully.

## Metric names

`reply rate`, `time-to-fill`, `bottleneck`. A reader without English has to know
what is being measured, so spell it out: частка відповідей, час закриття
вакансії, вузьке місце.

## Bureaucratese

The Soviet-era register: `з метою` for `щоб`, `у зв'язку з тим що` for `бо`,
`здійснювати` for `робити`, `являється` for `є`. `ProseUK.Bureaucratese` covers
it.

## Apostrophe

Ukrainian uses `ʼ` (U+02BC). Documents drift between that, the ASCII `'` and the
typographic `’`, often inside one file. Normalise to U+02BC everywhere.
`ProseUK.Apostrophe` finds the strays.

## Homoglyphs

Text pasted between editors picks up Latin `c`, `e`, `o`, `p`, `x` inside
Cyrillic words. Invisible to a reader, fatal to search and sorting.
`ProseCore.MixedScript` finds it.

## What Ukrainian keeps

Job grades in their settled form: Junior, Middle, Senior, Lead. Established
borrowings with no shorter native form: `конверсія`, `воронка`, `таргетована
реклама`. Long-naturalised words are not the target; recent transliterations are.
