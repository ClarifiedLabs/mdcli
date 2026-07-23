#!/usr/bin/env bash
set -euo pipefail

: "${VERSION:?VERSION is required (for example, v1.2.3)}"

if [[ "$VERSION" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
	version="${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.${BASH_REMATCH[3]}"
else
	echo "VERSION must be a v-prefixed semantic version (for example, v1.2.3)" >&2
	exit 2
fi

readme="${1:-README.md}"
if [[ ! -f "$readme" ]]; then
	echo "README not found: $readme" >&2
	exit 1
fi

block_file="$(mktemp)"
output_file="$(mktemp)"
trap 'rm -f "$block_file" "$output_file"' EXIT

releases="https://github.com/ClarifiedLabs/mdcli/releases/download/${VERSION}"

cat >"$block_file" <<EOF
Or download the latest release, ${VERSION}, directly:

- Apple silicon (arm64): [signed \`.pkg\`](${releases}/md_${VERSION}_darwin_arm64.pkg) · [\`.tar.gz\`](${releases}/md_${VERSION}_darwin_arm64.tar.gz)
- Intel (amd64): [\`.tar.gz\`](${releases}/md_${VERSION}_darwin_amd64.tar.gz)

### Linux

${VERSION} is available for amd64/x86_64 and arm64/aarch64:

| Format | amd64 / x86_64 | arm64 / aarch64 |
|---|---|---|
| Package | [\`.deb\`](${releases}/md_${version}_amd64.deb) · [\`.rpm\`](${releases}/md-${version}-1.x86_64.rpm) | [\`.deb\`](${releases}/md_${version}_arm64.deb) · [\`.rpm\`](${releases}/md-${version}-1.aarch64.rpm) |
| Tarball | [\`.tar.gz\`](${releases}/md_${VERSION}_linux_amd64.tar.gz) | [\`.tar.gz\`](${releases}/md_${VERSION}_linux_arm64.tar.gz) |

Every asset is listed with its SHA-256 in
[\`checksums.txt\`](${releases}/checksums.txt).
EOF

awk -v block_file="$block_file" '
	BEGIN {
		while ((getline line < block_file) > 0) {
			block = block line ORS
		}
		close(block_file)
	}
	$0 == "<!-- release-artifacts:start -->" {
		starts++
		print
		printf "%s", block
		replacing = 1
		next
	}
	$0 == "<!-- release-artifacts:end -->" {
		ends++
		replacing = 0
		print
		next
	}
	!replacing { print }
	END {
		if (starts != 1 || ends != 1 || replacing) {
			print "expected exactly one complete release-artifacts marker pair" > "/dev/stderr"
			exit 1
		}
	}
' "$readme" >"$output_file"

cp "$output_file" "$readme"
