#!/usr/bin/env bash
# Cuts measured English constructions into one style per register.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${1:-$ROOT/styles}"

# The output ships with the repository. Rebuilding is only for when the
# measurement changes, and needs an interpreter the runtime does not.
if [ -f "$ROOT/presets.conf" ] && [ "$ROOT/presets.conf" -nt "$ROOT/data/en-excess.tsv" ]; then
  exit 0
fi

command -v python3 >/dev/null 2>&1 || {
  echo "stet: python3 not found, keeping the preset rules that shipped" >&2
  exit 0
}

python3 "$ROOT/generators/lib/presets.py" \
  "$ROOT/data/en-patterns.tsv" \
  "$ROOT/data/en-excess.tsv" \
  "$OUT" EN
