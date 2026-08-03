#!/usr/bin/env bash
# Prints a path to a usable stet, producing one if the machine has none.
#
# This is the only script left. Everything it used to do now lives in the
# binary, and the binary has to arrive somehow: a release asset when there is
# one for this platform, a local build when there is a Go toolchain, and a clear
# refusal otherwise.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(dirname "$HERE")"
DATA="${CLAUDE_PLUGIN_DATA:-${XDG_DATA_HOME:-$HOME/.local/share}/stet}"
VERSION="${STET_VERSION:-0.4.0}" # x-release-please-version
QUIET=0

for arg in "$@"; do
  case "$arg" in
    --quiet) QUIET=1 ;;
  esac
done
say() { [ "$QUIET" -eq 1 ] || printf '%s\n' "$*" >&2; }

# The binary may live in the data directory, nowhere near the rules, and a
# caller outside Claude Code has no CLAUDE_PLUGIN_ROOT to point it back here.
mkdir -p "$DATA"
printf '%s\n' "$ROOT" > "$DATA/root"

for candidate in "$ROOT/bin/stet" "$DATA/bin/stet" "$DATA/bin/stet.exe"; do
  if [ -x "$candidate" ]; then
    say "stet: $candidate"
    printf '%s\n' "$candidate"
    exit 0
  fi
done
if command -v stet >/dev/null 2>&1; then
  say "stet: $(command -v stet)"
  command -v stet
  exit 0
fi

os="$(uname -s)"
arch="$(uname -m)"
case "$os:$arch" in
  Linux:x86_64)              asset="stet_${VERSION}_linux_amd64.tar.gz" ;;
  Linux:aarch64|Linux:arm64) asset="stet_${VERSION}_linux_arm64.tar.gz" ;;
  Darwin:x86_64)             asset="stet_${VERSION}_darwin_amd64.tar.gz" ;;
  Darwin:arm64)              asset="stet_${VERSION}_darwin_arm64.tar.gz" ;;
  MINGW*|MSYS*|CYGWIN*)      asset="stet_${VERSION}_windows_amd64.zip" ;;
  *)                         asset="" ;;
esac

if [ -n "$asset" ] && command -v curl >/dev/null 2>&1; then
  url="https://github.com/themaiby/stet/releases/download/v${VERSION}/${asset}"
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  if curl -fsSL "$url" -o "$tmp/archive" 2>/dev/null; then
    mkdir -p "$DATA/bin"
    case "$asset" in
      *.zip) unzip -q -o "$tmp/archive" stet.exe -d "$DATA/bin" 2>/dev/null || true ;;
      *)     tar -xzf "$tmp/archive" -C "$DATA/bin" stet 2>/dev/null || true ;;
    esac
    for candidate in "$DATA/bin/stet" "$DATA/bin/stet.exe"; do
      if [ -f "$candidate" ]; then
        chmod +x "$candidate"
        say "stet: $candidate ($VERSION, downloaded)"
        printf '%s\n' "$candidate"
        exit 0
      fi
    done
  fi
fi

if command -v go >/dev/null 2>&1; then
  say "stet: no release for this platform, building from source"
  ( cd "$ROOT" && go build -ldflags "-X github.com/themaiby/stet/internal/cli.Version=$VERSION" -o bin/stet ./cmd/stet )
  say "stet: $ROOT/bin/stet (built here)"
  printf '%s\n' "$ROOT/bin/stet"
  exit 0
fi

cat >&2 <<EOF
stet: no binary for $os/$arch and no Go toolchain to build one.

Either install Go and re-run this script, or fetch a release binary from
https://github.com/themaiby/stet/releases and put it on PATH.
EOF
exit 1
