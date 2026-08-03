#!/usr/bin/env bash
# prose-lint.sh [--lang CODES] [--preset NAME] [--config PATH] [--output FMT] PATH...
#   prose-lint.sh --list-presets
#
# Without --lang, a .vale.ini above the target wins, then every registered language.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(dirname "$HERE")"
REGISTRY="$ROOT/languages.conf"
CACHE="$ROOT/configs/.cache"

LANGS=""
PRESET=""
CONFIG=""
OUTPUT="line"
ARGS=()

while [ $# -gt 0 ]; do
  case "$1" in
    --list-presets)
      # The registry is generated and absent on a fresh clone. Building it reads
      # committed data only, so it costs a moment and needs no network.
      [ -f "$ROOT/presets.conf" ] || for g in "$ROOT"/generators/*-presets.sh; do
        [ -x "$g" ] && "$g" >/dev/null 2>&1 || true
      done
      printf 'Available presets for --preset:\n\n'
      grep -v '^[[:space:]]*#' "$ROOT/presets.conf" 2>/dev/null \
        | awk -F'|' '{printf "  %-10s %-4s %s%s\n", $2, $1, $4, ($5=="preliminary" ? "  (preliminary)" : "")}'
      printf '\nA preset holds for the language it was measured on, in the second column.\n'
      printf 'A language without one is checked by its register-independent rules alone,\n'
      printf 'which is what running without --preset does.\n'
      exit 0 ;;
    --preset) PRESET="$2"; shift 2 ;;
    --preset=*) PRESET="${1#*=}"; shift ;;
    --lang)   LANGS="$2"; shift 2 ;;
    --lang=*) LANGS="${1#*=}"; shift ;;
    --config) CONFIG="$2"; shift 2 ;;
    --config=*) CONFIG="${1#*=}"; shift ;;
    --output) OUTPUT="$2"; shift 2 ;;
    --output=*) OUTPUT="${1#*=}"; shift ;;
    --) shift; ARGS+=("$@"); break ;;
    *) ARGS+=("$1"); shift ;;
  esac
done

if [ "${#ARGS[@]}" -eq 0 ]; then
  echo "usage: prose-lint.sh [--lang CODES] [--preset NAME] [--config PATH] [--output FMT] PATH..." >&2
  echo "       prose-lint.sh --list-presets" >&2
  exit 2
fi

registered_codes() { grep -v '^[[:space:]]*#' "$REGISTRY" | grep -v '^[[:space:]]*$' | cut -d'|' -f1; }

row_for() { grep -v '^[[:space:]]*#' "$REGISTRY" | awk -F'|' -v c="$1" '$1==c {print; exit}'; }

generate_config() {
  local codes="$1" styles="ProseCore" packages="" policy="" code row base_style=""
  # Comments are written in English by convention, so the code section takes the
  # English row whatever --lang asked for. Applying another language's rules
  # there would quietly bless comments that should not be in that language.
  local en_row en_packages="" code_styles="ProseCore"
  local en_policy=""
  en_row="$(row_for en)"
  if [ -n "$en_row" ]; then
    code_styles="$code_styles, $(echo "$en_row" | cut -d'|' -f2)"
    en_packages="$(echo "$en_row" | cut -d'|' -f3)"
    en_policy="$(echo "$en_row" | cut -d'|' -f4)"
  fi
  if [ -n "$PRESET" ]; then
    local prow
    [ -f "$ROOT/presets.conf" ] || for g in "$ROOT"/generators/*-presets.sh; do
      [ -x "$g" ] && "$g" >/dev/null 2>&1 || true
    done
    # A code such as "docs" exists for several languages, so the row is found by
    # the pair. Falling back to the first match picked the wrong language.
    prow=""
    for plang_try in $codes; do
      prow="$(grep -v '^[[:space:]]*#' "$ROOT/presets.conf" 2>/dev/null \
        | awk -F'|' -v p="$PRESET" -v l="$plang_try" '$1==l && $2==p {print; exit}')"
      [ -n "$prow" ] && break
    done
    if [ -z "$prow" ]; then
      prow="$(grep -v '^[[:space:]]*#' "$ROOT/presets.conf" 2>/dev/null | awk -F'|' -v p="$PRESET" '$2==p {print; exit}')"
    fi
    if [ -n "$prow" ]; then
      local plang; plang="$(echo "$prow" | cut -d'|' -f1)"
      case " $codes " in
        *" $plang "*) ;;
        *)
          echo "stet: preset '$PRESET' was measured for '$plang', which --lang did not ask for." >&2
          echo "stet: its rules would not match this text. Add --lang $plang, or drop --preset." >&2
          exit 1 ;;
      esac
    fi
    if [ -z "$prow" ]; then
      echo "stet: unknown preset '$PRESET'. Available:" >&2
      grep -v '^[[:space:]]*#' "$ROOT/presets.conf" 2>/dev/null | awk -F'|' '{printf "  %-10s %s\n", $2, $4}' >&2
      exit 1
    fi
    base_style="$(echo "$prow" | cut -d'|' -f1 | tr 'a-z' 'A-Z')Base"
    # A language measured on one register has no base style to add.
    [ -d "$ROOT/styles/$base_style" ] || base_style=""
    styles="$styles${base_style:+, $base_style}, $(echo "$prow" | cut -d'|' -f3)"
  fi
  for code in $codes; do
    row="$(row_for "$code")"
    if [ -z "$row" ]; then
      echo "stet: unknown language '$code'. Registered: $(registered_codes | tr '\n' ' ')" >&2
      exit 1
    fi
    styles="$styles, $(echo "$row" | cut -d'|' -f2)"
    packages="$packages $(echo "$row" | cut -d'|' -f3)"
    policy="$policy$(echo "$row" | cut -d'|' -f4);"
  done

  local pkg_styles="" code_pkg_styles="" p
  for p in $packages; do
    pkg_styles="$pkg_styles, $(basename "$p" .zip)"
  done
  for p in $en_packages; do
    code_pkg_styles="$code_pkg_styles, $(basename "$p" .zip)"
    case " $packages " in *" $p "*) ;; *) packages="$packages $p" ;; esac
  done

  mkdir -p "$CACHE"
  local out="$CACHE/$(echo "$codes" | tr ' ' '-')${PRESET:+-$PRESET}.ini"

  {
    echo "# Generated by prose-lint.sh. Edit languages.conf, not this file."
    echo "StylesPath = ../../styles"
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
    echo "mts = js"
    echo "cts = js"
    echo
    echo "[*.{md,txt,csv,tsv,html,rst,adoc}]"
    echo "BasedOnStyles = ${styles}${pkg_styles}"
    echo "ProseCore.CommentLanguage = NO"
    if [ -n "${packages// /}" ]; then
      echo "Vale.Terms = NO"
      # ai-tells owns the dash where it is loaded; keeping ours counts it twice.
      echo "ProseCore.Typography = NO"
    fi
    echo "$policy" | tr ';' '\n' | sed '/^[[:space:]]*$/d;s/^[[:space:]]*//'

    # Comments only: Vale pulls them out and leaves the code alone. English
    # rules apply whatever --lang asked for, and presets do not apply at all,
    # since they were measured on registers of prose. Comments are terse by
    # design, which is why the wordiness rules come off too.
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
  } > "$out"

  printf '%s\n' "$out"
}

