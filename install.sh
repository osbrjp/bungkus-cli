#!/usr/bin/env bash
set -euo pipefail

REPO="osbrjp/bungkus-cli"
BIN_NAME="bungkus-cli"
USER_BIN="${HOME}/.local/bin"

err() { printf 'install: %s\n' "$*" >&2; exit 1; }
log() { printf '==> %s\n' "$*"; }

# resolve_link follows symlinks so an update replaces the real binary, not the link.
resolve_link() {
  local p="$1" t
  while [ -L "$p" ]; do
    t=$(readlink "$p")
    case "$t" in
      /*) p="$t" ;;
      *) p="$(dirname "$p")/$t" ;;
    esac
  done
  printf '%s\n' "$p"
}

# path_index prints the position of directory $1 in PATH, or nothing if absent.
path_index() {
  local i=0 d
  local IFS=:
  for d in $PATH; do
    i=$((i + 1))
    if [ "${d%/}" = "${1%/}" ]; then
      printf '%s\n' "$i"
      return
    fi
  done
}

# resolve_install_dir picks where the binary goes, without ever needing sudo
# when it can be avoided:
#   1. BUNGKUS_INSTALL_DIR when set;
#   2. else the folder of the installed binary (BUNGKUS_CURRENT_BIN from
#      `bungkus-cli update`, or the one on PATH), so updates replace it in place;
#   3. else ~/.local/bin for a fresh install.
# When the installed binary's folder isn't writable (a root-owned
# /usr/local/bin from older installers) and ~/.local/bin comes earlier on PATH,
# the new copy goes there and shadows the old one; otherwise it falls back to
# the old folder, which needs sudo as before. MIGRATED_FROM is set when the
# old copy is left behind.
MIGRATED_FROM=""
resolve_install_dir() {
  if [ -n "${BUNGKUS_INSTALL_DIR:-}" ]; then
    INSTALL_DIR="$BUNGKUS_INSTALL_DIR"
    return
  fi
  local current dir user_i dir_i
  current="${BUNGKUS_CURRENT_BIN:-$(command -v "$BIN_NAME" 2>/dev/null || true)}"
  if [ -z "$current" ]; then
    INSTALL_DIR="$USER_BIN"
    return
  fi
  current=$(resolve_link "$current")
  dir=$(dirname "$current")
  if [ -w "$dir" ]; then
    INSTALL_DIR="$dir"
    return
  fi
  user_i=$(path_index "$USER_BIN")
  dir_i=$(path_index "$dir")
  if [ -n "$user_i" ] && { [ -z "$dir_i" ] || [ "$user_i" -lt "$dir_i" ]; }; then
    INSTALL_DIR="$USER_BIN"
    MIGRATED_FROM="$current"
    return
  fi
  INSTALL_DIR="$dir"
}

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

sha256_of() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    err "neither shasum nor sha256sum found; cannot verify download"
  fi
}

resolve_latest_tag() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -m1 '"tag_name":' \
    | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/'
}

main() {
  resolve_install_dir
  if [ -n "${BUNGKUS_INSTALL_DRY_RUN:-}" ]; then
    printf '%s\n' "$INSTALL_DIR"
    [ -z "$MIGRATED_FROM" ] || printf 'migrated-from %s\n' "$MIGRATED_FROM"
    return
  fi

  require_cmd curl
  require_cmd uname

  local os arch tag asset url checksums_url tmp checksums_tmp expected actual dest
  os=$(detect_os)
  arch=$(detect_arch)

  log "resolving latest release for ${REPO}"
  tag=$(resolve_latest_tag)
  [ -n "$tag" ] || err "could not resolve latest release tag (is the repo public and does it have a published release?)"

  asset="${BIN_NAME}-${os}-${arch}"
  url="https://github.com/${REPO}/releases/download/${tag}/${asset}"
  checksums_url="https://github.com/${REPO}/releases/download/${tag}/checksums.txt"
  log "installing ${BIN_NAME} ${tag} (${os}/${arch})"

  tmp=$(mktemp -t "${BIN_NAME}.XXXXXX")
  checksums_tmp=$(mktemp -t "${BIN_NAME}-checksums.XXXXXX")
  trap 'rm -f "$tmp" "$checksums_tmp"' EXIT

  log "downloading ${url}"
  curl -fSL "$url" -o "$tmp"

  log "verifying checksum"
  curl -fSL "$checksums_url" -o "$checksums_tmp"
  expected=$(awk -v f="$asset" '$2==f {print $1}' "$checksums_tmp")
  [ -n "$expected" ] || err "no checksum entry for ${asset} in checksums.txt"
  actual=$(sha256_of "$tmp")
  [ "$expected" = "$actual" ] || err "checksum mismatch for ${asset} (expected ${expected}, got ${actual})"

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
  if [ -n "$MIGRATED_FROM" ]; then
    log "the old copy at ${MIGRATED_FROM} is now shadowed by ${dest}; remove it with: sudo rm ${MIGRATED_FROM}"
  fi
  if [ -z "$(path_index "$INSTALL_DIR")" ]; then
    log "${INSTALL_DIR} is not on your PATH; add it:"
    log "  fish: fish_add_path ${INSTALL_DIR}"
    log "  zsh/bash: echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc   (or ~/.bashrc)"
  fi
  log "verify: ${BIN_NAME} --help"
}

main "$@"
