#!/usr/bin/env bash
# Generates ProseUK/Calque.yml from the LanguageTool Ukrainian pattern rules.
#
# These are the multi-word rules that replace/soft lists cannot express. Rules
# needing a part-of-speech tagger are skipped and counted, since Vale matches
# text. Inflected tokens are expanded through the lemma index that
# dictionaries/uk.sh leaves behind, so this must run after it.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${1:-$ROOT/styles/ProseUK}"
LEMMAS="${STET_STYLES:-$ROOT/styles}/config/dictionaries/uk_lemmas.tsv"
REF="${LT_REF:-master}"
BASE="https://raw.githubusercontent.com/languagetool-org/languagetool/${REF}/languagetool-language-modules/uk/src/main/resources/org/languagetool/rules/uk"

if ! command -v python3 >/dev/null 2>&1; then
  echo "stet: python3 not found, skipping the pattern-rule converter" >&2
  exit 1
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "stet: fetching LanguageTool pattern rules (${REF})" >&2
curl -fsSL "$BASE/grammar-barbarism.xml" -o "$tmp/barbarism.xml"
curl -fsSL "$BASE/grammar-style.xml" -o "$tmp/style.xml"

mkdir -p "$OUT"
python3 "$ROOT/generators/lib/lt_xml.py" \
  "$OUT/Calque.yml" \
  "Калька або стилістична хиба: '%s'" \
  warning \
  "$LEMMAS" \
  "$tmp/barbarism.xml" "$tmp/style.xml"
