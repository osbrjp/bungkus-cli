#!/usr/bin/env bash
set -euo pipefail

REPO="spencer-osbrjp/bungkus-cli"
BIN_NAME="bungkus-cli"
INSTALL_DIR="${BUNGKUS_INSTALL_DIR:-/usr/local/bin}"

err() { printf 'install: %s\n' "$*" >&2; exit 1; }
log() { printf '==> %s\n' "$*"; }

detect_os() {
  case "$(uname -s)" in
    Darwin) echo darwin ;;
    Linux)  echo linux ;;
    *) err "unsupported OS: $(uname -s). bungkus-cli supports darwin and linux." ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    arm64|aarch64) echo arm64 ;;
    x86_64|amd64)  echo amd64 ;;
    *) err "unsupported architecture: $(uname -m). bungkus-cli supports arm64 and amd64." ;;
  esac
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"
}

resolve_latest_tag() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -m1 '"tag_name":' \
    | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/'
}

main() {
  require_cmd curl
  require_cmd uname

  local os arch tag asset url tmp dest
  os=$(detect_os)
  arch=$(detect_arch)

  log "resolving latest release for ${REPO}"
  tag=$(resolve_latest_tag)
  [ -n "$tag" ] || err "could not resolve latest release tag (is the repo public and does it have a published release?)"

  asset="${BIN_NAME}-${os}-${arch}"
  url="https://github.com/${REPO}/releases/download/${tag}/${asset}"
  log "installing ${BIN_NAME} ${tag} (${os}/${arch})"

  tmp=$(mktemp -t "${BIN_NAME}.XXXXXX")
  trap 'rm -f "$tmp"' EXIT

  log "downloading ${url}"
  curl -fSL "$url" -o "$tmp"
  chmod +x "$tmp"

  if [ "$os" = "darwin" ]; then
    xattr -d com.apple.quarantine "$tmp" 2>/dev/null || true
  fi

  dest="${INSTALL_DIR}/${BIN_NAME}"
  if [ -w "$INSTALL_DIR" ] || ([ ! -e "$INSTALL_DIR" ] && mkdir -p "$INSTALL_DIR" 2>/dev/null); then
    mv "$tmp" "$dest"
  else
    log "writing to ${INSTALL_DIR} requires sudo"
    sudo mv "$tmp" "$dest"
  fi

  trap - EXIT

  log "installed ${BIN_NAME} ${tag} -> ${dest}"
  log "verify: ${BIN_NAME} --help"
}

main "$@"
