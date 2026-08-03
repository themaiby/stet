#!/usr/bin/env bash
# Cuts measured constructions into UKBase and one style per register.
#
# Reads the committed measurement rather than the corpora, so this needs no
# network and no 15M of text. Regenerate the measurement with
# tools/compare-registers.py --tsv when the corpora change.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${1:-$ROOT/styles}"

# The output ships with the repository. Rebuilding is only for when the
# measurement changes, and needs an interpreter the runtime does not.
if [ -f "$ROOT/presets.conf" ] && [ "$ROOT/presets.conf" -nt "$ROOT/data/uk-excess.tsv" ]; then
  exit 0
fi

command -v python3 >/dev/null 2>&1 || {
  echo "stet: python3 not found, keeping the preset rules that shipped" >&2
  exit 0
}

python3 "$ROOT/generators/lib/presets.py" \
  "$ROOT/data/uk-patterns.tsv" \
  "$ROOT/data/uk-excess.tsv" \
  "$OUT" UK
