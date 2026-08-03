#!/usr/bin/env bash
# Removes everything this plugin downloaded or generated: uninstall.sh [--dry-run]
#
# Nothing outside the plugin directory and its own data directory is touched.
# Removing the plugin itself is the package manager's job.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(dirname "$HERE")"
DATA="${CLAUDE_PLUGIN_DATA:-${XDG_DATA_HOME:-$HOME/.local/share}/stet}"

DRY=0
[ "${1:-}" = "--dry-run" ] && DRY=1

targets=(
  "$DATA"
  "$ROOT/configs/.cache"
  "$ROOT/styles/config/dictionaries"
  "$ROOT/styles/ai-tells"
  "$ROOT/styles/write-good"
  "$ROOT/styles/ProseUK/Barbarism.yml"
  "$ROOT/styles/ProseUK/Preferred.yml"
  "$ROOT/styles/ProseUK/Calque.yml"
  "$ROOT/styles/ProseEN/Plain.yml"
  "$ROOT/presets.conf"
)
for d in "$ROOT"/styles/UK* "$ROOT"/styles/EN*; do
  [ -d "$d" ] && targets+=("$d")
done

freed=0
for t in "${targets[@]}"; do
  [ -e "$t" ] || continue
  size="$(du -sk "$t" 2>/dev/null | cut -f1)"
  freed=$((freed + ${size:-0}))
  if [ "$DRY" -eq 1 ]; then
    printf '  would remove  %s\n' "${t#$ROOT/}"
  else
    rm -rf "$t"
    printf '  removed  %s\n' "${t#$ROOT/}"
  fi
done

if [ "$freed" -eq 0 ]; then
  echo "stet: nothing to remove."
else
  printf 'stet: %s %s MB. Run the linter again to rebuild it.\n' \
    "$([ "$DRY" -eq 1 ] && echo 'would free' || echo 'freed')" "$((freed / 1024))"
fi
