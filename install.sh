#!/bin/sh
# Install PTAL - your GitHub pull requests on Telegram.
#
#   curl -fsSL https://raw.githubusercontent.com/Cristhianzl/telegram-PTAL/main/install.sh | sh
#
# Installs to ~/.local/bin by default. Set PREFIX to change that.
set -eu

REPO="Cristhianzl/telegram-PTAL"
BINARY="ptal"
PREFIX="${PREFIX:-$HOME/.local/bin}"

say()  { printf '%s\n' "$*"; }
die()  { printf 'error: %s\n' "$*" >&2; exit 1; }

detect_platform() {
  os=$(uname -s)
  arch=$(uname -m)

  case "$os" in
    Linux)  os=Linux ;;
    Darwin) os=Darwin ;;
    *) die "PTAL supports Linux and macOS, not $os.
       On Windows, run it inside WSL2 - the Linux path works unchanged." ;;
  esac

  case "$arch" in
    x86_64|amd64)  arch=x86_64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) die "unsupported architecture: $arch" ;;
  esac

  printf '%s_%s' "$os" "$arch"
}

latest_version() {
  # The releases/latest redirect carries the tag, which avoids depending on
  # jq or on an authenticated API call just to learn the version.
  #
  # With no release published the redirect lands on /releases instead, with
  # no /tag/ segment - so match on that rather than on the sed output, which
  # would otherwise pass the whole URL along as a "version".
  url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' \
    "https://github.com/$REPO/releases/latest" 2>/dev/null) || return 1
  case "$url" in
    */tag/*) printf '%s' "${url##*/tag/}" ;;
    *) return 1 ;;
  esac
}

main() {
  command -v curl >/dev/null 2>&1 || die "curl is required"
  command -v tar  >/dev/null 2>&1 || die "tar is required"

  platform=$(detect_platform)
  version=$(latest_version) || die "no release published yet for $REPO.
       Build from source instead:
         git clone https://github.com/$REPO && cd telegram-PTAL && make build"

  asset="${BINARY}_${platform}.tar.gz"
  url="https://github.com/$REPO/releases/download/$version/$asset"

  say "Installing $BINARY $version ($platform)"

  tmp=$(mktemp -d)
  # Clean up even if the download or extraction fails halfway.
  trap 'rm -rf "$tmp"' EXIT INT TERM

  curl -fsSL "$url" -o "$tmp/$asset" || die "download failed: $url"

  # Verify the checksum when the release publishes one. A silent skip would
  # defeat the point, so say so when it is missing.
  if curl -fsSL "https://github.com/$REPO/releases/download/$version/checksums.txt" \
       -o "$tmp/checksums.txt" 2>/dev/null; then
    if command -v sha256sum >/dev/null 2>&1; then
      (cd "$tmp" && grep " $asset\$" checksums.txt | sha256sum -c -) >/dev/null \
        || die "checksum mismatch for $asset"
      say "  checksum verified"
    elif command -v shasum >/dev/null 2>&1; then
      (cd "$tmp" && grep " $asset\$" checksums.txt | shasum -a 256 -c -) >/dev/null \
        || die "checksum mismatch for $asset"
      say "  checksum verified"
    fi
  else
    say "  warning: no checksums.txt published for this release"
  fi

  tar -xzf "$tmp/$asset" -C "$tmp"
  [ -f "$tmp/$BINARY" ] || die "the archive did not contain $BINARY"

  mkdir -p "$PREFIX"
  install -m 0755 "$tmp/$BINARY" "$PREFIX/$BINARY" 2>/dev/null \
    || { cp "$tmp/$BINARY" "$PREFIX/$BINARY" && chmod 0755 "$PREFIX/$BINARY"; }

  say "  installed to $PREFIX/$BINARY"
  say ""

  case ":$PATH:" in
    *":$PREFIX:"*) ;;
    *)
      say "$PREFIX is not on your PATH. Add this to your shell profile:"
      say ""
      say "    export PATH=\"$PREFIX:\$PATH\""
      say ""
      ;;
  esac

  say "Next:"
  say "    $BINARY setup      # connect Telegram"
  say "    $BINARY install    # start with your computer"
}

main "$@"