find_project_config() {
  local dir
  dir="$(cd "$(dirname "${ARGS[0]}")" 2>/dev/null && pwd || pwd)"
  while [ "$dir" != "/" ]; do
    if [ -f "$dir/.vale.ini" ]; then printf '%s\n' "$dir/.vale.ini"; return 0; fi
    dir="$(dirname "$dir")"
  done
  return 1
}

VALE="$("$HERE/ensure-vale.sh" --quiet)"

# Report the warm-up rather than blocking on it in silence, and never let a
# failed build pass as a clean report: the reader would trust an empty result
# from a rule that checked nothing.
state_line() { "$HERE/ensure-data.sh" --status; }

raw="$(state_line)"
phase="${raw%%|*}"

case "$phase" in
  cold)
    echo "stet: first run, building rule data. This takes about 30 seconds." >&2
    "$HERE/ensure-data.sh" "${LANGS:-all}" || true
    raw="$(state_line)"; phase="${raw%%|*}"
    ;;
  building)
    echo "stet: warm-up running (${raw##*|}). Waiting." >&2
    waited=0
    while [ "$waited" -lt 180 ]; do
      sleep 2; waited=$((waited + 2))
      raw="$(state_line)"; phase="${raw%%|*}"
      [ "$phase" = building ] || break
      [ $((waited % 20)) -eq 0 ] && echo "stet: still building (${raw##*|})" >&2
    done
    ;;
  *)
    "$HERE/ensure-data.sh" "${LANGS:-all}" || true
    raw="$(state_line)"; phase="${raw%%|*}"
    ;;
esac

if [ "$phase" = failed ]; then
  echo "stet: WARNING, ${raw##*|}" >&2
  echo "stet: this report is PARTIAL. Run scripts/ensure-data.sh to see the error." >&2
fi

if [ -z "$CONFIG" ]; then
  if [ -n "$LANGS" ]; then
    [ "$LANGS" = "all" ] && LANGS="$(registered_codes | paste -sd, -)"
    CONFIG="$(generate_config "$(echo "$LANGS" | tr ',' ' ')")"
  elif CONFIG="$(find_project_config)"; then
    :
  else
    CONFIG="$(generate_config "$(registered_codes | tr '\n' ' ')")"
  fi
fi

[ -f "$CONFIG" ] || { echo "stet: no such config: $CONFIG" >&2; exit 1; }

if grep -q '^Packages' "$CONFIG" 2>/dev/null; then
  marker="$(dirname "$CONFIG")/.synced-$(basename "$CONFIG")"
  if [ ! -f "$marker" ]; then
    "$VALE" --config="$CONFIG" sync >/dev/null 2>&1 && touch "$marker" || true
  fi
fi

exec "$VALE" --no-exit --config="$CONFIG" --output="$OUTPUT" "${ARGS[@]}"
