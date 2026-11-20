#!/usr/bin/env bash
set -euo pipefail

: "${TAP_DIR:?TAP_DIR is required}"
bottle_dir="$(cd "${1:?bottle metadata directory is required}" && pwd)"

shopt -s nullglob
bottle_jsons=("$bottle_dir"/*.json)
if [[ ${#bottle_jsons[@]} -eq 0 ]]; then
  echo "no bottle JSON artifacts found" >&2
  exit 1
fi

# Homebrew replaces the existing bottle block by default. Validate the
# whole matrix, then merge every input in one call so no platform is lost.
expected="$(
  jq -ers '
    [ .[] | to_entries[] |
      if (.key | split("/") | last) != "md" then error("unexpected formula: " + .key)
      else .value.bottle.tags | to_entries[] end
    ] | sort_by(.key)
      | if map(.key) != ["arm64_linux", "arm64_tahoe", "tahoe", "x86_64_linux"]
        then error("md bottle metadata must cover each expected platform exactly once")
        elif any(.[]; .value.sha256 | (type != "string") or (test("^[0-9a-f]{64}$") | not))
        then error("md bottle metadata contains an invalid SHA-256")
        else .[] | [.key, .value.sha256] | @tsv end
  ' "${bottle_jsons[@]}"
)"

cd "$TAP_DIR"
brew bottle --merge --write --no-commit "${bottle_jsons[@]}"
actual="$(
  sed -n '/^[[:space:]]*bottle do[[:space:]]*$/,/^[[:space:]]*end[[:space:]]*$/p' Formula/md.rb |
    sed -E -n 's/^[[:space:]]*sha256[[:space:]]+(cellar:[^,]+,[[:space:]]*)?([a-z0-9_]+):[[:space:]]*"([0-9a-f]{64})"[[:space:]]*$/\2\t\3/p' |
    LC_ALL=C sort
)"
if [[ "$actual" != "$expected" ]]; then
  echo "md bottle block does not match the expected platform checksums" >&2
  diff -u <(printf '%s\n' "$expected") <(printf '%s\n' "$actual") >&2 || true
  exit 1
fi
