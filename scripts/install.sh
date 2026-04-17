#!/bin/sh
# Armada one-line installer.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/DiegoDev2/Fleet/main/scripts/install.sh | sh
#
# Environment variables:
#   ARMADA_HOME      install root (default: $HOME/.armada)
#   ARMADA_VERSION   release tag to install (default: latest)
#   ARMADA_OWNER     GitHub owner (default: DiegoDev2)
#   ARMADA_REPO      GitHub repo  (default: Fleet)
set -eu

ARMADA_HOME="${ARMADA_HOME:-$HOME/.armada}"
ARMADA_VERSION="${ARMADA_VERSION:-latest}"
ARMADA_OWNER="${ARMADA_OWNER:-DiegoDev2}"
ARMADA_REPO="${ARMADA_REPO:-Fleet}"

log() { printf '==> %s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

detect_os() {
  case "$(uname -s)" in
    Linux)  echo linux ;;
    Darwin) echo darwin ;;
    *)      die "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    *) die "unsupported architecture: $(uname -m)" ;;
  esac
}

fetch_latest_tag() {
  api="https://api.github.com/repos/${ARMADA_OWNER}/${ARMADA_REPO}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$api" | grep -o '"tag_name": *"[^"]*"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/'
  else
    wget -qO- "$api" | grep -o '"tag_name": *"[^"]*"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/'
  fi
}

download() {
  url="$1"; out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$out" "$url"
  else
    wget -qO "$out" "$url"
  fi
}

verify_sha256() {
  file="$1"; sumfile="$2"
  if command -v sha256sum >/dev/null 2>&1; then
    expected="$(awk '{print $1}' "$sumfile")"
    actual="$(sha256sum "$file" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    expected="$(awk '{print $1}' "$sumfile")"
    actual="$(shasum -a 256 "$file" | awk '{print $1}')"
  else
    log "no sha256 tool available, skipping checksum verification"
    return 0
  fi
  [ "$expected" = "$actual" ] || die "checksum mismatch: expected $expected, got $actual"
}

main() {
  os="$(detect_os)"
  arch="$(detect_arch)"

  if [ "$ARMADA_VERSION" = "latest" ]; then
    ARMADA_VERSION="$(fetch_latest_tag)"
    [ -n "$ARMADA_VERSION" ] || die "could not determine latest release"
  fi

  log "installing armada $ARMADA_VERSION for $os/$arch into $ARMADA_HOME"

  base="https://github.com/${ARMADA_OWNER}/${ARMADA_REPO}/releases/download/${ARMADA_VERSION}"
  pkg="armada-${ARMADA_VERSION}-${os}-${arch}.tar.gz"
  sum="${pkg}.sha256"

  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  log "downloading $pkg"
  download "$base/$pkg" "$tmp/$pkg"
  download "$base/$sum" "$tmp/$sum"
  verify_sha256 "$tmp/$pkg" "$tmp/$sum"

  log "extracting"
  tar -xzf "$tmp/$pkg" -C "$tmp"

  mkdir -p "$ARMADA_HOME/bin"
  mv "$tmp/armada-${os}-${arch}" "$ARMADA_HOME/bin/armada"
  chmod +x "$ARMADA_HOME/bin/armada"

  log "installed to $ARMADA_HOME/bin/armada"
  case ":$PATH:" in
    *":$ARMADA_HOME/bin:"*) ;;
    *)
      log "add this line to your shell profile (~/.bashrc, ~/.zshrc, ...):"
      printf '    export PATH="%s/bin:$PATH"\n' "$ARMADA_HOME"
      ;;
  esac
}

main "$@"
