#!/usr/bin/env bash
# Writes uk_UA.dic and uk_UA.aff into $1, from brown-uk/dict_uk.
#
# Forms are expanded up front instead of using the upstream affix file: Vale
# compiles Hunspell conditions to Go regexes, and 647 of those rules need a
# lookbehind, which Go has not got.
set -euo pipefail

OUT="$1"
VERSION="${UK_DICT_VERSION:-6.8.0}"
URL="https://github.com/brown-uk/dict_uk/releases/download/v${VERSION}/dict_corp_vis.txt.bz2"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "stet: fetching Ukrainian dictionary ${VERSION}" >&2
curl -fsSL "$URL" -o "$tmp/corp.txt.bz2"
if command -v bunzip2 >/dev/null 2>&1; then
  bunzip2 -f "$tmp/corp.txt.bz2"
else
  python3 -c "import bz2,shutil,sys; shutil.copyfileobj(bz2.open(sys.argv[1]), open(sys.argv[2],'wb'))" \
    "$tmp/corp.txt.bz2" "$tmp/corp.txt"
fi

# Forms live on indented lines under their lemma; strip the indent or lose them.
# Emitted under all three apostrophes in use, else spelling fails on punctuation.
sed 's/^ *//' "$tmp/corp.txt" | cut -d' ' -f1 \
  | sed "s/'/\xca\xbc/g" > "$tmp/raw.txt"
{
  cat "$tmp/raw.txt"
  sed "s/\xca\xbc/'/g" "$tmp/raw.txt"
  sed "s/\xca\xbc/\xe2\x80\x99/g" "$tmp/raw.txt"
} | sort -u > "$tmp/forms.txt"

mkdir -p "$OUT"
{ wc -l < "$tmp/forms.txt"; cat "$tmp/forms.txt"; } > "$OUT/uk_UA.dic"
printf 'SET UTF-8\n' > "$OUT/uk_UA.aff"

# lemma<TAB>form form form ... , so the XML converter can expand inflected
# tokens without downloading this corpus a second time.
awk '{
  if ($0 ~ /^ /) { sub(/^ +/, ""); split($0, f, " "); if (lemma != "") forms = forms " " f[1] }
  else { if (lemma != "") print lemma "\t" forms; split($0, f, " "); lemma = f[1]; forms = f[1] }
} END { if (lemma != "") print lemma "\t" forms }' "$tmp/corp.txt" \
  | awk -F'\t' '!seen[$1]++' > "$OUT/uk_lemmas.tsv"

printf 'stet: %s lemmas indexed\n' "$(wc -l < "$OUT/uk_lemmas.tsv" | tr -d ' ')" >&2

printf 'stet: %s word forms\n' "$(wc -l < "$tmp/forms.txt" | tr -d ' ')" >&2
