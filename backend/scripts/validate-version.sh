#!/bin/sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
BACKEND_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
VERSION_FILE="$BACKEND_DIR/cmd/server/VERSION"
UPSTREAM_VERSION_FILE="$BACKEND_DIR/cmd/server/UPSTREAM_VERSION"

if ! awk 'END { exit (NR == 1 ? 0 : 1) }' "$VERSION_FILE"; then
  echo "VERSION must contain exactly one line" >&2
  exit 1
fi
if ! awk 'END { exit (NR == 1 ? 0 : 1) }' "$UPSTREAM_VERSION_FILE"; then
  echo "UPSTREAM_VERSION must contain exactly one line" >&2
  exit 1
fi

VERSION="$(cat "$VERSION_FILE")"
UPSTREAM_VERSION="$(cat "$UPSTREAM_VERSION_FILE")"
CR="$(printf '\r')"
VERSION="${VERSION%"$CR"}"
UPSTREAM_VERSION="${UPSTREAM_VERSION%"$CR"}"
UPSTREAM_PATTERN='^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
VERSION_PATTERN='^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)-([1-9][0-9]*)(\.[1-9][0-9]*)*$'

if ! printf '%s\n' "$UPSTREAM_VERSION" | grep -Eq "$UPSTREAM_PATTERN"; then
  echo "invalid UPSTREAM_VERSION: $UPSTREAM_VERSION" >&2
  exit 1
fi

if ! printf '%s\n' "$VERSION" | grep -Eq "$VERSION_PATTERN"; then
  echo "invalid AIWeLink VERSION: $VERSION" >&2
  exit 1
fi

case "$VERSION" in
  "$UPSTREAM_VERSION"-*) ;;
  *)
    echo "VERSION $VERSION does not match UPSTREAM_VERSION $UPSTREAM_VERSION" >&2
    exit 1
    ;;
esac

if [ "$#" -gt 1 ]; then
  echo "usage: $0 [release-tag]" >&2
  exit 1
fi

if [ "$#" -eq 1 ]; then
  if [ -z "$1" ]; then
    echo "release tag must not be empty" >&2
    exit 1
  fi
  if [ "$1" != "v$VERSION" ]; then
    echo "release tag ${1} does not match VERSION $VERSION" >&2
    exit 1
  fi
fi

printf 'AIWeLink %s (Sub2API %s)\n' "$VERSION" "$UPSTREAM_VERSION"
