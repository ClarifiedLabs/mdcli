#!/usr/bin/env bash
set -euo pipefail

: "${TAG:?TAG is required}"
: "${SOURCE_SHA256:?SOURCE_SHA256 is required}"

tap_dir="${TAP_DIR:?TAP_DIR is required}"
formula_dir="${tap_dir}/Formula"
version="${TAG#v}"
source_url="${SOURCE_URL:-https://github.com/ClarifiedLabs/mdcli/archive/refs/tags/${TAG}.tar.gz}"

mkdir -p "$formula_dir"

cat >"${formula_dir}/md.rb" <<FORMULA
class Md < Formula
  desc "Terminal Markdown viewer with ASCII Mermaid diagrams"
  homepage "https://github.com/ClarifiedLabs/mdcli"
  url "${source_url}"
  sha256 "${SOURCE_SHA256}"
  version "${version}"
  license "MIT"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X github.com/ClarifiedLabs/mdcli/internal/buildinfo.Version=v#{version}
    ]
    system "go", "build", "-trimpath", "-ldflags", ldflags.join(" "), "-o", bin/"md", "./cmd/md"
  end

  test do
    assert_match "md v#{version}", shell_output("#{bin}/md --version")
    (testpath/"doc.md").write("# Title\n\nSome **bold** text.\n")
    assert_match "Title", shell_output("#{bin}/md -color never -p never #{testpath}/doc.md")
  end
end
FORMULA
