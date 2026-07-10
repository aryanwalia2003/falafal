#!/usr/bin/env bash
# Installs the latest falafal release for Linux or macOS: downloads the
# matching binary and puts it in ~/.local/bin (creating it and adding it
# to PATH in your shell rc file if needed).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/aryanwalia2003/falafal/main/install.sh | bash

set -euo pipefail

REPO="aryanwalia2003/falafal"
INSTALL_DIR="$HOME/.local/bin"

os=$(uname -s)
arch=$(uname -m)

case "$os" in
  Linux) platform="linux" ;;
  Darwin) platform="darwin" ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) goarch="amd64" ;;
  arm64|aarch64) goarch="arm64" ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

echo "Fetching latest falafal release info..."
release_json=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest")
asset_url=$(echo "$release_json" | grep -o "\"browser_download_url\": *\"[^\"]*${platform}_${goarch}\.tar\.gz\"" | grep -o 'https://[^"]*')
tag=$(echo "$release_json" | grep -o '"tag_name": *"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$asset_url" ]; then
  echo "Could not find a ${platform}_${goarch} release asset." >&2
  exit 1
fi

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

echo "Downloading $tag for ${platform}_${goarch}..."
curl -fsSL "$asset_url" -o "$tmpdir/falafal.tar.gz"
tar -xzf "$tmpdir/falafal.tar.gz" -C "$tmpdir"

mkdir -p "$INSTALL_DIR"
binary=$(find "$tmpdir" -type f -name falafal | head -1)
cp "$binary" "$INSTALL_DIR/falafal"
chmod +x "$INSTALL_DIR/falafal"

echo "falafal $tag installed to $INSTALL_DIR/falafal"

case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    echo "Run: falafal --version"
    ;;
  *)
    shell_rc="$HOME/.bashrc"
    [ -n "${ZSH_VERSION:-}" ] && shell_rc="$HOME/.zshrc"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$shell_rc"
    echo "Added $INSTALL_DIR to PATH in $shell_rc."
    echo "Restart your terminal (or run: source $shell_rc), then run: falafal --version"
    ;;
esac
