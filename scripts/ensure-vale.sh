#!/usr/bin/env bash
# Prints a path to a usable Vale, downloading one only if the machine has none.
set -euo pipefail

VALE_VERSION="${VALE_VERSION:-3.16.0}"
DATA_DIR="${CLAUDE_PLUGIN_DATA:-${XDG_DATA_HOME:-$HOME/.local/share}/stet}"
BIN_DIR="$DATA_DIR/bin"
QUIET=0

for arg in "$@"; do
  case "$arg" in
    --quiet) QUIET=1 ;;
  esac
done

say() { [ "$QUIET" -eq 1 ] || printf '%s\n' "$*" >&2; }

if command -v vale >/dev/null 2>&1; then
  say "vale: $(command -v vale) ($(vale --version | awk '{print $NF}'))"
  command -v vale
  exit 0
fi

for cached in "$BIN_DIR/vale" "$BIN_DIR/vale.exe"; do
  if [ -x "$cached" ]; then
    say "vale: $cached (already downloaded)"
    printf '%s\n' "$cached"
    exit 0
  fi
done

os="$(uname -s)"
arch="$(uname -m)"

case "$os:$arch" in
  Linux:x86_64)          asset="vale_${VALE_VERSION}_Linux_64-bit.tar.gz" ;;
  Linux:aarch64|Linux:arm64) asset="vale_${VALE_VERSION}_Linux_arm64.tar.gz" ;;
  Darwin:x86_64)         asset="vale_${VALE_VERSION}_macOS_64-bit.tar.gz" ;;
  Darwin:arm64)          asset="vale_${VALE_VERSION}_macOS_arm64.tar.gz" ;;
  MINGW*|MSYS*|CYGWIN*) asset="vale_${VALE_VERSION}_Windows_64-bit.zip"; binary="vale.exe" ;;
  *)
    echo "stet: no prebuilt Vale for $os/$arch. Install Vale and keep it on PATH." >&2
    exit 1
    ;;
esac
binary="${binary:-vale}"

url="https://github.com/vale-cli/vale/releases/download/v${VALE_VERSION}/${asset}"
say "downloading $asset"

mkdir -p "$BIN_DIR"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

if ! curl -fsSL "$url" -o "$tmp/archive"; then
  echo "stet: download failed: $url" >&2
  exit 1
fi

case "$asset" in
  *.zip)
    # Windows ships a zip, and the tar in Git Bash cannot read one. Python is
    # already required, and zipfile is part of it.
    python3 -c "
import zipfile, shutil, sys
z = zipfile.ZipFile(sys.argv[1])
name = next(n.filename for n in z.infolist() if n.filename.endswith(sys.argv[3]))
with z.open(name) as src, open(sys.argv[2], 'wb') as dst:
    shutil.copyfileobj(src, dst)
" "$tmp/archive" "$BIN_DIR/$binary" "$binary" ;;
  *)
    tar -xzf "$tmp/archive" -C "$tmp" "$binary"
    mv "$tmp/$binary" "$BIN_DIR/$binary" ;;
esac
chmod +x "$BIN_DIR/$binary"

say "vale: $BIN_DIR/$binary ($VALE_VERSION)"
printf '%s\n' "$BIN_DIR/$binary"
