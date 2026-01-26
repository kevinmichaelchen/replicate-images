#!/bin/bash
set -euo pipefail

REPO="kevinmichaelchen/replicate-images"
BINARY="replicate-images"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  darwin|linux|freebsd) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Get latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release" >&2
  exit 1
fi

# Download URL
URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}_${OS}_${ARCH}.tar.gz"

echo "Downloading ${BINARY} ${LATEST} for ${OS}/${ARCH}..."

# Create temp directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download and extract
curl -fsSL "$URL" | tar -xz -C "$TMP_DIR"

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
else
  echo "Installing to $INSTALL_DIR (requires sudo)..."
  sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
fi

echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
echo "Run '${BINARY} --help' to get started"
