#!/usr/bin/env bash
set -euo pipefail

: "${VERSION:?VERSION is required}"
: "${COMMIT:?COMMIT is required}"
: "${DATE:?DATE is required}"
: "${GOOS:?GOOS is required}"
: "${GOARCH:?GOARCH is required}"

out_dir="${OUT_DIR:-dist/bin/${GOOS}-${GOARCH}}"
mkdir -p "$out_dir"

module="github.com/ClarifiedLabs/mdcli"
ldflags="-s -w"
ldflags+=" -X ${module}/internal/buildinfo.Version=${VERSION}"
ldflags+=" -X ${module}/internal/buildinfo.Commit=${COMMIT}"
ldflags+=" -X ${module}/internal/buildinfo.Date=${DATE}"

CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags "$ldflags" -o "${out_dir}/md" ./cmd/md
