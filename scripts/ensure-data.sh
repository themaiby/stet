#!/usr/bin/env bash
# Builds generated data and records progress: ensure-data.sh [--detach|--status] [CODES]
#
# A language opts in via dictionaries/<code>.sh and generators/<code>-*.sh.
# Sources are fetched at run time, never vendored: that keeps other projects'
# licences out of this repository and the rules current.
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(dirname "$HERE")"
REGISTRY="$ROOT/languages.conf"
DEST="${STET_STYLES:-$ROOT/styles}/config/dictionaries"
DATA="${CLAUDE_PLUGIN_DATA:-${XDG_DATA_HOME:-$HOME/.local/share}/stet}"
STATE="$DATA/warmup.state"
LOG="$DATA/warmup.log"

MODE=run
CODES=all
while [ $# -gt 0 ]; do
  case "$1" in
    --detach) MODE=detach; shift ;;
    --status) MODE=status; shift ;;
    *)        CODES="$1"; shift ;;
  esac
done

mkdir -p "$DATA"
state_write() { printf '%s\n' "$1" > "$STATE"; }

if [ "$MODE" = status ]; then
  if [ -f "$STATE" ]; then cat "$STATE"; else echo "cold|0|0|nothing built yet"; fi
  exit 0
fi

registered_codes() {
  grep -v '^[[:space:]]*#' "$REGISTRY" | grep -v '^[[:space:]]*$' | cut -d'|' -f1
}
[ "$CODES" = "all" ] && CODES="$(registered_codes | paste -sd, -)"

if [ "$MODE" = detach ]; then
  # A lock left by a killed run must not wedge the warm-up forever.
  if [ -f "$DATA/warmup.pid" ] && kill -0 "$(cat "$DATA/warmup.pid" 2>/dev/null)" 2>/dev/null; then
    exit 0
  fi
  nohup "$0" "$CODES" >>"$LOG" 2>&1 &
  echo $! > "$DATA/warmup.pid"
  exit 0
fi

tasks=()
for code in $(echo "$CODES" | tr ',' ' '); do
  for builder in "$ROOT/dictionaries/${code}.sh" "$ROOT"/generators/"${code}"-*.sh; do
    [ -x "$builder" ] || continue
    stamp="$DEST/.$(basename "$builder" .sh).built"
    if [ -f "$stamp" ] && [ "$builder" -ot "$stamp" ]; then continue; fi
    tasks+=("$builder")
  done
done

total=${#tasks[@]}
if [ "$total" -eq 0 ]; then
  state_write "ready|0|0|already built"
  exit 0
fi

mkdir -p "$DEST"
i=0
for builder in "${tasks[@]}"; do
  i=$((i + 1))
  name="$(basename "$builder" .sh)"
  case "$builder" in
    */dictionaries/*) target="$DEST"; hint="word forms, the slow one, around 20s" ;;
    *)                target="";      hint="rule pairs, a second or two" ;;
  esac

  state_write "building|$i|$total|$name ($hint)"
  printf '[%s] %s/%s building %s\n' "$(date +%H:%M:%S)" "$i" "$total" "$name" >&2

  if "$builder" $target; then
    date > "$DEST/.${name}.built"
  else
    # Silence here would let a rule check nothing at all, and the report would
    # read as clean because it was blind. Fail loudly instead.
    state_write "failed|$i|$total|$name failed; rules from it are inactive"
    echo "stet: FAILED to build $name. Rules that depend on it will not run." >&2
    exit 1
  fi
done

state_write "ready|$total|$total|built $total of $total"
