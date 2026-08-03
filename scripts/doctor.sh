#!/usr/bin/env bash
# Reports what this environment can do: doctor.sh
#
# Prints one line per capability rather than a single pass or fail, because most
# of what is missing costs a rule set, not the tool.
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(dirname "$HERE")"

have() { command -v "$1" >/dev/null 2>&1; }
line() { printf '  %-22s %s\n' "$1" "$2"; }

printf 'stet on %s %s\n\n' "$(uname -s)" "$(uname -m)"

if have vale; then
  line "Vale" "on PATH, $(vale --version | awk '{print $NF}')"
elif [ -x "${CLAUDE_PLUGIN_DATA:-${XDG_DATA_HOME:-$HOME/.local/share}/stet}/bin/vale" ]; then
  line "Vale" "downloaded earlier"
elif have curl && have tar; then
  line "Vale" "will download on first use"
else
  line "Vale" "MISSING, and curl or tar is absent to fetch it"
fi

have curl && line "curl" "yes" || line "curl" "MISSING, nothing can be fetched"
have tar  && line "tar"  "yes" || line "tar"  "MISSING, archives cannot be opened"

if have bunzip2 || have python3; then
  line "Ukrainian dictionary" "buildable"
else
  line "Ukrainian dictionary" "MISSING bunzip2 and python3, ProseUK.NotUkrainian stays off"
fi

if have python3; then
  line "LanguageTool patterns" "buildable, 136 rules"
else
  line "LanguageTool patterns" "needs python3, the other 9700 rules are unaffected"
fi

[ -f "$ROOT/presets.conf" ] \
  && line "presets" "$(grep -vc '^#' "$ROOT/presets.conf") shipped, no build needed" \
  || line "presets" "registry missing"

case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*)
    printf '\n  Windows: this path is reasoned from the platform, not tested.\n'
    printf '  WSL takes the Linux path and is the safer route.\n' ;;
esac
