#!/usr/bin/env sh
set -eu

repo="${DOTFILES_BOOTSTRAP_REPO:-mickmetalholic/dotfiles-bootstrap}"
version="${DOTFILES_BOOTSTRAP_VERSION:-latest}"

has() {
  command -v "$1" >/dev/null 2>&1
}

bootstrap_os() {
  uname_s="$(uname -s 2>/dev/null || printf unknown)"
  case "$uname_s" in
    Darwin) printf darwin ;;
    Linux) printf linux ;;
    *) return 1 ;;
  esac
}

bootstrap_arch() {
  uname_m="$(uname -m 2>/dev/null || printf unknown)"
  case "$uname_m" in
    x86_64|amd64) printf amd64 ;;
    arm64|aarch64) printf arm64 ;;
    *) return 1 ;;
  esac
}

download_file() {
  url="$1"
  output="$2"
  if has curl; then
    curl -fsSL "$url" -o "$output"
  elif has wget; then
    wget -qO "$output" "$url"
  else
    return 1
  fi
}

checksum_line() {
  file="$1"
  checksums="$2"
  archive="$(basename "$file")"
  line="$(grep "  $archive\$" "$checksums" 2>/dev/null || grep " $archive\$" "$checksums" 2>/dev/null || true)"
  [ -n "$line" ] || return 1
  printf '%s\n' "$line"
}

verify_checksum() {
  file="$1"
  checksums="$2"
  line="$(checksum_line "$file" "$checksums" || true)"
  [ -n "$line" ] || return 1

  if has shasum; then
    printf '%s\n' "$line" | (cd "$(dirname "$file")" && shasum -a 256 -c -)
  elif has sha256sum; then
    printf '%s\n' "$line" | (cd "$(dirname "$file")" && sha256sum -c -)
  else
    return 1
  fi
}

verify_with_public_checksums() {
  file="$1"
  primary_checksums="$2"
  fallback_checksums="$3"

  if checksum_line "$file" "$primary_checksums" >/dev/null; then
    verify_checksum "$file" "$primary_checksums"
    return $?
  fi
  download_file "$base_url/dot_checksums.txt" "$fallback_checksums" || {
    printf 'could not download dot_checksums.txt\n' >&2
    return 1
  }
  verify_checksum "$file" "$fallback_checksums"
}

install_dot() {
  archive_path="$1"
  install_dir="${DOTFILES_BIN_DIR:-$HOME/.local/share/dotfiles/bin}"
  extract_dir="$tmp_dir/dot"
  mkdir -p "$extract_dir" "$install_dir" || {
    printf 'could not create dot install directory\n' >&2
    return 1
  }
  tar -xzf "$archive_path" -C "$extract_dir" || {
    printf 'could not extract %s\n' "$(basename "$archive_path")" >&2
    return 1
  }
  if [ -f "$extract_dir/dot" ]; then
    dot_binary="$extract_dir/dot"
  else
    dot_binary="$(find "$extract_dir" -type f -name dot 2>/dev/null | head -n 1 || true)"
  fi
  if [ -z "$dot_binary" ]; then
    printf 'archive did not contain executable dot\n' >&2
    return 1
  fi
  cp "$dot_binary" "$install_dir/dot" || {
    printf 'could not install dot to %s\n' "$install_dir/dot" >&2
    return 1
  }
  chmod 0755 "$install_dir/dot" || {
    printf 'could not mark dot executable at %s\n' "$install_dir/dot" >&2
    return 1
  }
  printf 'installed dot to %s\n' "$install_dir/dot"
}

tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/dotfiles-bootstrap.XXXXXX")" || exit 1
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

os="$(bootstrap_os)" || {
  printf 'unsupported OS for dotfiles-bootstrap\n' >&2
  exit 1
}
arch="$(bootstrap_arch)" || {
  printf 'unsupported architecture for dotfiles-bootstrap\n' >&2
  exit 1
}

archive="dotfiles-bootstrap_${os}_${arch}.tar.gz"
dot_archive="dot_${os}_${arch}.tar.gz"
base_url="https://github.com/$repo/releases/$version/download"
archive_path="$tmp_dir/$archive"
dot_archive_path="$tmp_dir/$dot_archive"
checksums_path="$tmp_dir/checksums.txt"
dot_checksums_path="$tmp_dir/dot_checksums.txt"

download_file "$base_url/$archive" "$archive_path" || {
  printf 'could not download %s\n' "$archive" >&2
  exit 1
}
download_file "$base_url/checksums.txt" "$checksums_path" || {
  printf 'could not download checksums.txt\n' >&2
  exit 1
}
verify_checksum "$archive_path" "$checksums_path" || {
  printf 'checksum verification failed for %s\n' "$archive" >&2
  exit 1
}
download_file "$base_url/$dot_archive" "$dot_archive_path" || {
  printf 'could not download %s\n' "$dot_archive" >&2
  exit 1
}
verify_with_public_checksums "$dot_archive_path" "$checksums_path" "$dot_checksums_path" || {
  printf 'checksum verification failed for %s\n' "$dot_archive" >&2
  exit 1
}
install_dot "$dot_archive_path" || exit 1

tar -xzf "$archive_path" -C "$tmp_dir" || {
  printf 'could not extract %s\n' "$archive" >&2
  exit 1
}

if [ ! -x "$tmp_dir/dotfiles-bootstrap" ]; then
  printf 'archive did not contain executable dotfiles-bootstrap\n' >&2
  exit 1
fi

"$tmp_dir/dotfiles-bootstrap" "$@"
