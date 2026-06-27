#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-latest}"
INSTALL_DIR="${HOME}/.local/bin"
REPO="HolyPrapor/memento"

need_cmd() { command -v "$1" &>/dev/null || { echo "Missing: $1" >&2; exit 1; }; }
need_cmd curl
need_cmd tar

case "$(uname -s)" in
    Linux)  os="linux" ;;
    Darwin) os="darwin" ;;
    *)      echo "Unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
    aarch64|arm64) arch="arm64" ;;
    x86_64|amd64)  arch="amd64" ;;
    *)             echo "Unsupported arch: $(uname -m)" >&2; exit 1 ;;
esac

ext=".tar.gz"

if [ "$VERSION" = "latest" ]; then
    release=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest")
else
    tag="$VERSION"
    [[ "$tag" != v* ]] && tag="v$tag"
    release=$(curl -s "https://api.github.com/repos/${REPO}/releases/tags/${tag}")
fi

url=$(echo "$release" | grep -o "\"browser_download_url\": *\"[^\"]*${os}_${arch}${ext}[^\"]*\"" | head -1 | grep -o '"[^"]*"$' | tr -d '"')

if [ -z "$url" ]; then
    echo "No release asset found for ${os}/${arch}" >&2
    exit 1
fi

echo "Downloading memento from $url ..."

tmp="$(mktemp -d)"
curl -sL "$url" -o "${tmp}/memento.tar.gz"
tar xzf "${tmp}/memento.tar.gz" -C "$tmp"

mkdir -p "$INSTALL_DIR"
cp "$(find "$tmp" -name memento -type f | head -1)" "${INSTALL_DIR}/memento"
chmod +x "${INSTALL_DIR}/memento"
rm -rf "$tmp"

if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo "Add ${INSTALL_DIR} to your PATH:"
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc"
fi

echo "memento installed to ${INSTALL_DIR}/memento"
