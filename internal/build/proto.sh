#!/bin/sh

set -eu

# Use PROTOC environment variable or default to protoc
# shellcheck disable=SC2223
: ${PROTOC:=protoc}

# Test if protoc is available
# shellcheck disable=SC2086
if ! $PROTOC --version >/dev/null 2>&1; then
	echo "Warning: $PROTOC not found, skipping protocol buffer generation" >&2
	echo "To generate protocol buffers, install protoc or set PROTOC environment variable" >&2
	exit 0
fi

DIR="$PWD"
cd "$(git rev-parse --show-toplevel)"
PKGDIR="${DIR#"$PWD"/}"

# shellcheck disable=SC2086
$PROTOC -Iproto/nanopb -Iproto/nanorpc -Iproto/vendor \
	"--go_out=$PKGDIR" \
	--go_opt=paths=source_relative \
	"proto/$GOPACKAGE/${GOFILE%.go}.proto"
