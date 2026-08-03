#!/usr/bin/env bash
# init-project.sh [--lang CODES] [--force] [TARGET_DIR]
#
# Leaves the project self-contained: after this it lints with plain vale.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(dirname "$HERE")"
REGISTRY="$ROOT/languages.conf"

LANGS="all"
FORCE=0
TARGET="$PWD"

while [ $# -gt 0 ]; do
  case "$1" in
    --lang)  LANGS="$2"; shift 2 ;;
    --lang=*) LANGS="${1#*=}"; shift ;;
    --force) FORCE=1; shift ;;
    *)       TARGET="$1"; shift ;;
  esac
done

registered_codes() { grep -v '^[[:space:]]*#' "$REGISTRY" | grep -v '^[[:space:]]*$' | cut -d'|' -f1; }

[ "$LANGS" = "all" ] && LANGS="$(registered_codes | paste -sd, -)"

styles="ProseCore"
packages=""
policy=""

# Comments are written in English by convention, so the code section takes the
# English row whatever --lang asked for.
en_row="$(grep -v '^[[:space:]]*#' "$REGISTRY" | awk -F'|' '$1=="en" {print; exit}')"
code_styles="ProseCore$([ -n "$en_row" ] && echo ", $(echo "$en_row" | cut -d'|' -f2)")"
en_packages="$(echo "$en_row" | cut -d'|' -f3)"
en_policy="$(echo "$en_row" | cut -d'|' -f4)"
for code in $(echo "$LANGS" | tr ',' ' '); do
  row="$(grep -v '^[[:space:]]*#' "$REGISTRY" | awk -F'|' -v c="$code" '$1==c {print; exit}')"
  if [ -z "$row" ]; then
    echo "init-project: unknown language '$code'. Registered: $(registered_codes | tr '\n' ' ')" >&2
    exit 1
  fi
  styles="$styles, $(echo "$row" | cut -d'|' -f2)"
  packages="$packages $(echo "$row" | cut -d'|' -f3)"
  policy="$policy$(echo "$row" | cut -d'|' -f4);"
done

pkg_styles=""; code_pkg_styles=""
for p in $packages; do pkg_styles="$pkg_styles, $(basename "$p" .zip)"; done
for p in $en_packages; do
  code_pkg_styles="$code_pkg_styles, $(basename "$p" .zip)"
  case " $packages " in *" $p "*) ;; *) packages="$packages $p" ;; esac
done

if [ -f "$TARGET/.vale.ini" ] && [ "$FORCE" -eq 0 ]; then
  echo "$TARGET/.vale.ini already exists. Re-run with --force to replace it." >&2
  exit 1
fi

mkdir -p "$TARGET/.vale/styles"
# Generated artefacts are not copied: the dictionary alone is 96 MB, and a copy
# would rot with no way to rebuild it. The builders come along instead.
rsync -a --exclude 'config/dictionaries' \
         --exclude 'ai-tells' --exclude 'write-good' \
         --exclude 'ProseUK/Barbarism.yml' --exclude 'ProseUK/Preferred.yml' \
         --exclude 'ProseUK/Calque.yml' --exclude 'ProseEN/Plain.yml' \
         "$ROOT/styles/" "$TARGET/.vale/styles/" 2>/dev/null \
  || cp -R "$ROOT/styles/." "$TARGET/.vale/styles/"

cp -R "$ROOT/dictionaries" "$ROOT/generators" "$TARGET/.vale/"
cp "$ROOT/languages.conf" "$TARGET/.vale/"
sed -e 's|^ROOT=.*|ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" \&\& pwd)"|' \
    -e 's|\$ROOT/styles|$ROOT/styles|' \
    "$ROOT/scripts/ensure-data.sh" > "$TARGET/.vale/refresh.sh"
chmod +x "$TARGET/.vale/refresh.sh"

{
  echo "StylesPath = .vale/styles"
  echo "MinAlertLevel = suggestion"
  echo "Vocab = Project"
  if [ -n "${packages// /}" ]; then
    echo "Packages = $(echo $packages | tr ' ' ',' | sed 's/,/, /g')"
  fi
  echo
  echo "[formats]"
  echo "csv = txt"
  echo "tsv = txt"
  echo "ts = js"
  echo "tsx = js"
  echo
  echo "[*.{md,txt,csv,tsv,html,rst,adoc}]"
  echo "BasedOnStyles = ${styles}${pkg_styles}"
  echo "ProseCore.CommentLanguage = NO"
  [ -n "${packages// /}" ] && echo "Vale.Terms = NO"
  echo "$policy" | tr ';' '\n' | sed '/^[[:space:]]*$/d;s/^[[:space:]]*//'

  # Comments only. English rules whatever --lang asked for, and no preset:
  # presets were measured on registers of prose, which a comment is not.
  echo
  echo "[*.{ts,tsx,js,jsx,mjs,cjs,go,py,rb,java,kt,cs,php,rs,c,h,cpp,hpp,swift,scala,lua}]"
  echo "BasedOnStyles = ${code_styles}${code_pkg_styles}"
  # Measured on a 240-file TypeScript project: of 547 findings, 545 came from
  # doc-comment formatting rather than the prose. Vale leaves the leading "*"
  # of a JSDoc block in the text and lints code inside @example as prose, so
  # everything keyed to punctuation, spacing or line structure misfires.
  # What survives are the rules that read word sequences.
  echo "ProseCore.Formatting = NO"
  echo "ProseCore.Typography = NO"
  if [ -n "${en_packages// /}" ]; then
    echo "Vale.Terms = NO"
    echo "write-good.TooWordy = NO"
    echo "write-good.Weasel = NO"
    echo "write-good.Passive = NO"
    echo "write-good.E-Prime = NO"
    echo "ai-tells.FormalRegister = NO"
    echo "ai-tells.SemicolonUsage = NO"
    echo "ai-tells.ColonUsage = NO"
    echo "ai-tells.EmDashUsage = NO"
    echo "ai-tells.VerbTricolon = NO"
  fi
} > "$TARGET/.vale.ini"

cat >&2 <<EOF
Scaffolded in $TARGET  (languages: $LANGS)

  .vale.ini                                     policy: which styles, which severity
  .vale/styles/                                 the rules themselves
  .vale/styles/config/vocabularies/Project/     accept.txt: terms this project allows

Tune the policy, not the rules:
  ProseUK.Morphology = suggestion               demote
  ProseCore.Typography = NO                     switch off
  echo 'сорсинг.*' >> .vale/styles/config/vocabularies/Project/accept.txt

Generated data is not copied. Build it here:
  .vale/refresh.sh all

Packages, if any, are fetched by: vale sync
EOF
