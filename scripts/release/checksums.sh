#!/usr/bin/env bash
set -euo pipefail

dist_dir="${DIST_DIR:-dist}"
out="${CHECKSUMS_FILE:-${dist_dir}/checksums.txt}"

# Stage outside the checksummed directory: a temp file inside it exists before
# find runs and would end up checksumming itself.
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

# shasum ships with macOS, sha256sum with most Linux images; both print
# "<sha256>  <name>".
if command -v shasum >/dev/null 2>&1; then
	sha256() { shasum -a 256 "$1"; }
else
	sha256() { sha256sum "$1"; }
fi

find "$dist_dir" -maxdepth 1 -type f ! -name "$(basename "$out")" -print |
	sort |
	while IFS= read -r file; do
		(cd "$(dirname "$file")" && sha256 "$(basename "$file")")
	done >"$tmp"

cp "$tmp" "$out"
chmod 0644 "$out"
