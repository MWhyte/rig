#!/bin/sh
# rig.fm installer
#
# Usage:
#   curl -fsSL https://rig.fm/install.sh | sh
#
# Environment overrides:
#   RIG_VERSION   Tag to install (default: latest)
#   BIN_DIR       Install directory (default: /usr/local/bin, falls back to ~/.local/bin)

set -eu

REPO="MWhyte/rig"
BIN_NAME="rig"

info() { printf '\033[1;34m==>\033[0m %s\n' "$*" >&2; }
warn() { printf '\033[1;33m!!\033[0m %s\n'  "$*" >&2; }
err()  { printf '\033[1;31mxx\033[0m %s\n'  "$*" >&2; exit 1; }

detect_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    linux|darwin) printf '%s' "$os" ;;
    *) err "unsupported OS: $os (linux and darwin only)" ;;
  esac
}

detect_arch() {
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64)  printf 'amd64' ;;
    aarch64|arm64) printf 'arm64' ;;
    *) err "unsupported architecture: $arch (amd64 and arm64 only)" ;;
  esac
}

fetch() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$1" -o "$2"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$1" -O "$2"
  else
    err "need curl or wget to download"
  fi
}

resolve_version() {
  v=${RIG_VERSION:-latest}
  if [ "$v" = "latest" ]; then
    if command -v curl >/dev/null 2>&1; then
      v=$(curl -fsSLI -o /dev/null -w '%{url_effective}' \
        "https://github.com/$REPO/releases/latest" \
        | sed 's#.*/tag/##' | tr -d '\r\n')
    else
      v=$(wget -q --server-response --max-redirect=0 \
        "https://github.com/$REPO/releases/latest" 2>&1 \
        | awk '/Location:/ {print $2}' | sed 's#.*/tag/##' | tr -d '\r\n')
    fi
    [ -n "$v" ] || err "could not determine latest version"
  fi
  printf '%s' "$v"
}

choose_bin_dir() {
  if [ -n "${BIN_DIR:-}" ]; then
    printf '%s' "$BIN_DIR"
    return
  fi
  if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
    printf '/usr/local/bin'
    return
  fi
  # /usr/local/bin will need sudo; if sudo is unavailable, fall back to ~/.local/bin
  if command -v sudo >/dev/null 2>&1 && [ -d /usr/local/bin ]; then
    printf '/usr/local/bin'
    return
  fi
  printf '%s' "$HOME/.local/bin"
}

sha_check() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c -
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 -c -
  else
    err "need sha256sum or shasum to verify checksum"
  fi
}

install_binary() {
  src=$1
  dst=$2
  dst_dir=$(dirname "$dst")
  mkdir -p "$dst_dir" 2>/dev/null || true
  if [ -w "$dst_dir" ]; then
    install -m 0755 "$src" "$dst"
  else
    info "installing to $dst (requires sudo)"
    sudo install -m 0755 "$src" "$dst"
  fi
}

check_mpv() {
  if command -v mpv >/dev/null 2>&1; then
    return
  fi
  warn "mpv is not installed — rig needs mpv for audio playback."
  cat >&2 <<EOF
    Install mpv with your package manager:
      Debian/Ubuntu: sudo apt install mpv
      Fedora:        sudo dnf install mpv
      Arch:          sudo pacman -S mpv
      Alpine:        sudo apk add mpv
      macOS:         brew install mpv
EOF
}

main() {
  os=$(detect_os)
  arch=$(detect_arch)
  version=$(resolve_version)
  bin_dir=$(choose_bin_dir)

  asset="${BIN_NAME}_${os}_${arch}.tar.gz"
  base="https://github.com/$REPO/releases/download/$version"

  info "rig.fm installer"
  info "version:     $version"
  info "platform:    $os/$arch"
  info "install dir: $bin_dir"

  tmp=$(mktemp -d 2>/dev/null || mktemp -d -t rig-install)
  trap 'rm -rf "$tmp"' EXIT INT TERM

  info "downloading $asset"
  fetch "$base/$asset" "$tmp/$asset"
  fetch "$base/checksums.txt" "$tmp/checksums.txt"

  info "verifying checksum"
  ( cd "$tmp" && grep "  $asset$" checksums.txt | sha_check >/dev/null ) \
    || err "checksum verification failed for $asset"

  info "extracting"
  tar -xzf "$tmp/$asset" -C "$tmp"

  install_binary "$tmp/$BIN_NAME" "$bin_dir/$BIN_NAME"
  info "installed $BIN_NAME to $bin_dir/$BIN_NAME"

  check_mpv

  case ":$PATH:" in
    *":$bin_dir:"*) ;;
    *)
      warn "$bin_dir is not on your PATH. Add it with:"
      printf '    export PATH=%s:$PATH\n' "$bin_dir" >&2
      ;;
  esac

  info "done. run: rig"
}

main
