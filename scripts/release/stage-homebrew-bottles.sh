#!/usr/bin/env bash
set -euo pipefail

published_url=
if [[ "${1:-}" == "--verify-published" ]]; then
  published_url="${2:?published release URL required}"
  shift 2
fi
dist_dir="${1:?release asset directory required}"
shopt -s nullglob
json_files=("$dist_dir"/*.bottle.json)
if (( ${#json_files[@]} == 0 )); then
  echo "no Homebrew bottle metadata found" >&2
  exit 1
fi

# Homebrew writes a local filename, which can differ from the filename its
# downloader requests. Keep the JSON and stage the archive under the latter.
rows="$(
  jq -ers '
    def valid_name: if type == "string" then test("^md-[A-Za-z0-9+_.-]+[.]bottle([.][0-9]+)?[.]tar[.]gz$") else false end;
    def valid_digest: if type == "string" then test("^[0-9a-f]{64}$") else false end;
    [
      .[] | to_entries[] |
      if (.key | split("/") | last) != "md" then error("unexpected bottle formula: " + .key)
      else .value.bottle.tags | to_entries[] end |
      {tag: .key, source: .value.local_filename, target: .value.filename, digest: .value.sha256}
    ] | sort_by(.tag) |
    if map(.tag) != ["arm64_linux", "arm64_tahoe", "tahoe", "x86_64_linux"]
      then error("md bottle metadata must cover each expected platform exactly once")
    elif any(.[]; ((.source | valid_name) and (.target | valid_name) and (.digest | valid_digest)) | not)
      then error("invalid md bottle filename or SHA-256")
    elif (map(.source) | unique | length) != length or (map(.target) | unique | length) != length
      then error("duplicate md bottle filenames")
    else .[] | [.source, .target, .digest] | @tsv end
  ' "${json_files[@]}"
)"

if [[ -n "$published_url" ]]; then
  while IFS=$'\t' read -r _ target _; do
    curl --head --location --fail --silent --show-error --retry 3 --retry-all-errors \
      --max-time 20 --proto '=https' --proto-redir '=https' --output /dev/null \
      "${published_url%/}/$target"
  done <<<"$rows"
  exit 0
fi

# Verify all inputs before renaming any file, and never overwrite an asset.
while IFS=$'\t' read -r source target digest; do
  if [[ ! -f "$dist_dir/$source" || ( "$source" != "$target" && -e "$dist_dir/$target" ) ]]; then
    printf 'missing input or occupied output for Homebrew bottle %s -> %s\n' "$source" "$target" >&2
    exit 1
  fi
  actual="$(sha256sum -- "$dist_dir/$source")"
  if [[ "${actual%% *}" != "$digest" ]]; then
    printf 'SHA-256 mismatch for Homebrew bottle %s\n' "$source" >&2
    exit 1
  fi
done <<<"$rows"

while IFS=$'\t' read -r source target _; do
  if [[ "$source" != "$target" ]]; then
    mv -- "$dist_dir/$source" "$dist_dir/$target"
  fi
done <<<"$rows"
