# Shared helper: locate the Go toolchain and prepare the environment.
# Sourced by the bin/* commands; not executable on its own.
# shellcheck shell=bash

WRITINGS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

GO_BIN=""
if command -v go >/dev/null 2>&1; then
  GO_BIN="go"
else
  for candidate in /usr/local/go/bin/go "$WRITINGS_ROOT/.tools/go/bin/go" "$HOME/go/bin/go" /opt/homebrew/bin/go; do
    if [ -x "$candidate" ]; then
      GO_BIN="$candidate"
      break
    fi
  done
fi
if [ -z "$GO_BIN" ]; then
  echo "error: Go toolchain not found; install it from https://go.dev/dl/ (or: brew install go)" >&2
  exit 1
fi

# Never auto-download a different toolchain version.
export GOTOOLCHAIN=local

# Offline module cache support (used by sandboxed environments):
# if .gomodcache exists in the repository, use it and stay offline.
if [ -d "$WRITINGS_ROOT/.gomodcache" ]; then
  export GOMODCACHE="$WRITINGS_ROOT/.gomodcache"
  export GOFLAGS="-mod=mod"
  export GOPROXY=off
  export GOSUMDB=off
fi

cd "$WRITINGS_ROOT" || exit 1

# First run: record dependency hashes in go.sum (needs network once,
# unless an offline .gomodcache is present).
if [ ! -f "$WRITINGS_ROOT/go.sum" ]; then
  "$GO_BIN" mod tidy || exit 1
fi
