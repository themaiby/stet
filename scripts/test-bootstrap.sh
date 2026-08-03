#!/usr/bin/env bash
# Checks how bootstrap.sh picks a binary. Two bugs have hidden here already, and
# nothing else reads this file.
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failed=0

# The script is run from a copy whose root holds no bin/, because a build in the
# working tree wins ahead of everything these cases are about.
mkdir -p "$work/root/scripts"
cp "$HERE/bootstrap.sh" "$work/root/scripts/bootstrap.sh"
: > "$work/root/languages.conf"
BOOTSTRAP="$work/root/scripts/bootstrap.sh"

check() {
  if [ "$2" = "$3" ]; then
    printf '  ok    %s\n' "$1"
  else
    printf '  FAIL  %s\n    want %s\n    got  %s\n' "$1" "$3" "$2"
    failed=1
  fi
}

stub() {
  mkdir -p "$(dirname "$1")"
  printf '#!/bin/sh\necho "stet %s"\n' "$2" > "$1"
  chmod +x "$1"
}

data="$work/data"
stub "$data/bin/stet" 1.0.0

got="$(STET_VERSION=1.0.0 CLAUDE_PLUGIN_DATA="$data" "$BOOTSTRAP" --quiet 2>/dev/null)"
check "a cached copy at the right version is used" "$got" "$data/bin/stet"

# A version behind the rules is worse than none: it reads them wrongly or not at
# all. The download that follows has no release to fetch, so only the removal is
# asserted here.
STET_VERSION=99.0.0 CLAUDE_PLUGIN_DATA="$data" "$BOOTSTRAP" --quiet >/dev/null 2>&1
present=present
[ -e "$data/bin/stet" ] || present=removed
check "a cached copy behind the rules is dropped" "$present" "removed"

path="$work/path"
stub "$path/stet" 2.0.0
empty="$work/empty"
mkdir -p "$empty"
got="$(PATH="$path:$PATH" STET_VERSION=1.0.0 CLAUDE_PLUGIN_DATA="$empty" "$BOOTSTRAP" --quiet 2>/dev/null)"
check "a binary somebody else installed is left alone" "$got" "$path/stet"

exit "$failed"
