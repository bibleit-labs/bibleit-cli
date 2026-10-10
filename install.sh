#!/usr/bin/env bash
# Downloaded scripts execute only after the complete function has been read.
main() (
  set -euo pipefail
  fail() { printf 'bibleit installer: %s\n' "$*" >&2; exit 1; }
  version="${1:-}"
  [[ $# -le 1 ]] || fail 'Usage: bash install.sh [vMAJOR.MINOR.PATCH]'
  repo='https://github.com/bibleit-labs/bibleit-cli'
  assets="${BIBLEIT_INSTALL_ASSETS:-}"
  destination="${BIBLEIT_INSTALL_DIR:-$HOME/.local/bin}"
  if [[ -z "$version" ]]; then
    [[ -z "$assets" ]] || fail 'Specify a version when testing local assets'
    resolved=$(curl --proto '=https' --proto-redir '=https' -fsSLI -o /dev/null -w '%{url_effective}' "$repo/releases/latest")
    [[ "$resolved" == "$repo/releases/tag/"* ]] || fail 'Could not resolve latest release'
    version="${resolved##*/}"
  fi
  [[ "$version" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[A-Za-z0-9]+([.-][A-Za-z0-9]+)*)?$ ]] || fail 'Invalid version'
  case "$(uname -s)" in Darwin) system=darwin ;; Linux) system=linux ;; *) fail 'Use install.ps1 on Windows' ;; esac
  case "$(uname -m)" in x86_64|amd64) arch=amd64 ;; arm64|aarch64) arch=arm64 ;; *) fail 'Unsupported CPU architecture' ;; esac
  if [[ "$system" == darwin && "$arch" == amd64 ]] && [[ "$(sysctl -n sysctl.proc_translated 2>/dev/null || true)" == 1 ]]; then arch=arm64; fi
  name="bibleit_${version#v}_${system}_${arch}.tar.gz"
  temporary=$(mktemp -d)
  trap 'rm -rf "$temporary"' EXIT
  fetch() {
    if [[ -n "$assets" ]]; then
      cp "$assets/$1" "$temporary/$1"
    else
      curl --proto '=https' --proto-redir '=https' -fsSL "$repo/releases/download/$version/$1" -o "$temporary/$1"
    fi
  }
  fetch "$name"
  fetch checksums.txt
  expected=$(awk -v name="$name" '$2 == name {print $1}' "$temporary/checksums.txt")
  [[ "$expected" =~ ^[0-9a-f]{64}$ ]] || fail 'Missing or duplicate SHA-256 checksum'
  if command -v sha256sum >/dev/null; then
    actual=$(sha256sum "$temporary/$name" | awk '{print $1}')
  else
    actual=$(shasum -a 256 "$temporary/$name" | awk '{print $1}')
  fi
  [[ "$actual" == "$expected" ]] || fail 'Checksum mismatch; existing installation unchanged'
  # Extract only the expected binary, never arbitrary archive paths.
  tar -xOzf "$temporary/$name" bibleit > "$temporary/bibleit"
  [[ -s "$temporary/bibleit" ]] || fail 'Archive contains no binary'
  chmod 755 "$temporary/bibleit"
  [[ "$("$temporary/bibleit" version)" == "bibleit CLI version ${version#v}" ]] || fail 'Binary version mismatch'
  mkdir -p "$destination"
  staged=$(mktemp "$destination/.bibleit.XXXXXX")
  trap 'rm -rf "$temporary"; rm -f "${staged:-}"' EXIT
  cp "$temporary/bibleit" "$staged"
  chmod 755 "$staged"
  mv -f "$staged" "$destination/bibleit"
  printf 'Installed Bibleit %s at %s/bibleit\nAdd this directory to PATH if needed.\n' "$version" "$destination"
)
main "$@"
