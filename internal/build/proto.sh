#!/bin/sh

set -eu

DIR="$PWD"
cd "$(git rev-parse --show-toplevel)"
PKGDIR="${DIR#"$PWD"/}"

protoc -Iproto/nanopb -Iproto/nanorpc -Iproto/vendor \
	"--go_out=$PKGDIR" \
	--go_opt=paths=source_relative \
	"proto/$GOPACKAGE/${GOFILE%.go}.proto"
